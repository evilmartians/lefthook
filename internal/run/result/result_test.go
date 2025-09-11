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
