package git

// Repository is an interface to work with git repo.
// It's realization might change
type Repository interface {
	HooksPath() (string, error)
	RootPath() string
	GitPath() string
	OperationInProgress() bool
}
