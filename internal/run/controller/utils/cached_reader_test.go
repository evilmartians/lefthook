package utils

import (
	"bytes"
	"io"
	"sync"
	"testing"
)

func TestCachedReader(t *testing.T) {
	testSlice := []byte("Some example string\nMultiline")

	cachedReader := NewCachedReader(bytes.NewReader(testSlice))

	for range 5 {
		res, err := io.ReadAll(cachedReader.Reader())
		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}

		if !bytes.Equal(res, testSlice) {
			t.Errorf("expected %v to be equal to %v", res, testSlice)
		}
	}
}

func TestCachedReaderConcurrent(t *testing.T) {
	testSlice := []byte("refs/heads/main 0000 refs/heads/main 1111\n")

	cachedReader := NewCachedReader(bytes.NewReader(testSlice))

	const consumers = 16
	var wg sync.WaitGroup
	results := make([][]byte, consumers)
	for i := range consumers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			res, err := io.ReadAll(cachedReader.Reader())
			if err != nil {
				t.Errorf("unexpected err: %s", err)
			}
			results[i] = res
		}(i)
	}
	wg.Wait()

	for i, res := range results {
		if !bytes.Equal(res, testSlice) {
			t.Errorf("consumer %d: expected %v to be equal to %v", i, res, testSlice)
		}
	}
}
