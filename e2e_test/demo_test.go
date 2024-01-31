package e2e_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/gofri/go-github-pagination/github_pagination/github_pagination"
	"github.com/google/go-github/v58/github"
)

func tryLoad(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	value, err := os.ReadFile(fmt.Sprintf("/tmp/github/%s", key))
	if err == nil {
		return strings.TrimSpace(string(value))
	}
	return ""
}

func TestDemo(t *testing.T) {
	token := tryLoad("GITHUB_PAT")
	if token == "" {
		t.Fatal("skipping test, no token provided")
	}

	pager := github_pagination.NewGithubPaginationClient(nil)
	gh := github.NewClient(pager).WithAuthToken(token)

	per_page := 3

	repos, _, err := gh.Repositories.ListByAuthenticatedUser(context.Background(),
		&github.RepositoryListByAuthenticatedUserOptions{
			ListOptions: github.ListOptions{
				PerPage: per_page,
			},
			Visibility: "public",
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
