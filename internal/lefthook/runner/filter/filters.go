package filter

import (
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gobwas/glob"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
)

type typeMask int

const (
	typeExecutable typeMask = 1 << iota
	typeNotExecutable
	typeSymlink
	typeNotSymlink
	typeText
	typeNotText
	typeBinary
	typeNotBinary

	detectTypes   = typeText | typeNotText | typeBinary | typeNotBinary
	detectBufSize = 1024
)

func Apply(fs afero.Fs, command *config.Command, files []string) []string {
	if len(files) == 0 {
		return nil
	}

	log.Debug("[lefthook] files before filters:\n", files)

	files = byGlob(files, command.Glob)
	files = byExclude(files, command.Exclude)
	files = byRoot(files, command.Root)
	files = byType(fs, files, command.Types)

	log.Debug("[lefthook] files after filters:\n", files)

	return files
}

func byGlob(vs []string, matcher string) []string {
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

func byExclude(vs []string, matcher string) []string {
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

func byRoot(vs []string, matcher string) []string {
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

func byType(fs afero.Fs, vs []string, types []string) []string {
	if len(types) == 0 {
		return vs
	}

	mask := fillTypeMask(types)

	vsf := make([]string, 0)
	for _, v := range vs {
		fileInfo, err := fs.Stat(v)
		if err != nil {
			continue
		}

		if mask&typeSymlink != 0 && fileInfo.Mode()&os.ModeSymlink == 0 {
			continue
		}
		if mask&typeNotSymlink != 0 && fileInfo.Mode()&os.ModeSymlink != 0 {
			continue
		}
		if mask&typeExecutable != 0 && fileInfo.Mode().Perm()&0o111 == 0 {
			continue
		}
		if mask&typeNotExecutable != 0 && fileInfo.Mode().Perm()&0o111 != 0 {
			continue
		}

		if mask&detectTypes != 0 {
			text := fileInfo.Mode().IsRegular() && isText(fs, v)
			binary := fileInfo.Mode().IsRegular() && !text

			if mask&typeText != 0 && binary {
				continue
			}
			if mask&typeNotText != 0 && text {
				continue
			}
			if mask&typeBinary != 0 && text {
				continue
			}
			if mask&typeNotBinary != 0 && binary {
				continue
			}
		}

		vsf = append(vsf, v)
	}

	return vsf
}

func fillTypeMask(types []string) typeMask {
	var mask typeMask

	for _, t := range types {
		switch t {
		case "executable":
			mask |= typeExecutable
		case "symlink":
			mask |= typeSymlink
		case "not executable":
			mask |= typeNotExecutable
		case "not symlink":
			mask |= typeNotSymlink
		case "binary":
			mask |= typeBinary
		case "not binary":
			mask |= typeNotBinary
		case "text":
			mask |= typeText
		case "not text":
			mask |= typeNotText
		default:
			log.Warn("Unknown filter type: ", t)
		}
	}

	return mask
}

func isText(fs afero.Fs, filepath string) bool {
	file, err := fs.Open(filepath)
	if err != nil {
		panic("Could not open the file")
	}

	var buf []byte = make([]byte, 0, detectBufSize)
	_, err = io.ReadFull(file, buf)
	if err != nil {
		panic("Could not read the file")
	}

	mime := mimetype.Detect(buf)
	for mtype := mime; mtype != nil; mtype = mtype.Parent() {
		if mtype.Is("text/plain") {
			return true
		}
	}

	return false
}
