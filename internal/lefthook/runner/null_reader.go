package runner

import "io"

// nullReader always returns `io.EOF`.
type nullReader struct{}

func NewNullReader() io.Reader {
	return nullReader{}
}

// Implements io.Reader interface.
func (nullReader) Read(b []byte) (int, error) {
	return 0, io.EOF
}
