package github_pagination_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"testing"

	"github.com/gofri/go-github-pagination/github_pagination/github_pagination"
)

const totalItems = 20

type ClosableBody struct {
	body   bytes.Buffer
	closed *bool
}

func (c *ClosableBody) Read(p []byte) (n int, err error) {
	return c.body.Read(p)
}
func (c *ClosableBody) Close() error {
	*c.closed = true
	return nil
}

type server struct {
	t      *testing.T
	closed bool
}

func (s *server) CompleteData() []int {
	var data []int
	for i := 0; i < totalItems; i++ {
		data = append(data, i)
	}
	return data
}

func (s *server) getBody(page int, perPage int) []byte {
	data := s.CompleteData()
	start := (page - 1) * perPage
	if start >= len(data) {
		return nil
	}
	end := start + perPage
	if end > len(data) {
		end = len(data)
	}
	body := data[start:end]
	asBytes, err := json.Marshal(body)
	if err != nil {
		s.t.Fatalf("failed to marshal body: %v", err)
	}
	return asBytes
}

func (s *server) getHeader(page int, perPage int) http.Header {
	if page*perPage < totalItems {
		return http.Header{
			"Link": []string{
				fmt.Sprintf(`<http://example.com?page=%d&per_page=%d>; rel="next"`,
					page+1, perPage),
			},
		}
	} else {
		return http.Header{}
	}
}

func (s *server) RoundTrip(req *http.Request) (*http.Response, error) {
	page := req.URL.Query().Get("page")
	perPage := req.URL.Query().Get("per_page")
	pageInt, err := strconv.Atoi(page)
	if page == "" {
		pageInt = 1
	} else if err != nil {
		s.t.Fatalf("failed to convert page to int: %v", err)
	}
	perPageInt, err := strconv.Atoi(perPage)
	if err != nil {
		s.t.Fatalf("failed to convert per_page to int: %v", err)
	}
	body := s.getBody(pageInt, perPageInt)
	closable := &ClosableBody{
		body:   *bytes.NewBuffer(body),
		closed: &s.closed,
	}
	if body == nil {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       closable,
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       closable,
		Header:     s.getHeader(pageInt, perPageInt),
	}, nil
}

func TestBasic(t *testing.T) {
	server := &server{t: t}
	pagination := github_pagination.NewGithubPaginationClient(server)
	resp, err := pagination.Get("http://example.com?per_page=5")
	if err != nil {
		t.Fatalf("failed to get response: %v", err)
	}
	if !server.closed {
		t.Fatalf("server not closed")
	}
	var body []int
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if len(body) != totalItems {
		t.Fatalf("expected %d items, got %d", totalItems, len(body))
	}
	if got, want := body, server.CompleteData(); slices.Compare(got, want) != 0 {
		t.Fatalf("expected %v, got %v", want, got)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("failed to close response body: %v", err)
	}
}
