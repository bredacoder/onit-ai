package understanding

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTaskState_ValidStates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		input string
		want  TaskState
	}{
		{"created", "created", TaskCreated},
		{"understood", "understood", TaskUnderstood},
		{"acting", "acting", TaskActing},
		{"awaiting_approval", "awaiting_approval", TaskAwaitingApproval},
		{"confirmed", "confirmed", TaskConfirmed},
		{"cancelled", "cancelled", TaskCancelled},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseTaskState(tc.input)
			require.NoError(t, err, "ParseTaskState(%q)", tc.input)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestParseTaskState_UnknownReturnsError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"invalid pending", "pending"},
		{"uppercase CREATED", "CREATED"},
		{"title case Created", "Created"},
		{"completely unknown", "unknown"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseTaskState(tc.input)
			require.Error(t, err, "ParseTaskState(%q) should return error", tc.input)
			require.ErrorIs(t, err, ErrUnknownTaskState)
		})
	}
}

func TestTask_Attribute_NilAttributeMap(t *testing.T) {
	t.Parallel()

	task := Task{}

	got, ok := task.Attribute("any-key")
	require.False(t, ok, "Attribute() on nil map should return ok=false")
	require.Nil(t, got)
}

func TestTask_Attribute_PopulatedMap(t *testing.T) {
	t.Parallel()

	task := Task{
		Attributes: map[string]any{
			"preferred_hours": "morning",
			"pets":            true,
		},
	}

	got, ok := task.Attribute("preferred_hours")
	require.True(t, ok)
	require.Equal(t, "morning", got)

	_, ok = task.Attribute("nonexistent")
	require.False(t, ok)
}
