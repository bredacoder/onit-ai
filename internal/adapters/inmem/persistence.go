// Package inmem provides in-memory fakes for every core port. These fakes are
// intended for use in core unit tests only; they must never be imported by
// production build code. Dependencies point inward: this package imports core
// packages but no adapter packages.
package inmem

import (
	"context"

	"github.com/bredacoder/onit-ai/internal/core"
	"github.com/bredacoder/onit-ai/internal/core/ids"
	"github.com/bredacoder/onit-ai/internal/core/understanding"
)

// Compile-time assertion: Persistence must satisfy core.Persistence.
var _ core.Persistence = (*Persistence)(nil)

// Persistence is an in-memory fake for core.Persistence. Tasks holds all seeded
// tasks across all users; ListTasks returns only those matching the requested
// UserID, enforcing multi-tenant isolation in tests.
type Persistence struct {
	// Tasks is the backing store. Callers may populate it directly before
	// calling ListTasks, or use AddTask for a more fluent API.
	Tasks []understanding.Task
}

// NewPersistence returns a Persistence fake with no tasks seeded.
func NewPersistence() *Persistence {
	return &Persistence{}
}

// AddTask appends t to the internal task slice. It is provided as a
// convenience so tests can seed tasks without accessing the field directly.
func (p *Persistence) AddTask(t understanding.Task) {
	p.Tasks = append(p.Tasks, t)
}

// ListTasks returns only the tasks whose UserID equals userID. It never
// returns tasks belonging to another user and never returns an error.
func (p *Persistence) ListTasks(_ context.Context, userID ids.UserID) ([]understanding.Task, error) {
	result := make([]understanding.Task, 0)
	for _, t := range p.Tasks {
		if t.UserID == userID {
			result = append(result, t)
		}
	}

	return result, nil
}
