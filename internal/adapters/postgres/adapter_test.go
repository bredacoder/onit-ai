package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	pgadapter "github.com/bredacoder/onit-ai/internal/adapters/postgres"
	"github.com/bredacoder/onit-ai/internal/core/ids"
	"github.com/bredacoder/onit-ai/internal/core/understanding"
)

// mustPool returns a connected pgxpool or skips the test if DATABASE_URL is
// not set. All integration tests call this guard first so that
// `go test ./internal/core/...` (no DATABASE_URL) never runs this file.
func mustPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("integration test requires DATABASE_URL")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err, "pgxpool.New")

	t.Cleanup(pool.Close)

	return pool
}

// insertUser inserts a minimal user row so tasks can reference it via FK.
func insertUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	t.Helper()

	_, err := pool.Exec(
		context.Background(),
		`INSERT INTO users (id) VALUES ($1) ON CONFLICT DO NOTHING`,
		userID,
	)
	require.NoError(t, err, "insert user")
}

// insertTask inserts a task row directly.
// Pass "" for attributesJSON to store the schema default (empty JSON object).
// budgetCap may be nil.
func insertTask(
	t *testing.T,
	pool *pgxpool.Pool,
	taskID, userID, serviceType, state string,
	budgetCap *int64,
	attributesJSON string,
) {
	t.Helper()

	// The attributes column is NOT NULL with DEFAULT '{}'. When attributesJSON
	// is empty we pass the literal empty object so the INSERT is accepted.
	attrs := attributesJSON
	if attrs == "" {
		attrs = "{}"
	}

	_, err := pool.Exec(
		context.Background(),
		`INSERT INTO tasks (id, user_id, service_type, state, budget_cap, attributes)
		 VALUES ($1, $2, $3, $4, $5, $6::jsonb)`,
		taskID,
		userID,
		serviceType,
		state,
		budgetCap,
		attrs,
	)
	require.NoError(t, err, "insert task")
}

// cleanupTaskThenUser registers deferred DELETE statements for a task and its
// owning user. Because defer is LIFO, calling this helper registers the user
// cleanup first (runs last) so the FK constraint is never violated.
func cleanupTaskThenUser(t *testing.T, pool *pgxpool.Pool, taskID, userID string) {
	t.Helper()

	// Register user cleanup first; it will run LAST (LIFO).
	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
		if err != nil {
			t.Logf("cleanup: delete user %q: %v", userID, err)
		}
	})

	// Register task cleanup second; it will run FIRST (LIFO).
	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), `DELETE FROM tasks WHERE id = $1`, taskID)
		if err != nil {
			t.Logf("cleanup: delete task %q: %v", taskID, err)
		}
	})
}

// TestListTasks_InsertAndRetrieve verifies that a task inserted for a user is
// returned by ListTasks with all mapped fields correct, including:
//   - non-nil attributes round-tripped through JSONB,
//   - nil budget_cap remaining nil in the domain type.
func TestListTasks_InsertAndRetrieve(t *testing.T) {
	pool := mustPool(t)

	userID := "test-user-t14-retrieve"
	taskID := "test-task-t14-retrieve"

	insertUser(t, pool, userID)

	cleanupTaskThenUser(t, pool, taskID, userID)

	insertTask(
		t, pool,
		taskID, userID,
		"plumbing",
		"created",
		nil,                                   // nil budget_cap
		`{"priority":"high","note":"urgent"}`, // attributes
	)

	adapter := pgadapter.New(pool)

	tasks, err := adapter.ListTasks(context.Background(), ids.UserID(userID))
	require.NoError(t, err)
	require.Len(t, tasks, 1)

	got := tasks[0]
	require.Equal(t, ids.TaskID(taskID), got.ID)
	require.Equal(t, ids.UserID(userID), got.UserID)
	require.Equal(t, "plumbing", got.ServiceType)
	require.Equal(t, understanding.TaskCreated, got.State)
	require.Nil(t, got.BudgetCap, "budget_cap must be nil when not set")

	// Attributes round-trip through JSONB.
	require.NotNil(t, got.Attributes)
	require.Equal(t, "high", got.Attributes["priority"])
	require.Equal(t, "urgent", got.Attributes["note"])
}

