package inmem

import (
	"context"
	"time"

	"github.com/bredacoder/onit-ai/internal/core/ids"
	"github.com/bredacoder/onit-ai/internal/core/scheduling"
)

// Compile-time assertion: Calendar must satisfy scheduling.CalendarPort.
var _ scheduling.CalendarPort = (*Calendar)(nil)

// Calendar is an in-memory fake for scheduling.CalendarPort. It returns a fixed
// slice of busy slots on every FreeBusy call. Slots may be configured before
// use; the zero value returns an empty result (window entirely free).
type Calendar struct {
	Slots []scheduling.Slot
}

// NewCalendar returns a Calendar fake that always returns slots on FreeBusy.
func NewCalendar(slots []scheduling.Slot) *Calendar {
	return &Calendar{Slots: slots}
}

// FreeBusy returns the canned Slots regardless of userID or the time window.
func (c *Calendar) FreeBusy(_ context.Context, _ ids.UserID, _, _ time.Time) ([]scheduling.Slot, error) {
	return c.Slots, nil
}
