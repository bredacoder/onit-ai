// Package scheduling declares the CalendarPort for checking a user's calendar
// availability. The concrete adapter (gcal) lives in internal/adapters/gcal and
// implements this interface; the core never imports it directly.
package scheduling

import (
	"context"
	"time"

	"github.com/bredacoder/onit-ai/internal/core/ids"
)

// Slot is a contiguous busy or free time window on the user's calendar.
type Slot struct {
	Start time.Time
	End   time.Time
}

// CalendarPort is the port through which the core reads a user's calendar.
// Implementations must filter results by userID to uphold multi-tenant isolation.
type CalendarPort interface {
	// FreeBusy returns the busy slots for the given user within [from, to).
	// An empty slice means the window is entirely free.
	FreeBusy(ctx context.Context, userID ids.UserID, from, to time.Time) ([]Slot, error)
}
