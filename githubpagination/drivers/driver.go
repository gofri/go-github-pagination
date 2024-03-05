package drivers

import (
	"errors"
	"net/http"
)

var ErrStopPagination = errors.New("stop pagination")

type Driver interface {
	OnNextRequest(request *http.Request, pageCount int) error
	OnNextResponse(resp *http.Response, nextRequest *http.Request, pageCount int) error
	OnFinish(resp *http.Response, pageCount int) error
	OnBadResponse(resp *http.Response, err error)
}

func ShouldStop(err error) bool {
	return errors.Is(err, ErrStopPagination)
}

func isNonPaginatedRequest(nextRequest *http.Request, pageCount int) bool {
	return nextRequest == nil && pageCount == 1
}
