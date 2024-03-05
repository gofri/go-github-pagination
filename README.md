# go-github-pagination

[![Go Report Card](https://goreportcard.com/badge/github.com/gofri/go-github-pagination)](https://goreportcard.com/report/github.com/gofri/go-github-pagination)

Package `go-github-pagination` provides an http.RoundTripper implementation that handles [Pagination](https://docs.github.com/en/rest/using-the-rest-api/using-pagination-in-the-rest-api) for the GitHub API.  
`go-github-pagination` can be used with any HTTP client communicating with GitHub API.  
It is meant to complement [go-github](https://github.com/google/go-github), but this repository is not associated with go-github repository nor Google.  

## Recommended: Rate Limit Handling

Please checkout my other repository, [go-github-ratelimit](https://github.com/gofri/go-github-ratelimit).  
It supports rate limit handling out of the box, and plays well with the pagination round-tripper.  
It is best to stack the pagination round-tripper on top of the ratelimit round-tripper.  

## Installation

```go get github.com/gofri/go-github-pagination```

## Usage Example (with [go-github](https://github.com/google/go-github))

```go
import "github.com/google/go-github/v58/github"
import "github.com/gofri/go-github-pagination/githubpagination"

func main() {
  paginator := githubpagination.NewClient(nil,
    githubpagination.WithPerPage(100), // default to 100 results per page
  )
  client := github.NewClient(paginator).WithAuthToken("your personal access token")

  // now use the client as you please
}
```

## Client Options

The RoundTripper accepts a set of options to configure its behavior.
The options are:

- `WithPaginationEnabled` / `WithPaginationDisabled`: enable/disable pagination. default: enabled.
- `WithPerPage`: Set the default `per_page` value for requests. recommended: 100. default: not set, server-side decision.
- `WithMaxNumOfPages`: Set the maximum number of pages to return. default: unlimited.
- `WithDriver`: Use a custom pagination driver (see async pagination comment). default: sync.

## Per-Request Options

Use `WithOverrideConfig(opts...)` to override the configuration for a specific request (using the request context).  
Per-request configurations are especially useful if you want to enable/disable/limit pagination for specific requests.

## Async Pagination

Async pagination enables users to handle pages concurrently.  
Since the interfaces of both `http.Client` & `go_github.Client` are sync,  
the interface for async pagination uses wrappers.
The wrapper is designed to support go-github out of the box.  
You can find useful examples in the e2e-tests for different use cases.  
In addition, there are lower-level primitives for plumbers who want to implement their own pagination driver.  
Please dive into the code or open an issue for help with that.

_Note: please open an issue if you think that a channel-based interface would work better for you._

Usage example:

```go
  paginator := githubpagination.NewClient(nil,
    githubpagination.WithPerPage(100), // default to 100 results per page
  )
  client := github.NewClient(paginator).WithAuthToken("your personal access token")
  handler := func(resp *http.Response, repos []*github.Repository) error {
    fmt.Printf("found repos: %+v\n", repos)
    return nil
  }
  ctx := githubpagination.WithOverrideConfig(context.Background(),
    githubpagination.WithMaxNumOfPages(3), // e.g, limit number of pages for this request
  )
  async := githubpagination.NewAsync(handler)
  err := async.Paginate(client.Repositories.ListByUser, ctx, "gofri", nil)
  if err != nil {
    panic(err)
  }
```

## Search API Pagination - Incomplete Results

According to the (obscure) API documentation, some endpoints may return a dictionary instead of an array.
This return scheme is used to report incomplete results (due to timeouts).

The result is expected to be of the following structure (the actual items dictionary differs per endpoint):

```json
{
  "total_count": 0,
  "incomplete_results": false,
  "items": [{}]
}
```

The merge strategy used is to summarize the total_count, OR the incomplete_results, and join the items.
In practice, this special case appears to only occur with the Search API.  
Please report incidents with a different behaviour if you face them.

## Known Limitations

The following features may be implemented in the future, per request.
Please open an issue or a pull request if you need any.  

- Callbacks.
- GraphQL pagination.

## GitHub Pagination API Documentation References

- [Using pagination in the rest api](https://docs.github.com/en/rest/using-the-rest-api/using-pagination-in-the-rest-api)
- [Using pagination in the graphql api](https://docs.github.com/en/graphql/guides/using-pagination-in-the-graphql-api)
- [Search API - incomplete results](https://docs.github.com/en/rest/search/search#timeouts-and-incomplete-results)

## License

This package is distributed under the MIT license found in the LICENSE file.  
Contribution and feedback is welcome.
