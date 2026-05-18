package controller

import "sync"

type stageFixedFiles struct {
	mu    sync.Mutex
	files map[string]struct{}
}

func newStageFixedFiles() *stageFixedFiles {
	return &stageFixedFiles{
		files: make(map[string]struct{}),
	}
}

func (f *stageFixedFiles) Add(files []string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, file := range files {
		f.files[file] = struct{}{}
	}
}

func (f *stageFixedFiles) Files() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	result := make([]string, 0, len(f.files))

	for file := range f.files {
		result = append(result, file)
	}

	return result
}
