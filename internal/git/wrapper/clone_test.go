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

func TestWrapper_Clone(t *testing.T) {
	fs := afero.NewMemMapFs()
	errClone := errors.New("clone failed")
	errNotRef := errors.New("no matching ref")

	const (
		url    = "https://example.com/repo.git"
		sha1   = "13d66a090ac6e73f3ff44652761facb02b00e91c"
		sha256 = "13d66a090ac6e73f3ff44652761facb02b00e91c13d66a090ac6e73f3ff44652"
	)

	for name, tt := range map[string]struct {
		args       wrapper.CloneArgs
		outs       []cmdtest.Out
		err        error
		errContain string
	}{
		"without-ref": {
			args: wrapper.CloneArgs{Root: "/root", Dest: "dest", Url: url},
			outs: []cmdtest.Out{
				{Command: "git -C /root clone --quiet --origin origin --depth 1 " + url + " dest"},
			},
		},
		"with-ref": {
			args: wrapper.CloneArgs{Root: "/root", Dest: "dest", Url: url, Ref: "v1.0.0"},
			outs: []cmdtest.Out{
				{Command: "git -C /root clone --quiet --origin origin --depth 1 --branch v1.0.0 " + url + " dest"},
			},
		},
		"with-commit-sha": {
			args: wrapper.CloneArgs{Root: "/root", Dest: "dest", Url: url, Ref: sha1},
			outs: []cmdtest.Out{
				{Command: "git ls-remote --refs --exit-code " + url + " " + sha1, Err: errNotRef},
				{Command: "git -C /root init --quiet dest"},
				{Command: "git -C /root/dest remote add origin " + url},
			},
		},
		"with-sha256-commit": {
			args: wrapper.CloneArgs{Root: "/root", Dest: "dest", Url: url, Ref: sha256},
			outs: []cmdtest.Out{
				{Command: "git ls-remote --refs --exit-code " + url + " " + sha256, Err: errNotRef},
				{Command: "git -C /root init --quiet --object-format=sha256 dest"},
				{Command: "git -C /root/dest remote add origin " + url},
			},
		},
		"with-hex-branch-name": {
			args: wrapper.CloneArgs{Root: "/root", Dest: "dest", Url: url, Ref: sha1},
			outs: []cmdtest.Out{
				{Command: "git ls-remote --refs --exit-code " + url + " " + sha1, Output: sha1 + "\trefs/heads/" + sha1},
				{Command: "git -C /root clone --quiet --origin origin --depth 1 --branch " + sha1 + " " + url + " dest"},
			},
		},
		"with-suffixed-hex-ref-only": {
			args: wrapper.CloneArgs{Root: "/root", Dest: "dest", Url: url, Ref: sha1},
			outs: []cmdtest.Out{
				{Command: "git ls-remote --refs --exit-code " + url + " " + sha1, Output: sha1 + "\trefs/heads/feature/" + sha1},
				{Command: "git -C /root init --quiet dest"},
				{Command: "git -C /root/dest remote add origin " + url},
			},
		},
		"with-abbreviated-sha": {
			args: wrapper.CloneArgs{Root: "/root", Dest: "dest", Url: url, Ref: "13d66a0"},
			outs: []cmdtest.Out{
				{Command: "git ls-remote --refs --exit-code " + url + " 13d66a0", Err: errNotRef},
			},
			errContain: "abbreviated commit SHA",
		},
		"clone-fails": {
			args: wrapper.CloneArgs{Root: "/root", Dest: "dest", Url: url},
			outs: []cmdtest.Out{
				{Command: "git -C /root clone --quiet --origin origin --depth 1 " + url + " dest", Err: errClone},
			},
			err: errClone,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := cmdtest.NewFakeCmd(t, tt.outs)
			w := wrapper.New(fs, cmd, loggertest.New())

			err := w.Clone(tt.args)

			switch {
			case tt.err != nil:
				if !errors.Is(err, tt.err) {
					t.Errorf("wrapper.Clone() error = %v, want %v", err, tt.err)
				}
			case len(tt.errContain) > 0:
				if err == nil || !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("wrapper.Clone() error = %v, want it to contain %q", err, tt.errContain)
				}
			case err != nil:
				t.Errorf("wrapper.Clone() error = %v, want nil", err)
			}
		})
	}
}
