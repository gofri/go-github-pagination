package pagination_drivers

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
)

type asyncPaginationRawHandler interface {
	HandleRawPage(response *http.Response) error
	HandleRawError(err error, response *http.Response)
	HandleRawFinish(response *http.Response, pageCount int)
}

type asyncPaginationRawDriver struct {
	handler   asyncPaginationRawHandler
	waiter    sync.WaitGroup
	respError atomic.Pointer[error]
}

func NewAsyncPaginationRawDriver(handler asyncPaginationRawHandler) *asyncPaginationRawDriver {
	return &asyncPaginationRawDriver{
		handler: handler,
	}
}

func (d *asyncPaginationRawDriver) OnNextRequest(request *http.Request, pageCount int) error {
	if err := d.respError.Load(); err != nil {
		return *err
	}

	return nil
}

func (d *asyncPaginationRawDriver) OnNextResponse(response *http.Response, nextRequest *http.Request, pageCount int) (err error) {
	d.waiter.Add(1)
	go func(response *http.Response) {
		defer d.waiter.Done()
		defer func() {
			response.Body.Close()
			response.Body = io.NopCloser(bytes.NewReader([]byte{}))
		}()
		if err := d.handler.HandleRawPage(response); err != nil {
			d.respError.Store(&err)
			d.handler.HandleRawError(err, response)
		}
	}(response)

	// non-paginated requests still have to go through the handler,
	// so only stop AFTER the first one
	if isNonPaginatedRequest(nextRequest, pageCount) {
		return StopPaginationError
	}

	return nil
}

func (d *asyncPaginationRawDriver) OnFinish(response *http.Response, pageCount int) error {
	// wait BEFORE calling the finish handler,
	// so that errors from page handlers are handled (instead of nil)
	d.waiter.Wait()
	d.handler.HandleRawFinish(response, pageCount)
	return nil
}

func (d *asyncPaginationRawDriver) OnBadResponse(response *http.Response, err error) {
	d.handler.HandleRawError(err, response)
}
