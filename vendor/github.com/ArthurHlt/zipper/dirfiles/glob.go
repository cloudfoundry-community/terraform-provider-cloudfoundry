package dirfiles

import "github.com/gobwas/glob"

type Glob struct {
	wrap    glob.Glob
	pattern string
}

func (g Glob) Match(p string) bool {
	return g.wrap.Match(p)
}

func (g Glob) String() string {
	return g.pattern
}

func MustCompileGlob(p string) Glob {
	return Glob{glob.MustCompile(p), p}
}
