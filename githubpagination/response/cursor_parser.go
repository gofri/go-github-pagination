package response

import "net/url"

const cursorKey = "cursor"

type cursorSubParser struct {
	paginationOptions
}

func (p *cursorSubParser) Parse(query *url.Values, relType RelType) bool {
	cursor := query.Get(cursorKey)
	if cursor == "" {
		return false
	}
	p.SetRel(relType, cursor)
	return true
}

func (p *cursorSubParser) GetNextQueryParams() map[string]string {
	return p.GetNextAs(cursorKey)
}
