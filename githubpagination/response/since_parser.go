package response

import "net/url"

const sinceNextKey = "since"

type sinceSubParser struct {
	paginationOptions
}

func (p *sinceSubParser) Parse(query *url.Values, relType RelType) bool {
	since := query.Get(sinceNextKey)
	if since == "" {
		return false
	}
	p.SetRel(relType, since)
	return true
}

func (p *sinceSubParser) GetNextQueryParams() map[string]string {
	return p.GetNextAs(sinceNextKey)
}
