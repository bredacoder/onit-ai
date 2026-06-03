// Package identity holds the User value object for the onit core domain.
package identity

import "github.com/bredacoder/onit-ai/internal/core/ids"

// User represents an onit user.
// This is intentionally minimal for M0 (single-tenant, Wizard of Oz).
// Preferences, credentials, and notification settings will be added in later slices.
type User struct {
	ID ids.UserID
}
