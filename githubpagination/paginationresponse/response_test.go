package paginationresponse_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gofri/go-github-pagination/githubpagination/paginationresponse"
)

type linkTestSample struct {
	Title           string
	EndpointExample string
	Links           map[paginationresponse.RelType]string
	Expected        map[string]string
}

func (l linkTestSample) Test(t *testing.T) {
	t.Run(l.Title, func(t *testing.T) {
		t.Logf("testing %s which corresponds to %s", l.Title, l.EndpointExample)
		parser := paginationresponse.NewParser()
		request, err := http.NewRequest(`GET`, `https://api.github.com/example`, nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		request = parser.GetNextRequest(request, l.getResponse())
		if request == nil {
			if len(l.Expected) == 0 {
				return
			} else {
				t.Fatalf("expected %v query params, got none", len(l.Expected))
			}
		}
		query := request.URL.Query()
		if len(l.Expected) != len(query) {
			t.Fatalf("expected %v query params, got %v", len(l.Expected), len(query))
		}
		if len(l.Expected) == 0 {
			return
		}
		for key, expected := range l.Expected {
			if got, want := query.Get(key), expected; got != want {
				t.Fatalf("expected %v, got %v", want, got)
			}
		}
	})
}

func (l linkTestSample) getLinks() []string {
	links := []string{}
	for rel, link := range l.Links {
		links = append(links, fmt.Sprintf(`<%s>; %s`, link, rel))
	}
	return links
}

func (l linkTestSample) getResponse() *http.Response {
	header := http.Header{}
	header.Set(`Link`, strings.Join(l.getLinks(), `, `))
	return &http.Response{
		Header: header,
	}
}

// different test cases, based on observed responses from GitHub
var responseLinks = []linkTestSample{
	{
		Title:           "page next only",
		EndpointExample: `https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-organization-repositories`,
		Links: map[paginationresponse.RelType]string{
			paginationresponse.RelTypeNext: `https://api.github.com/example?page=2`,
		},
		Expected: map[string]string{`page`: `2`},
	},
	{
		Title:           "page prev only",
		EndpointExample: `https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-organization-repositories`,
		Links: map[paginationresponse.RelType]string{
			paginationresponse.RelTypePrev: `https://api.github.com/example?page=1`,
		},
		Expected: nil,
	},
	{
		Title:           "page next and prev",
		EndpointExample: `https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-organization-repositories`,
		Links: map[paginationresponse.RelType]string{
			paginationresponse.RelTypeNext: `https://api.github.com/example?page=3`,
			paginationresponse.RelTypePrev: `https://api.github.com/example?page=1`,
		},
		Expected: map[string]string{`page`: `3`},
	},
	{
		Title:           "page token",
		EndpointExample: `https://docs.github.com/en/enterprise-cloud@latest/rest/teams/team-sync?apiVersion=2022-11-28#list-idp-groups-for-an-organization`,
		Links: map[paginationresponse.RelType]string{
			paginationresponse.RelTypeNext: `https://api.github.com/example?page=ABC`,
		},
		Expected: map[string]string{`page`: `ABC`},
	},
	{
		Title:           "cursor next only",
		EndpointExample: `https://docs.github.com/en/rest/orgs/webhooks?apiVersion=2022-11-28#list-deliveries-for-an-organization-webhook`,
		Links: map[paginationresponse.RelType]string{
			paginationresponse.RelTypeNext: `https://api.github.com/example?cursor=ABC`,
		},
		Expected: map[string]string{`cursor`: `ABC`},
	},
	{
		Title:           "cursor prev only",
		EndpointExample: `https://docs.github.com/en/rest/orgs/webhooks?apiVersion=2022-11-28#list-deliveries-for-an-organization-webhook`,
		Links: map[paginationresponse.RelType]string{
			paginationresponse.RelTypePrev: `https://api.github.com/example?cursor=ABC`,
		},
		Expected: nil,
	},
	{
		Title:           "cursor next and prev",
		EndpointExample: `https://docs.github.com/en/rest/orgs/webhooks?apiVersion=2022-11-28#list-deliveries-for-an-organization-webhook`,
		Links: map[paginationresponse.RelType]string{
			paginationresponse.RelTypeNext: `https://api.github.com/example?cursor=ABC`,
			paginationresponse.RelTypePrev: `https://api.github.com/example?cursor=DEF`,
		},
		Expected: map[string]string{`cursor`: `ABC`},
	},
	{
		Title:           "after only",
		EndpointExample: `https://docs.github.com/en/enterprise-cloud@latest/rest/orgs/orgs?apiVersion=2022-11-28#get-the-audit-log-for-an-organization`,
		Links: map[paginationresponse.RelType]string{
			paginationresponse.RelTypeNext: `https://api.github.com/example?after=ABC`,
		},
		Expected: map[string]string{`after`: `ABC`},
	},
	{
		Title:           "before only",
		EndpointExample: `https://docs.github.com/en/enterprise-cloud@latest/rest/orgs/orgs?apiVersion=2022-11-28#get-the-audit-log-for-an-organization`,
		Links: map[paginationresponse.RelType]string{
			paginationresponse.RelTypePrev: `https://api.github.com/example?before=ABC`,
		},
		Expected: nil,
	},
	{
		Title:           "before and after",
		EndpointExample: `https://docs.github.com/en/enterprise-cloud@latest/rest/orgs/orgs?apiVersion=2022-11-28#get-the-audit-log-for-an-organization`,
		Links: map[paginationresponse.RelType]string{
			paginationresponse.RelTypeNext: `https://api.github.com/example?after=ABC`,
			paginationresponse.RelTypePrev: `https://api.github.com/example?before=ABC`,
		},
		Expected: map[string]string{`after`: `ABC`},
	},
	{
		Title:           "page and before after",
		EndpointExample: `https://docs.github.com/en/rest/code-scanning/code-scanning?apiVersion=2022-11-28`,
		Links: map[paginationresponse.RelType]string{
			paginationresponse.RelTypeNext: `https://api.github.com/example?page=5&after=ABC`,
			paginationresponse.RelTypePrev: `https://api.github.com/example?before=ABC`,
		},
		Expected: map[string]string{`after`: `ABC`},
	},
	{
		Title:           "since instead of page",
		EndpointExample: `https://docs.github.com/en/rest/orgs/orgs?apiVersion=2022-11-28#list-organizations`,
		Links: map[paginationresponse.RelType]string{
			paginationresponse.RelTypeNext: `https://api.github.com/example?since=ABC`,
		},
		Expected: map[string]string{`since`: `ABC`},
	},
	{
		Title:           "since not for paging",
		EndpointExample: `https://docs.github.com/en/rest/commits/commits?apiVersion=2022-11-28#list-commits`,
		Links: map[paginationresponse.RelType]string{
			paginationresponse.RelTypeNext: `https://api.github.com/example?page=10&since=ABC`,
		},
		Expected: map[string]string{`page`: `10`},
	},
}

func TestParser(t *testing.T) {
	for _, link := range responseLinks {
		link.Test(t)
	}
}
