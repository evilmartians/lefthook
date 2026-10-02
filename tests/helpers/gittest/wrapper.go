package gittest

import "github.com/evilmartians/lefthook/v2/internal/git/wrapper"

type FakeWrapper struct {
	StubVersion                   func() (string, error)
	StubPaths                     func() (*wrapper.PathsResult, error)
	StubLocalHooksPath            func() string
	StubUnsetLocalHooksPath       func() error
	StubGlobalHooksPath           func() string
	StubUnsetGlobalHooksPath      func() error
	StubAllFiles                  func() ([]string, error)
	StubStagedFiles               func() ([]string, error)
	StubStagedFilesWithDeleted    func() ([]string, error)
	StubPushFiles                 func() ([]string, error)
	StubFilesByCommandRelative    func(string, string) ([]string, error)
	StubStatusShort               func() ([]wrapper.FileStatus, error)
	StubDiff                      func([]string, bool) (string, error)
	StubSaveUnstagedDiff          func([]string) error
	StubUnstagedDiffApplicable    func() bool
	StubApplyUnstagedDiff         func(bool) error
	StubStoreStash                func() error
	StubDropStash                 func() error
	StubDiscardUnstagedChanges    func([]string) error
	StubDiscardAllUnstagedChanges func() error
	StubStageFiles                func([]string) error
	StubHashObjects               func([]string) ([]string, error)
	StubState                     func() wrapper.State
	StubClone                     func(wrapper.CloneArgs) error
	StubPull                      func(string) error
	StubFetch                     func(string, string) error
}

func NewFakeWrapper() *FakeWrapper {
	return &FakeWrapper{}
}

func (w *FakeWrapper) Version() (string, error)             { return w.StubVersion() }
func (w *FakeWrapper) Paths() (*wrapper.PathsResult, error) { return w.StubPaths() }
func (w *FakeWrapper) LocalHooksPath() string               { return w.StubLocalHooksPath() }
func (w *FakeWrapper) UnsetLocalHooksPath() error           { return w.StubUnsetLocalHooksPath() }
func (w *FakeWrapper) GlobalHooksPath() string              { return w.StubGlobalHooksPath() }
func (w *FakeWrapper) UnsetGlobalHooksPath() error          { return w.StubUnsetLocalHooksPath() }
func (w *FakeWrapper) AllFiles() ([]string, error)          { return w.StubAllFiles() }
func (w *FakeWrapper) StagedFiles() ([]string, error)       { return w.StubStagedFiles() }
func (w *FakeWrapper) StagedFilesWithDeleted() ([]string, error) {
	return w.StubStagedFilesWithDeleted()
}
func (w *FakeWrapper) PushFiles() ([]string, error) { return w.StubPushFiles() }
func (w *FakeWrapper) FilesByCommandRelative(a string, b string) ([]string, error) {
	return w.StubFilesByCommandRelative(a, b)
}
func (w *FakeWrapper) StatusShort() ([]wrapper.FileStatus, error) { return w.StubStatusShort() }
func (w *FakeWrapper) Diff(s []string, b bool) (string, error)    { return w.StubDiff(s, b) }
func (w *FakeWrapper) SaveUnstagedDiff(s []string) error          { return w.StubSaveUnstagedDiff(s) }
func (w *FakeWrapper) UnstagedDiffApplicable() bool               { return w.StubUnstagedDiffApplicable() }
func (w *FakeWrapper) ApplyUnstagedDiff(b bool) error             { return w.StubApplyUnstagedDiff(b) }
func (w *FakeWrapper) StoreStash() error                          { return w.StubStoreStash() }
func (w *FakeWrapper) DropStash() error                           { return w.StubDropStash() }
func (w *FakeWrapper) DiscardUnstagedChanges(s []string) error {
	return w.StubDiscardUnstagedChanges(s)
}
func (w *FakeWrapper) DiscardAllUnstagedChanges() error         { return w.StubDiscardAllUnstagedChanges() }
func (w *FakeWrapper) StageFiles(s []string) error              { return w.StubStageFiles(s) }
func (w *FakeWrapper) HashObjects(s []string) ([]string, error) { return w.StubHashObjects(s) }
func (w *FakeWrapper) State() wrapper.State                     { return w.StubState() }
func (w *FakeWrapper) Clone(a wrapper.CloneArgs) error          { return w.StubClone(a) }
func (w *FakeWrapper) Pull(s string) error                      { return w.StubPull(s) }
func (w *FakeWrapper) Fetch(a string, b string) error           { return w.StubFetch(a, b) }
