package inmem

import (
	"context"

	"github.com/bredacoder/onit-ai/internal/core"
)

// Compile-time assertion: LLM must satisfy core.LLM.
var _ core.LLM = (*LLM)(nil)

// LLM is an in-memory fake for core.LLM. It returns a fixed Response on every
// call. Response may be configured before use; the zero value returns an empty
// text response with no tool calls.
type LLM struct {
	Response core.Response
}

// NewLLM returns an LLM fake whose Complete always returns the given response.
func NewLLM(resp core.Response) *LLM {
	return &LLM{Response: resp}
}

// Complete returns the canned Response regardless of the request.
func (l *LLM) Complete(_ context.Context, _ core.Request) (core.Response, error) {
	return l.Response, nil
}
