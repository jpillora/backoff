package backoff

import (
	"time"
)

// Retry extends Backoff to provide a simpler, higher-level API, for
// implementing exponential backoff/retry logic.
//
// It is not safe to use concurrently. Additionally, mutating the Backoff
// or Next fields directly may have unintended consequences, e.g. affecting the
// Last value.
type Retry struct {
	// Backoff is the internal [Backoff].
	Backoff *Backoff

	// Next is the next allowed time. As that suggests, it is set on
	// [Retry.Allow], each time that function returns a non-zero value.
	// Another attempt will be allowed if this is a time <= now, or the zero
	// value. Mutating this field directly may cause unexpected behavior,
	// use [Retry.Reset], instead.
	Next time.Time
}

var timeNow = time.Now // monkey patchable for testing

// Allow acts as a limiter, returning the next allowed time, or `time.Time{}`,
// if the next attempt should commence, inclusive of the first attempt.
// If `time.Time{}` is returned, [Backoff.Duration] will be called, which will
// increment the [Backoff.Attempt] count, and then used to update [Retry.Next].
func (x *Retry) Allow() time.Time {
	now := timeNow()
	if x.Next != (time.Time{}) && now.Before(x.Next) {
		return x.Next
	}
	x.Next = now.Add(x.Backoff.Duration())
	return time.Time{}
}

// Reset resets the attempt count, and clears the next allowed time, removing
// any applied limit.
// This method will typically be called after each success.
func (x *Retry) Reset() {
	x.Backoff.Reset()
	x.Next = time.Time{}
}

// Last is the last time Allow was called.
//
// WARNING: Mutating the Backoff directly may cause this value to be incorrect.
func (x *Retry) Last() time.Time {
	if x.Next == (time.Time{}) {
		return time.Time{}
	}
	return x.Next.Add(-x.Backoff.ForAttempt(x.Backoff.Attempt() - 1))
}
