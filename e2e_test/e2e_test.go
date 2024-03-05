package e2e_test

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofri/go-github-pagination/githubpagination"
	"github.com/gofri/go-github-pagination/githubpagination/drivers"
	"github.com/gofri/go-github-pagination/githubpagination/searchresult"
	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v58/github"
)

func getRateLimitHandler() http.RoundTripper {
	client, err := github_ratelimit.NewRateLimitWaiter(nil,
		github_ratelimit.WithLimitDetectedCallback(func(cb *github_ratelimit.CallbackContext) {
			log.Printf("secondary rate limit detected: %v", cb)
		},
		))
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func getGithubClient(httpClient *http.Client) *github.Client {
	client := github.NewClient(httpClient)
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		client = client.WithAuthToken(token)
	}
	return client
}

func getAuthGithubClientOrSkip(t *testing.T, httpClient *http.Client) *github.Client {
	client := github.NewClient(httpClient)
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		client = client.WithAuthToken(token)
	} else {
		t.Skip("skipping test; GITHUB_TOKEN not set")
	}
	return client
}

func TestSlice(t *testing.T) {
	pager := githubpagination.NewClient(getRateLimitHandler())
	gh := getGithubClient(pager)
	perPage := 3

	repos, _, err := gh.Repositories.ListByUser(context.Background(),
		"gofri",
		&github.RepositoryListByUserOptions{
			ListOptions: github.ListOptions{
				PerPage: perPage,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if count := len(repos); count <= perPage {
		t.Fatalf("expected more than %d repos, got %d", perPage, count)
	}
	log.Printf("found %v repos: \n", len(repos))
}

func TestNoPagination(t *testing.T) {
	pager := githubpagination.NewClient(getRateLimitHandler())
	gh := getGithubClient(pager)

	issue, _, err := gh.Issues.Get(context.Background(), "gofri", "go-github-pagination", 1)
	if err != nil {
		t.Fatal(err)
	}
	if issue == nil {
		t.Fatal("expected an issue")
	}
}

func TestSearch(t *testing.T) {
	perPage := 3
	maxPages := 2
	pager := githubpagination.NewClient(getRateLimitHandler(),
		githubpagination.WithPerPage(perPage),
		githubpagination.WithMaxNumOfPages(maxPages),
	)
	gh := getAuthGithubClientOrSkip(t, pager)
	if gh == nil {
		t.Skipf("skipping test; GITHUB_TOKEN not set")
	}
	result, resp, err := gh.Search.Code(context.Background(), "go_github", &github.SearchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if result == nil {
		t.Fatal("expected a result")
	}
	if got, want := len(result.CodeResults), perPage*maxPages; got != want {
		t.Fatalf("expected %d code results, got %d", want, got)
	}
}

type customRawHandler struct {
	t *testing.T
}

func (c *customRawHandler) HandleRawPage(resp *http.Response) error {
	if resp == nil {
		c.t.Fatal("expected a response")
	}
	if resp.StatusCode != http.StatusOK {
		c.t.Fatalf("expected status code 200, got %d", resp.StatusCode)
	}
	if resp.Body == nil {
		c.t.Fatal("expected a response body")
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatal(err)
	}
	if len(bytes) == 0 {
		c.t.Fatal("expected a non-empty response body")
	}
	return nil
}
func (c *customRawHandler) HandleRawError(err error, response *http.Response)      {}
func (c *customRawHandler) HandleRawFinish(response *http.Response, pageCount int) {}
func TestAsync(t *testing.T) {

	t.Run("async-no-pagination", func(t *testing.T) {
		pager := githubpagination.NewClient(getRateLimitHandler())
		gh := getGithubClient(pager)

		handler := func(resp *http.Response, issues []*github.Issue) error {
			return nil
		}
		async := githubpagination.NewAsync(handler)
		err := async.Paginate(gh.Issues.Get, context.Background(), "gofri", "go-github-pagination", 1)
		if err == nil {
			t.Fatal("expected an error -- this is not a valid return type")
		}

		// now let's test with the raw driver - we DO expect a result
		customHandler := &customRawHandler{t: t}
		ctx := githubpagination.WithOverrideConfig(context.Background(),
			githubpagination.WithDriver(
				drivers.NewAsyncPaginationRawDriver(customHandler),
			),
		)
		_, _, err = gh.Issues.Get(ctx, "gofri", "go-github-pagination", 1)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("async-slice", func(t *testing.T) {
		perPage := 3
		maxPages := 2
		pager := githubpagination.NewClient(getRateLimitHandler(),
			githubpagination.WithPerPage(perPage),
			githubpagination.WithMaxNumOfPages(maxPages),
		)
		gh := getGithubClient(pager)

		var total atomic.Int64
		handler := func(resp *http.Response, repos []*github.Repository) error {
			total.Add(int64(len(repos)))
			return nil
		}
		err := githubpagination.NewAsync(handler).Paginate(gh.Repositories.ListByUser,
			context.Background(),
			"gofri",
			nil,
		)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := int(total.Load()), perPage*maxPages; got != want {
			t.Fatalf("expected %d repos, got %d", want, got)
		}
	})

	t.Run("async-search", func(t *testing.T) {
		perPage := 5
		maxPages := 2
		pager := githubpagination.NewClient(getRateLimitHandler())
		gh := getAuthGithubClientOrSkip(t, pager)

		// let's override the config in the request,
		// so that we test the double-override functionality
		ctx := githubpagination.WithOverrideConfig(context.Background(),
			githubpagination.WithPerPage(perPage),
			githubpagination.WithMaxNumOfPages(maxPages),
		)

		var total atomic.Int64
		handler := func(resp *http.Response, result *searchresult.Typed[github.CodeResult]) error {
			if result == nil {
				t.Fatalf("expected a result, got %v", result)
			} else if result.TotalCount < len(result.Items) {
				t.Fatalf("expected total count to be greater than or equal to the number of items, got %d", result.TotalCount)
			}

			total.Add(int64(len(result.Items)))
			return nil
		}
		if err := githubpagination.NewAsyncSearch(handler).Paginate(gh.Search.Code, ctx, "go_github", nil); err != nil {
			t.Fatal(err)
		}
		if got, want := int(total.Load()), perPage*maxPages; got != want {
			t.Fatalf("expected %d repos, got %d", want, got)
		}
	})

	t.Run("async-search-as-slice", func(t *testing.T) {
		// using the slice handler (ignoring total count and incomplete results)
		// works just as well as the search-result handler
		perPage := 5
		maxPages := 2
		pager := githubpagination.NewClient(getRateLimitHandler(),
			githubpagination.WithPerPage(perPage),
			githubpagination.WithMaxNumOfPages(maxPages),
		)
		gh := getAuthGithubClientOrSkip(t, pager)
		var total atomic.Int64
		sliceHandler := func(resp *http.Response, codeResults []*github.CodeResult) error {
			time.Sleep(500 * time.Millisecond) // simulate slow response (for testing purposes)
			total.Add(int64(len(codeResults)))
			return nil
		}
		ctx := context.Background()
		async := githubpagination.NewAsync(sliceHandler)
		if err := async.Paginate(gh.Search.Code, ctx, "go_github", nil); err != nil {
			t.Fatal(err)
		}
		if got, want := int(total.Load()), perPage*maxPages; got != want {
			t.Fatalf("expected %d repos, got %d", want, got)
		}
	})

	t.Run("async-early-exit", func(t *testing.T) {
		customError := errors.New("custom error")
		maxPages := 10
		gh := getGithubClient(githubpagination.NewClient(getRateLimitHandler()))
		var count atomic.Int64
		handler := func(resp *http.Response, result *searchresult.Typed[github.Repository]) error {
			if len(result.Items) == 0 {
				t.Fatalf("expected at least one repo, got %d", len(result.Items))
			}
			count.Add(1)
			return customError
		}
		ctx := githubpagination.WithOverrideConfig(context.Background(),
			githubpagination.WithPerPage(1),
			githubpagination.WithMaxNumOfPages(maxPages),
		)
		// we use the AsyncSearch constructor although this is a slice request,
		// just to show that it is possible
		err := githubpagination.NewAsyncSearch(handler).Paginate(gh.Repositories.ListByUser,
			ctx, "gofri", nil,
		)
		if err == nil {
			t.Fatal(err)
		} else if !errors.Is(err, customError) {
			t.Fatalf("expected customError, got %v", err)
		}
		if count.Load() > 3 { // flakey test (non-deterministic by design, but 3 is arbitrary)
			t.Fatalf("expected early exit, got %d", count.Load())
		}
	})

}
