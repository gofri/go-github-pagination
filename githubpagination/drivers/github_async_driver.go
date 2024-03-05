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

func (h *githubRawHandler[DataType]) HandleRawPage(resp *http.Response) error {
	data, err := h.parseResponse(resp)
	if err != nil {
		return err
	}
	if err := h.handler.HandlePage(data, resp); err != nil {
		return err
	}
	return nil
}

func (h *githubRawHandler[DataType]) HandleRawFinish(resp *http.Response, pageCount int) {
	h.handler.HandleFinish(resp, pageCount)
}

func (h *githubRawHandler[DataType]) HandleRawError(err error, resp *http.Response) {
	h.handler.HandleError(resp, err)
}

func (h *githubRawHandler[DataType]) parseResponse(resp *http.Response) (*searchresult.Typed[DataType], error) {
	if h.isSearchResponseType {
		return h.parseSearchResponse(resp)
	}
	return h.parseSliceResponse(resp)
}

func (h *githubRawHandler[DataType]) parseSearchResponse(resp *http.Response) (*searchresult.Typed[DataType], error) {
	var untyped searchresult.Untyped
	if err := json.NewDecoder(resp.Body).Decode(&untyped); err != nil {
		return nil, err
	}
	typed, err := searchresult.UntypedToTyped[DataType](&untyped)
	if err != nil {
		return nil, err
	}
	return typed, nil
}

func (h *githubRawHandler[DataType]) parseSliceResponse(resp *http.Response) (*searchresult.Typed[DataType], error) {
	untyped := make([]*DataType, 0)
	if err := json.NewDecoder(resp.Body).Decode(&untyped); err != nil {
		return nil, err
	}
	return searchresult.FromSlice(untyped), nil
}
