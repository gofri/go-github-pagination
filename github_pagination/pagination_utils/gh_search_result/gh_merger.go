package gh_search_result

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// see the main README.md for an explanation of this file.
type merger struct {
	totalCount        int
	incompleteResults bool
}

func NewMerger() *merger {
	return &merger{}
}

func (g *merger) Digest(reader io.Reader) (slice json.RawMessage, err error) {
	var result Untyped
	err = json.NewDecoder(reader).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to digest next map part: %w", err)
	}

	g.totalCount += result.TotalCount
	g.incompleteResults = g.incompleteResults || result.IncompleteResults

	return *result.Items, nil
}

func (g *merger) Finalize(sliceReader io.Reader) (mapReader io.Reader) {
	if g.totalCount == 0 {
		return bytes.NewReader([]byte{})
	}
	preSlice := g.getPreSliceReader()
	postSlice := g.getPostSliceReader()
	return io.MultiReader(preSlice, sliceReader, postSlice)
}

func (g *merger) getPreSliceReader() io.Reader {
	preSliceText := fmt.Sprintf(`{"total_count": %d, "incomplete_results": %v, "items": `,
		g.totalCount, g.incompleteResults)
	return strings.NewReader(preSliceText)
}

func (g *merger) getPostSliceReader() io.Reader {
	postSliceText := "}"
	return strings.NewReader(postSliceText)
}