// TestListTasks_UserIDIsolation inserts tasks for two users and asserts that
// ListTasks for user A never returns user B's tasks (multi-tenant boundary).
func TestListTasks_UserIDIsolation(t *testing.T) {
	pool := mustPool(t)

	userA := "test-user-t14-isolation-a"
	userB := "test-user-t14-isolation-b"
	taskA := "test-task-t14-isolation-a"
	taskB := "test-task-t14-isolation-b"

	insertUser(t, pool, userA)
	insertUser(t, pool, userB)

	cleanupTaskThenUser(t, pool, taskA, userA)
	cleanupTaskThenUser(t, pool, taskB, userB)

	insertTask(t, pool, taskA, userA, "cleaning", "created", nil, "")
	insertTask(t, pool, taskB, userB, "painting", "created", nil, "")

	adapter := pgadapter.New(pool)
	ctx := context.Background()

	tasksA, err := adapter.ListTasks(ctx, ids.UserID(userA))
	require.NoError(t, err)
	require.Len(t, tasksA, 1, "user A must see exactly their own task")
	require.Equal(t, ids.TaskID(taskA), tasksA[0].ID)

	tasksB, err := adapter.ListTasks(ctx, ids.UserID(userB))
	require.NoError(t, err)
	require.Len(t, tasksB, 1, "user B must see exactly their own task")
	require.Equal(t, ids.TaskID(taskB), tasksB[0].ID)

	// Confirm user B's task never appears in user A's result set.
	for _, task := range tasksA {
		require.NotEqual(t, ids.TaskID(taskB), task.ID, "user B task leaked into user A results")
	}
}

// TestListTasks_UnknownState verifies that a row carrying an unrecognised state
// string causes ListTasks to return an explicit error wrapping
// understanding.ErrUnknownTaskState rather than silently using a default (FND-07).
func TestListTasks_UnknownState(t *testing.T) {
	pool := mustPool(t)

	userID := "test-user-t14-badstate"
	taskID := "test-task-t14-badstate"

	insertUser(t, pool, userID)

	cleanupTaskThenUser(t, pool, taskID, userID)

	insertTask(t, pool, taskID, userID, "gardening", "bogus_state", nil, "")

	adapter := pgadapter.New(pool)

	_, err := adapter.ListTasks(context.Background(), ids.UserID(userID))
	require.Error(t, err, "unknown state must return an error")
	require.True(
		t,
		errors.Is(err, understanding.ErrUnknownTaskState),
		"error must wrap understanding.ErrUnknownTaskState; got: %v", err,
	)
}

// TestListTasks_EmptyAttributes verifies that a row whose attributes column
// contains an empty JSON object '{}' is mapped to a non-nil, empty map
// (not nil, not a panic).
func TestListTasks_EmptyAttributes(t *testing.T) {
	pool := mustPool(t)

	userID := "test-user-t14-emptyattrs"
	taskID := "test-task-t14-emptyattrs"

	insertUser(t, pool, userID)

	cleanupTaskThenUser(t, pool, taskID, userID)

	// Pass explicit empty JSON object; schema default would also produce this.
	insertTask(t, pool, taskID, userID, "moving", "created", nil, "{}")

	adapter := pgadapter.New(pool)

	tasks, err := adapter.ListTasks(context.Background(), ids.UserID(userID))
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	require.NotNil(t, tasks[0].Attributes, "Attributes must be a non-nil empty map, not nil")
	require.Empty(t, tasks[0].Attributes)
}

// TestListTasks_ClosedPool verifies that a closed pool causes ListTasks to
// return a wrapped error and not panic (FND-03).
func TestListTasks_ClosedPool(t *testing.T) {
	pool := mustPool(t)

	// Close the pool immediately before using it.
	pool.Close()

	adapter := pgadapter.New(pool)

	_, err := adapter.ListTasks(context.Background(), ids.UserID(fmt.Sprintf("user-%d", 1)))
	require.Error(t, err, "closed pool must return an error")
	// The error must be wrapped — it must NOT be nil and must NOT panic.
}
