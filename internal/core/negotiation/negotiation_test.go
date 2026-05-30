package negotiation_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bredacoder/onit-ai/internal/core/negotiation"
)

func TestParseNegotiationState_ValidStates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  negotiation.NegotiationState
	}{
		{"draft", "draft", negotiation.StateDraft},
		{"awaiting_response", "awaiting_response", negotiation.StateAwaitingResponse},
		{"counteroffer", "counteroffer", negotiation.StateCounteroffer},
		{"human_approval", "human_approval", negotiation.StateHumanApproval},
		{"confirmed", "confirmed", negotiation.StateConfirmed},
		{"declined", "declined", negotiation.StateDeclined},
		{"expired", "expired", negotiation.StateExpired},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := negotiation.ParseNegotiationState(tc.input)
			require.NoError(t, err, "ParseNegotiationState(%q)", tc.input)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestParseNegotiationState_UnknownState(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"uppercase DRAFT", "DRAFT"},
		{"title case Draft", "Draft"},
		{"invalid pending", "pending"},
		{"completely unknown", "unknown"},
		{"invalid open", "open"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := negotiation.ParseNegotiationState(tc.input)
			require.Error(t, err, "ParseNegotiationState(%q) should return error", tc.input)
			require.ErrorIs(t, err, negotiation.ErrUnknownNegotiationState)
		})
	}
}
