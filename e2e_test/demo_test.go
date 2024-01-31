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
	token := tryLoad("GITHUB_TOKEN")
	user := tryLoad("GITHUB_USER")
	if token == "" || user == "" {
		t.Skip("skipping test, no token or user")
	}

	pager := github_pagination.NewGithubPaginationClient(nil)
	gh := github.NewClient(pager).WithAuthToken(token)

	repos, _, err := gh.Repositories.ListByUser(context.Background(), "gofri",
		&github.RepositoryListByUserOptions{
			ListOptions: github.ListOptions{
				PerPage: 5,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if count := len(repos); count <= 5 {
		t.Fatal("expected more than 5 repos, got ", count)
	}
	log.Printf("found %v repos: \n", len(repos))
}
