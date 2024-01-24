package backoff

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// This example demonstrates how to use [Retry] as a safety mechanism for
// calling an action that may fail, with exponential backoff, that persists
// across calls.
//
// Note that is a simple example, that does not consider concurrency.
func ExampleRetry_doSomethingWithRetry() {
	var doSomething func() error // the action being attempted

	// retry is reused, to maintain state across calls to `doSomethingWithRetry`
	retry := &Retry{Backoff: &Backoff{
		Min:    100 * time.Millisecond,
		Max:    400 * time.Millisecond,
		Factor: 2,
	}}

	// doSomethingWithRetry enforces a rate limit on calling `doSomething`, which
	// persists across calls to `doSomethingWithRetry`, resetting on success
	doSomethingWithRetry := func(maxAttempts int) (err error) {
		for attempt := 0; attempt < maxAttempts; {
			if next := retry.Allow(); next != (time.Time{}) {
				// if an attempt is not yet allowed, wait until it is
				atmpt := attempt
				if atmpt == 0 {
					atmpt = 1
				}
				fmt.Printf("attempt %d of %d: limited for %s\n", atmpt, maxAttempts, retry.Backoff.ForAttempt(retry.Backoff.Attempt()-1))
				time.Sleep(time.Until(next))
				continue
			}

			attempt++

			err = doSomething() // the action being attempted
			fmt.Printf("attempt %d of %d: doSomething(): %v\n", attempt, maxAttempts, err)
			if err != nil {
				// in practice you might have some logging here, or just handle the
				// returned (last) error
				continue
			}

			// success!

			// allow another call immediately, reset attempts
			retry.Reset()

			return nil
		}

		return err
	}

	// Test case: Success on the first attempt
	doSomething = func() error {
		return nil
	}
	err := doSomethingWithRetry(3)
	if err != nil {
		panic(fmt.Sprintf("Expected success on the first attempt, got error: %v", err))
	}

	// Test case: Success on the third attempt
	attempts := 0
	doSomething = func() error {
		attempts++
		if attempts < 3 {
			return errors.New("simulated error")
		}
		return nil
	}
	err = doSomethingWithRetry(3)
	if err != nil {
		panic(fmt.Sprintf("Expected success on the third attempt, got error: %v", err))
	}

	// Test case: Failure after all attempts
	doSomething = func() error {
		return errors.New("simulated error")
	}
	err = doSomethingWithRetry(3)
	if err == nil {
		panic(fmt.Sprintf("Expected error after all attempts, got success"))
	}

	// Test case: does not reset the backoff attempts
	err = doSomethingWithRetry(1)
	if err == nil {
		panic(fmt.Sprintf("Expected error after all attempts, got success"))
	}

	//output:
	//attempt 1 of 3: doSomething(): <nil>
	//attempt 1 of 3: doSomething(): simulated error
	//attempt 1 of 3: limited for 100ms
	//attempt 2 of 3: doSomething(): simulated error
	//attempt 2 of 3: limited for 200ms
	//attempt 3 of 3: doSomething(): <nil>
	//attempt 1 of 3: doSomething(): simulated error
	//attempt 1 of 3: limited for 100ms
	//attempt 2 of 3: doSomething(): simulated error
	//attempt 2 of 3: limited for 200ms
	//attempt 3 of 3: doSomething(): simulated error
	//attempt 1 of 1: limited for 400ms
	//attempt 1 of 1: doSomething(): simulated error
}

