package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bredacoder/onit-ai/internal/adapters/inmem"
	"github.com/bredacoder/onit-ai/internal/core/ids"
)

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
