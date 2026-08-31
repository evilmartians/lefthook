package result

import "time"

type status int8

const (
	success status = iota
	failure
	skip
)

// Result contains name of a command/script, an optional fail string, and execution duration.
type Result struct {
	Sub      []Result
	Name     string
	text     string
	status   status
	exitCode int
	Duration time.Duration
}

func (r Result) Success() bool {
	return r.status == success
}

func (r Result) Failure() bool {
	return r.status == failure
}

func (r Result) Text() string {
	return r.text
}

// ExitCode returns the subprocess exit code for failed jobs (defaults to 1).
func (r Result) ExitCode() int {
	if r.exitCode > 0 {
		return r.exitCode
	}
	if r.Failure() {
		return 1
	}
	return 0
}

func Skip(name string) Result {
	return Result{Name: name, status: skip}
}

func Success(name string, duration time.Duration) Result {
	return Result{Name: name, status: success, Duration: duration}
}

func Failure(name, text string, duration time.Duration) Result {
	return FailureWithCode(name, text, duration, 1)
}

func FailureWithCode(name, text string, duration time.Duration, exitCode int) Result {
	if exitCode == 0 {
		exitCode = 1
	}
	return Result{Name: name, status: failure, text: text, Duration: duration, exitCode: exitCode}
}

func Group(name string, results []Result) Result {
	stat := success
	allSkip := true
	var totalDuration time.Duration
	maxExitCode := 0
	for _, res := range results {
		switch res.status {
		case success:
			allSkip = false
		case failure:
			stat = failure
			allSkip = false
			if code := res.ExitCode(); code > maxExitCode {
				maxExitCode = code
			}
		case skip:
		}
		totalDuration += res.Duration
	}

	if allSkip {
		stat = skip
	}

	return Result{Name: name, status: stat, Sub: results, Duration: totalDuration, exitCode: maxExitCode}
}
