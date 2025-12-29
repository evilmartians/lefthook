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

func TestCommandsToJobsWithTimeout(t *testing.T) {
	commands := map[string]*Command{
		"test": {
			Run:     "npm test",
			Timeout: "60s",
		},
		"build": {
			Run:     "npm build",
			Timeout: "5m",
		},
		"no-timeout": {
			Run: "echo hello",
		},
	}

	jobs := CommandsToJobs(commands)

	// Find jobs by name since ordering may vary
	var testJob, buildJob, noTimeoutJob *Job
	for _, job := range jobs {
		switch job.Name {
		case "test":
			testJob = job
		case "build":
			buildJob = job
		case "no-timeout":
			noTimeoutJob = job
		}
	}

	assert.NotNil(t, testJob)
	assert.Equal(t, "npm test", testJob.Run)
	assert.Equal(t, "60s", testJob.Timeout)

	assert.NotNil(t, buildJob)
	assert.Equal(t, "npm build", buildJob.Run)
	assert.Equal(t, "5m", buildJob.Timeout)

	assert.NotNil(t, noTimeoutJob)
	assert.Equal(t, "echo hello", noTimeoutJob.Run)
	assert.Equal(t, "", noTimeoutJob.Timeout)
}
