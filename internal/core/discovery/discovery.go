package discovery

import "context"

// Query describes what the caller is looking for.
// Kept minimal for M0; additional filters (radius, budget) will be added in
// later slices as the agent grows.
type Query struct {
	ServiceType string
	Location    string
}

// DiscoverySource is a single upstream provider of search results (e.g. Google
// Places, Yelp). Each source is responsible for populating provenance fields on
// the Provider it returns.
type DiscoverySource interface {
	// Name returns a stable identifier for this source, used to populate
	// Provider.Source.
	Name() string
	// Find returns providers that match q from this source.
	Find(ctx context.Context, q Query) ([]Provider, error)
}

// Port is the registry that fans a Query out across all registered
// DiscoverySources and returns the merged result set.
// Adapters implement this interface; the core declares it.
type Port interface {
	Find(ctx context.Context, q Query) ([]Provider, error)
}
