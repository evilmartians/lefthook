package runner

import (
	"bytes"
	"io"
)

// cachedReader reads from the provided `io.Reader` until `io.EOF` and saves
// the read content into the inner buffer.
//
// After `io.EOF` it will provide the read data again and again.
type cachedReader struct {
	in        io.Reader
	useBuffer bool
	buf       []byte
	reader    *bytes.Reader
}

func NewCachedReader(in io.Reader) *cachedReader {
	return &cachedReader{
		in:     in,
		buf:    []byte{},
		reader: bytes.NewReader([]byte{}),
	}
}

func (r *cachedReader) Read(p []byte) (int, error) {
	if r.useBuffer {
		n, err := r.reader.Read(p)
		if err == io.EOF {
			_, seekErr := r.reader.Seek(0, io.SeekStart)
			if seekErr != nil {
				panic(seekErr)
			}

			return n, err
		}

		return n, err
	}

	n, err := r.in.Read(p)
	r.buf = append(r.buf, p[:n]...)
	if err == io.EOF {
		r.useBuffer = true
		r.reader = bytes.NewReader(r.buf)
	}
	return n, err
}
