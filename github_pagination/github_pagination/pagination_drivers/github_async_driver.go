package pagination_drivers

import (
	"encoding/json"
	"net/http"

	"github.com/gofri/go-github-pagination/github_pagination/pagination_utils/gh_search_result"
)

type githubAsyncPaginationHandler[DataType any] interface {
	HandlePage(data *gh_search_result.Typed[DataType], response *http.Response) error
	HandleError(response *http.Response, err error)
	HandleFinish(response *http.Response, pageCount int)
}

// githubAsyncPaginationDriver is a wrapper around the raw driver.
// it is used to translate the raw responses to go-github styled responses.
// both sliced and search responses are translated to gh_search_result.Typed,
// so that the interface is simpler and unified.
type githubAsyncPaginationDriver[DataType any] struct {
	asyncPaginationRawDriver
}

func NewGithubAsyncPaginationDriver[DataType any](handler githubAsyncPaginationHandler[DataType], isSearchResponse bool) *githubAsyncPaginationDriver[DataType] {
	return &githubAsyncPaginationDriver[DataType]{
		asyncPaginationRawDriver: asyncPaginationRawDriver{
			handler: &githubRawHandler[DataType]{
				handler:              handler,
				isSearchResponseType: isSearchResponse,
			},
		},
	}
}

type githubRawHandler[DataType any] struct {
	handler              githubAsyncPaginationHandler[DataType]
	isSearchResponseType bool
}

func (h *githubRawHandler[DataType]) HandleRawPage(response *http.Response) error {
	data, err := h.parseResponse(response)
	if err != nil {
		return err
	}
	if err := h.handler.HandlePage(data, response); err != nil {
		return err
	}
	return nil
}

func (h *githubRawHandler[DataType]) HandleRawFinish(response *http.Response, pageCount int) {
	h.handler.HandleFinish(response, pageCount)
}

func (h *githubRawHandler[DataType]) HandleRawError(err error, response *http.Response) {
	h.handler.HandleError(response, err)
}

func (h *githubRawHandler[DataType]) parseResponse(response *http.Response) (*gh_search_result.Typed[DataType], error) {
	if h.isSearchResponseType {
		return h.parseSearchResponse(response)
	} else {
		return h.parseSliceResponse(response)
	}
}

func (h *githubRawHandler[DataType]) parseSearchResponse(response *http.Response) (*gh_search_result.Typed[DataType], error) {
	var untyped gh_search_result.Untyped
	if err := json.NewDecoder(response.Body).Decode(&untyped); err != nil {
		return nil, err
	}
	typed, err := gh_search_result.UntypedToTyped[DataType](&untyped)
	if err != nil {
		return nil, err
	}
	return typed, nil
}

func (h *githubRawHandler[DataType]) parseSliceResponse(response *http.Response) (*gh_search_result.Typed[DataType], error) {
	untyped := make([]*DataType, 0)
	if err := json.NewDecoder(response.Body).Decode(&untyped); err != nil {
		return nil, err
	}
	return gh_search_result.FromSlice(untyped), nil
}
