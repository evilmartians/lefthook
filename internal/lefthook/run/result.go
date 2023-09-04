package run

const (
	StatusOk status = iota
	StatusErr
)

type Result struct {
	Name   string
	Text   string
	Status status
}

func resultSuccess(name string) Result {
	return Result{Name: name, Status: StatusOk}
}

func resultFail(name, text string) Result {
	return Result{Name: name, Text: text, Status: StatusErr}
}
