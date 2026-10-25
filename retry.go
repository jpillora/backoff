package backoff

import (
	"time"
)

// Retry extends [Backoff] to provide a simpler, higher-level API, for
// implementing exponential backoff/retry logic.
//
// It is not safe to use concurrently.
type Retry struct {
	// Backoff is the internal [Backoff].
	Backoff *Backoff

	// Next is the next allowed time. As that suggests, it is set on
	// [Retry.Allow], each time an attempt is allowed.
	// Another attempt will be allowed if Next is a time <= now, or the zero
	// value.
	Next time.Time
}

var timeNow = time.Now // monkey patchable for testing

// Allow acts as a limiter, returning the next allowed time, or `time.Time{}`,
// if the next attempt should commence, inclusive of the first attempt.
// If `time.Time{}` is returned, [Backoff.Duration] will be called, which will
// increment the [Backoff.Attempt] count. The value of [Retry.Next] will be
// updated to reflect the next allowed time (the returned duration, from now).
func (x *Retry) Allow() time.Time {
	now := timeNow()
	if x.Next != (time.Time{}) && now.Before(x.Next) {
		return x.Next
	}
	x.Next = now.Add(x.Backoff.Duration())
	return time.Time{}
}

// Reset resets the attempt count and clears the next allowed time, thereby
// removing any applied backoff delay. This method is typically invoked after
// a successful operation. By doing so, it ensures that if a subsequent
// failure occurs, the backoff timer restarts from the minimum duration. This
// approach is useful for scenarios where intermittent issues are resolved,
// allowing the system to promptly react to new errors without being delayed
// by the increased backoff time accumulated from previous failures.
//
// Clearing the Next field directly is an alternative, that will allow the next
// attempt immediately, without resetting the attempt count.
func (x *Retry) Reset() {
	x.Backoff.Reset()
	x.Next = time.Time{}
}
