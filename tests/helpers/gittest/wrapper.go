package gittest

import "github.com/evilmartians/lefthook/v2/internal/git/wrapper"

type StubWrapper struct {
	VersionFunc                   func() (string, error)
	PathsFunc                     func() (*wrapper.PathsResult, error)
	LocalHooksPathFunc            func() string
	UnsetLocalHooksPathFunc       func() error
	StubGlobalHooksPath           func() string
	UnsetGlobalHooksPathFunc      func() error
	AllFilesFunc                  func() ([]string, error)
	StagedFilesFunc               func() ([]string, error)
	StagedFilesWithDeletedFunc    func() ([]string, error)
	PushFilesFunc                 func() ([]string, error)
	FilesByCommandRelativeFunc    func(string, string) ([]string, error)
	StatusShortFunc               func() ([]wrapper.FileStatus, error)
	DiffFunc                      func([]string, bool) (string, error)
	SaveUnstagedDiffFunc          func([]string) error
	UnstagedDiffApplicableFunc    func() bool
	ApplyUnstagedDiffFunc         func(bool) error
	StoreStashFunc                func() error
	DropStashFunc                 func() error
	DiscardUnstagedChangesFunc    func([]string) error
	DiscardAllUnstagedChangesFunc func() error
	StageFilesFunc                func([]string) error
	HashObjectsFunc               func([]string) ([]string, error)
	StateFunc                     func() wrapper.State
	CloneFunc                     func(wrapper.CloneArgs) error
	PullFunc                      func(string) error
	FetchFunc                     func(string, string) error
}

func NewStubWrapper() *StubWrapper {
	return &StubWrapper{}
}

func (w *StubWrapper) Version() (string, error)             { return w.VersionFunc() }
func (w *StubWrapper) Paths() (*wrapper.PathsResult, error) { return w.PathsFunc() }
func (w *StubWrapper) LocalHooksPath() string               { return w.LocalHooksPathFunc() }
func (w *StubWrapper) UnsetLocalHooksPath() error           { return w.UnsetLocalHooksPathFunc() }
func (w *StubWrapper) GlobalHooksPath() string              { return w.StubGlobalHooksPath() }
func (w *StubWrapper) UnsetGlobalHooksPath() error          { return w.UnsetLocalHooksPathFunc() }
func (w *StubWrapper) AllFiles() ([]string, error)          { return w.AllFilesFunc() }
func (w *StubWrapper) StagedFiles() ([]string, error)       { return w.StagedFilesFunc() }
func (w *StubWrapper) StagedFilesWithDeleted() ([]string, error) {
	return w.StagedFilesWithDeletedFunc()
}
func (w *StubWrapper) PushFiles() ([]string, error) { return w.PushFilesFunc() }
func (w *StubWrapper) FilesByCommandRelative(a string, b string) ([]string, error) {
	return w.FilesByCommandRelativeFunc(a, b)
}
func (w *StubWrapper) StatusShort() ([]wrapper.FileStatus, error) { return w.StatusShortFunc() }
func (w *StubWrapper) Diff(s []string, b bool) (string, error)    { return w.DiffFunc(s, b) }
func (w *StubWrapper) SaveUnstagedDiff(s []string) error          { return w.SaveUnstagedDiffFunc(s) }
func (w *StubWrapper) UnstagedDiffApplicable() bool               { return w.UnstagedDiffApplicableFunc() }
func (w *StubWrapper) ApplyUnstagedDiff(b bool) error             { return w.ApplyUnstagedDiffFunc(b) }
func (w *StubWrapper) StoreStash() error                          { return w.StoreStashFunc() }
func (w *StubWrapper) DropStash() error                           { return w.DropStashFunc() }
func (w *StubWrapper) DiscardUnstagedChanges(s []string) error {
	return w.DiscardUnstagedChangesFunc(s)
}
func (w *StubWrapper) DiscardAllUnstagedChanges() error         { return w.DiscardAllUnstagedChangesFunc() }
func (w *StubWrapper) StageFiles(s []string) error              { return w.StageFilesFunc(s) }
func (w *StubWrapper) HashObjects(s []string) ([]string, error) { return w.HashObjectsFunc(s) }
func (w *StubWrapper) State() wrapper.State                     { return w.StateFunc() }
func (w *StubWrapper) Clone(a wrapper.CloneArgs) error          { return w.CloneFunc(a) }
func (w *StubWrapper) Pull(s string) error                      { return w.PullFunc(s) }
func (w *StubWrapper) Fetch(a string, b string) error           { return w.FetchFunc(a, b) }
