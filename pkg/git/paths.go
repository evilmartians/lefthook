package git

import (
	"os"

	git2go "github.com/libgit2/git2go/v33"
)

type Repository struct {
	repo *git2go.Repository
}

func NewRepository() (*Repository, error) {
	r, err := openRepo()
	if err != nil {
		return nil, err
	}

	return &Repository{repo: r}, nil
}

func (r *Repository) HooksPath() (string, error) {
	return r.repo.ItemPath(git2go.RepositoryItemHooks)
}

func (r *Repository) RootPath() string {
	return r.repo.Workdir()
}

func (r *Repository) GitPath() string {
	return r.repo.Path()
}

func (r *Repository) OperationInProgress() bool {
	return r.repo.State() != git2go.RepositoryStateNone
}

func openRepo() (*git2go.Repository, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	r, err := git2go.OpenRepositoryExtended(currentPath, 0, "")
	if err != nil {
		return nil, err
	}

	return r, nil
}
