package backoff

import (
	"testing"
	"time"
)

func Test1(t *testing.T) {

	b := &Backoff{
		Min:    100 * time.Millisecond,
		Max:    10 * time.Second,
		Factor: 2,
	}

	equals(t, b.Duration(), 100*time.Millisecond)
	equals(t, b.Duration(), 200*time.Millisecond)
	equals(t, b.Duration(), 400*time.Millisecond)
	b.Reset()
	equals(t, b.Duration(), 100*time.Millisecond)
}

func Test2(t *testing.T) {

	b := &Backoff{
		Min:    100 * time.Millisecond,
		Max:    10 * time.Second,
		Factor: 1.5,
	}

	equals(t, b.Duration(), 100*time.Millisecond)
	equals(t, b.Duration(), 150*time.Millisecond)
	equals(t, b.Duration(), 225*time.Millisecond)
	b.Reset()
	equals(t, b.Duration(), 100*time.Millisecond)
}

func Test3(t *testing.T) {

	b := &Backoff{
		Min:    100 * time.Nanosecond,
		Max:    10 * time.Second,
		Factor: 1.75,
	}

	equals(t, b.Duration(), 100*time.Nanosecond)
	equals(t, b.Duration(), 175*time.Nanosecond)
	equals(t, b.Duration(), 306*time.Nanosecond)
	b.Reset()
	equals(t, b.Duration(), 100*time.Nanosecond)
}

func TestJitter(t *testing.T) {
	b := &Backoff{
		Min:    100 * time.Millisecond,
		Max:    10 * time.Second,
		Factor: 2,
		Jitter: true,
	}

	equals(t, b.Duration(), 100*time.Millisecond)
	between(t, b.Duration(), 100*time.Millisecond, 200*time.Millisecond)
	between(t, b.Duration(), 100*time.Millisecond, 400*time.Millisecond)
	b.Reset()
	equals(t, b.Duration(), 100*time.Millisecond)
}

func between(t *testing.T, actual, low, high time.Duration) {
	if actual < low {
		t.Fatalf("Got %s, Expecting >= %s", actual, low)
	}
	if actual > high {
		t.Fatalf("Got %s, Expecting <= %s", actual, high)
	}
}

func equals(t *testing.T, d1, d2 time.Duration) {
	if d1 != d2 {
		t.Fatalf("Got %s, Expecting %s", d1, d2)
	}
}

func notequals(t *testing.T, d1, d2 time.Duration) {
	if d1 == d2 {
		t.Fatalf("Got %s, Expecting %s", d1, d2)
	}
}

func Test4(t *testing.T) {

	b := &Backoff{
		Min:    0,
		Max:    0,
		Factor: 0,
	}

	equals(t, b.Duration(), 100*time.Millisecond)
	equals(t, b.Duration(), 200*time.Millisecond)
	equals(t, b.Duration(), 400*time.Millisecond)
	b.Reset()
	equals(t, b.Duration(), 100*time.Millisecond)
}

func Test5(t *testing.T) {

	b := &Backoff{
		Min:    100 * time.Millisecond,
		Max:    200 * time.Millisecond,
		Factor: 2,
	}

	equals(t, b.Duration(), 100*time.Millisecond)
	equals(t, b.Duration(), 200*time.Millisecond)
	equals(t, b.Duration(), 200*time.Millisecond)
	b.Reset()
	equals(t, b.Duration(), 100*time.Millisecond)
}

func Test6(t *testing.T) {

	b := &Backoff{
		Min:    100 * time.Nanosecond,
		Max:    10 * time.Second,
		Factor: 1.75,
	}

	equals(t, b.DurationForAttempt(0), 100*time.Nanosecond)
	equals(t, b.DurationForAttempt(1), 175*time.Nanosecond)
	equals(t, b.DurationForAttempt(2), 306*time.Nanosecond)
	b.Reset()
	equals(t, b.DurationForAttempt(0), 100*time.Nanosecond)
}

func Test7(t *testing.T) {
	var vals [5]time.Duration
	for i := 0; i < 5; i++ {
		b := &Backoff{
			Min:    100 * time.Nanosecond,
			Max:    10 * time.Second,
			Factor: 1.75,
			Jitter: true,
		}
		vals[i] = b.DurationForAttempt(0)
	}

	notequals(t, vals[0], vals[1])
	notequals(t, vals[1], vals[2])
	notequals(t, vals[2], vals[3])
	notequals(t, vals[3], vals[4])
}
