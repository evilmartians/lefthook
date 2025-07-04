package filters

import (
	"fmt"
	"testing"
)

func TestDetectText(t *testing.T) {
	for i, tt := range [...]struct {
		bytes  []byte
		result bool
	}{
		{
			bytes:  []byte{},
			result: true,
		},
		{
			bytes:  []byte{0xEF, 0xBB, 0xBF}, // utf-8 BOM
			result: true,
		},
		{
			bytes:  []byte{0x00, 0x00, 0xFE, 0xFF}, // utf-32be BOM
			result: true,
		},
		{
			bytes:  []byte{0xFF, 0xFE, 0x00, 0x00}, // utf-32le BOM
			result: true,
		},
		{
			bytes:  []byte{0xFE, 0xFF}, // utf-16be BOM
			result: true,
		},
		{
			bytes:  []byte{0xFF, 0xFE}, // utf-16le BOM
			result: true,
		},
		{
			bytes:  []byte{0xFA, 0xCF, 0xFE, 0xED, 0x00, 0x0C},
			result: false,
		},
		{
			bytes:  []byte{0x70, 0x5B, 0x65, 0x72, 0x63, 0x2D}, // .lefthook.toml
			result: true,
		},
		{
			bytes:  []byte{0x5B, 0x21, 0x75, 0x42, 0x6C, 0x69, 0x20, 0x64, 0x74, 0x53, 0x74, 0x61, 0x73, 0x75, 0x28, 0x5D}, // README.md
			result: true,
		},
	} {
		t.Run(fmt.Sprintf("#%d:", i), func(t *testing.T) {
			if detectText(tt.bytes) != tt.result {
				t.Error("results don't match; expected", tt.result)
			}
		})
	}
}
