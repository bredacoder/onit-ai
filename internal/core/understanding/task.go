// Package understanding holds the Task aggregate and its related types for the
// onit core domain. A Task represents a user's intent to hire a local-service
// provider; it carries a typed spine of well-known fields plus an open jsonb
// attribute map for extensibility (PRD §15.3).
package understanding

import (
	"errors"
	"fmt"

	"github.com/bredacoder/onit-ai/internal/core/ids"
)

// ErrUnknownTaskState is returned by ParseTaskState when the input string does
// not match any known TaskState value.
var ErrUnknownTaskState = errors.New("understanding: unknown task state")

// TaskState is the lifecycle state of a Task.
// The set below is PROVISIONAL — it will be finalised in the Understand slice.
type TaskState string

const (
	// TaskCreated is the initial state after the user's free-text intent has been
	// recorded but not yet parsed into structured fields.
	TaskCreated TaskState = "created"

	// TaskUnderstood indicates the agent has extracted all required spine fields
	// from the user's intent.
	TaskUnderstood TaskState = "understood"

	// TaskActing indicates the agent is actively searching for providers or
	// negotiating on the user's behalf.
	TaskActing TaskState = "acting"

	// TaskAwaitingApproval indicates the agent has a proposed outcome and is
	// waiting for explicit human confirmation before proceeding.
	TaskAwaitingApproval TaskState = "awaiting_approval"

	// TaskConfirmed indicates the user has approved the outcome and the task is
	// considered complete.
	TaskConfirmed TaskState = "confirmed"

	// TaskCancelled indicates the task was abandoned before completion.
	TaskCancelled TaskState = "cancelled"
)

// knownTaskStates is the lookup table used by ParseTaskState.
var knownTaskStates = map[string]TaskState{
	string(TaskCreated):          TaskCreated,
	string(TaskUnderstood):       TaskUnderstood,
	string(TaskActing):           TaskActing,
	string(TaskAwaitingApproval): TaskAwaitingApproval,
	string(TaskConfirmed):        TaskConfirmed,
	string(TaskCancelled):        TaskCancelled,
}

// ParseTaskState parses a raw string (e.g. from the database) into a TaskState.
// It returns ErrUnknownTaskState (wrapped) for any input that does not match the
// known value set, so callers can distinguish this condition with errors.Is.
func ParseTaskState(s string) (TaskState, error) {
	if ts, ok := knownTaskStates[s]; ok {
		return ts, nil
	}

	return "", fmt.Errorf("%w: %q", ErrUnknownTaskState, s)
}

// Task is the central aggregate of the understanding bounded context.
// It captures what the user wants done (the typed spine) and leaves room for
// domain-specific attributes via an open jsonb boundary (Attributes).
//
// UserID is present on every transactional type so that multi-tenant isolation
// can be enforced at every layer (PRD §7–§8).
type Task struct {
	ID          ids.TaskID
	UserID      ids.UserID
	ServiceType string
	Location    string
	// TimeWindow is stored as a raw string in M0 and will be refined to a
	// dedicated value object in the Understand slice.
	TimeWindow  string
	BudgetCap   *int64 // cents; nil means unset
	Constraints string
	State       TaskState
	// Attributes is the open jsonb boundary (PRD §15.3). A nil map is treated
	// identically to an empty map — use Attribute() for nil-safe reads.
	Attributes map[string]any
}

// Attribute returns the value stored under key in the open Attributes map.
// It is nil-safe: a nil or empty map returns (nil, false) without panicking,
// so callers can always use this accessor regardless of how the Task was built.
func (t Task) Attribute(key string) (any, bool) {
	if t.Attributes == nil {
		return nil, false
	}

	v, ok := t.Attributes[key]

	return v, ok
}
