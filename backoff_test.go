package backoff

import (
	"testing"
	"time"
)

func TestMaxAttempts(t *testing.T) {

	b := &Backoff{
		Min:         100 * time.Millisecond,
		Max:         10 * time.Second,
		Factor:      2,
		MaxAttempts: 2,
	}

	checkDuration(t, b, 100*time.Millisecond)
	checkDuration(t, b, 200*time.Millisecond)
	checkError(t, b, ErrMaxAttemptsExceeded)
}

func Test1(t *testing.T) {

	b := &Backoff{
		Min:    100 * time.Millisecond,
		Max:    10 * time.Second,
		Factor: 2,
	}

	checkDuration(t, b, 100*time.Millisecond)
	checkDuration(t, b, 200*time.Millisecond)
	checkDuration(t, b, 400*time.Millisecond)
	b.Reset()
	checkDuration(t, b, 100*time.Millisecond)
}

func Test2(t *testing.T) {

	b := &Backoff{
		Min:    100 * time.Millisecond,
		Max:    10 * time.Second,
		Factor: 1.5,
	}

	checkDuration(t, b, 100*time.Millisecond)
	checkDuration(t, b, 150*time.Millisecond)
	checkDuration(t, b, 225*time.Millisecond)
	b.Reset()
	checkDuration(t, b, 100*time.Millisecond)
}

func Test3(t *testing.T) {

	b := &Backoff{
		Min:    100 * time.Nanosecond,
		Max:    10 * time.Second,
		Factor: 1.75,
	}

	checkDuration(t, b, 100*time.Nanosecond)
	checkDuration(t, b, 175*time.Nanosecond)
	checkDuration(t, b, 306*time.Nanosecond)
	b.Reset()
	checkDuration(t, b, 100*time.Nanosecond)
}

func TestJitter(t *testing.T) {
	b := &Backoff{
		Min:    100 * time.Millisecond,
		Max:    10 * time.Second,
		Factor: 2,
		Jitter: true,
	}

	checkDuration(t, b, 100*time.Millisecond)
	checkDurationJitter(t, b, 100*time.Millisecond, 200*time.Millisecond)
	checkDurationJitter(t, b, 100*time.Millisecond, 400*time.Millisecond)
	b.Reset()
	checkDuration(t, b, 100*time.Millisecond)
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

func checkDurationJitter(t *testing.T, b *Backoff, min, max time.Duration) {
	d, err := b.Duration()
	between(t, d, min, max)
	if err != nil {
		t.Fatal("Error returned")
	}
}

func checkDuration(t *testing.T, b *Backoff, expected time.Duration) {
	d, err := b.Duration()
	equals(t, d, expected)
	if err != nil {
		t.Fatal("Error returned")
	}
}

func checkError(t *testing.T, b *Backoff, expected error) {
	d, err := b.Duration()
	equals(t, d, 0)
	if err != expected {
		t.Fatalf("Got %s, Expecting %v", err, expected)
	}
}
