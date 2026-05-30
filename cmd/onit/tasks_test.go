//go:build integration

package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	pgadapter "github.com/bredacoder/onit-ai/internal/adapters/postgres"
	"github.com/bredacoder/onit-ai/internal/adapters/inmem"
	"github.com/bredacoder/onit-ai/internal/core"
	"github.com/bredacoder/onit-ai/internal/core/ids"
)

// mustPool returns a pgxpool connected to DATABASE_URL, or skips the test if
// that env var is not set. All e2e tests in this file call this guard so that
// `go test ./internal/core/...` (offline, no DATABASE_URL) is unaffected.
func mustPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("e2e test requires DATABASE_URL")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err, "pgxpool.New")

	t.Cleanup(pool.Close)

	return pool
}

// insertUser creates a minimal user row so tasks can reference it via FK.
func insertUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	t.Helper()

	_, err := pool.Exec(
		context.Background(),
		`INSERT INTO users (id) VALUES ($1) ON CONFLICT DO NOTHING`,
		userID,
	)
	require.NoError(t, err, "insert user %q", userID)
}

// insertTask inserts a task row directly.
func insertTask(
	t *testing.T,
	pool *pgxpool.Pool,
	taskID, userID, serviceType, state string,
) {
	t.Helper()

	_, err := pool.Exec(
		context.Background(),
		`INSERT INTO tasks (id, user_id, service_type, state, attributes)
		 VALUES ($1, $2, $3, $4, '{}')`,
		taskID, userID, serviceType, state,
	)
	require.NoError(t, err, "insert task %q", taskID)
}

// cleanupTaskThenUser registers deferred DELETEs for task then user.
// Defer is LIFO, so user cleanup (registered first) runs last — satisfying FK.
func cleanupTaskThenUser(t *testing.T, pool *pgxpool.Pool, taskID, userID string) {
	t.Helper()

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM tasks WHERE id = $1`, taskID)
	})
}

// cleanupUser registers a deferred DELETE for a user row.
func cleanupUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	t.Helper()

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})
}

// runTasksCmd wires a root command with the given persistence layer and userID,
// sets args to "tasks", captures stdout and stderr into separate buffers,
// and returns (outBuf, errBuf, executeError).
func runTasksCmd(p core.Persistence, userID ids.UserID) (string, string, error) {
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	cmd := newRootCmd(p, userID)
	cmd.SetArgs([]string{"tasks"})
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()

	return outBuf.String(), errBuf.String(), err
}

// TestTasksCmd_EmptyState verifies that when the DB is migrated but no tasks
// exist for the current user, the command prints the empty-state message and
// exits with a nil error (acceptance criterion 1).
func TestTasksCmd_EmptyState(t *testing.T) {
	pool := mustPool(t)

	userID := "t15-e2e-empty-user"
	insertUser(t, pool, userID)
	cleanupUser(t, pool, userID)

	adapter := pgadapter.New(pool)

	out, _, err := runTasksCmd(adapter, ids.UserID(userID))

	require.NoError(t, err)
	require.Contains(t, out, "no tasks yet")
}

// TestTasksCmd_RowsFiltered verifies that:
//   - tasks for the current user are listed with id, service type, and state,
//   - a task belonging to a different user is NOT shown (multi-tenant boundary).
//
// Acceptance criterion 2.
func TestTasksCmd_RowsFiltered(t *testing.T) {
	pool := mustPool(t)

	userA := "t15-e2e-filter-a"
	userB := "t15-e2e-filter-b"
	taskA := "t15-task-filter-a"
	taskB := "t15-task-filter-b"

	insertUser(t, pool, userA)
	insertUser(t, pool, userB)
	cleanupTaskThenUser(t, pool, taskA, userA)
	cleanupTaskThenUser(t, pool, taskB, userB)

	insertTask(t, pool, taskA, userA, "plumbing", "created")
	insertTask(t, pool, taskB, userB, "cleaning", "created")

	adapter := pgadapter.New(pool)

	out, _, err := runTasksCmd(adapter, ids.UserID(userA))

	require.NoError(t, err)

	// userA's task must appear.
	require.Contains(t, out, taskA, "own task id must be in output")
	require.Contains(t, out, "plumbing", "own service type must be in output")
	require.Contains(t, out, "created", "task state must be in output")

	// userB's task must NOT appear.
	require.NotContains(t, out, taskB, "another user's task must not be listed")
	require.NotContains(t, out, "cleaning", "another user's service type must not be listed")
}

// TestTasksCmd_DBUnavailable verifies that when the adapter is backed by a
// closed (unavailable) pool the command returns an error, prints a clean error
// message, and does not produce a raw stacktrace or panic (acceptance
// criterion 3 / FND-03).
func TestTasksCmd_DBUnavailable(t *testing.T) {
	pool := mustPool(t)

	// Close the pool before use to simulate the DB being unavailable.
	pool.Close()

	adapter := pgadapter.New(pool)

	out, errOut, err := runTasksCmd(adapter, ids.UserID("any-user"))

	require.Error(t, err, "unavailable DB must return an error")

	// The combined output must not contain Go stacktrace markers.
	combined := out + errOut
	require.NotContains(t, combined, "goroutine ", "output must not contain a raw stacktrace")
	require.NotContains(t, combined, "panic:", "output must not contain a panic")

	// The error message itself must be a non-empty, human-readable string.
	require.NotEmpty(t, fmt.Sprintf("%v", err))
}

// TestTasksCmd_NoCurrentUser verifies that when userID is empty the command
// returns a clear error (edge case: no current user configured). No panic must
// occur (spec edge case + Decision B).
// This test runs offline using inmem.Persistence since the error path
// (empty userID check) never reaches the persistence layer.
func TestTasksCmd_NoCurrentUser(t *testing.T) {
	p := inmem.NewPersistence()

	_, _, err := runTasksCmd(p, ids.UserID(""))

	require.Error(t, err, "empty userID must return an error")
	require.ErrorIs(t, err, errNoCurrentUser)

	// Ensure the error message mentions how to fix it.
	require.Contains(
		t,
		strings.ToLower(err.Error()),
		"onit_user_id",
		"error must mention ONIT_USER_ID so the user knows what to set",
	)
}
