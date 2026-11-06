package wrapper_test

import (
	"errors"
	"strings"
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
	errNotRef := errors.New("no matching ref")

	const sha = "13d66a090ac6e73f3ff44652761facb02b00e91c"

	for name, tt := range map[string]struct {
		ref        string
		outs       []cmdtest.Out
		err        error
		errContain string
	}{
		"fetches-and-checkouts": {
			ref: "main",
			outs: []cmdtest.Out{
				{Command: "git -C /repo fetch --quiet --depth 1 origin -- main"},
				{Command: "git -C /repo checkout FETCH_HEAD"},
			},
		},
		"fetches-and-checkouts-commit-sha": {
			ref: sha,
			outs: []cmdtest.Out{
				{Command: "git -C /repo ls-remote --refs --exit-code origin " + sha, Err: errNotRef},
				{Command: "git -C /repo fetch --quiet --depth 1 origin -- " + sha},
				{Command: "git -C /repo rev-parse --verify --quiet " + sha + "^{commit}"},
				{Command: "git -C /repo checkout FETCH_HEAD"},
			},
		},
		"fetches-hex-branch-by-full-name": {
			ref: sha,
			outs: []cmdtest.Out{
				{
					Command: "git -C /repo ls-remote --refs --exit-code origin " + sha,
					Output:  sha + "\trefs/heads/feature/" + sha + "\n" + sha + "\trefs/heads/" + sha,
				},
				{Command: "git -C /repo fetch --quiet --depth 1 origin -- refs/heads/" + sha},
				{Command: "git -C /repo checkout FETCH_HEAD"},
			},
		},
		"commit-sha-not-found": {
			ref: sha,
			outs: []cmdtest.Out{
				{Command: "git -C /repo ls-remote --refs --exit-code origin " + sha, Err: errNotRef},
				{Command: "git -C /repo fetch --quiet --depth 1 origin -- " + sha},
				{Command: "git -C /repo rev-parse --verify --quiet " + sha + "^{commit}", Err: errRevParse},
			},
			err: errRevParse,
		},
		"rejects-abbreviated-sha": {
			ref: "13d66a0",
			outs: []cmdtest.Out{
				{Command: "git -C /repo ls-remote --refs --exit-code origin 13d66a0", Err: errNotRef},
			},
			errContain: "abbreviated commit SHA",
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

			switch {
			case tt.err != nil:
				if !errors.Is(err, tt.err) {
					t.Errorf("wrapper.Fetch() error = %v, want %v", err, tt.err)
				}
			case len(tt.errContain) > 0:
				if err == nil || !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("wrapper.Fetch() error = %v, want it to contain %q", err, tt.errContain)
				}
			case err != nil:
				t.Errorf("wrapper.Fetch() error = %v, want nil", err)
			}
		})
	}
}
