package json_merger

import (
	"fmt"
	"io"
)

type JsonMerger interface {
	ReadNext(io.ReadCloser) error
	Merged() io.Reader
}

// merger is a JsonMerger that auto-detects the type of the json data and delegates to the appropriate merger.
// it covers both the slice and map cases.
type merger struct {
	mergerType   JsonType
	actualMerger JsonMerger
}

func NewMerger() JsonMerger {
	return &merger{
		mergerType:   JsonTypeUnknown,
		actualMerger: nil,
	}
}

func (m *merger) ReadNext(reader io.ReadCloser) error {
	newReader, err := m.initMerger(reader)
	if err != nil {
		return fmt.Errorf("failed to init merger: %v", err)
	}

	if err := m.actualMerger.ReadNext(newReader); err != nil {
		return err
	}

	return nil
}

func (m *merger) Merged() io.Reader {
	return m.actualMerger.Merged()
}

func (m *merger) initMerger(reader io.ReadCloser) (io.ReadCloser, error) {
	detected, newReader, err := DetectJsonType(reader)
	if err != nil {
		return newReader, err
	}

	// if this is the first time we are detecting the type, set it.
	if m.mergerType == JsonTypeUnknown {
		m.mergerType = detected
		switch detected {
		case JsonTypeArray:
			m.actualMerger = NewUnprocessedSlice()
		case JsonTypeDictionary:
			m.actualMerger = NewGitHubUnprocessedMap()
		default:
			return newReader, fmt.Errorf("unexpected json type %v", detected)
		}
		return newReader, nil
	}

	// for all consecutive calls, make sure the type is the same.
	if m.mergerType != detected {
		return newReader, fmt.Errorf("expected %v, got %v", m.mergerType, detected)
	}

	return newReader, nil
}
