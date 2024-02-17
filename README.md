# go-github-pagination

[![Go Report Card](https://goreportcard.com/badge/github.com/gofri/go-github-pagination)](https://goreportcard.com/report/github.com/gofri/go-github-pagination)

Package `go-github-pagination` provides an http.RoundTripper implementation that handles [Pagination](https://docs.github.com/en/rest/using-the-rest-api/using-pagination-in-the-rest-api) for the GitHub API.  
`go-github-pagination` can be used with any HTTP client communicating with GitHub API.  
It is meant to complement [go-github](https://github.com/google/go-github), but this repository is not associated with go-github repository nor Google.  

## Installation

```go get github.com/gofri/go-github-pagination```

## Usage Example (with [go-github](https://github.com/google/go-github))

```go
import "github.com/google/go-github/v58/github"
import "github.com/gofri/go-github-pagination/github_pagination/github_pagination"

func main() {
  paginator := github_pagination.NewGithubPaginationClient(nil,
    github_pagination.WithPerPage(100), // default to 100 results per page
  )
  client := github.NewClient(paginator).WithAuthToken("your personal access token")

  // now use the client as you please
}
```

## Client Options

The RoundTripper accepts a set of options to configure its behavior.
The options are:

- `WithPaginationEnabled` / `WithPaginationDisabled`: enable/disable pagination (default: enabled).
- `WithPerPage`: Set the default `per_page` value for requests (recommended: 100).
- `WithMaxNumOfPages`: Set the maximum number of pages to return.

## Per-Request Options

Use `WithOverrideConfig(opts...)` to override the configuration for a specific request (using the request context).  
Per-request configurations are especially useful if you want to enable/disable/limit pagination for specific requests.

## Known Limitations

All of these may be developed in the future (some are definitly on the roadmap).  
Please open an issue or a pull request if you need any.  
Unsupported features (at this point):

- Async interface (see below).
- Custom strategy in case of primary/secondary rate limits / errors.
- Callbacks.
- GraphQL pagination.

## Async Pagination

Async pagination refers to handling pages while fetching the next pages.  
Unfortunately, the interfaces of both http.Client & go_github.Client have a sync nature.  
This fact makes total sense for itself, but it makes it impossible to shove async pagination under the hood without abusing the interface.  
As a result, async pagination must be supported via an additional interface.
Specifically, I intend to provide an interface that accepts go-github functions for usability.
Please feel free to share you thoughts/needs for this interface.

## Incomplete Results

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

## Internals - How Does It Work?

The implementation consists of a few building blocks:

- `json_merger`: merges the response body (slice/map) of the pages.
- `pagination_utils`: utilities to handle the pagination API used by GitHub.
- `github_pagination`: the main package that glues everything into an http.RoundTripper.

## GitHub Pagination API Documentation References

- [using pagination in the rest api](https://docs.github.com/en/rest/using-the-rest-api/using-pagination-in-the-rest-api)
- [using-pagination-in-the-graphql-api](https://docs.github.com/en/graphql/guides/using-pagination-in-the-graphql-api)

## License

This package is distributed under the MIT license found in the LICENSE file.  
Contribution and feedback is welcome.
