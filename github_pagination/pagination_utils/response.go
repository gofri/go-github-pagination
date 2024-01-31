package pagination_utils

import (
	"net/http"

	"github.com/gofri/go-github-pagination/github_pagination/pagination_utils/pagination_response"
)

func GetNextRequest(prevRequest *http.Request, prevResponse *http.Response) *http.Request {
	return pagination_response.NewParser().GetNextRequest(prevRequest, prevResponse)
}
