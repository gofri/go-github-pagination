package json_merger_test

import (
	"testing"

	"github.com/gofri/go-github-pagination/github_pagination/json_merger"
)

func TestMerger(t *testing.T) {
	_TestMultipleSlices(t, json_merger.NewMerger())
	_TestMultipleMaps(t, json_merger.NewMerger())
}
