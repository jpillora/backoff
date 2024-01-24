# Backoff

A simple exponential backoff counter in Go (Golang)

[![GoDoc](https://godoc.org/github.com/jpillora/backoff?status.svg)](https://godoc.org/github.com/jpillora/backoff)
[![Build Status](https://github.com/jpillora/backoff/actions/workflows/build.yml/badge.svg)](https://github.com/jpillora/backoff/actions/workflows/build.yml)

### Install

```
$ go get -v github.com/jpillora/backoff
```

### Usage

Backoff is a `time.Duration` counter. It starts at `Min`. After every call to `Duration()` it is  multiplied by `Factor`. It is capped at `Max`. It returns to `Min` on every call to `Reset()`. `Jitter` adds randomness ([see below](#example-using-jitter)). Used in conjunction with the `time` package.

For a higher-level API, see [Retry](#example-using-retry).

#### Documentation

https://pkg.go.dev/github.com/jpillora/backoff

---

#### Simple example

``` go

b := &backoff.Backoff{
	//These are the defaults
	Min:    100 * time.Millisecond,
	Max:    10 * time.Second,
	Factor: 2,
	Jitter: false,
}

fmt.Printf("%s\n", b.Duration())
fmt.Printf("%s\n", b.Duration())
fmt.Printf("%s\n", b.Duration())

fmt.Printf("Reset!\n")
b.Reset()

fmt.Printf("%s\n", b.Duration())
```

```
100ms
200ms
400ms
Reset!
100ms
```

---

#### Example using `net` package

``` go
b := &backoff.Backoff{
    Max:    5 * time.Minute,
}

for {
	conn, err := net.Dial("tcp", "example.com:5309")
	if err != nil {
		d := b.Duration()
		fmt.Printf("%s, reconnecting in %s", err, d)
		time.Sleep(d)
		continue
	}
	//connected
	b.Reset()
	conn.Write([]byte("hello world!"))
	// ... Read ... Write ... etc
	conn.Close()
	//disconnected
}

```

---

#### Example using `Jitter`

Enabling `Jitter` adds some randomization to the backoff durations. [See Amazon's writeup of performance gains using jitter](http://www.awsarchitectureblog.com/2015/03/backoff.html). Seeding is not necessary but doing so gives repeatable results.

```go
import "math/rand"

b := &backoff.Backoff{
	Jitter: true,
}

rand.Seed(42)

fmt.Printf("%s\n", b.Duration())
fmt.Printf("%s\n", b.Duration())
fmt.Printf("%s\n", b.Duration())

fmt.Printf("Reset!\n")
b.Reset()

fmt.Printf("%s\n", b.Duration())
fmt.Printf("%s\n", b.Duration())
fmt.Printf("%s\n", b.Duration())
```

```
100ms
106.600049ms
281.228155ms
Reset!
100ms
104.381845ms
214.957989ms
```

---

#### Example using `Retry`

The `Retry` type extends `Backoff` to provide a simpler, higher-level API for implementing exponential backoff/retry logic.
Here's an example of how you can use it:

```go
// retry is reused, to maintain state across calls to `doSomethingWithRetry`
retry := &backoff.Retry{Backoff: &backoff.Backoff{
	Min:    100 * time.Millisecond,
	Max:    10 * time.Second,
	Factor: 2,
}}

// doSomethingWithRetry enforces a rate limit on calling `doSomething`, which
// persists across calls to `doSomethingWithRetry`, resetting on success
doSomethingWithRetry := func(maxAttempts int) (err error) {
	for attempt := 0; attempt < maxAttempts; {
		if next := retry.Allow(); next != (time.Time{}) {
			// if an attempt is not yet allowed, wait until it is
			time.Sleep(time.Until(next))
			continue
		}

		attempt++

		err = doSomething() // the action being attempted
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

// call `doSomethingWithRetry`, whenever you want to `doSomething`
```

---

#### Credits

Forked from [some JavaScript](https://github.com/segmentio/backo) written by [@tj](https://github.com/tj)
