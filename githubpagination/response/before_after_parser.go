package response

import "net/url"

const afterKey = "after"
const beforeKey = "before"

type beforeAfterSubParser struct {
	paginationOptions
}

func (p *beforeAfterSubParser) Parse(query *url.Values, relType RelType) bool {
	value := p.getValue(query, relType)
	if value == "" {
		return false
	}
	p.SetRel(relType, value)
	return true
}

func (p *beforeAfterSubParser) GetNextQueryParams() map[string]string {
	return p.GetNextAs(afterKey)
}

func (p *beforeAfterSubParser) getValue(query *url.Values, relType RelType) string {
	switch relType {
	case RelTypeNext:
		return query.Get(afterKey)
	case RelTypePrev:
		return query.Get(beforeKey)
	default:
		return ""
	}
}
