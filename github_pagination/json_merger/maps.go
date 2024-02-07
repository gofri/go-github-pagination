package json_merger

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

type unprocessedMapCombiner interface {
	Digest(io.Reader) (slice json.RawMessage, err error)
	Finalize(sliceReader io.Reader) io.Reader
}

type NotPaginatableDictError struct{}

func (e *NotPaginatableDictError) Error() string {
	return "not a paginatable dictionary"
}

type UnprocessedMap struct {
	slice               *UnprocessedSlice
	combiner            unprocessedMapCombiner
	unpaginatedResponse *bytes.Buffer
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
	var unpaginatedResponse bytes.Buffer
	tee := io.TeeReader(reader, &unpaginatedResponse)
	nextSlice, err := m.combiner.Digest(tee)
	if errors.Is(err, &NotPaginatableDictError{}) {
		// TODO add test for not paginatable endpoints
		m.unpaginatedResponse = &unpaginatedResponse
		return nil
	}
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
	if m.unpaginatedResponse != nil {
		return m.unpaginatedResponse
	}
	mergedSlice := m.slice.Merged()
	return m.combiner.Finalize(mergedSlice)
}
