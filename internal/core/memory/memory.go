// Package memory declares the Port for the three-layer agent memory described in
// PRD §15.4. The three layers are:
//
//   - Factual    — observed domain facts: provider history, prices seen, episode log.
//   - Behavioral — per-user editable instructions loaded into the system prompt each turn.
//   - Procedural — per-category playbooks stored as parameterised data (not compiled code).
//
// The concrete adapter (pgvector) lives in internal/adapters/memory; the core never
// imports it directly. A trivial in-memory fake is used for offline core tests.
package memory

import (
	"context"
	"time"

	"github.com/bredacoder/onit-ai/internal/core/ids"
)

// Layer identifies one of the three memory layers.
type Layer string

const (
	// LayerFactual holds observed domain facts (episodic, provider, price history).
	LayerFactual Layer = "factual"
	// LayerBehavioral holds per-user editable behavioral instructions.
	LayerBehavioral Layer = "behavioral"
	// LayerProcedural holds per-category service playbooks stored as data.
	LayerProcedural Layer = "procedural"
)

// Record is a single memory entry within a layer.
type Record struct {
	// ID is an opaque identifier assigned by the store on Remember.
	ID string
	// Content is the natural-language or structured text of the memory entry.
	Content string
	// CreatedAt is when the record was first stored.
	CreatedAt time.Time
}

// Port is the memory port through which the core reads and writes layered memory.
// All operations are scoped to a single user to uphold multi-tenant isolation.
type Port interface {
	// Recall retrieves records from the given layer that are semantically relevant
	// to query. Implementations may use vector similarity or keyword search.
	Recall(ctx context.Context, userID ids.UserID, layer Layer, query string) ([]Record, error)

	// Remember persists a new record in the given layer for the user.
	Remember(ctx context.Context, userID ids.UserID, layer Layer, rec Record) error
}
