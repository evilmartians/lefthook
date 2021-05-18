package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterGlob(t *testing.T) {
	files := []string{"path/to/file.jpg", "path/to/file.png", "path/to/file.go", "path/to/another-file.JPG"}
	pattern := "*.{jpg,PNG}"
	result := FilterGlob(files, pattern)

	expected := []string{"path/to/file.jpg", "path/to/file.png", "path/to/another-file.JPG"}
	assert.ElementsMatch(t, expected, result)
}
