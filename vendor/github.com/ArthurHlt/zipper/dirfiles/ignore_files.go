package dirfiles

import (
	"path"
	"strings"
)

type IgnoreFiles interface {
	FileShouldBeIgnored(path string) bool
}

func NewIgnoreFiles(text string) IgnoreFiles {
	patterns := []ignorePattern{}
	lines := strings.Split(text, "\n")
	lines = append(defaultIgnoreLines, lines...)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		ignore := true
		if strings.HasPrefix(line, "!") {
			line = line[1:]
			ignore = false
		}

		for _, p := range globsForPattern(path.Clean(line)) {
			patterns = append(patterns, ignorePattern{ignore, p})
		}
	}

	return ignoreFile(patterns)
}

func (ignore ignoreFile) FileShouldBeIgnored(path string) bool {
	result := false

	for _, pattern := range ignore {
		if strings.HasPrefix(pattern.glob.String(), "/") && !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		if pattern.glob.Match(path) {
			result = pattern.exclude
		}
	}

	return result
}

func globsForPattern(pattern string) (globs []Glob) {
	globs = append(globs, MustCompileGlob(pattern))
	globs = append(globs, MustCompileGlob(path.Join(pattern, "*")))
	globs = append(globs, MustCompileGlob(path.Join(pattern, "**", "*")))

	if !strings.HasPrefix(pattern, "/") {
		globs = append(globs, MustCompileGlob(path.Join("**", pattern)))
		globs = append(globs, MustCompileGlob(path.Join("**", pattern, "*")))
		globs = append(globs, MustCompileGlob(path.Join("**", pattern, "**", "*")))
	}

	return
}

type ignorePattern struct {
	exclude bool
	glob    Glob
}

type ignoreFile []ignorePattern

var defaultIgnoreLines = []string{
	".zipignore",
	".cfignore",
	".gitignore",
	".cloudignore",
	".git",
	".hg",
	".svn",
	"_darcs",
	".DS_Store",
}
