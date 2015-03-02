package backoff

import (
	"math"
	"time"
)

//Backoff is a time.Duration counter. It starts at Min.
//After every call to Duration() it is  multiplied by Factor.
//It is capped at Max. It returns to Min on every call to Reset().
//Used in conjunction with the time package.
type Backoff struct {
	//Factor is the multiplying factor for each increment step
	attempts, Factor int
	//Min and Max are the minimum and maximum values of the counter
	curr, Min, Max time.Duration
}

//Returns the current value of the counter and then
//multiplies it Factor
func (b *Backoff) Duration() time.Duration {
	//Zero-values are nonsensical, so we use
	//them to apply defaults
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

//Resets the current value of the counter back to Min
func (b *Backoff) Reset() {
	b.attempts = 0
}
