package result

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGroup(t *testing.T) {
	for i, tt := range [...]struct {
		name     string
		results  []Result
		expected Result
	}{
		{
			name:    "empty results",
			results: []Result{},
			expected: Result{
				Name:     "test-group",
				status:   skip,
				Sub:      []Result{},
				Duration: 0,
			},
		},
		{
			name: "all success results",
			results: []Result{
				Success("cmd1", 100*time.Millisecond),
				Success("cmd2", 200*time.Millisecond),
				Success("cmd3", 150*time.Millisecond),
			},
			expected: Result{
				Name:   "test-group",
				status: success,
				Sub: []Result{
					Success("cmd1", 100*time.Millisecond),
					Success("cmd2", 200*time.Millisecond),
					Success("cmd3", 150*time.Millisecond),
				},
				Duration: 450 * time.Millisecond,
			},
		},
		{
			name: "all skip results",
			results: []Result{
				Skip("cmd1"),
				Skip("cmd2"),
				Skip("cmd3"),
			},
			expected: Result{
				Name:   "test-group",
				status: skip,
				Sub: []Result{
					Skip("cmd1"),
					Skip("cmd2"),
					Skip("cmd3"),
				},
				Duration: 0,
			},
		},
		{
			name: "all failure results",
			results: []Result{
				Failure("cmd1", "error 1", 50*time.Millisecond),
				Failure("cmd2", "error 2", 75*time.Millisecond),
			},
			expected: Result{
				Name:   "test-group",
				status: failure,
				Sub: []Result{
					Failure("cmd1", "error 1", 50*time.Millisecond),
					Failure("cmd2", "error 2", 75*time.Millisecond),
				},
				Duration: 125 * time.Millisecond,
			},
		},
		{
			name: "mixed success and skip",
			results: []Result{
				Success("cmd1", 100*time.Millisecond),
				Skip("cmd2"),
				Success("cmd3", 200*time.Millisecond),
			},
			expected: Result{
				Name:   "test-group",
				status: success,
				Sub: []Result{
					Success("cmd1", 100*time.Millisecond),
					Skip("cmd2"),
					Success("cmd3", 200*time.Millisecond),
				},
				Duration: 300 * time.Millisecond,
			},
		},
		{
			name: "mixed success and failure",
			results: []Result{
				Success("cmd1", 100*time.Millisecond),
				Failure("cmd2", "failed", 50*time.Millisecond),
				Success("cmd3", 75*time.Millisecond),
			},
			expected: Result{
				Name:   "test-group",
				status: failure,
				Sub: []Result{
					Success("cmd1", 100*time.Millisecond),
					Failure("cmd2", "failed", 50*time.Millisecond),
					Success("cmd3", 75*time.Millisecond),
				},
				Duration: 225 * time.Millisecond,
			},
		},
		{
			name: "mixed skip and failure",
			results: []Result{
				Skip("cmd1"),
				Failure("cmd2", "failed", 100*time.Millisecond),
				Skip("cmd3"),
			},
			expected: Result{
				Name:   "test-group",
				status: failure,
				Sub: []Result{
					Skip("cmd1"),
					Failure("cmd2", "failed", 100*time.Millisecond),
					Skip("cmd3"),
				},
				Duration: 100 * time.Millisecond,
			},
		},
		{
			name: "all three statuses mixed",
			results: []Result{
				Success("cmd1", 50*time.Millisecond),
				Skip("cmd2"),
				Failure("cmd3", "error", 25*time.Millisecond),
				Success("cmd4", 125*time.Millisecond),
			},
			expected: Result{
				Name:   "test-group",
				status: failure,
				Sub: []Result{
					Success("cmd1", 50*time.Millisecond),
					Skip("cmd2"),
					Failure("cmd3", "error", 25*time.Millisecond),
					Success("cmd4", 125*time.Millisecond),
				},
				Duration: 200 * time.Millisecond,
			},
		},
		{
			name: "single success result",
			results: []Result{
				Success("single-cmd", 300*time.Millisecond),
			},
			expected: Result{
				Name:   "test-group",
				status: success,
				Sub: []Result{
					Success("single-cmd", 300*time.Millisecond),
				},
				Duration: 300 * time.Millisecond,
			},
		},
		{
			name: "single skip result",
			results: []Result{
				Skip("single-cmd"),
			},
			expected: Result{
				Name:   "test-group",
				status: skip,
				Sub: []Result{
					Skip("single-cmd"),
				},
				Duration: 0,
			},
		},
		{
			name: "single failure result",
			results: []Result{
				Failure("single-cmd", "single error", 150*time.Millisecond),
			},
			expected: Result{
				Name:   "test-group",
				status: failure,
				Sub: []Result{
					Failure("single-cmd", "single error", 150*time.Millisecond),
				},
				Duration: 150 * time.Millisecond,
			},
		},
	} {
		t.Run(fmt.Sprintf("test %d: %s", i, tt.name), func(t *testing.T) {
			assert := assert.New(t)
			result := Group("test-group", tt.results)

			assert.Equal(tt.expected.Name, result.Name)
			assert.Equal(tt.expected.status, result.status)
			assert.Equal(tt.expected.Duration, result.Duration)
			assert.EqualValues(tt.expected.Sub, result.Sub)

			// Test the status methods
			switch tt.expected.status {
			case success:
				assert.True(result.Success())
				assert.False(result.Failure())
			case failure:
				assert.False(result.Success())
				assert.True(result.Failure())
			case skip:
				assert.False(result.Success())
				assert.False(result.Failure())
			}
		})
	}
}

