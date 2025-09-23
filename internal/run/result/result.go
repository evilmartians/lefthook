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

func Skip(name string) Result {
	return Result{Name: name, status: skip}
}

func Success(name string, duration time.Duration) Result {
	return Result{Name: name, status: success, Duration: duration}
}

func Failure(name, text string, duration time.Duration) Result {
	return Result{Name: name, status: failure, text: text, Duration: duration}
}

func Group(name string, results []Result) Result {
	stat := success
	allSkip := true
	var totalDuration time.Duration
	for _, res := range results {
		switch res.status {
		case success:
			allSkip = false
		case failure:
			stat = failure
			allSkip = false
		case skip:
		}
		totalDuration += res.Duration
	}

	if allSkip {
		stat = skip
	}

	return Result{Name: name, status: stat, Sub: results, Duration: totalDuration}
}
