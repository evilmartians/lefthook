package config

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/evilmartians/lefthook/v2/internal/git"
)

type Remote struct {
	GitURL string `json:"git_url,omitempty" jsonschema:"description=A URL to Git repository. It will be accessed with privileges of the machine lefthook runs on." koanf:"git_url" mapstructure:"git_url" toml:"git_url" yaml:"git_url"`

	Ref string `json:"ref,omitempty" jsonschema:"description=An optional *branch* or *tag* name" mapstructure:"ref,omitempty" toml:"ref,omitempty" yaml:",omitempty"`

	Configs []string `json:"configs,omitempty" jsonschema:"description=An optional array of config paths from remote's root,default=lefthook.yml" mapstructure:"configs,omitempty" toml:"configs,omitempty" yaml:",omitempty"`

	Refetch bool `json:"refetch,omitempty" jsonschema:"description=Set to true if you want to always refetch the remote" mapstructure:"refetch,omitempty" toml:"refetch,omitempty" yaml:",omitempty"`

	RefetchFrequency string `json:"refetch_frequency,omitempty" jsonschema:"description=Provide a frequency for the remotes refetches,example=24h" koanf:"refetch_frequency" mapstructure:"refetch_frequency,omitempty" toml:"refetch_frequency,omitempty" yaml:",omitempty"`
}

func (r *Remote) Configured() bool {
	if r == nil {
		return false
	}

	return len(r.GitURL) > 0
}

// DirName returns the name of the folder this remote is checked out to. It is
// the repository name and the ref, plus a short hash of git_url when another
// remote in remotes has the same repository name but a different git_url.
func (r *Remote) DirName(remotes []*Remote) string {
	return git.RemoteDirectoryName(r.GitURL, r.Ref, r.dirSuffix(remotes))
}

// DirNameWithoutRef is DirName for any ref of this remote's repository.
func (r *Remote) DirNameWithoutRef(remotes []*Remote) string {
	return git.RemoteDirectoryName(r.GitURL, "", r.dirSuffix(remotes))
}

func (r *Remote) dirSuffix(remotes []*Remote) string {
	name := git.RemoteDirectoryName(r.GitURL, "", "")
	for _, other := range remotes {
		if other.Configured() && other.GitURL != r.GitURL && git.RemoteDirectoryName(other.GitURL, "", "") == name {
			sum := sha256.Sum256([]byte(r.GitURL))
			return hex.EncodeToString(sum[:4])
		}
	}

	return ""
}
