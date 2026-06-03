package inmem_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bredacoder/onit-ai/internal/adapters/inmem"
	"github.com/bredacoder/onit-ai/internal/core/ids"
	"github.com/bredacoder/onit-ai/internal/core/memory"
)

// TestMemory_RememberAndRecall verifies that Remember stores records and Recall
// retrieves them for the exact (userID, layer) combination provided.
func TestMemory_RememberAndRecall(t *testing.T) {
	t.Parallel()

	const userA ids.UserID = "user-a"
	const layerFactual memory.Layer = memory.LayerFactual

	rec1 := memory.Record{Content: "fact 1"}
	rec2 := memory.Record{Content: "fact 2"}

	m := inmem.NewMemory()
	ctx := context.Background()

	// Remember two records.
	err := m.Remember(ctx, userA, layerFactual, rec1)
	require.NoError(t, err)

	err = m.Remember(ctx, userA, layerFactual, rec2)
	require.NoError(t, err)

	// Recall both records.
	got, err := m.Recall(ctx, userA, layerFactual, "")
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, rec1, got[0])
	require.Equal(t, rec2, got[1])
}

// TestMemory_UserIDIsolation verifies that Recall for user B never returns
// records stored for user A (multi-tenant boundary).
func TestMemory_UserIDIsolation(t *testing.T) {
	t.Parallel()

	const userA ids.UserID = "user-a"
	const userB ids.UserID = "user-b"
	const layer memory.Layer = memory.LayerFactual

	m := inmem.NewMemory()
	ctx := context.Background()

	// Store a record for user A.
	recA := memory.Record{Content: "user A's fact"}
	err := m.Remember(ctx, userA, layer, recA)
	require.NoError(t, err)

	// Recall for user B must not see user A's record.
	gotB, err := m.Recall(ctx, userB, layer, "")
	require.NoError(t, err)
	require.Nil(t, gotB, "Recall for user B with no records stored must return nil")

	// Recall for user A must see their own record.
	gotA, err := m.Recall(ctx, userA, layer, "")
	require.NoError(t, err)
	require.Len(t, gotA, 1)
	require.Equal(t, recA, gotA[0])
}

// TestMemory_LayerIsolation verifies that Recall respects layer boundaries:
// records stored in one layer are not returned when recalling from a different layer.
func TestMemory_LayerIsolation(t *testing.T) {
	t.Parallel()

	const user ids.UserID = "user-1"

	m := inmem.NewMemory()
	ctx := context.Background()

	// Store in LayerFactual.
	recFactual := memory.Record{Content: "a fact"}
	err := m.Remember(ctx, user, memory.LayerFactual, recFactual)
	require.NoError(t, err)

	// Recall from LayerBehavioral must not see the factual record.
	gotBehavioral, err := m.Recall(ctx, user, memory.LayerBehavioral, "")
	require.NoError(t, err)
	require.Nil(t, gotBehavioral, "Recall from a different layer must not return records from another layer")

	// Recall from LayerFactual must see the factual record.
	gotFactual, err := m.Recall(ctx, user, memory.LayerFactual, "")
	require.NoError(t, err)
	require.Len(t, gotFactual, 1)
	require.Equal(t, recFactual, gotFactual[0])
}
