module github.com/gofri/go-github-pagination-e2e

replace github.com/gofri/go-github-pagination => ../

go 1.21

require (
	github.com/gofri/go-github-pagination v0.0.0-00010101000000-000000000000
	github.com/google/go-querystring v1.1.0 // indirect
)

require (
	github.com/gofri/go-github-ratelimit v1.1.0
	github.com/google/go-github/v58 v58.0.0
)
