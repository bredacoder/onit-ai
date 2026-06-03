package negotiation

import (
	"time"

	"github.com/bredacoder/onit-ai/internal/core/ids"
)

// Direction indicates whether a message was sent to or received from a provider.
type Direction string

const (
	// DirectionOutbound means the message was sent by the agent to the provider.
	DirectionOutbound Direction = "outbound"

	// DirectionInbound means the message was received from the provider.
	DirectionInbound Direction = "inbound"
)

// Message is a single turn in the conversation thread of a Negotiation.
type Message struct {
	ID            ids.MessageID
	UserID        ids.UserID
	NegotiationID ids.NegotiationID
	Direction     Direction
	Body          string
	SentAt        time.Time
}
