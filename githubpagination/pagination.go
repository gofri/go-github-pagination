package githubpagination

import (
	"net/http"

	"github.com/gofri/go-github-pagination/githubpagination/drivers"
	github_response "github.com/gofri/go-github-pagination/githubpagination/response"
)

type PaginationDriver = drivers.Driver

type GitHubPagination struct {
	Base   http.RoundTripper
	config *Config
}

func New(base http.RoundTripper, opts ...Option) *GitHubPagination {
	if base == nil {
		base = http.DefaultTransport
	}
	return &GitHubPagination{
		Base:   base,
		config: newConfig(opts...),
	}
}

func NewClient(base http.RoundTripper, opts ...Option) *http.Client {
	return &http.Client{
		Transport: New(base, opts...),
	}
}

func (g *GitHubPagination) RoundTrip(request *http.Request) (*http.Response, error) {
	reqConfig := g.config.GetRequestConfig(request)
	if reqConfig.Disabled {
		return g.Base.RoundTrip(request)
	}
	driver := reqConfig.GetDriver()

	// it is enough to call update-request once,
	// since query parameters are kept through the pagination.
	request = reqConfig.UpdateRequest(request)

	pageCount := 1
	var resp *http.Response
	for {
		var err error

		// send the request
		resp, err = g.Base.RoundTrip(request)
		if err != nil {
			driver.OnBadResponse(resp, err)
			return nil, err
		}

		// only paginate through successful requests.
		if resp.StatusCode != http.StatusOK {
			driver.OnBadResponse(resp, err)
			break
		}

		// get the next request for pagination
		request = github_response.GetNextRequest(request, resp)
		if err := driver.OnNextRequest(request, pageCount); err != nil {
			if drivers.ShouldStop(err) {
				break
			}
			return nil, err
		}

		if err := driver.OnNextResponse(resp, request, pageCount); err != nil {
			if drivers.ShouldStop(err) {
				break
			}
			return nil, err
		}

		// stop paginating if there are no more pages
		if request == nil {
			break
		}

		// update the count and check if we should stop paginating
		pageCount++
		if reqConfig.IsPaginationOverflow(pageCount) {
			break
		}
	}

	if err := driver.OnFinish(resp, pageCount); err != nil {
		return nil, err
	}
	return resp, nil
}
