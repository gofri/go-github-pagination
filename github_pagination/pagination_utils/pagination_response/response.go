package pagination_response

import (
	"net/http"
	"net/url"
	"strings"
)

type paginationSubParser interface {
	Parse(query *url.Values, relType RelType) bool
	GetNextQueryParams() map[string]string
}

type paginationResponseParser struct {
	subparsers []paginationSubParser
}

func NewParser() *paginationResponseParser {
	return &paginationResponseParser{
		subparsers: []paginationSubParser{
			&cursorSubParser{},
			&beforeAfterSubParser{},
			&pageSubParser{},
			&sinceSubParser{},
		},
	}
}

func (p *paginationResponseParser) parse(response *http.Response) map[string]string {
	if response == nil {
		return nil
	}
	if response.Header == nil {
		return nil
	}
	linkHeader, ok := response.Header["Link"]
	if !ok || len(linkHeader) == 0 {
		return nil
	}
	for _, links := range linkHeader {
		for _, link := range strings.Split(links, ",") {
			p.parseLink(link)
		}
	}
	return p.getNextQueryParams()
}

func (p *paginationResponseParser) GetNextRequest(prevRequest *http.Request, prevResponse *http.Response) *http.Request {
	params := p.parse(prevResponse)
	if params == nil {
		return nil
	}

	request := prevRequest.Clone(prevRequest.Context())
	query := request.URL.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	request.URL.RawQuery = query.Encode()

	return request
}

func (p *paginationResponseParser) getNextQueryParams() map[string]string {
	for _, subparser := range p.subparsers {
		if params := subparser.GetNextQueryParams(); params != nil {
			return params
		}
	}
	return nil
}

func (p *paginationResponseParser) parseLink(link string) {
	segments := strings.Split(strings.TrimSpace(link), ";")
	if len(segments) < 2 {
		return
	}
	query := p.hrefToQeury(segments[0])
	for _, segment := range segments[1:] {
		p.parseSegment(segment, query)
	}
}

func (p *paginationResponseParser) parseSegment(segment string, query *url.Values) {
	relType := getRelType(segment)
	if relType == RelTypeUnknown {
		return
	}
	for _, subparser := range p.subparsers {
		if subparser.Parse(query, relType) {
			break
		}
	}
}

func (p *paginationResponseParser) hrefToQeury(formattedHref string) *url.Values {
	formattedHref = strings.TrimSpace(formattedHref)
	formattedHrefLen := len(formattedHref)
	if formattedHrefLen < 2 {
		return nil
	}
	prefix, href, suffix := formattedHref[:1], formattedHref[1:formattedHrefLen-1], formattedHref[formattedHrefLen-1:]
	if prefix != "<" || suffix != ">" {
		return nil
	}
	url, err := url.Parse(href)
	if err != nil {
		return nil
	}
	query := url.Query()
	return &query
}
