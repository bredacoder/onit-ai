// Package negotiation contains the Negotiation aggregate and its FSM state type.
package negotiation

import (
	"errors"
	"fmt"

	"github.com/bredacoder/onit-ai/internal/core/ids"
)

// NegotiationState is the named type for the negotiation FSM states (PRD §10.2).
type NegotiationState string

const (
	// StateDraft is the initial state: a negotiation that has been created but not yet sent.
	StateDraft NegotiationState = "draft"

	// StateAwaitingResponse means the opening message has been sent and we are waiting for the provider to reply.
	StateAwaitingResponse NegotiationState = "awaiting_response"

	// StateCounteroffer means the provider replied with a counteroffer.
	StateCounteroffer NegotiationState = "counteroffer"

	// StateHumanApproval means the negotiation reached a point that requires the user's explicit approval.
	StateHumanApproval NegotiationState = "human_approval"

	// StateConfirmed means the user approved and the negotiation is complete.
	StateConfirmed NegotiationState = "confirmed"

	// StateDeclined means the negotiation was declined (by the provider or the user).
	StateDeclined NegotiationState = "declined"

	// StateExpired means the negotiation timed out without reaching a terminal outcome.
	StateExpired NegotiationState = "expired"
)

// ErrUnknownNegotiationState is returned by ParseNegotiationState when the
// supplied string does not match any known NegotiationState value.
var ErrUnknownNegotiationState = errors.New("negotiation: unknown state")

// ParseNegotiationState converts a raw string (e.g. read from the database) into
// a NegotiationState. It returns ErrUnknownNegotiationState for any unrecognised
// value so that the caller can never silently accept an invalid state.
func ParseNegotiationState(s string) (NegotiationState, error) {
	switch NegotiationState(s) {
	case StateDraft,
		StateAwaitingResponse,
		StateCounteroffer,
		StateHumanApproval,
		StateConfirmed,
		StateDeclined,
		StateExpired:
		return NegotiationState(s), nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownNegotiationState, s)
	}
}

// Negotiation is the aggregate root for a provider negotiation. Transition methods
// are declared in the Negotiation slice; this slice declares the type only.
type Negotiation struct {
	ID         ids.NegotiationID
	UserID     ids.UserID
	TaskID     ids.TaskID
	ProviderID ids.ProviderID
	State      NegotiationState
}
