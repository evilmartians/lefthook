package runner

type status int8

const (
	success status = iota
	failure
	skip
)

// Result contains name of a command/script and an optional fail string.
type Result struct {
	Sub    []Result
	Name   string
	text   string
	status status
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

func skipped(name string) Result {
	return Result{Name: name, status: skip}
}

func succeeded(name string) Result {
	return Result{Name: name, status: success}
}

func failed(name, text string) Result {
	return Result{Name: name, status: failure, text: text}
}

func groupResult(name string, results []Result) Result {
	var stat status = success
	for _, res := range results {
		if res.status == failure {
			stat = failure
			break
		}
		if res.status == skip {
			stat = skip
		}
	}

	return Result{Name: name, status: stat, Sub: results}
}
