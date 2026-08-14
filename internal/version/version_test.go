package version

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	for i, tt := range [...]struct {
		wanted, given string
		err           error
	}{
		{
			wanted: "1.0.0",
			given:  "1.0.1",
			err:    nil,
		},
		{
			wanted: "v1.0.0",
			given:  "1.0.1",
			err:    nil,
		},
		{
			wanted: "1.0.0",
			given:  "v1.0.1",
			err:    nil,
		},
		{
			wanted: "1",
			given:  "1.2",
			err:    nil,
		},
		{
			wanted: "3.0.0",
			given:  "1.1",
			err:    ErrUncoveredVersion,
		},
		{
			wanted: "13",
			given:  "10",
			err:    ErrUncoveredVersion,
		},
		{
			wanted: "10--.0-best",
			given:  "10",
			err:    ErrInvalidVersion,
		},
		{
			wanted: "10",
			given:  "vv10.0.0-best",
			err:    ErrInvalidVersion,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.err, Check(tt.wanted, tt.given))
		})
	}
}

func TestVersion(t *testing.T) {
	assert.Equal(t, version, Version(false))
	assert.Equal(t, version+" "+commit, Version(true))
}
