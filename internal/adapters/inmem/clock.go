package inmem

import (
	"time"

	"github.com/bredacoder/onit-ai/internal/core"
)

// Compile-time assertion: Clock must satisfy core.Clock.
var _ core.Clock = (*Clock)(nil)

// Clock is an in-memory fake for core.Clock. It returns a fixed point in time
// so that core logic that depends on the clock is deterministic in tests.
type Clock struct {
	fixed time.Time
}

// NewClock returns a Clock fake that always returns t from Now.
func NewClock(t time.Time) *Clock {
	return &Clock{fixed: t}
}

// Now returns the fixed time set at construction.
func (c *Clock) Now() time.Time {
	return c.fixed
}
