package main

import (
	"sync"
	"time"
)

// FixedWindowLimiter limits requests to a fixed number per time window.
//
// How it works:
// - Time is divided into fixed windows (e.g., every 60 seconds).
// - Each window has a counter that tracks how many requests have been made.
// - If the counter exceeds the limit, requests are rejected until the next window starts.
// - When a new window begins, the counter resets to zero.
//
// Example: limit = 5, window = 60s
//   Window 1 [0:00 - 1:00] → allows 5 requests, rejects the rest
//   Window 2 [1:00 - 2:00] → counter resets, allows 5 again
//
// Trade-off (burst at edges):
//   A user could send 5 requests at 0:59 and 5 more at 1:01,
//   getting 10 requests in 2 seconds. This is the known weakness
//   of fixed window — sliding window fixes this but is more complex.

type FixedWindowLimiter struct {
	mu          sync.Mutex    // protects concurrent access to fields below
	limit       int           // max allowed requests per window
	window      time.Duration // length of each time window
	count       int           // how many requests have been made in the current window
	windowStart time.Time     // when the current window began
}

// New creates a limiter that allows `limit` requests per `window` duration.
func New(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		limit:       limit,
		window:      window,
		count:       0,
		windowStart: time.Now(), // first window starts now
	}
}

// Allow checks if a request should be permitted.
// Returns true if under the limit, false if rate limited.
func (f *FixedWindowLimiter) Allow() bool {
	f.mu.Lock()         // lock so only one goroutine modifies state at a time
	defer f.mu.Unlock() // unlock when this function returns

	now := time.Now()

	// Check if we've moved past the current window.
	// If so, reset the counter and start a new window.
	if now.Sub(f.windowStart) >= f.window {
		f.count = 0         // reset the request counter
		f.windowStart = now // new window begins now
	}

	// Check if we're still under the limit
	if f.count < f.limit {
		f.count++   // count this request
		return true // allowed
	}

	// Over the limit — reject
	return false
}
