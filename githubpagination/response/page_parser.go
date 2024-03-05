package response

import "net/url"

const pageKey = "page"

type pageSubParser struct {
	paginationOptions
}

func (p *pageSubParser) Parse(query *url.Values, relType RelType) bool {
	page := query.Get(pageKey)
	if page == "" {
		return false
	}
	p.SetRel(relType, page)
	return true
}

func (p *pageSubParser) GetNextQueryParams() map[string]string {
	return p.GetNextAs(pageKey)
}
