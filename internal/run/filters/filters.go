package filters

import (
	"errors"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gobwas/glob"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/internal/log"
)

type fileTypeFilter struct {
	simpleTypes int
	mimeTypes   []string
}

const (
	typeExecutable int = 1 << iota
	typeNotExecutable
	typeSymlink
	typeNotSymlink
	typeText
	typeBinary

	detectTypes    = typeText | typeBinary
	detectBufSize  = 1024
	executableMask = 0o111
)

type Params struct {
	Glob      []string
	Root      string
	FileTypes []string
	Exclude   interface{}
}

func Apply(fs afero.Fs, files []string, params Params) []string {
	if len(files) == 0 {
		return nil
	}

	b := log.Builder(log.DebugLevel, "[lefthook] ").
		Add("filtered [ ]: ", files)

	files = byGlob(files, params.Glob)
	files = byExclude(files, params.Exclude)
	files = byRoot(files, params.Root)
	files = byType(fs, files, params.FileTypes)

	b.Add("filtered [x]: ", files).
		Log()

	return files
}

func byGlob(vs []string, matchers []string) []string {
	if len(matchers) == 0 {
		return vs
	}

	var hasNonEmpty bool
	vsf := make([]string, 0)
	for _, matcher := range matchers {
		if len(matcher) == 0 {
			continue
		}

		hasNonEmpty = true

		g := glob.MustCompile(strings.ToLower(matcher))

		for _, v := range vs {
			if res := g.Match(strings.ToLower(v)); res {
				vsf = append(vsf, v)
			}
		}
	}

	if !hasNonEmpty {
		return vs
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

		var foundMatch bool
		vsf := make([]string, 0)
		for _, v := range vs {
			for _, g := range globs {
				if ok := g.Match(v); ok {
					foundMatch = true
					break
				}
			}

			if !foundMatch {
				vsf = append(vsf, v)
			}
			foundMatch = false
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

	filter := parseFileTypeFilter(types)

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
		if filter.simpleTypes&typeSymlink != 0 && !isSymlink {
			continue
		}
		if filter.simpleTypes&typeNotSymlink != 0 && isSymlink {
			continue
		}
		if filter.simpleTypes&typeExecutable != 0 && (!isExecutable || isSymlink) {
			continue
		}
		if filter.simpleTypes&typeNotExecutable != 0 && (isExecutable && !isSymlink) {
			continue
		}

		if filter.simpleTypes&detectTypes != 0 {
			if !fileInfo.Mode().IsRegular() {
				continue
			}

			text := checkIsText(fs, v)
			binary := !text

			if filter.simpleTypes&typeText != 0 && binary {
				continue
			}
			if filter.simpleTypes&typeBinary != 0 && text {
				continue
			}
		}

		if len(filter.mimeTypes) != 0 {
			if !fileInfo.Mode().IsRegular() {
				continue
			}

			var found bool
			fileMimeType, err := mimetype.DetectFile(v)
			if err != nil {
				log.Errorf("Couldn't check mime type of file %s: %s", v, err)
				continue
			}
			for _, mime := range filter.mimeTypes {
				if fileMimeType.Is(mime) {
					found = true
				}
			}
			if !found {
				continue
			}
		}

		vsf = append(vsf, v)
	}

	return vsf
}

func parseFileTypeFilter(types []string) fileTypeFilter {
	var filter fileTypeFilter

	for _, t := range types {
		switch {
		case t == "executable":
			filter.simpleTypes |= typeExecutable
		case t == "symlink":
			filter.simpleTypes |= typeSymlink
		case t == "not executable":
			filter.simpleTypes |= typeNotExecutable
		case t == "not symlink":
			filter.simpleTypes |= typeNotSymlink
		case t == "binary":
			filter.simpleTypes |= typeBinary
		case t == "text":
			filter.simpleTypes |= typeText
		case (strings.HasPrefix(t, "text/") || strings.HasPrefix(t, "application/")) && mimetype.Lookup(t) != nil:
			filter.mimeTypes = append(filter.mimeTypes, t)
		default:
			log.Warn("Unknown filter type: ", t)
		}
	}

	return filter
}

func checkIsText(fs afero.Fs, filepath string) bool {
	file, err := fs.Open(filepath)
	if err != nil {
		log.Error("Couldn't open file for content detecting: ", err)
		return false
	}

	buf := make([]byte, detectBufSize)
	n, err := io.ReadFull(file, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		log.Error("Couldn't read file for content detecting: ", err)
		return false
	}

	return detectText(buf[:n])
}
