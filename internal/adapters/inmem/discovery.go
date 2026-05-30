package inmem

import (
	"context"

	"github.com/bredacoder/onit-ai/internal/core/discovery"
)

// Compile-time assertion: Discovery must satisfy discovery.Port.
var _ discovery.Port = (*Discovery)(nil)

// Discovery is an in-memory fake for discovery.Port. It returns a fixed slice
// of providers on every Find call. Providers may be configured before use; the
// zero value returns an empty result.
type Discovery struct {
	Providers []discovery.Provider
}

// NewDiscovery returns a Discovery fake that always returns providers on Find.
func NewDiscovery(providers []discovery.Provider) *Discovery {
	return &Discovery{Providers: providers}
}

// Find returns the canned Providers slice regardless of the query.
func (d *Discovery) Find(_ context.Context, _ discovery.Query) ([]discovery.Provider, error) {
	return d.Providers, nil
}