// This example demonstrates the correctness of the backoff behavior, by
// repeatedly calling [Retry.Allow].
func ExampleRetry_Allow_backoffBehavior() {
	defer func() { timeNow = time.Now }()

	retry := &Retry{Backoff: &Backoff{Min: time.Second, Max: time.Second * 60}}

	attemptAllowAt := func(at time.Duration) {
		currentAttempt := retry.Backoff.Attempt() - 1
		currentBackoff := retry.Backoff.ForAttempt(currentAttempt)
		var lastAttempt time.Time
		if retry.Next != (time.Time{}) {
			lastAttempt = retry.Next.Add(-retry.Backoff.ForAttempt(retry.Backoff.Attempt() - 1))
		}
		now := time.Unix(0, int64(at))
		timeNow = func() time.Time { return now }
		limitedUntil := retry.Allow()
		var s string
		if limitedUntil != (time.Time{}) {
			if limitedUntil != retry.Next {
				panic(`unexpected return value`)
			}
			s = `limited (next at ` + time.Duration(limitedUntil.UnixNano()).String() + `)`
		} else {
			s = `ok (next at ` + time.Duration(retry.Next.UnixNano()).String() + `)`
		}
		fmt.Printf("at %s last %s (%.0f, %s): %s\n", at, time.Duration(lastAttempt.UnixNano()), currentAttempt, currentBackoff, s)
	}

	fmt.Println("at OFFSET last LAST_OFFSET (CUR_ATTEMPT, CUR_BACKOFF): ALLOW_RESULT")
	attemptAllowAt(0)
	attemptAllowAt(time.Millisecond * 100)
	attemptAllowAt(time.Millisecond * 700)
	attemptAllowAt(time.Second)
	attemptAllowAt(time.Second)
	attemptAllowAt(time.Second * 2)
	attemptAllowAt(time.Second * 3)
	attemptAllowAt(time.Second * 3)
	attemptAllowAt(time.Second * 4)
	attemptAllowAt(time.Second * 5)
	attemptAllowAt(time.Second * 8)
	attemptAllowAt(time.Second * 20)
	attemptAllowAt(time.Second * 55)
	attemptAllowAt(time.Second * 80)
	attemptAllowAt(time.Second * 120)
	attemptAllowAt(time.Second * 500)
	attemptAllowAt(time.Second * 550)
	attemptAllowAt(time.Second * 560)

	//output:
	//at OFFSET last LAST_OFFSET (CUR_ATTEMPT, CUR_BACKOFF): ALLOW_RESULT
	//at 0s last -1887601h16m18.871345152s (-1, 1s): ok (next at 1s)
	//at 100ms last 0s (0, 1s): limited (next at 1s)
	//at 700ms last 0s (0, 1s): limited (next at 1s)
	//at 1s last 0s (0, 1s): ok (next at 3s)
	//at 1s last 1s (1, 2s): limited (next at 3s)
	//at 2s last 1s (1, 2s): limited (next at 3s)
	//at 3s last 1s (1, 2s): ok (next at 7s)
	//at 3s last 3s (2, 4s): limited (next at 7s)
	//at 4s last 3s (2, 4s): limited (next at 7s)
	//at 5s last 3s (2, 4s): limited (next at 7s)
	//at 8s last 3s (2, 4s): ok (next at 16s)
	//at 20s last 8s (3, 8s): ok (next at 36s)
	//at 55s last 20s (4, 16s): ok (next at 1m27s)
	//at 1m20s last 55s (5, 32s): limited (next at 1m27s)
	//at 2m0s last 55s (5, 32s): ok (next at 3m0s)
	//at 8m20s last 2m0s (6, 1m0s): ok (next at 9m20s)
	//at 9m10s last 8m20s (7, 1m0s): limited (next at 9m20s)
	//at 9m20s last 8m20s (7, 1m0s): ok (next at 10m20s)
}

