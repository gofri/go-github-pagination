package drivers

import (
	"encoding/json"
	"net/http"

	"github.com/gofri/go-github-pagination/githubpagination/searchresult"
)

type githubAsyncPaginationHandler[DataType any] interface {
	HandlePage(data *searchresult.Typed[DataType], resp *http.Response) error
	HandleError(resp *http.Response, err error)
	HandleFinish(resp *http.Response, pageCount int)
}

// GithubAsyncPaginationDriver is a wrapper around the raw driver.
// it is used to translate the raw responses to go-github styled responses.
// both sliced and search responses are translated to searchresult.Typed,
// so that the interface is simpler and unified.
type GithubAsyncPaginationDriver[DataType any] struct {
	AsyncPaginationRawDriver
}

func NewGithubAsyncPaginationDriver[DataType any](handler githubAsyncPaginationHandler[DataType], isSearchResponse bool) *GithubAsyncPaginationDriver[DataType] {
	return &GithubAsyncPaginationDriver[DataType]{
		AsyncPaginationRawDriver: AsyncPaginationRawDriver{
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

func (h *githubRawHandler[DataType]) parseResponse(response *http.Response) (*searchresult.Typed[DataType], error) {
	if h.isSearchResponseType {
		return h.parseSearchResponse(response)
	}
	return h.parseSliceResponse(response)
}

func (h *githubRawHandler[DataType]) parseSearchResponse(response *http.Response) (*searchresult.Typed[DataType], error) {
	var untyped searchresult.Untyped
	if err := json.NewDecoder(response.Body).Decode(&untyped); err != nil {
		return nil, err
	}
	typed, err := searchresult.UntypedToTyped[DataType](&untyped)
	if err != nil {
		return nil, err
	}
	return typed, nil
}

func (h *githubRawHandler[DataType]) parseSliceResponse(response *http.Response) (*searchresult.Typed[DataType], error) {
	untyped := make([]*DataType, 0)
	if err := json.NewDecoder(response.Body).Decode(&untyped); err != nil {
		return nil, err
	}
	return searchresult.FromSlice(untyped), nil
}
