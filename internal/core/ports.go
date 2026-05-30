// Package core declares the cross-cutting ports used by multiple bounded
// contexts in onit's hexagonal architecture (ADR-008). Context-specific ports
// (discovery.Port, scheduling.CalendarPort, memory.Port) live in their own
// sub-packages; the ports here are consumed by more than one context and have
// no single natural home.
package core

import (
	"context"
	"time"

	"github.com/bredacoder/onit-ai/internal/core/ids"
	"github.com/bredacoder/onit-ai/internal/core/understanding"
)

// Persistence is the storage port for the onit core.
// It grows per slice — only ListTasks is required for the Foundation slice.
type Persistence interface {
	ListTasks(ctx context.Context, userID ids.UserID) ([]understanding.Task, error)
}

// Message is a single turn in a conversation sent to the LLM.
type Message struct {
	// Role is "user" or "assistant".
	Role    string
	Content string
}

// ToolSpec describes a tool the LLM may call.
type ToolSpec struct {
	Name        string
	Description string
	// InputSchema is a JSON Schema object describing the tool's parameters.
	InputSchema map[string]any
}

// ToolCall is a single tool invocation returned by the LLM.
type ToolCall struct {
	// ID is the opaque call identifier returned by the LLM; must be echoed back
	// when reporting the tool result.
	ID    string
	Name  string
	Input map[string]any
}

// Request is the input to a single LLM turn (ADR-003).
type Request struct {
	System   string
	Messages []Message
	Tools    []ToolSpec
}

// Response is the output of a single LLM turn (ADR-003).
// Exactly one of Text or ToolCalls is non-zero per turn.
type Response struct {
	Text      string
	ToolCalls []ToolCall
}

// LLM is the single-turn language-model port (ADR-003).
// The agent loop (iterate until no more tool calls) is implemented once in the
// core and is provider-agnostic; the adapter translates Request/Response to the
// wire format of the concrete provider. The SDK toolrunner is NOT used.
type LLM interface {
	Complete(ctx context.Context, req Request) (Response, error)
}

// Clock isolates the core from the wall clock, making time-dependent logic
// deterministic in tests.
type Clock interface {
	Now() time.Time
}
