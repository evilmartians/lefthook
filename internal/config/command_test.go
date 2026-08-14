package config

import (
	"testing"
	"time"

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

func TestCommandsToJobsWithTimeout(t *testing.T) {
	commands := map[string]*Command{
		"lint": {
			Run:      "echo lint",
			Timeout:  60 * time.Second,
			Priority: 1,
		},
		"test": {
			Run:     "echo test",
			Timeout: 5 * time.Minute,
		},
	}

	jobs := CommandsToJobs(commands)

	assert.Equal(t, jobs, []*Job{
		{Name: "lint", Run: "echo lint", Timeout: 60 * time.Second},
		{Name: "test", Run: "echo test", Timeout: 5 * time.Minute},
	})
}
