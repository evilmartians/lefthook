// Package report turns hook results into machine-readable output.
//
// Two formats are supported: `json` writes one summary document for the whole
// run, and `jsonl` writes one line per job as soon as the job finishes. Both
// describe jobs with the same records, so a consumer only needs one decoder.
package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/evilmartians/lefthook/v2/internal/run/result"
)

// Format is the output format requested on the command line.
type Format string

const (
	// FormatText is the default human-readable output.
	FormatText Format = "text"
	// FormatJSON writes one summary document when the run ends.
	FormatJSON Format = "json"
	// FormatJSONL writes one line per job as the job finishes.
	FormatJSONL Format = "jsonl"
)

// ErrUnknownFormat is returned for an unsupported --output-format value.
var ErrUnknownFormat = errors.New("invalid 'output-format' value, supported: 'text', 'json', 'jsonl'")

// ParseFormat maps the command line value to a Format. An empty value means the
// default human-readable output.
func ParseFormat(value string) (Format, error) {
	switch Format(value) {
	case "", FormatText:
		return FormatText, nil
	case FormatJSON:
		return FormatJSON, nil
	case FormatJSONL:
		return FormatJSONL, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownFormat, value)
	}
}

// Machine reports whether the format replaces the human-readable output.
func (f Format) Machine() bool {
	return f == FormatJSON || f == FormatJSONL
}

// Job is the record of a single job in the structured output.
type Job struct {
	Name       string        `json:"name"`
	GroupPath  []string      `json:"group_path,omitempty"`
	Status     result.Status `json:"status"`
	ExitCode   *int          `json:"exit_code,omitempty"`
	DurationMS int64         `json:"duration_ms"`
	SkipReason string        `json:"skip_reason,omitempty"`
	Error      string        `json:"error,omitempty"`
}

type summary struct {
	Hook       string        `json:"hook"`
	Status     result.Status `json:"status"`
	DurationMS int64         `json:"duration_ms"`
	Error      string        `json:"error,omitempty"`
	Jobs       []Job         `json:"jobs"`
}

// Reporter writes job results in a machine-readable format.
//
// It is safe for concurrent use: jobs of a parallel group finish in different
// goroutines.
type Reporter struct {
	format Format
	out    io.Writer
	start  time.Time

	mu      sync.Mutex
	jobs    []Job
	written bool
}

// New returns a reporter for the given format and writer. It returns nil for the
// human-readable format, which has no structured output.
func New(format Format, out io.Writer) *Reporter {
	if !format.Machine() {
		return nil
	}

	return &Reporter{format: format, out: out, start: time.Now()}
}

// JobDone records a finished job. For `jsonl` this also writes the job line
// immediately, so a consumer can follow the run.
func (r *Reporter) JobDone(groupPath []string, res result.Result) {
	if r == nil {
		return
	}

	job := newJob(groupPath, res)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.jobs = append(r.jobs, job)

	if r.format == FormatJSONL {
		r.writeJSON(job)
	}
}

// Summary writes the run summary for the `json` format. For `jsonl` the job
// lines already carry everything, so nothing is written.
func (r *Reporter) Summary(hook string, status result.Status, runErr error) {
	if r == nil || r.format != FormatJSON {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.written {
		return
	}
	r.written = true

	doc := summary{
		Hook:       hook,
		Status:     status,
		DurationMS: time.Since(r.start).Milliseconds(),
		Jobs:       r.jobs,
	}
	if runErr != nil {
		doc.Error = runErr.Error()
	}

	r.writeJSON(doc)
}

func (r *Reporter) writeJSON(value any) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	_, _ = r.out.Write(append(data, '\n'))
}

func newJob(groupPath []string, res result.Result) Job {
	job := Job{
		Name:       res.Name,
		Status:     res.Status(),
		DurationMS: res.Duration.Milliseconds(),
	}
	if len(groupPath) > 0 {
		job.GroupPath = append([]string(nil), groupPath...)
	}

	if code, ok := res.ExitCode(); ok {
		job.ExitCode = &code
	}

	if res.Skip() {
		job.SkipReason = res.Reason()
	}

	if job.Status == result.StatusError {
		job.Error = res.Text()
	}

	return job
}
