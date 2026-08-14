package system

import (
	"bytes"
	"io"
	"testing"
)

func TestNullReader(t *testing.T) {
	res, err := io.ReadAll(NullReader)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
	}

	if !bytes.Equal(res, []byte{}) {
		t.Errorf("expected %v to be equal to %v", res, []byte{})
	}
}
