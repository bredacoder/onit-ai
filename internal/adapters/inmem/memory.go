package inmem

import (
	"context"

	"github.com/bredacoder/onit-ai/internal/core/ids"
	coreMemory "github.com/bredacoder/onit-ai/internal/core/memory"
)

// Compile-time assertion: Memory must satisfy memory.Port.
var _ coreMemory.Port = (*Memory)(nil)

// Memory is an in-memory fake for memory.Port. Remember stores records in an
// in-memory map keyed by (userID, layer); Recall returns the stored records for
// that (userID, layer) combination. If a query filter is needed in the future it
// can be added without changing the interface.
type Memory struct {
	records map[memKey][]coreMemory.Record
}

type memKey struct {
	userID ids.UserID
	layer  coreMemory.Layer
}

// NewMemory returns a Memory fake with an empty backing store.
func NewMemory() *Memory {
	return &Memory{records: make(map[memKey][]coreMemory.Record)}
}

// Remember stores rec under the given userID and layer.
func (m *Memory) Remember(_ context.Context, userID ids.UserID, layer coreMemory.Layer, rec coreMemory.Record) error {
	k := memKey{userID: userID, layer: layer}
	m.records[k] = append(m.records[k], rec)

	return nil
}

// Recall returns all records stored for the given userID and layer. The query
// argument is ignored by this fake — all stored records are returned.
func (m *Memory) Recall(_ context.Context, userID ids.UserID, layer coreMemory.Layer, _ string) ([]coreMemory.Record, error) {
	k := memKey{userID: userID, layer: layer}

	return m.records[k], nil
}
