package cmdtest

import (
	"io"
)

// NewTracking returns executor that collects the called commands.
func NewTracking(cb func(string, string, io.Writer) error) *TrackingCmd {
	return &TrackingCmd{
		Commands: make([]string, 0),
		callback: cb,
	}
}

// NewDumb returns executor that does simply nothing.
func NewDumb() *DumbCmd {
	return &DumbCmd{}
}
