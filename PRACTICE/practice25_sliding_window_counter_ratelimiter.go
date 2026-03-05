package main

import (
	"sync"
	"time"
)

// SlidingWindowLimiter limits requests using a sliding window approach.
//
// How it works:
// - Instead of hard-resetting at window boundaries (like fixed window),
//   we look at a "sliding" window that moves with the current time.
// - We blend the previous window's count with the current window's count
//   using a weight based on how far we are into the current window.
//
// Example: limit = 10, window = 60s, we're 15s into the current window
//
//   Previous window had 8 requests, current window has 3 so far.
//   We are 15/60 = 25% into the current window.
//   So 75% of the previous window still "overlaps" our sliding view.
//
//   Estimated count = (prev * overlap%) + current
//                   = (8    * 0.75)     + 3
//                   = 6 + 3 = 9   ← under 10, so ALLOW
//
//   |---- prev window ----|---- curr window ----|
//   |                  [=======sliding=======]  |
//                      ^75%          ^25%
//                      of prev       of curr
//
// Why this is better than fixed window:
//   Fixed window allows bursts at boundaries (e.g., 10 at 0:59 + 10 at 1:01).
//   Sliding window smooths this out by considering the overlap,
//   so those 10 requests at 0:59 still "count" when checking at 1:01.
//
// Trade-off:
//   This is an approximation — not an exact count of requests in the
//   last 60s. But it's very cheap (only 2 counters, no per-request storage)
//   and good enough for most real-world use cases.

type SlidingWindowLimiter struct {
	mu          sync.Mutex    // protects all fields below
	limit       int           // max requests allowed per window
	window      time.Duration // size of the window (e.g., 60s)
	prevCount   int           // total requests in the previous window
	currCount   int           // requests so far in the current window
	windowStart time.Time     // when the current window started
}

// New creates a limiter allowing `limit` requests per `window` duration.
func NewSWRL(limit int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		limit:       limit,
		window:      window,
		prevCount:   0,
		currCount:   0,
		windowStart: time.Now(),
	}
}

// Allow checks whether the next request should be permitted.
func (s *SlidingWindowLimiter) Allow() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(s.windowStart)

	// --- Window rotation ---
	// If we've moved past the current window, rotate:
	// current becomes previous, and we start a fresh current window.
	if elapsed >= s.window {
		// How many full windows have passed? Usually 1, but could be more
		// if no requests came in for a long time.
		windowsPassed := int(elapsed / s.window)

		if windowsPassed == 1 {
			// Normal case: just finished one window
			s.prevCount = s.currCount // current becomes previous
		} else {
			// Multiple windows passed with no activity — previous is empty
			s.prevCount = 0
		}

		s.currCount = 0
		// Advance windowStart by the number of full windows passed
		s.windowStart = s.windowStart.Add(time.Duration(windowsPassed) * s.window)
		elapsed = now.Sub(s.windowStart) // recalculate elapsed in new window
	}

	// --- Calculate the weighted estimate ---
	// How far are we into the current window? (0.0 to 1.0)
	progress := float64(elapsed) / float64(s.window)

	// The overlap of the previous window in our sliding view
	// e.g., if we're 25% into current, 75% of previous still overlaps
	prevWeight := 1.0 - progress

	// Estimated request count in the sliding window
	estimate := (float64(s.prevCount) * prevWeight) + float64(s.currCount)

	// --- Decision ---
	if estimate < float64(s.limit) {
		s.currCount++ // count this request
		return true   // allowed
	}

	return false // rate limited
}
