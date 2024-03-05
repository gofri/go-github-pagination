package drivers

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
)

type asyncPaginationRawHandler interface {
	HandleRawPage(resp *http.Response) error
	HandleRawError(err error, resp *http.Response)
	HandleRawFinish(resp *http.Response, pageCount int)
}

type AsyncPaginationRawDriver struct {
	handler   asyncPaginationRawHandler
	waiter    sync.WaitGroup
	respError atomic.Pointer[error]
}

func NewAsyncPaginationRawDriver(handler asyncPaginationRawHandler) *AsyncPaginationRawDriver {
	return &AsyncPaginationRawDriver{
		handler: handler,
	}
}

func (d *AsyncPaginationRawDriver) OnNextRequest(request *http.Request, pageCount int) error {
	if err := d.respError.Load(); err != nil {
		return *err
	}

	return nil
}

func (d *AsyncPaginationRawDriver) OnNextResponse(resp *http.Response, nextRequest *http.Request, pageCount int) (err error) {
	d.waiter.Add(1)
	go func(resp *http.Response) {
		defer d.waiter.Done()
		defer func() {
			resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewReader([]byte{}))
		}()
		if err := d.handler.HandleRawPage(resp); err != nil {
			d.respError.Store(&err)
			d.handler.HandleRawError(err, resp)
		}
	}(resp)

	// non-paginated requests still have to go through the handler,
	// so only stop AFTER the first one
	if isNonPaginatedRequest(nextRequest, pageCount) {
		return ErrStopPagination
	}

	return nil
}

func (d *AsyncPaginationRawDriver) OnFinish(resp *http.Response, pageCount int) error {
	// wait BEFORE calling the finish handler,
	// so that errors from page handlers are handled (instead of nil)
	d.waiter.Wait()
	d.handler.HandleRawFinish(resp, pageCount)
	return nil
}

func (d *AsyncPaginationRawDriver) OnBadResponse(resp *http.Response, err error) {
	d.handler.HandleRawError(err, resp)
}
