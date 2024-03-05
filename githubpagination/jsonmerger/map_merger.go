package jsonmerger

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gofri/go-github-pagination/githubpagination/searchresult"
)

type unprocessedMapCombiner interface {
	Digest(io.Reader) (slice json.RawMessage, err error)
	Finalize(sliceReader io.Reader) io.Reader
}

type UnprocessedMap struct {
	slice    *UnprocessedSlice
	combiner unprocessedMapCombiner
}

func NewUnprocessedMap(combiner unprocessedMapCombiner) *UnprocessedMap {
	return &UnprocessedMap{
		slice:    NewUnprocessedSlice(),
		combiner: combiner,
	}
}

func NewGitHubUnprocessedMap() *UnprocessedMap {
	return NewUnprocessedMap(searchresult.NewMerger())
}

func (m *UnprocessedMap) ReadNext(reader io.ReadCloser) error {
	defer reader.Close()
	nextSlice, err := m.combiner.Digest(reader)
	if err != nil {
		return err
	}

	sliceReader := bytes.NewReader(nextSlice)
	if err := m.slice.ReadNext(io.NopCloser(sliceReader)); err != nil {
		return err
	}

	return nil
}

func (m *UnprocessedMap) Merged() io.Reader {
	mergedSlice := m.slice.Merged()
	return m.combiner.Finalize(mergedSlice)
}
