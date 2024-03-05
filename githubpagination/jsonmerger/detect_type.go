package jsonmerger

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONType represents the type of the json data.
// Only arrays and dictionaries are supported.
type JSONType int

const (
	JSONTypeUnknown JSONType = iota
	JSONTypeArray
	JSONTypeDictionary
)

// DetectJSONType detects the type of the json data in the decoder.
// The decoder is expected to be at the beginning of the json data.
// only arrays and dictionaries are supported.
func DetectJSONType(inputStream io.ReadCloser) (JSONType, io.ReadCloser, error) {
	lazy := NewLazyTeeReader(inputStream)
	reader := lazy.GetNextReader()
	jsonType, err := DetectJSONTypeUnsafe(reader)
	receoveredReader := lazy.GetFinalReader()
	return jsonType, receoveredReader, err
}

// DetectJSONTypeUnsafe is like DetectJSONType, but it pops off the bytes from the input stream.
func DetectJSONTypeUnsafe(inputStream io.Reader) (JSONType, error) {
	decoder := json.NewDecoder(inputStream)
	token, err := decoder.Token()
	if err != nil {
		return JSONTypeUnknown, err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return JSONTypeUnknown, fmt.Errorf("expected json.Delim, got %T (%v)", token, token)
	}

	if delim == '[' {
		return JSONTypeArray, nil
	} else if delim == '{' {
		return JSONTypeDictionary, nil
	} else {
		return JSONTypeUnknown, fmt.Errorf("unexpected json.Delim %v", delim)
	}
}
