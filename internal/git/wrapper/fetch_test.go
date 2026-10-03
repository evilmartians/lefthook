package wrapper_test

import (
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestWrapper_Fetch(t *testing.T) {
	fs := afero.NewMemMapFs()
	errFetch := errors.New("fetch failed")
	errCheckout := errors.New("checkout failed")

	for name, tt := range map[string]struct {
		outs []cmdtest.Out
		err  error
	}{
		"fetches-and-checkouts": {
			outs: []cmdtest.Out{
				{Command: "git -C /repo fetch --quiet --depth 1 origin -- main"},
				{Command: "git -C /repo checkout FETCH_HEAD"},
			},
		},
		"fetch-fails": {
			outs: []cmdtest.Out{
				{Command: "git -C /repo fetch --quiet --depth 1 origin -- main", Err: errFetch},
			},
			err: errFetch,
		},
		"checkout-fails": {
			outs: []cmdtest.Out{
				{Command: "git -C /repo fetch --quiet --depth 1 origin -- main"},
				{Command: "git -C /repo checkout FETCH_HEAD", Err: errCheckout},
			},
			err: errCheckout,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewOrdered(t, tt.outs)
			w := wrapper.New(fs, cmd, loggertest.New())

			err := w.Fetch("main", "/repo")

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.Fetch() error = %v, want %v", err, tt.err)
			}
		})
	}
}
