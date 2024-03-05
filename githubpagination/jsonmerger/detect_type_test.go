package jsonmerger_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/gofri/go-github-pagination/githubpagination/jsonmerger"
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
		ExpectedType jsonmerger.JSONType
		FromType     bool
	}{
		{
			Data:         []string{"a", "b"},
			ExpectedType: jsonmerger.JSONTypeArray,
			FromType:     true,
		},
		{
			Data:         map[string]string{"a": "b"},
			ExpectedType: jsonmerger.JSONTypeDictionary,
			FromType:     true,
		},
		{
			Data:         "a",
			ExpectedType: jsonmerger.JSONTypeUnknown,
			FromType:     true,
		},
		{
			Data:         []byte("invalid json"),
			ExpectedType: jsonmerger.JSONTypeUnknown,
			FromType:     false,
		},
	}

	for i, testCase := range testCases {
		reader, err := toReader(testCase.Data, testCase.FromType)
		if err != nil {
			t.Fatalf("%d) unexpected error: %v", i, err)
		}
		actualType, err := jsonmerger.DetectJSONTypeUnsafe(reader)
		if testCase.ExpectedType == jsonmerger.JSONTypeUnknown && err == nil {
			t.Fatalf("%d) expected error, got %v", i, actualType)
		} else if testCase.ExpectedType != jsonmerger.JSONTypeUnknown && err != nil {
			t.Fatalf("%d) unexpected error: %v", i, err)
		}
	}
}
