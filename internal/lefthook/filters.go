package lefthook

import (
	"regexp"
	"strings"

	"github.com/gobwas/glob"
)

func FilterGlob(vs []string, matcher string) []string {
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

func FilterExclude(vs []string, matcher string) []string {
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

func FilterRelative(vs []string, matcher string) []string {
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
