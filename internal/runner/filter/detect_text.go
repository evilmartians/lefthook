package filter

import (
	"bytes"
)

// See: https://github.com/gabriel-vasile/mimetype/blob/6e3aeb1/internal/charset/charset.go

var boms = [][]byte{
	{0xEF, 0xBB, 0xBF},       // utf-8
	{0x00, 0x00, 0xFE, 0xFF}, // utf-32be
	{0xFF, 0xFE, 0x00, 0x00}, // utf-32le
	{0xFE, 0xFF},             // utf-16be
	{0xFF, 0xFE},             // utf-16le
}

// hasBOM returns true if the charset declared in the BOM of content.
func hasBOM(content []byte) bool {
	for _, bom := range boms {
		if bytes.HasPrefix(content, bom) {
			return true
		}
	}
	return false
}

// detectText checks if a sequence contains of a plain text bytes.
//
// This function does not parse BOM-less UTF16 and UTF32 files. Not really
// sure it should. Linux file utility also requires a BOM for UTF16 and UTF32.
func detectText(bytes []byte) bool {
	if hasBOM(bytes) {
		return true
	}

	// Binary data bytes as defined here: https://mimesniff.spec.whatwg.org/#binary-data-byte
	for _, b := range bytes {
		if b <= 0x08 ||
			b == 0x0B ||
			0x0E <= b && b <= 0x1A ||
			0x1C <= b && b <= 0x1F {
			return false
		}
	}

	return true
}
