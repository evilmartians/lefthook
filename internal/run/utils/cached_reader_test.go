package utils

import (
	"bytes"
	"io"
	"testing"
)

func TestCachedReader(t *testing.T) {
	testSlice := []byte("Some example string\nMultiline")

	cachedReader := NewCachedReader(bytes.NewReader(testSlice))

	for range 5 {
		res, err := io.ReadAll(cachedReader)
		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}

		if !bytes.Equal(res, testSlice) {
			t.Errorf("expected %v to be equal to %v", res, testSlice)
		}
	}
}
