package github_pagination

import (
	"io"
	"net/http"

	"github.com/gofri/go-github-pagination/github_pagination/json_merger"
	"github.com/gofri/go-github-pagination/github_pagination/pagination_utils"
)

type GitHubPagination struct {
	Base   http.RoundTripper
	config *Config
}

func NewGithubPagination(base http.RoundTripper, opts ...Option) *GitHubPagination {
	if base == nil {
		base = http.DefaultTransport
	}
	return &GitHubPagination{
		Base:   base,
		config: newConfig(opts...),
	}
}

func NewGithubPaginationClient(base http.RoundTripper, opts ...Option) *http.Client {
	return &http.Client{
		Transport: NewGithubPagination(base, opts...),
	}
}

func (g *GitHubPagination) RoundTrip(request *http.Request) (resp *http.Response, err error) {
	reqConfig := g.config.GetRequestConfig(request)
	if reqConfig.Disabled {
		return g.Base.RoundTrip(request)
	}

	// it is enough to call the update request once,
	// since query parameters are kept through the pagination.
	request = reqConfig.UpdateRequest(request)

	merger := json_merger.NewMerger()
	pageCount := 1
	for {
		resp, err = g.Base.RoundTrip(request)
		if err != nil {
			return resp, err
		}

		// only paginate through successful requests.
		// having a non-successful request would drop all previous results.
		// TODO this is gonna have a config for strictness (i.e., whether to drop all previous results or not)
		if resp.StatusCode != http.StatusOK {
			return resp, nil
		}

		request = pagination_utils.GetNextRequest(request, resp)

		// early-exit for non-paginated requests
		if request == nil && pageCount == 1 {
			break
		}

		if err := merger.ReadNext(resp.Body); err != nil {
			return resp, err
		}

		if request == nil {
			break
		}
		pageCount++
		if reqConfig.IsPaginationOverflow(pageCount) {
			break
		}
	}

	// only merge if we paginated.
	// otherwise, we just return the response as is.
	if pageCount > 1 {
		resp.Body = io.NopCloser(merger.Merged())
	}

	return resp, nil
}
