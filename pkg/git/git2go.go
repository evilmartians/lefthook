package git

import (
	"os"

	git2go "github.com/libgit2/git2go/v33"
)

type Git2GoRepository struct {
	repo *git2go.Repository
}

func NewGit2GoRepository() (*Git2GoRepository, error) {
	r, err := openRepo()
	if err != nil {
		return nil, err
	}

	return &Git2GoRepository{repo: r}, nil
}

func (r *Git2GoRepository) HooksPath() (string, error) {
	return r.repo.ItemPath(git2go.RepositoryItemHooks)
}

func (r *Git2GoRepository) RootPath() string {
	return r.repo.Workdir()
}

func (r *Git2GoRepository) GitPath() string {
	return r.repo.Path()
}

func (r *Git2GoRepository) OperationInProgress() bool {
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