func TestRetry_Allow(t *testing.T) {
	defer func() { timeNow = time.Now }()

	tests := []struct {
		name        string
		initial     Retry
		setTime     time.Time // Time to set as the current time
		advance     int       // Number of times to advance the internal backoff
		expected    time.Time // Expected result from Allow
		nextTime    time.Time // Expected time after Allow
		attempt     float64   // Expected attempt number after Allow
		shouldReset bool      // Indicates if Retry should be reset before test
	}{
		{
			name: "Initial allow call",
			initial: Retry{
				Backoff: &Backoff{
					Min:    100 * time.Millisecond,
					Max:    10 * time.Second,
					Factor: 2,
				},
			},
			setTime:     time.Unix(0, 0),
			advance:     0,
			expected:    time.Time{},
			nextTime:    time.Unix(0, 0).Add(100 * time.Millisecond),
			attempt:     1,
			shouldReset: false,
		},
		{
			name: "Allow call after duration elapsed",
			initial: Retry{
				Backoff: &Backoff{
					Min:    100 * time.Millisecond,
					Max:    10 * time.Second,
					Factor: 2,
				},
			},
			setTime:     time.Unix(0, 0).Add(200 * time.Millisecond),
			advance:     1,
			expected:    time.Time{},
			nextTime:    time.Unix(0, 0).Add(200 * time.Millisecond).Add(200 * time.Millisecond),
			attempt:     2,
			shouldReset: true,
		},
		{
			name: "Allow call before initial duration elapsed",
			initial: Retry{
				Backoff: &Backoff{
					Min:    100 * time.Millisecond,
					Max:    10 * time.Second,
					Factor: 2,
				},
				Next: time.Unix(0, 0).Add(100 * time.Millisecond), // Set initial time to a specific value
			},
			setTime:     time.Unix(0, 0).Add(50 * time.Millisecond),  // Set current time before the initial time elapses
			advance:     1,                                           // No advance since we are testing the initial state
			expected:    time.Unix(0, 0).Add(100 * time.Millisecond), // Expected to return the initial time since it has not elapsed
			nextTime:    time.Unix(0, 0).Add(100 * time.Millisecond), // The next time remains the same as initial time is not elapsed
			attempt:     1,                                           // Attempt remains 1 as we are within the initial duration
			shouldReset: false,
		},
		{
			name: "Multiple allow calls, no reset",
			initial: Retry{
				Backoff: &Backoff{
					Min:    100 * time.Millisecond,
					Max:    500 * time.Millisecond,
					Factor: 2,
				},
			},
			setTime:     time.Unix(0, 0).Add(250 * time.Millisecond),
			advance:     2,
			expected:    time.Time{},
			nextTime:    time.Unix(0, 0).Add(250 * time.Millisecond).Add(400 * time.Millisecond),
			attempt:     3,
			shouldReset: false,
		},
		{
			name: "Max duration exceeded",
			initial: Retry{
				Backoff: &Backoff{
					Min:    100 * time.Millisecond,
					Max:    300 * time.Millisecond,
					Factor: 2,
				},
			},
			setTime:     time.Unix(0, 0).Add(1000 * time.Millisecond),
			advance:     5,
			expected:    time.Time{},
			nextTime:    time.Unix(0, 0).Add(1000 * time.Millisecond).Add(300 * time.Millisecond),
			attempt:     6,
			shouldReset: true,
		},
		{
			name: "Reset before allow call",
			initial: Retry{
				Backoff: &Backoff{
					Min:    100 * time.Millisecond,
					Max:    10 * time.Second,
					Factor: 2,
				},
			},
			setTime:     time.Unix(0, 0).Add(200 * time.Millisecond),
			advance:     1,
			expected:    time.Time{},
			nextTime:    time.Unix(0, 0).Add(200 * time.Millisecond).Add(200 * time.Millisecond),
			attempt:     2,
			shouldReset: true,
		},
		{
			name: "No advance, with reset",
			initial: Retry{
				Backoff: &Backoff{
					Min:    100 * time.Millisecond,
					Max:    10 * time.Second,
					Factor: 2,
				},
			},
			setTime:     time.Unix(0, 0),
			advance:     0,
			expected:    time.Time{},
			nextTime:    time.Unix(0, 0).Add(100 * time.Millisecond),
			attempt:     1,
			shouldReset: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldReset {
				tc.initial.Reset()
			}
			// Set timeNow to a controlled time
			timeNow = func() time.Time {
				return tc.setTime
			}

			// Advance the internal backoff as needed
			for i := 0; i < tc.advance; i++ {
				tc.initial.Backoff.Duration()
			}

			result := tc.initial.Allow()
			if !timeEqualIncludingIfZero(result, tc.expected) {
				t.Errorf("%s: unexpected Allow result, got %v, want %v", tc.name, result, tc.expected)
			}
			if !timeEqualIncludingIfZero(tc.initial.Next, tc.nextTime) {
				t.Errorf("%s: unexpected next time, got %v, want %v", tc.name, tc.initial.Next, tc.nextTime)
			}
			if tc.initial.Backoff.Attempt() != tc.attempt {
				t.Errorf("%s: unexpected attempt number, got %f, want %f", tc.name, tc.initial.Backoff.Attempt(), tc.attempt)
			}
		})
	}
}

func TestRetry_Reset(t *testing.T) {
	defer func() { timeNow = time.Now }()

	initialTime := time.Unix(0, 0)
	timeNow = func() time.Time {
		return initialTime
	}

	x := &Retry{
		Backoff: &Backoff{
			Min:    100 * time.Millisecond,
			Max:    10 * time.Second,
			Factor: 2,
			Jitter: false,
		},
	}

	// Call Allow to set an initial time
	initialTime = timeNow()
	x.Allow()

	// Reset the backoff
	x.Reset()

	// Next time should be reset to zero
	nextTime := x.Allow()
	if nextTime != (time.Time{}) {
		t.Error("After reset, next time should be time.Time{}, got", nextTime)
	}
}

func timeEqualIncludingIfZero(a, b time.Time) bool {
	return a.Equal(b) && (a == (time.Time{})) == (b == time.Time{})
}
