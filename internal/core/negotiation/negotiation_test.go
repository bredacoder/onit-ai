package negotiation_test

import (
	"errors"
	"testing"

	"github.com/bredacoder/onit-ai/internal/core/negotiation"
)

func TestParseNegotiationState_ValidStates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  negotiation.NegotiationState
	}{
		{"draft", negotiation.StateDraft},
		{"awaiting_response", negotiation.StateAwaitingResponse},
		{"counteroffer", negotiation.StateCounteroffer},
		{"human_approval", negotiation.StateHumanApproval},
		{"confirmed", negotiation.StateConfirmed},
		{"declined", negotiation.StateDeclined},
		{"expired", negotiation.StateExpired},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got, err := negotiation.ParseNegotiationState(tc.input)
			if err != nil {
				t.Fatalf("ParseNegotiationState(%q) returned unexpected error: %v", tc.input, err)
			}

			if got != tc.want {
				t.Errorf("ParseNegotiationState(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseNegotiationState_UnknownState(t *testing.T) {
	t.Parallel()

	unknowns := []string{"", "DRAFT", "Draft", "pending", "unknown", "open"}

	for _, s := range unknowns {
		t.Run(s, func(t *testing.T) {
			t.Parallel()

			_, err := negotiation.ParseNegotiationState(s)
			if err == nil {
				t.Fatalf("ParseNegotiationState(%q) expected an error, got nil", s)
			}

			if !errors.Is(err, negotiation.ErrUnknownNegotiationState) {
				t.Errorf("ParseNegotiationState(%q) error = %v; want to wrap ErrUnknownNegotiationState", s, err)
			}
		})
	}
}
