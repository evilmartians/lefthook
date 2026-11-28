package config

import "testing"

func TestRemoteDirName(t *testing.T) {
	a := &Remote{GitURL: "https://github.com/scop/lefthook-test.git"}
	b := &Remote{GitURL: "https://github.com/evilmartians/lefthook-test.git", Ref: "v1"}
	c := &Remote{GitURL: "https://github.com/evilmartians/lefthook.git"}
	a2 := &Remote{GitURL: a.GitURL, Ref: "main"}

	for name, tt := range map[string]struct {
		remote  *Remote
		remotes []*Remote
		want    string
	}{
		"unique name keeps plain dir": {c, []*Remote{a, b, c}, "lefthook"},
		"same name, different urls":   {a, []*Remote{a, b}, "lefthook-test-85e079c5"},
		"same name with ref":          {b, []*Remote{a, b}, "lefthook-test-6d4f7157-v1"},
		"same url, different refs":    {a2, []*Remote{a, a2}, "lefthook-test-main"},
	} {
		t.Run(name, func(t *testing.T) {
			if got := tt.remote.DirName(tt.remotes); got != tt.want {
				t.Errorf("DirName() = %q, want %q", got, tt.want)
			}
		})
	}

	if a.DirName([]*Remote{a, b}) == b.DirName([]*Remote{a, b}) {
		t.Error("remotes with the same repository name must get different dirs")
	}
}
