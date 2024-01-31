package json_merger

import (
	"encoding/json"
	"fmt"
	"io"
)

// JsonType represents the type of the json data.
// Only arrays and dictionaries are supported.
type JsonType int

const (
	JsonTypeUnknown JsonType = iota
	JsonTypeArray
	JsonTypeDictionary
)

// DetectJsonType detects the type of the json data in the decoder.
// The decoder is expected to be at the beginning of the json data.
// only arrays and dictionaries are supported.
func DetectJsonType(inputStream io.ReadCloser) (JsonType, io.ReadCloser, error) {
	lazy := NewLazyTeeReader(inputStream)
	reader := lazy.GetNextReader()
	jsonType, err := DetectJsonTypeUnsafe(reader)
	receoveredReader := lazy.GetFinalReader()
	return jsonType, receoveredReader, err
}

// DetectJsonTypeUnsafe is like DetectJsonType, but it pops off the bytes from the input stream.
func DetectJsonTypeUnsafe(inputStream io.Reader) (JsonType, error) {
	decoder := json.NewDecoder(inputStream)
	token, err := decoder.Token()
	if err != nil {
		return JsonTypeUnknown, err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return JsonTypeUnknown, fmt.Errorf("expected json.Delim, got %T (%v)", token, token)
	}

	if delim == '[' {
		return JsonTypeArray, nil
	} else if delim == '{' {
		return JsonTypeDictionary, nil
	} else {
		return JsonTypeUnknown, fmt.Errorf("unexpected json.Delim %v", delim)
	}
}
