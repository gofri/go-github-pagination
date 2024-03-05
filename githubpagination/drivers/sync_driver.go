package drivers

import (
	"io"
	"net/http"

	"github.com/gofri/go-github-pagination/githubpagination/jsonmerger"
)

type SyncPaginationDriver struct {
	merger jsonmerger.JSONMerger
}

func NewSyncPaginationDriver() *SyncPaginationDriver {
	return &SyncPaginationDriver{
		merger: jsonmerger.NewMerger(),
	}
}

func (d *SyncPaginationDriver) OnNextRequest(request *http.Request, pageCount int) error {
	// early-exit for non-paginated requests
	if isNonPaginatedRequest(request, pageCount) {
		return ErrStopPagination
	}
	return nil
}

func (d *SyncPaginationDriver) OnNextResponse(response *http.Response, nextRequest *http.Request, pageCount int) error {
	if err := d.merger.ReadNext(response.Body); err != nil {
		return err
	}
	return nil
}

func (d *SyncPaginationDriver) OnFinish(response *http.Response, pageCount int) error {
	if pageCount > 1 {
		response.Body = io.NopCloser(d.merger.Merged())
	}
	return nil
}

func (d *SyncPaginationDriver) OnBadResponse(response *http.Response, err error) {
}
