package json_merger_test

import (
	"io"
	"testing"

	"github.com/gofri/go-github-pagination/github_pagination/json_merger"
)

type oneTimeReader struct {
	internalIndex int
	Closed        bool
}

func (r *oneTimeReader) Read(p []byte) (n int, err error) {
	src := []byte("hello")
	if r.internalIndex >= len(src) {
		return 0, io.EOF
	}
	space := cap(p)
	remaining := len(src) - r.internalIndex
	actual := min(space, remaining)
	copy(p, src[r.internalIndex:r.internalIndex+actual])
	r.internalIndex += actual
	return actual, nil
}

func (r *oneTimeReader) Close() error {
	r.Closed = true
	return nil
}

func TestCloner(t *testing.T) {
	src := &oneTimeReader{}

	lazy := json_merger.NewLazyTeeReader(src)
	r := lazy.GetNextReader()
	smallChunk := make([]byte, 2)
	count, err := r.Read(smallChunk)
	if err != nil {
		t.Fatalf("error reading: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected to read 2 bytes, got %d", count)
	}
	if string(smallChunk) != "he" {
		t.Fatalf("expected to read 'he', got %s", string(smallChunk))
	}

	nextReader := lazy.GetFinalReader()
	count, err = nextReader.Read(smallChunk)
	if err != nil {
		t.Fatalf("error reading: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected to read 2 bytes, got %d", count)
	}
	if string(smallChunk) != "he" {
		t.Fatalf("expected to read 'he' again, got %s", string(smallChunk))
	}
	rest := make([]byte, 10)
	count, err = nextReader.Read(rest)
	if err != nil {
		t.Fatalf("error reading: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected to read 3 bytes, got %d", count)
	}
	if string(rest[:count]) != "llo" {
		t.Fatalf("expected to read 'llo', got %s", string(rest[:count]))
	}
	count, err = nextReader.Read(rest)
	if count != 0 || err != io.EOF {
		t.Fatalf("expected to read 0 bytes and EOF, got %d and %v", count, err)
	}

	if err := nextReader.Close(); err != nil {
		t.Fatalf("error closing: %v", err)
	}
	if !src.Closed {
		t.Fatalf("expected to close the source reader")
	}
}
