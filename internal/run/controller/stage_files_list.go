package controller

import "sync"

type stageFilesList struct {
	mu    sync.Mutex
	files map[string]struct{}
}

func newStageFilesList() *stageFilesList {
	return &stageFilesList{
		files: make(map[string]struct{}),
	}
}

func (f *stageFilesList) Add(files []string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, file := range files {
		f.files[file] = struct{}{}
	}
}

func (f *stageFilesList) Files() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	result := make([]string, 0, len(f.files))

	for file := range f.files {
		result = append(result, file)
	}

	return result
}
