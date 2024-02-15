package json_merger

import (
	"bytes"
	"encoding/json"
	"io"
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
	return NewUnprocessedMap(&githubMapCombiner{})
}

func (m *UnprocessedMap) ReadNext(reader io.ReadCloser) error {
	nextSlice, err := m.combiner.Digest(reader)
	if err != nil {
		return err
	}
	if err := reader.Close(); err != nil {
		return err // TODO should close on failure too (defer and set error) ?
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
