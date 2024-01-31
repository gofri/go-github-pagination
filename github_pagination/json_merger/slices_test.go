package json_merger_test

import (
	"bytes"
	"encoding/json"
	"io"
	"slices"
	"testing"

	"github.com/gofri/go-github-pagination/github_pagination/json_merger"
)

func MergeInto(merger json_merger.JsonMerger, v any) error {
	return json.NewDecoder(merger.Merged()).Decode(v)
}

func _TestMultipleSlices(t *testing.T, merger json_merger.JsonMerger) {
	var result []int
	inputSlices := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	expectedMerged := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	var rawInputs []bytes.Buffer
	for _, input := range inputSlices {
		var rawSlice bytes.Buffer
		if err := json.NewEncoder(&rawSlice).Encode(input); err != nil {
			t.Fatal(err)
		}
		rawInputs = append(rawInputs, rawSlice)
	}

	for _, rawInput := range rawInputs {
		if err := merger.ReadNext(io.NopCloser(&rawInput)); err != nil {
			t.Fatal(err)
		}
	}

	if err := MergeInto(merger, &result); err != nil {
		t.Fatalf("expected %v, got %v", expectedMerged, result)
	}
	if slices.Compare(expectedMerged, result) != 0 {
		t.Fatalf("expected %v, got %v", expectedMerged, result)
	}
}

func TestMultipleSlices(t *testing.T) {
	_TestMultipleSlices(t, json_merger.NewUnprocessedSlice())
}
