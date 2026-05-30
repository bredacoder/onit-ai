package understanding

import (
	"errors"
	"testing"
)

func TestParseTaskState_ValidStates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  TaskState
	}{
		{"created", TaskCreated},
		{"understood", TaskUnderstood},
		{"acting", TaskActing},
		{"awaiting_approval", TaskAwaitingApproval},
		{"confirmed", TaskConfirmed},
		{"cancelled", TaskCancelled},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got, err := ParseTaskState(tc.input)
			if err != nil {
				t.Fatalf("ParseTaskState(%q) returned unexpected error: %v", tc.input, err)
			}

			if got != tc.want {
				t.Errorf("ParseTaskState(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseTaskState_UnknownReturnsError(t *testing.T) {
	t.Parallel()

	unknowns := []string{"", "pending", "CREATED", "Created", "unknown"}

	for _, s := range unknowns {
		t.Run(s, func(t *testing.T) {
			t.Parallel()

			_, err := ParseTaskState(s)
			if err == nil {
				t.Fatalf("ParseTaskState(%q) expected an error, got nil", s)
			}

			if !errors.Is(err, ErrUnknownTaskState) {
				t.Errorf("ParseTaskState(%q) error = %v; want errors.Is(err, ErrUnknownTaskState) == true", s, err)
			}
		})
	}
}

func TestTask_Attribute_NilAttributeMap(t *testing.T) {
	t.Parallel()

	// A zero-value Task has a nil Attributes map. Attribute() must not panic
	// and must return (nil, false) for any key.
	task := Task{}

	got, ok := task.Attribute("any-key")
	if ok {
		t.Errorf("Attribute() on nil map returned ok=true; want false")
	}

	if got != nil {
		t.Errorf("Attribute() on nil map returned value=%v; want nil", got)
	}
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
	if !ok {
		t.Fatal("Attribute(\"preferred_hours\") returned ok=false; want true")
	}

	if got != "morning" {
		t.Errorf("Attribute(\"preferred_hours\") = %v; want \"morning\"", got)
	}

	_, ok = task.Attribute("nonexistent")
	if ok {
		t.Error("Attribute(\"nonexistent\") returned ok=true; want false")
	}
}
