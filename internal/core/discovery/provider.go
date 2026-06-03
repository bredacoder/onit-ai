// Package discovery declares the Provider domain type and the ports for
// multi-source provider discovery (PRD §15.2).
package discovery

import "github.com/bredacoder/onit-ai/internal/core/ids"

// Provider is a local-service provider surfaced by the discovery subsystem.
//
// Provenance fields (Source, Confidence, Evidence) carry the origin of each
// record per PRD §15.2: the agent must never assert a provider fact it cannot
// attribute to a specific source with a measured confidence.
type Provider struct {
	ID     ids.ProviderID
	UserID ids.UserID
	Name   string
	Phone  string
	Rating float64
	Hours  string

	// Provenance — required by PRD §15.2.
	// Source identifies the discovery source that returned this provider
	// (e.g. "google_places", "yelp").
	Source string
	// Confidence is a [0, 1] score reflecting how certain the source is that
	// this provider matches the query.
	Confidence float64
	// Evidence is a human-readable summary of the signal that produced this
	// provider (e.g. review excerpts, matched keywords).
	Evidence string
}
