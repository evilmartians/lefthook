package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScriptsToJobs(t *testing.T) {
	scripts := map[string]*Script{
		"check.sh": {
			Runner:   "bash",
			Priority: 150,
		},
		"10_test.sh": {
			Runner:     "bash",
			StageFixed: true,
		},
		"2_test.sh": {
			Runner:     "bash",
			StageFixed: true,
		},
		"first.sh": {
			Runner:   "bash",
			Priority: 1,
		},
		"last.sh": {
			Runner: "bash",
		},
	}

	jobs := ScriptsToJobs(scripts)

	assert.Equal(t, jobs, []*Job{
		{Name: "first.sh", Script: "first.sh", Runner: "bash"},
		{Name: "check.sh", Script: "check.sh", Runner: "bash"},
		{Name: "2_test.sh", Script: "2_test.sh", Runner: "bash", StageFixed: true},
		{Name: "10_test.sh", Script: "10_test.sh", Runner: "bash", StageFixed: true},
		{Name: "last.sh", Script: "last.sh", Runner: "bash"},
	})
}

func TestScriptsToJobsWithTimeout(t *testing.T) {
	scripts := map[string]*Script{
		"test.sh": {
			Runner:  "bash",
			Timeout: "30s",
		},
		"build.sh": {
			Runner:  "bash",
			Timeout: "10m",
		},
		"no-timeout.sh": {
			Runner: "bash",
		},
	}

	jobs := ScriptsToJobs(scripts)

	// Find jobs by name since ordering may vary
	var testJob, buildJob, noTimeoutJob *Job
	for _, job := range jobs {
		switch job.Name {
		case "test.sh":
			testJob = job
		case "build.sh":
			buildJob = job
		case "no-timeout.sh":
			noTimeoutJob = job
		}
	}

	assert.NotNil(t, testJob)
	assert.Equal(t, "test.sh", testJob.Script)
	assert.Equal(t, "30s", testJob.Timeout)

	assert.NotNil(t, buildJob)
	assert.Equal(t, "build.sh", buildJob.Script)
	assert.Equal(t, "10m", buildJob.Timeout)

	assert.NotNil(t, noTimeoutJob)
	assert.Equal(t, "no-timeout.sh", noTimeoutJob.Script)
	assert.Equal(t, "", noTimeoutJob.Timeout)
}
