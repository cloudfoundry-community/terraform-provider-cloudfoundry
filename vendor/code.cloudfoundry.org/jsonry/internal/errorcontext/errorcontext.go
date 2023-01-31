// Package errorcontext represents a marshalling or unmarshalling call stack for use in error messages
package errorcontext

import (
	"fmt"
	"reflect"
)

type sort uint

const (
	field sort = iota
	index
	key
)

type ErrorContext []segment

func (ctx ErrorContext) WithField(n string, t reflect.Type) ErrorContext {
	return ctx.push(segment{sort: field, name: n, typ: t})
}

func (ctx ErrorContext) WithIndex(i int, t reflect.Type) ErrorContext {
	return ctx.push(segment{sort: index, index: i, typ: t})
}

func (ctx ErrorContext) WithKey(k string, t reflect.Type) ErrorContext {
	return ctx.push(segment{sort: key, name: k, typ: t})
}

func (ctx ErrorContext) String() string {
	switch len(ctx) {
	case 0:
		return "root path"
	case 1:
		return ctx.leaf().String()
	default:
		return fmt.Sprintf("%s path %s", ctx.leaf(), ctx.path())
	}
}

func (ctx ErrorContext) leaf() segment {
	return ctx[len(ctx)-1]
}

func (ctx ErrorContext) path() string {
	var path string
	for _, s := range ctx {
		switch s.sort {
		case index:
			path = fmt.Sprintf("%s[%d]", path, s.index)
		case field:
			if len(path) > 0 {
				path = path + "."
			}
			path = path + s.name
		case key:
			path = fmt.Sprintf(`%s["%s"]`, path, s.name)
		}
	}

	return path
}

func (ctx ErrorContext) push(s segment) ErrorContext {
	return append([]segment{s}, ctx...)
}
