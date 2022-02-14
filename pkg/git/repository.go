package git

type Repository struct {
	HooksPath string
	RootPath  string
	GitPath   string
}

func NewRepository() (*Repository, error) {
	repo, err := NewGit2GoRepository()
	if err != nil {
		return nil, err
	}

	hooksPath, err := repo.HooksPath()
	if err != nil {
		return nil, err
	}

	return &Repository{
		HooksPath: hooksPath,
		RootPath:  repo.RootPath(),
		GitPath:   repo.GitPath(),
	}, nil
}
