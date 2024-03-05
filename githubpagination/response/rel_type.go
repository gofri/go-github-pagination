package response

import "strings"

// RelType represents the type of the link relation.
type RelType string

// RelType constants.
const (
	RelTypeNext    RelType = `rel="next"`
	RelTypePrev    RelType = `rel="prev"`
	RelTypeFirst   RelType = `rel="first"`
	RelTypeLast    RelType = `rel="last"`
	RelTypeUnknown RelType = ``
)

func getRelType(segment string) RelType {
	switch strings.TrimSpace(segment) {
	case string(RelTypeNext):
		return RelTypeNext
	case string(RelTypePrev):
		return RelTypePrev
	case string(RelTypeFirst):
		return RelTypeFirst
	case string(RelTypeLast):
		return RelTypeLast
	default:
		return RelTypeUnknown
	}
}

type paginationOptions struct {
	Next  string
	Prev  string
	First string
	Last  string
}

func (p *paginationOptions) SetRel(relType RelType, value string) {
	switch relType {
	case RelTypeNext:
		p.Next = value
	case RelTypePrev:
		p.Prev = value
	case RelTypeFirst:
		p.First = value
	case RelTypeLast:
		p.Last = value
	}
}

func (p *paginationOptions) GetNextAs(key string) map[string]string {
	if p.Next == "" {
		return nil
	}
	return map[string]string{
		key: p.Next,
	}
}
