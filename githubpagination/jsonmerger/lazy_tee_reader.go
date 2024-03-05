package jsonmerger

import (
	"bytes"
	"io"
)

// LazyTeeReader is a reader that reads from the input stream and writes to the output stream.
// it provides the first and second readers to use.
// a more general version would probably just be supporting Reset() without limitation.
type LazyTeeReader struct {
	inputStream io.ReadCloser
	buffer      bytes.Buffer
}

func NewLazyTeeReader(inputStream io.ReadCloser) *LazyTeeReader {
	return &LazyTeeReader{
		inputStream: inputStream,
		buffer:      bytes.Buffer{},
	}
}

// GetNextReader returns a reader that reads from the input stream,
// in such a way that the bytes are not lost.
func (r *LazyTeeReader) GetNextReader() io.Reader {
	teeReader := io.TeeReader(r.inputStream, &r.buffer)
	return io.MultiReader(&r.buffer, teeReader)
}

// GetFinalReader returns a reader that reads from the input stream,
// but loses the bytes that read from the original input stream.
// this is useful when the input stream is not needed anymore.
func (r *LazyTeeReader) GetFinalReader() io.ReadCloser {
	return &FinalReader{
		inputStream:    r.inputStream,
		combinedReader: io.MultiReader(&r.buffer, r.inputStream),
	}
}

type FinalReader struct {
	inputStream    io.ReadCloser
	combinedReader io.Reader
}

func (r *FinalReader) Read(p []byte) (n int, err error) {
	return r.combinedReader.Read(p)
}

func (r *FinalReader) Close() error {
	return r.inputStream.Close()
}
