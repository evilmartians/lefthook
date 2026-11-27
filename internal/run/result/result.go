package result

import "time"

type status int8

const (
	success status = iota
	failure
	skip
)

// Status is the machine-readable state of a Result.
type Status string

const (
	// StatusPassed means the job finished successfully.
	StatusPassed Status = "passed"
	// StatusFailed means the job ran and its own command exited with a non-zero code.
	StatusFailed Status = "failed"
	// StatusSkipped means the job did not run.
	StatusSkipped Status = "skipped"
	// StatusError means the job could not run, or did not run to completion.
	StatusError Status = "error"
)

// Result contains name of a command/script, an optional fail string, and execution duration.
type Result struct {
	Sub      []Result
	Name     string
	text     string
	reason   string
	status   status
	exitCode *int
	Duration time.Duration
}

func (r Result) Success() bool {
	return r.status == success
}

func (r Result) Failure() bool {
	return r.status == failure
}

func (r Result) Skip() bool {
	return r.status == skip
}

func (r Result) Text() string {
	return r.text
}

// Reason returns why the job was skipped, when that is known.
func (r Result) Reason() string {
	return r.reason
}

// ExitCode returns the exit code of the command the job ran, when it is known.
func (r Result) ExitCode() (int, bool) {
	if r.exitCode == nil {
		return 0, false
	}

	return *r.exitCode, true
}

// Status returns the machine-readable state of the result. A group takes the
// combined state of its sub-results.
//
// A failure that carries an exit code comes from the job's own command, so it is
// reported as failed. Other failures — an invalid job, a timeout, an error while
// building the command — mean the job did not run to completion.
func (r Result) Status() Status {
	if len(r.Sub) > 0 {
		return Aggregate(r.Sub)
	}

	switch r.status {
	case success:
		return StatusPassed
	case failure:
		if r.exitCode != nil {
			return StatusFailed
		}

		return StatusError
	case skip:
		return StatusSkipped
	}

	return StatusError
}

func Skip(name string) Result {
	return Result{Name: name, status: skip}
}

// SkipWithReason returns a skipped result that also records why it was skipped.
func SkipWithReason(name, reason string) Result {
	return Result{Name: name, status: skip, reason: reason}
}

func Success(name string, duration time.Duration) Result {
	return Result{Name: name, status: success, Duration: duration}
}

func Failure(name, text string, duration time.Duration) Result {
	return Result{Name: name, status: failure, text: text, Duration: duration}
}

// WithExitCode records the exit code of the command that produced the result.
// Only set it for a command that actually ran.
func (r Result) WithExitCode(code int) Result {
	r.exitCode = &code

	return r
}

func Group(name string, results []Result) Result {
	var totalDuration time.Duration
	for _, res := range results {
		totalDuration += res.Duration
	}

	stat := success
	switch Aggregate(results) {
	case StatusSkipped:
		stat = skip
	case StatusFailed, StatusError:
		stat = failure
	case StatusPassed:
	}

	return Result{Name: name, status: stat, Sub: results, Duration: totalDuration}
}

// Aggregate returns the combined status of the given results: an error wins over
// a failure, a failure over a pass, and a run where nothing passed is skipped.
func Aggregate(results []Result) Status {
	var anyPassed, anyFailed bool

	for _, res := range results {
		switch res.Status() {
		case StatusError:
			return StatusError
		case StatusFailed:
			anyFailed = true
		case StatusPassed:
			anyPassed = true
		case StatusSkipped:
		}
	}

	switch {
	case anyFailed:
		return StatusFailed
	case anyPassed:
		return StatusPassed
	}

	return StatusSkipped
}
