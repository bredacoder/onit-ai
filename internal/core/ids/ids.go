// Package ids declares typed string identifiers used across onit's core domain.
// It imports nothing so that any package in the core tree can import it without
// creating an import cycle.
package ids

// UserID uniquely identifies a user (tenant).
type UserID string

// TaskID uniquely identifies a task aggregate.
type TaskID string

// NegotiationID uniquely identifies a negotiation aggregate.
type NegotiationID string

// ProviderID uniquely identifies a provider.
type ProviderID string

// MessageID uniquely identifies a message within a negotiation.
type MessageID string
