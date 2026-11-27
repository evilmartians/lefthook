package report

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/evilmartians/lefthook/v2/internal/run/result"
)

func TestParseFormat(t *testing.T) {
	for name, tt := range map[string]struct {
		value   string
		want    Format
		wantErr bool
	}{
		"default":      {value: "", want: FormatText},
		"text":         {value: "text", want: FormatText},
		"json":         {value: "json", want: FormatJSON},
		"jsonl":        {value: "jsonl", want: FormatJSONL},
		"unknown":      {value: "xml", wantErr: true},
		"wrong-case":   {value: "JSON", wantErr: true},
		"leading-from": {value: " json", wantErr: true},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := ParseFormat(tt.value)

			if tt.wantErr {
				if !errors.Is(err, ErrUnknownFormat) {
					t.Fatalf("ParseFormat(%q) error = %v, want ErrUnknownFormat", tt.value, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("ParseFormat(%q) unexpected error: %v", tt.value, err)
			}

			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestFormat_Machine(t *testing.T) {
	for name, tt := range map[string]struct {
		format Format
		want   bool
	}{
		"text":  {format: FormatText},
		"json":  {format: FormatJSON, want: true},
		"jsonl": {format: FormatJSONL, want: true},
	} {
		t.Run(name, func(t *testing.T) {
			if got := tt.format.Machine(); got != tt.want {
				t.Errorf("Format(%q).Machine() = %v, want %v", tt.format, got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	if reporter := New(FormatText, new(bytes.Buffer)); reporter != nil {
		t.Error("New(FormatText) = non-nil, want nil: text output has no structured report")
	}

	if reporter := New(FormatJSON, new(bytes.Buffer)); reporter == nil {
		t.Error("New(FormatJSON) = nil, want a reporter")
	}
}

func TestReporter_JobDone_jsonl(t *testing.T) {
	var buf bytes.Buffer
	reporter := New(FormatJSONL, &buf)

	reporter.JobDone(nil, result.Success("lint", 1500*time.Millisecond))
	reporter.JobDone([]string{"checks"}, result.Failure("test", "failed", 250*time.Millisecond).WithExitCode(2))
	reporter.JobDone(nil, result.SkipWithReason("docs", "no matching staged files"))

	want := []Job{
		{Name: "lint", Status: result.StatusPassed, DurationMS: 1500},
		{Name: "test", GroupPath: []string{"checks"}, Status: result.StatusFailed, ExitCode: new(2), DurationMS: 250},
		{Name: "docs", Status: result.StatusSkipped, DurationMS: 0, SkipReason: "no matching staged files"},
	}

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != len(want) {
		t.Fatalf("job lines = %d, want %d: %q", len(lines), len(want), buf.String())
	}

	for i, line := range lines {
		var got Job
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatalf("line %d is not valid JSON: %v", i, err)
		}

		if !cmp.Equal(got, want[i]) {
			t.Errorf("job %d = %+v, want %+v", i, got, want[i])
		}
	}
}

func TestReporter_JobDone_jsonlIsNotASummary(t *testing.T) {
	var buf bytes.Buffer
	reporter := New(FormatJSONL, &buf)

	reporter.JobDone(nil, result.Success("lint", time.Millisecond))
	reporter.Summary("pre-commit", result.StatusPassed, nil)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("jsonl wrote %d lines, want 1 (jobs only, no summary): %q", len(lines), buf.String())
	}

	var job Job
	if err := json.Unmarshal([]byte(lines[0]), &job); err != nil {
		t.Fatalf("line is not a job record: %v", err)
	}

	if job.Name != "lint" {
		t.Errorf("job name = %q, want %q", job.Name, "lint")
	}
}

func TestReporter_Summary_json(t *testing.T) {
	var buf bytes.Buffer
	reporter := New(FormatJSON, &buf)

	reporter.JobDone(nil, result.Success("lint", 10*time.Millisecond))
	reporter.JobDone(nil, result.Failure("test", "boom", 20*time.Millisecond).WithExitCode(1))
	reporter.Summary("pre-commit", result.StatusFailed, nil)

	var doc struct {
		Hook       string        `json:"hook"`
		Status     result.Status `json:"status"`
		DurationMS int64         `json:"duration_ms"`
		Error      string        `json:"error"`
		Jobs       []Job         `json:"jobs"`
	}
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("summary is not valid JSON: %v: %q", err, buf.String())
	}

	if doc.Hook != "pre-commit" {
		t.Errorf("hook = %q, want %q", doc.Hook, "pre-commit")
	}

	if doc.Status != result.StatusFailed {
		t.Errorf("status = %q, want %q", doc.Status, result.StatusFailed)
	}

	if len(doc.Jobs) != 2 {
		t.Fatalf("jobs = %d, want 2", len(doc.Jobs))
	}

	if doc.Error != "" {
		t.Errorf("error = %q, want empty", doc.Error)
	}

	if strings.Count(buf.String(), "\n") != 1 {
		t.Errorf("summary should be a single line of JSON, got %q", buf.String())
	}
}

func TestReporter_Summary_recordsError(t *testing.T) {
	var buf bytes.Buffer
	reporter := New(FormatJSON, &buf)

	reporter.Summary("pre-commit", result.StatusError, errors.New("hook pre-commit doesn't exist"))

	var doc struct {
		Status result.Status `json:"status"`
		Error  string        `json:"error"`
	}
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("summary is not valid JSON: %v", err)
	}

	if doc.Status != result.StatusError {
		t.Errorf("status = %q, want %q", doc.Status, result.StatusError)
	}

	if doc.Error != "hook pre-commit doesn't exist" {
		t.Errorf("error = %q, want the run error", doc.Error)
	}
}

func TestReporter_Summary_writtenOnce(t *testing.T) {
	var buf bytes.Buffer
	reporter := New(FormatJSON, &buf)

	reporter.Summary("pre-commit", result.StatusPassed, nil)
	reporter.Summary("pre-commit", result.StatusError, errors.New("late error"))

	var docs int
	decoder := json.NewDecoder(&buf)
	for decoder.More() {
		var doc map[string]any
		if err := decoder.Decode(&doc); err != nil {
			t.Fatalf("unexpected JSON: %v", err)
		}
		docs++
	}

	if docs != 1 {
		t.Errorf("wrote %d summary documents, want 1", docs)
	}

	if strings.Contains(buf.String(), "late error") {
		t.Error("second summary was written, want only the first one")
	}
}

func TestReporter_Summary_noJobsIsAnEmptyList(t *testing.T) {
	var buf bytes.Buffer
	reporter := New(FormatJSON, &buf)

	reporter.Summary("pre-commit", result.StatusSkipped, nil)

	var doc struct {
		Jobs []Job `json:"jobs"`
	}
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("summary is not valid JSON: %v", err)
	}

	if len(doc.Jobs) != 0 {
		t.Errorf("jobs = %v, want an empty list", doc.Jobs)
	}
}

// A nil reporter is the human-readable mode: every method must be a no-op so the
// runner can call it unconditionally.
func TestReporter_nilIsNoOp(t *testing.T) {
	reporter := New(FormatText, new(bytes.Buffer))

	reporter.JobDone(nil, result.Success("lint", time.Millisecond))
	reporter.Summary("pre-commit", result.StatusPassed, nil)
}
