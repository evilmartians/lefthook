package exec

import "io"

// nullReader always returns EOF.
type nullReader struct{}

func (nullReader) Read(b []byte) (int, error) {
	return 0, io.EOF
}
