package runner

import (
	"os"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
	"github.com/spf13/afero"
)

func filterGlob(vs []string, matcher string) []string {
	if matcher == "" {
		return vs
	}

	g := glob.MustCompile(strings.ToLower(matcher))

	vsf := make([]string, 0)
	for _, v := range vs {
		if res := g.Match(strings.ToLower(v)); res {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func filterExclude(vs []string, matcher string) []string {
	if matcher == "" {
		return vs
	}

	vsf := make([]string, 0)
	for _, v := range vs {
		if res, _ := regexp.MatchString(matcher, v); !res {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func filterRelative(vs []string, matcher string) []string {
	if matcher == "" {
		return vs
	}

	vsf := make([]string, 0)
	for _, v := range vs {
		if strings.HasPrefix(v, matcher) {
			vsf = append(vsf, strings.Replace(v, matcher, "./", 1))
		}
	}
	return vsf
}

func filterSymlinks(vs []string, fs afero.Fs) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		stat, err := fs.Stat(v)
		if err != nil {
			continue
		}

		if stat.Mode()&os.ModeSymlink != 0 {
			continue
		}

		vsf = append(vsf, v)
	}
	return vsf
}
