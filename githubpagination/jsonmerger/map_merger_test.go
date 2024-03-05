package jsonmerger_test

import (
	"bytes"
	"encoding/json"
	"io"
	"slices"
	"testing"

	"github.com/gofri/go-github-pagination/githubpagination/jsonmerger"
)

type mappedDataType struct {
	TotalCount       int   `json:"total_count"`
	IncompleteResult bool  `json:"incomplete_results"`
	Items            []int `json:"items"`
}

func TestMultipleMaps(t *testing.T) {
	_TestMultipleMaps(t, jsonmerger.NewGitHubUnprocessedMap())
}

func _TestMultipleMaps(t *testing.T, merger jsonmerger.JSONMerger) {
	t.Parallel()

	t.Run("test github map merger", func(t *testing.T) {
		inputMaps := []mappedDataType{
			{
				TotalCount:       1,
				IncompleteResult: false,
				Items:            []int{10, 20},
			},
			{
				TotalCount:       2,
				IncompleteResult: true,
				Items:            []int{30, 40},
			},
		}
		expectedMerged := mappedDataType{
			TotalCount:       3,
			IncompleteResult: true,
			Items:            []int{10, 20, 30, 40},
		}
		var rawInputs []bytes.Buffer
		for _, input := range inputMaps {
			var rawSlice bytes.Buffer
			if err := json.NewEncoder(&rawSlice).Encode(input); err != nil {
				t.Fatal(err)
			}
			rawInputs = append(rawInputs, rawSlice)
		}

		merger := jsonmerger.NewGitHubUnprocessedMap()
		for _, rawInput := range rawInputs {
			if err := merger.ReadNext(io.NopCloser(&rawInput)); err != nil {
				t.Fatal(err)
			}
		}

		var result mappedDataType
		if err := MergeInto(merger, &result); err != nil {
			t.Fatalf("expected %v, got %v", expectedMerged, result)
		}

		if result.TotalCount != expectedMerged.TotalCount {
			t.Fatalf("expected %v, got %v", expectedMerged.TotalCount, result.TotalCount)
		}
		if result.IncompleteResult != expectedMerged.IncompleteResult {
			t.Fatalf("expected %v, got %v", expectedMerged.IncompleteResult, result.IncompleteResult)
		}
		if slices.Compare(result.Items, expectedMerged.Items) != 0 {
			t.Fatalf("expected %v, got %v", expectedMerged.Items, result.Items)
		}
	})
}
