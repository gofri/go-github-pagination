package json_merger_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/gofri/go-github-pagination/github_pagination/json_merger"
)

func typeToReader(data any) (io.Reader, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytesToReader(bytes), nil
}

func bytesToReader(data []byte) io.Reader {
	return bytes.NewReader(data)
}

func toReader(data any, fromType bool) (io.Reader, error) {
	if fromType {
		return typeToReader(data)
	} else {
		return bytesToReader(data.([]byte)), nil
	}
}

func TestTypeDetection(t *testing.T) {
	testCases := []struct {
		Data         any
		ExpectedType json_merger.JsonType
		FromType     bool
	}{
		{
			Data:         []string{"a", "b"},
			ExpectedType: json_merger.JsonTypeArray,
			FromType:     true,
		},
		{
			Data:         map[string]string{"a": "b"},
			ExpectedType: json_merger.JsonTypeDictionary,
			FromType:     true,
		},
		{
			Data:         "a",
			ExpectedType: json_merger.JsonTypeUnknown,
			FromType:     true,
		},
		{
			Data:         []byte("invalid json"),
			ExpectedType: json_merger.JsonTypeUnknown,
			FromType:     false,
		},
	}

	for i, testCase := range testCases {
		reader, err := toReader(testCase.Data, testCase.FromType)
		if err != nil {
			t.Fatalf("%d) unexpected error: %v", i, err)
		}
		actualType, err := json_merger.DetectJsonTypeUnsafe(reader)
		if testCase.ExpectedType == json_merger.JsonTypeUnknown && err == nil {
			t.Fatalf("%d) expected error, got %v", i, actualType)
		} else if testCase.ExpectedType != json_merger.JsonTypeUnknown && err != nil {
			t.Fatalf("%d) unexpected error: %v", i, err)
		}
	}
}
