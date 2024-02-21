package pagination_drivers

import (
	"io"
	"net/http"

	"github.com/gofri/go-github-pagination/github_pagination/json_merger"
)

type syncPaginationDriver struct {
	merger json_merger.JsonMerger
}

func NewSyncPaginationDriver() *syncPaginationDriver {
	return &syncPaginationDriver{
		merger: json_merger.NewMerger(),
	}
}

func (d *syncPaginationDriver) OnNextRequest(request *http.Request, pageCount int) error {
	// early-exit for non-paginated requests
	if isNonPaginatedRequest(request, pageCount) {
		return StopPaginationError
	}
	return nil
}

func (d *syncPaginationDriver) OnNextResponse(response *http.Response, nextRequest *http.Request, pageCount int) error {
	if err := d.merger.ReadNext(response.Body); err != nil {
		return err
	}
	return nil
}

func (d *syncPaginationDriver) OnFinish(response *http.Response, pageCount int) error {
	if pageCount > 1 {
		response.Body = io.NopCloser(d.merger.Merged())
	}
	return nil
}

func (d *syncPaginationDriver) OnBadResponse(response *http.Response, err error) {
}
