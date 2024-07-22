package filters

import (
	"errors"
	"io"
	"os"
	"regexp"
	"strings"

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
	typeBinary

	detectTypes    = typeText | typeBinary
	detectBufSize  = 1024
	executableMask = 0o111
)

func Apply(fs afero.Fs, command *config.Command, files []string) []string {
	if len(files) == 0 {
		return nil
	}

	log.Debug("[lefthook] files before filters:\n", files)

	files = byGlob(files, command.Glob)
	files = byExclude(files, command.Exclude)
	files = byRoot(files, command.Root)
	files = byType(fs, files, command.FileTypes)

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

func byExclude(vs []string, matcher interface{}) []string {
	switch exclude := matcher.(type) {
	case nil:
		return vs
	case string:
		if len(exclude) == 0 {
			return vs
		}

		vsf := make([]string, 0)
		for _, v := range vs {
			if res, _ := regexp.MatchString(exclude, v); !res {
				vsf = append(vsf, v)
			}
		}

		return vsf
	case []interface{}:
		if len(exclude) == 0 {
			return vs
		}

		globs := make([]glob.Glob, 0, len(exclude))
		for _, name := range exclude {
			globs = append(globs, glob.MustCompile(name.(string)))
		}

		vsf := make([]string, 0)
		for _, v := range vs {
			for _, g := range globs {
				if ok := g.Match(v); !ok {
					vsf = append(vsf, v)
				}
			}
		}

		return vsf
	}

	log.Warn("invalid value for exclude option")

	return vs
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
		var err error
		var fileInfo os.FileInfo
		lfs, ok := fs.(afero.Lstater)
		if ok {
			fileInfo, _, err = lfs.LstatIfPossible(v)
		} else {
			fileInfo, err = fs.Stat(v)
		}
		if err != nil {
			log.Errorf("Couldn't check file type of %s: %s", v, err)
			continue
		}

		isSymlink := fileInfo.Mode()&os.ModeSymlink != 0
		isExecutable := fileInfo.Mode().Perm()&executableMask != 0
		if mask&typeSymlink != 0 && !isSymlink {
			continue
		}
		if mask&typeNotSymlink != 0 && isSymlink {
			continue
		}
		if mask&typeExecutable != 0 && (!isExecutable || isSymlink) {
			continue
		}
		if mask&typeNotExecutable != 0 && (isExecutable && !isSymlink) {
			continue
		}

		if mask&detectTypes != 0 {
			if !fileInfo.Mode().IsRegular() {
				continue
			}

			text := checkIsText(fs, v)
			binary := !text

			if mask&typeText != 0 && binary {
				continue
			}
			if mask&typeBinary != 0 && text {
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
		case "text":
			mask |= typeText
		default:
			log.Warn("Unknown filter type: ", t)
		}
	}

	return mask
}

func checkIsText(fs afero.Fs, filepath string) bool {
	file, err := fs.Open(filepath)
	if err != nil {
		log.Error("Couldn't open file for content detecting: ", err)
		return false
	}

	var buf []byte = make([]byte, detectBufSize)
	n, err := io.ReadFull(file, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		log.Error("Couldn't read file for content detecting: ", err)
		return false
	}

	return detectText(buf[:n])
}
