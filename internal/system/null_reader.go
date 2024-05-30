package system

import "io"

// nullReader always returns `io.EOF`.
type nullReader struct{}

var NullReader = nullReader{}

// Implements io.Reader interface.
func (nullReader) Read(b []byte) (int, error) {
	return 0, io.EOF
}
