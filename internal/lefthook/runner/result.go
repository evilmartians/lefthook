package runner

const (
	StatusOk status = iota
	StatusErr
)

type Result struct {
	Name   string
	Status status
}

func resultSuccess(name string) Result {
	return Result{Name: name, Status: StatusOk}
}

func resultFail(name string) Result {
	return Result{Name: name, Status: StatusErr}
}
