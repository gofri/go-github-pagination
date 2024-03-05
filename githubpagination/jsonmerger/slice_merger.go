package jsonmerger

import (
	"encoding/json"
	"io"
)

// UnprocessedSlice represents a slice of json.RawMessage.
// It can be used to merge consecutive slices of json data, i.e.,
// without double-parsing the slice elements, but rather,
// only parsing the top-level array.
// note that we are not aware of the type/pointer to which it will be unmarshalled,
// so we cannot simply rely on unmarshall to gradually grow it.
//
// implementation note:
// the approach taken here is to json-parse the data into a slice of json.RawMessage on the way in,
// and then to construct a new json stream (on the fly) on the way out.
// a different approach would be to simply use a json.RawMessage on the way in,
// decoding only the head and tail tokens (using a reverse reader for the tail).
// alternative's pros:
// - more efficient (json parsing is somewhat costly).
// alternative's cons:
// - more complicated (a plumber's approach).
// - does not work for the map case, so it also adds implementation complexity.
//
// all in all, the current approach is simpler and more robust.
// unless someone has a good reason to change it (namely, performance pain),
// we should stick with the existing implementation.

type UnprocessedSlice struct {
	subSlices []json.RawMessage
}

func NewUnprocessedSlice() *UnprocessedSlice {
	return &UnprocessedSlice{
		subSlices: nil,
	}
}

func (slice *UnprocessedSlice) ReadNext(reader io.ReadCloser) error {
	defer reader.Close()
	var toAppend []json.RawMessage
	if err := json.NewDecoder(reader).Decode(&toAppend); err != nil {
		return err
	}
	slice.subSlices = append(slice.subSlices, toAppend...)

	return nil
}

func (slice *UnprocessedSlice) Merged() io.Reader {
	return newSlicesReader(slice)
}

type slicesReader struct {
	slice           *UnprocessedSlice
	index           int
	positionInSlice int
}

func newSlicesReader(slice *UnprocessedSlice) *slicesReader {
	return &slicesReader{
		slice:           slice,
		index:           -1,
		positionInSlice: 0,
	}
}

func (r *slicesReader) Read(p []byte) (n int, err error) {
	data := r.getNextDataToRead()
	if data == nil {
		return 0, io.EOF
	}

	n = copy(p, data)
	if n < len(data) {
		r.positionInSlice += n
	} else {
		r.index++
		r.positionInSlice = 0
	}

	return n, nil
}

func (r *slicesReader) getNextDataToRead() []byte {
	curIndex := r.index
	numOfSlices := len(r.slice.subSlices)
	if numOfSlices == 0 {
		return nil
	}
	switch {
	case curIndex == -1: // first read - open the array
		return []byte{'['}
	case curIndex == numOfSlices: // last read - close the array
		return []byte{']'}
	case curIndex > numOfSlices: // after last read - EOF
		return nil
	default: // read the next slice
		sliceBytes := r.slice.subSlices[curIndex][r.positionInSlice:]
		if curIndex < numOfSlices-1 { // add a comma between slices
			sliceBytes = append(sliceBytes, ',')
		}
		return sliceBytes
	}
}
