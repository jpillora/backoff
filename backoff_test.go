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

	if b.Duration() != 100*time.Millisecond {
		t.Error("Should be 100ms")
	}

	if b.Duration() != 200*time.Millisecond {
		t.Error("Should be 200ms")
	}

	if b.Duration() != 400*time.Millisecond {
		t.Error("Should be 400ms")
	}

	b.Reset()

	if b.Duration() != 100*time.Millisecond {
		t.Error("Should be 100ms again")
	}
}
