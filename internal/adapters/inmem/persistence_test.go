package inmem_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bredacoder/onit-ai/internal/adapters/inmem"
	"github.com/bredacoder/onit-ai/internal/core/ids"
	"github.com/bredacoder/onit-ai/internal/core/understanding"
)

// TestPersistence_ListTasks_UserIDIsolation asserts that ListTasks never
// returns tasks belonging to a different user (the core multi-tenant invariant).
func TestPersistence_ListTasks_UserIDIsolation(t *testing.T) {
	t.Parallel()

	const userA ids.UserID = "user-a"
	const userB ids.UserID = "user-b"

	taskA1 := understanding.Task{ID: "task-a1", UserID: userA, ServiceType: "plumbing", State: understanding.TaskCreated}
	taskA2 := understanding.Task{ID: "task-a2", UserID: userA, ServiceType: "cleaning", State: understanding.TaskUnderstood}
	taskB1 := understanding.Task{ID: "task-b1", UserID: userB, ServiceType: "painting", State: understanding.TaskCreated}

	p := inmem.NewPersistence()
	p.AddTask(taskA1)
	p.AddTask(taskA2)
	p.AddTask(taskB1)

	ctx := context.Background()

	t.Run("user A sees only their own tasks", func(t *testing.T) {
		t.Parallel()

		got, err := p.ListTasks(ctx, userA)
		require.NoError(t, err)
		require.Len(t, got, 2)

		ids := make([]ids.TaskID, 0, len(got))
		for _, task := range got {
			ids = append(ids, task.ID)
		}

		require.Contains(t, ids, taskA1.ID)
		require.Contains(t, ids, taskA2.ID)

		// user B's task must never appear for user A.
		require.NotContains(t, ids, taskB1.ID)
	})

	t.Run("user B sees only their own tasks and never user A tasks", func(t *testing.T) {
		t.Parallel()

		got, err := p.ListTasks(ctx, userB)
		require.NoError(t, err)
		require.Len(t, got, 1)
		require.Equal(t, taskB1.ID, got[0].ID)
		require.Equal(t, userB, got[0].UserID)
	})
}

// TestPersistence_ListTasks_UnknownUserReturnsEmpty asserts that querying with a
// user ID that has no tasks returns an empty slice without error.
func TestPersistence_ListTasks_UnknownUserReturnsEmpty(t *testing.T) {
	t.Parallel()

	p := inmem.NewPersistence()
	p.AddTask(understanding.Task{ID: "task-1", UserID: "some-user", State: understanding.TaskCreated})

	ctx := context.Background()

	got, err := p.ListTasks(ctx, "unknown-user")
	require.NoError(t, err)
	require.NotNil(t, got, "ListTasks must return a non-nil empty slice, not nil")
	require.Empty(t, got)
}
