package utils

import (
	"bytes"
	"io"
	"sync"
)

// CachedReader reads from the provided `io.Reader` to `io.EOF` once and caches
// the content. Each call to `Reader` returns an independent reader over the
// cached bytes, so concurrent consumers each receive the full input.
type CachedReader struct {
	mu     sync.Mutex
	in     io.Reader
	cached bool
	buf    []byte
}

func NewCachedReader(in io.Reader) *CachedReader {
	return &CachedReader{in: in}
}

func (r *CachedReader) Reader() io.Reader {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.cached {
		r.buf, _ = io.ReadAll(r.in)
		r.cached = true
	}

	return bytes.NewReader(r.buf)
}
