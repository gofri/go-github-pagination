package jsonmerger_test

import (
	"testing"

	"github.com/gofri/go-github-pagination/githubpagination/jsonmerger"
)

func TestMerger(t *testing.T) {
	_TestMultipleSlices(t, jsonmerger.NewMerger())
	_TestMultipleMaps(t, jsonmerger.NewMerger())
}
