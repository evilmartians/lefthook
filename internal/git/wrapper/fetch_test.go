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
	errRevParse := errors.New("rev-parse failed")

	for name, tt := range map[string]struct {
		ref  string
		outs []cmdtest.Out
		err  error
	}{
		"fetches-and-checkouts": {
			ref: "main",
			outs: []cmdtest.Out{
				{Command: "git -C /repo fetch --quiet --depth 1 origin -- main"},
				{Command: "git -C /repo checkout FETCH_HEAD"},
			},
		},
		"fetches-and-checkouts-commit-sha": {
			ref: "13d66a090ac6e73f3ff44652761facb02b00e91c",
			outs: []cmdtest.Out{
				{Command: "git -C /repo fetch --quiet --depth 1 origin -- 13d66a090ac6e73f3ff44652761facb02b00e91c"},
				{Command: "git -C /repo rev-parse --verify --quiet 13d66a090ac6e73f3ff44652761facb02b00e91c^{commit}"},
				{Command: "git -C /repo checkout FETCH_HEAD"},
			},
		},
		"commit-sha-not-found": {
			ref: "13d66a090ac6e73f3ff44652761facb02b00e91c",
			outs: []cmdtest.Out{
				{Command: "git -C /repo fetch --quiet --depth 1 origin -- 13d66a090ac6e73f3ff44652761facb02b00e91c"},
				{Command: "git -C /repo rev-parse --verify --quiet 13d66a090ac6e73f3ff44652761facb02b00e91c^{commit}", Err: errRevParse},
			},
			err: errRevParse,
		},
		"fetch-fails": {
			ref: "main",
			outs: []cmdtest.Out{
				{Command: "git -C /repo fetch --quiet --depth 1 origin -- main", Err: errFetch},
			},
			err: errFetch,
		},
		"checkout-fails": {
			ref: "main",
			outs: []cmdtest.Out{
				{Command: "git -C /repo fetch --quiet --depth 1 origin -- main"},
				{Command: "git -C /repo checkout FETCH_HEAD", Err: errCheckout},
			},
			err: errCheckout,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewFakeCmd(t, tt.outs)
			w := wrapper.New(fs, cmd, loggertest.New())

			err := w.Fetch(tt.ref, "/repo")

			if !errors.Is(err, tt.err) {
				t.Errorf("wrapper.Fetch() error = %v, want %v", err, tt.err)
			}
		})
	}
}
