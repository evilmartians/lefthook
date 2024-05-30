package runner

import (
	"bytes"
	"io"
	"testing"
)

func TestNullReader(t *testing.T) {
	nullReader := NewNullReader()

	res, err := io.ReadAll(nullReader)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
	}

	if !bytes.Equal(res, []byte{}) {
		t.Errorf("expected %v to be equal to %v", res, []byte{})
	}
}
