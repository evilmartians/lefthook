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
		"lint.sh": {
			Runner:   "bash",
			Timeout:  "30s",
			Priority: 1,
		},
		"test.sh": {
			Runner:  "bash",
			Timeout: "10m",
		},
	}

	jobs := ScriptsToJobs(scripts)

	assert.Equal(t, jobs, []*Job{
		{Name: "lint.sh", Script: "lint.sh", Runner: "bash", Timeout: "30s"},
		{Name: "test.sh", Script: "test.sh", Runner: "bash", Timeout: "10m"},
	})
}
