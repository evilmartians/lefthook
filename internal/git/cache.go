package git

import (
	"sync"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
)

type Cache struct {
	wrapper Wrapper

	stagedFilesOnce            func() ([]string, error)
	stagedFilesWithDeletedOnce func() ([]string, error)
	statusShortOnce            func() ([]wrapper.FileStatus, error)
	stateOnce                  func() wrapper.State
}

func newCache(wrapper Wrapper) *Cache {
	c := &Cache{wrapper: wrapper}
	c.Reset()

	return c
}

// WarmUp starts warming up the cache, returns a waiter func. The status is
// only needed to hide unstaged changes, and is the costliest to compute in a
// large repository, so it is warmed up only on request.
func (c *Cache) WarmUp(withStatus bool) func() {
	var wg sync.WaitGroup

	wg.Go(func() {
		_, _ = c.stagedFilesOnce()
	})

	wg.Go(func() {
		_, _ = c.stagedFilesWithDeletedOnce()
	})

	if withStatus {
		wg.Go(func() {
			_, _ = c.statusShortOnce()
		})
	}

	return wg.Wait
}

// Reset cleans up the cached values.
func (c *Cache) Reset() {
	c.stagedFilesOnce = sync.OnceValues(func() ([]string, error) {
		return c.wrapper.StagedFiles()
	})

	c.stagedFilesWithDeletedOnce = sync.OnceValues(func() ([]string, error) {
		return c.wrapper.StagedFilesWithDeleted()
	})

	c.statusShortOnce = sync.OnceValues(func() ([]wrapper.FileStatus, error) {
		return c.wrapper.StatusShort()
	})

	c.stateOnce = sync.OnceValue(func() wrapper.State {
		return c.wrapper.State()
	})
}
