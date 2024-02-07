package json_merger

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// see the main README.md for an explanation of this file.

type pagedMapResponse struct {
	TotalCount       int              `json:"total_count"`
	IncompleteResult bool             `json:"incomplete_results"`
	Items            *json.RawMessage `json:"items"`
}

type githubMapCombiner struct {
	totalCount        int
	incompleteResults bool
}

func (g *githubMapCombiner) Digest(reader io.Reader) (slice json.RawMessage, err error) {
	var result pagedMapResponse
	err = json.NewDecoder(reader).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to digest next map part: %w", err)
	} else if result.Items == nil {
		return nil, &NotPaginatableDictError{}
	}

	g.totalCount += result.TotalCount
	g.incompleteResults = g.incompleteResults || result.IncompleteResult

	return *result.Items, nil
}

func (g *githubMapCombiner) Finalize(sliceReader io.Reader) (mapReader io.Reader) {
	preSlice := g.getPreSliceReader()
	postSlice := g.getPostSliceReader()
	return io.MultiReader(preSlice, sliceReader, postSlice)
}

func (g *githubMapCombiner) getPreSliceReader() io.Reader {
	preSliceText := fmt.Sprintf(`{"total_count": %d, "incomplete_results": %v, "items": `,
		g.totalCount, g.incompleteResults)
	return strings.NewReader(preSliceText)
}

func (g *githubMapCombiner) getPostSliceReader() io.Reader {
	postSliceText := "}"
	return strings.NewReader(postSliceText)
}
