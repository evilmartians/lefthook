package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandsToJobs(t *testing.T) {
	commands := map[string]*Command{
		"check": {
			Run:      "echo",
			Priority: 150,
		},
		"10lint": {
			Run:        "echo",
			StageFixed: true,
		},
		"first": {
			Run:      "echo",
			Priority: 1,
		},
		"2lint": {
			Run:        "echo",
			StageFixed: true,
		},
		"last": {
			Run: "echo",
		},
	}

	jobs := CommandsToJobs(commands)

	assert.Equal(t, jobs, []*Job{
		{Name: "first", Run: "echo"},
		{Name: "check", Run: "echo"},
		{Name: "2lint", Run: "echo", StageFixed: true},
		{Name: "10lint", Run: "echo", StageFixed: true},
		{Name: "last", Run: "echo"},
	})
}
