package github_pagination_test

import (
	"bytes"
	"context"
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
	body     bytes.Buffer
	closeCnt *int
}

func (c *ClosableBody) Read(p []byte) (n int, err error) {
	return c.body.Read(p)
}
func (c *ClosableBody) Close() error {
	*c.closeCnt += 1
	return nil
}

type server struct {
	t          *testing.T
	CloseCnt   int
	Iterations int
}

func (s *server) Reset() {
	s.CloseCnt = 0
	s.Iterations = 0
}

func (s *server) CompleteData() []int {
	var data []int
	for i := 0; i < totalItems; i++ {
		data = append(data, i)
	}
	return data
}

func (s *server) TestFullResponse(resp *http.Response, numPages int) {
	s.TestPartialResponse(resp, numPages, totalItems)
}
func (s *server) TestPartialResponse(resp *http.Response, numPages int, total int) {
	defer s.Reset()
	if got, want := s.CloseCnt, numPages; got != want {
		s.t.Fatalf("expected %d close calls, got %d", want, got)
	}

	var body []int
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		s.t.Fatalf("failed to decode response body: %v", err)
	}
	data := s.CompleteData()
	data = data[:total]
	if len(body) != total {
		s.t.Fatalf("expected %d items, got %d", totalItems, len(body))
	}
	if got, want := body, data; slices.Compare(got, want) != 0 {
		s.t.Fatalf("expected %v, got %v", want, got)
	}
	if s.Iterations != numPages {
		s.t.Fatalf("expected %d iterations, got %d", numPages, s.Iterations)
	}
	if err := resp.Body.Close(); err != nil {
		s.t.Fatalf("failed to close response body: %v", err)
	}
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
	s.Iterations += 1
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
		body:     *bytes.NewBuffer(body),
		closeCnt: &s.CloseCnt,
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
	t.Parallel()
	numPages := 4
	perPage := totalItems / numPages
	server := &server{t: t}
	pagination := github_pagination.NewGithubPaginationClient(server)
	resp, err := pagination.Get(fmt.Sprintf("http://example.com?per_page=%d", perPage))
	if err != nil {
		t.Fatalf("failed to get response: %v", err)
	}
	server.TestFullResponse(resp, numPages)
}

func TestConfig(t *testing.T) {
	t.Parallel()
	server := &server{t: t}

	t.Run("PerPage", func(t *testing.T) {
		pagination := github_pagination.NewGithubPaginationClient(server,
			github_pagination.WithPerPage(10),
		)
		body, err := pagination.Get("http://example.com?per_page=5")
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		server.TestFullResponse(body, 2)
	})

	t.Run("MaxPages", func(t *testing.T) {
		pagination := github_pagination.NewGithubPaginationClient(server,
			github_pagination.WithPerPage(3),
			github_pagination.WithMaxNumOfPages(2))
		body, err := pagination.Get("http://example.com?per_page=5")
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		server.TestPartialResponse(body, 2, 6)
	})

	t.Run("Disabled", func(t *testing.T) {
		pagination := github_pagination.NewGithubPaginationClient(server,
			github_pagination.WithPaginationDisabled())
		resp, err := pagination.Get("http://example.com?per_page=5")
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		defer server.Reset()
		if got, want := server.CloseCnt, 0; got != want {
			t.Fatalf("expected %d close calls, got %d", want, got)
		}

		var body []int
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		data := server.CompleteData()[:5]
		if len(body) != 5 {
			t.Fatalf("expected %d items, got %d", totalItems, len(body))
		}
		if got, want := body, data; slices.Compare(got, want) != 0 {
			t.Fatalf("expected %v, got %v", want, got)
		}
		if server.Iterations != 1 {
			t.Fatalf("expected %d iterations, got %d", 1, server.Iterations)
		}
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}

	})

	t.Run("PerRequestConfig", func(t *testing.T) {
		numPages := 4
		perPage := totalItems / numPages
		pagination := github_pagination.NewGithubPaginationClient(server,
			github_pagination.WithPaginationDisabled())

		req, err := http.NewRequestWithContext(
			github_pagination.WithOverrideConfig(
				context.Background(),
				github_pagination.WithPaginationEnabled(),
			),
			"GET",
			fmt.Sprintf("http://example.com?per_page=%d", perPage),
			nil,
		)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		body, err := pagination.Do(req)
		if err != nil {
			t.Fatalf("failed to get response: %v", err)
		}
		server.TestFullResponse(body, 4)
	})
}
