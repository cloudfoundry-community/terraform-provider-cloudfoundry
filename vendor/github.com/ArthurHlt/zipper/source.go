package zipper

import (
	"context"
	"net/http"
)

const (
	HttpClientContextKey SourceContextKey = iota
)

type SourceContextKey int

type Source struct {
	// Path for zip handler
	Path string
	// ctx is either the client or server context. It should only
	// be modified via copying the whole Request using WithContext.
	// It is unexported to prevent people from using Context wrong
	// and mutating the contexts held by callers of the same request.
	ctx context.Context
}

// Create a new source
func NewSource(path string) *Source {
	return &Source{Path: path}
}

// Context returns the request's context. To change the context, use
// WithContext.
//
// The returned context is always non-nil; it defaults to the
// background context.
//
// For outgoing client requests, the context controls cancelation.
//
// For incoming server requests, the context is canceled when the
// client's connection closes, the request is canceled (with HTTP/2),
// or when the ServeHTTP method returns.
func (s *Source) Context() context.Context {
	if s.ctx != nil {
		return s.ctx
	}
	return context.Background()
}

// WithContext returns a shallow copy of r with its context changed
// to ctx. The provided ctx must be non-nil.
func (s *Source) WithContext(ctx context.Context) *Source {
	if ctx == nil {
		panic("nil context")
	}
	s2 := new(Source)
	*s2 = *s
	s2.ctx = ctx
	return s2
}

// Set http client in the context of a source
// This could be use for a zip handler
func SetCtxHttpClient(src *Source, client *http.Client) {
	parentContext := src.Context()
	ctxValueReq := src.WithContext(context.WithValue(parentContext, HttpClientContextKey, client))
	*src = *ctxValueReq
}

// Retrieve http client set in context
// This could be use for a zip handler
func CtxHttpClient(src *Source) *http.Client {
	val := src.Context().Value(HttpClientContextKey)
	if val == nil {
		return nil
	}
	return val.(*http.Client)
}
