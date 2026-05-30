// Package postgres provides a Postgres-backed implementation of core.Persistence
// using pgx/v5 and sqlc-generated queries. The sqlc struct types are an
// anti-corruption layer (ACL): they are never exposed outside this package.
// All mapping from gen.Task to understanding.Task happens here.
package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bredacoder/onit-ai/internal/adapters/postgres/gen"
	"github.com/bredacoder/onit-ai/internal/core"
	"github.com/bredacoder/onit-ai/internal/core/ids"
	"github.com/bredacoder/onit-ai/internal/core/understanding"
)

// Compile-time assertion: Adapter must satisfy core.Persistence.
var _ core.Persistence = (*Adapter)(nil)

// Adapter implements core.Persistence over a pgx connection pool. It is the
// only place in the codebase that imports the sqlc-generated code; it maps
// every gen.* type to a core domain type before returning.
type Adapter struct {
	pool    *pgxpool.Pool
	queries *gen.Queries
}

// New creates an Adapter backed by pool. It also builds the sqlc Queries
// handle so callers do not need to import the gen package.
func New(pool *pgxpool.Pool) *Adapter {
	return &Adapter{
		pool:    pool,
		queries: gen.New(pool),
	}
}

// ListTasks returns all tasks owned by userID from the database, mapped to
// core domain types. It guarantees:
//   - Only rows belonging to userID are returned (enforced in SQL).
//   - An empty (non-nil) slice is returned when there are no rows.
//   - A row whose state is not in the known set returns an explicit error
//     (wrapping understanding.ErrUnknownTaskState) rather than a silent
//     default (FND-07).
//   - Database and query errors are wrapped with contextual information
//     so callers can use errors.Is / errors.As without losing the chain
//     (FND-03).
func (a *Adapter) ListTasks(ctx context.Context, userID ids.UserID) ([]understanding.Task, error) {
	rows, err := a.queries.ListTasksByUser(ctx, string(userID))
	if err != nil {
		return nil, fmt.Errorf("postgres adapter: ListTasks for user %q: %w", userID, err)
	}

	tasks := make([]understanding.Task, 0, len(rows))

	for _, row := range rows {
		t, err := mapTask(row)
		if err != nil {
			return nil, fmt.Errorf("postgres adapter: ListTasks for user %q: %w", userID, err)
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

// mapTask converts a sqlc-generated gen.Task (ACL boundary) into the core
// domain understanding.Task. All field translations are explicit so that a
// schema change in the generated code surfaces as a compile-time error here.
func mapTask(row gen.Task) (understanding.Task, error) {
	state, err := understanding.ParseTaskState(row.State)
	if err != nil {
		return understanding.Task{}, fmt.Errorf("mapping task %q: %w", row.ID, err)
	}

	attrs, err := decodeAttributes(row.Attributes)
	if err != nil {
		return understanding.Task{}, fmt.Errorf("mapping task %q attributes: %w", row.ID, err)
	}

	var budgetCap *int64
	if row.BudgetCap.Valid {
		v := row.BudgetCap.Int64
		budgetCap = &v
	}

	return understanding.Task{
		ID:          ids.TaskID(row.ID),
		UserID:      ids.UserID(row.UserID),
		ServiceType: row.ServiceType.String,
		Location:    row.Location.String,
		TimeWindow:  row.TimeWindow.String,
		BudgetCap:   budgetCap,
		Constraints: row.TaskConstraints.String,
		State:       state,
		Attributes:  attrs,
	}, nil
}

// decodeAttributes deserialises the raw JSONB bytes from Postgres into the
// open attribute map. Nil or empty bytes are treated as an empty map (no
// error), matching the spec (FND-07 / design.md error-handling table).
func decodeAttributes(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return make(map[string]any), nil
	}

	var attrs map[string]any
	if err := json.Unmarshal(raw, &attrs); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return attrs, nil
}