func TestResult_Status(t *testing.T) {
	code := 2

	for name, tt := range map[string]struct {
		result Result
		want   Status
	}{
		"success": {
			result: Success("ok", time.Second),
			want:   StatusPassed,
		},
		"skip": {
			result: Skip("skipped"),
			want:   StatusSkipped,
		},
		"failure-from-command": {
			result: Failure("cmd", "boom", time.Second).WithExitCode(code),
			want:   StatusFailed,
		},
		"failure-without-exit-code": {
			result: Failure("cmd", "invalid job", 0),
			want:   StatusError,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if got := tt.result.Status(); got != tt.want {
				t.Errorf("Result.Status() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResult_ExitCode(t *testing.T) {
	if _, ok := Success("ok", time.Second).ExitCode(); ok {
		t.Error("Success().ExitCode() reports a code, want none")
	}

	code, ok := Failure("cmd", "", time.Second).WithExitCode(42).ExitCode()
	if !ok || code != 42 {
		t.Errorf("ExitCode() = (%d, %v), want (42, true)", code, ok)
	}
}

func TestResult_Reason(t *testing.T) {
	result := SkipWithReason("docs", "no matching staged files")

	if result.Reason() != "no matching staged files" {
		t.Errorf("Reason() = %q, want the skip reason", result.Reason())
	}

	if !result.Skip() {
		t.Error("SkipWithReason() is not a skip")
	}
}

func TestAggregate(t *testing.T) {
	for name, tt := range map[string]struct {
		results []Result
		want    Status
	}{
		"empty": {
			results: nil,
			want:    StatusSkipped,
		},
		"all-passed": {
			results: []Result{Success("a", time.Second), Success("b", time.Second)},
			want:    StatusPassed,
		},
		"all-skipped": {
			results: []Result{Skip("a"), SkipWithReason("b", "no files")},
			want:    StatusSkipped,
		},
		"failure-wins-over-pass": {
			results: []Result{Success("a", time.Second), Failure("b", "", time.Second).WithExitCode(1)},
			want:    StatusFailed,
		},
		"error-wins-over-failure": {
			results: []Result{
				Failure("a", "", time.Second).WithExitCode(1),
				Failure("b", "timeout", time.Second),
			},
			want: StatusError,
		},
		"pass-wins-over-skip": {
			results: []Result{Skip("a"), Success("b", time.Second)},
			want:    StatusPassed,
		},
		"group-nested-in-results": {
			results: []Result{Group("grp", []Result{Success("a", time.Second)})},
			want:    StatusPassed,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if got := Aggregate(tt.results); got != tt.want {
				t.Errorf("Aggregate() = %q, want %q", got, tt.want)
			}
		})
	}
}
