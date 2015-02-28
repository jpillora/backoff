package backoff

import (
	"math"
	"time"
)

type Backoff struct {
	attempts, Factor int
	curr, Min, Max   time.Duration
}

func (b *Backoff) Duration() time.Duration {
	//abit hacky though, if zero-value, apply defaults
	if b.Min == 0 {
		b.Min = 100 * time.Millisecond
	}
	if b.Max == 0 {
		b.Max = 10 * time.Second
	}
	if b.Factor == 0 {
		b.Factor = 2
	}
	if b.curr == 0 {
		b.curr = b.Min
	}

	//calculate next duration in ms
	ms := float64(b.curr) * math.Pow(float64(b.Factor), float64(b.attempts))
	//bump attempts count
	b.attempts++
	//return as a time.Duration
	return time.Duration(math.Min(ms, float64(b.Max)))
}

func (b *Backoff) Reset() {
	b.attempts = 0
}
