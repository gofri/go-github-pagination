package e2e_test

import (
	"context"
	"log"
	"testing"

	"github.com/gofri/go-github-pagination/github_pagination/github_pagination"
	"github.com/google/go-github/v58/github"
)

func TestDemo(t *testing.T) {
	pager := github_pagination.NewGithubPaginationClient(nil)
	gh := github.NewClient(pager)

	per_page := 3

	repos, _, err := gh.Repositories.ListByUser(context.Background(),
		"gofri",
		&github.RepositoryListByUserOptions{
			ListOptions: github.ListOptions{
				PerPage: per_page,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if count := len(repos); count <= per_page {
		t.Fatalf("expected more than %d repos, got %d", per_page, count)
	}
	log.Printf("found %v repos: \n", len(repos))
}

func TestNonpagination(t *testing.T) {
	pager := github_pagination.NewGithubPaginationClient(nil)
	gh := github.NewClient(pager)

	issue, _, err := gh.Issues.Get(context.Background(), "gofri", "go-github-pagination", 1)
	if err != nil {
		t.Fatal(err)
	}
	if issue == nil {
		t.Fatal("expected an issue")
	}
}
