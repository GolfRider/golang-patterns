package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Circuit Breaker States:
//
//   CLOSED (normal) ──── failures hit threshold count────► OPEN (reject all)
//      ▲                                                  │
//      │                                            wait cooldown time
//      │                                                  │
//      └──── success ◄──── HALF-OPEN (allow 1 test) ◄────┘
//            resets         if test fails → back to OPEN

type State int

const (
	Closed   State = iota // normal, requests flow through
	Open                  // broken, reject everything
	HalfOpen              // testing, allow one request
)

func (s State) String() string {
	return [...]string{"CLOSED", "OPEN", "HALF-OPEN"}[s]
}

type CircuitBreaker struct {
	mu          sync.Mutex
	state       State
	failures    int
	maxFailures int           // failures before opening
	cooldown    time.Duration // how long to wait before trying again
	openedAt    time.Time     // when we entered OPEN state
}

func NewCircuitBreaker(maxFailures int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       Closed,
		maxFailures: maxFailures,
		cooldown:    cooldown,
	}
}

// Call wraps your function with circuit breaker protection.
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	switch cb.state {
	case Open:
		// Check if cooldown has passed
		if time.Since(cb.openedAt) > cb.cooldown {
			// Give it one more chance
			cb.state = HalfOpen
			fmt.Println("  [CB] OPEN → HALF-OPEN: testing...")
		} else {
			cb.mu.Unlock()
			return errors.New("circuit is OPEN, request rejected")
		}
	}

	cb.mu.Unlock()

	// Actually call the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		if cb.state == HalfOpen {
			// Test failed, back to open
			cb.state = Open
			cb.openedAt = time.Now()
			fmt.Println("  [CB] HALF-OPEN → OPEN: test failed")
		} else if cb.failures >= cb.maxFailures {
			cb.state = Open
			cb.openedAt = time.Now()
			fmt.Printf("  [CB] CLOSED → OPEN: %d failures\n", cb.failures)
		}
		return err
	}

	// Success!
	if cb.state == HalfOpen {
		fmt.Println("  [CB] HALF-OPEN → CLOSED: recovered!")
	}
	cb.state = Closed
	cb.failures = 0
	return nil
}

// -------- Retry with Circuit Breaker --------

func CallWithRetry(cb *CircuitBreaker, fn func() error, maxRetries int, delay time.Duration) error {
	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err = cb.Call(fn)
		if err == nil {
			return nil
		}
		fmt.Printf("  attempt %d failed: %v\n", attempt+1, err)
		if attempt < maxRetries {
			time.Sleep(delay)
			delay *= 2 // exponential backoff
		}
	}
	return fmt.Errorf("all %d retries failed: %w", maxRetries+1, err)
}

// -------- Demo --------

func problem9() {
	cb := NewCircuitBreaker(3, 2*time.Second)

	callCount := 0
	// Simulate: first 4 calls fail, then succeed
	unreliableService := func() error {
		callCount++
		if callCount <= 4 {
			return fmt.Errorf("service error")
		}
		fmt.Println("  service: OK!")
		return nil
	}

	// Phase 1: Failures trip the breaker
	fmt.Println("=== Phase 1: Failing requests ===")
	for i := 1; i <= 5; i++ {
		err := cb.Call(unreliableService)
		fmt.Printf("call %d: state=%s err=%v\n\n", i, cb.state, err)
	}

	// Phase 2: Wait for cooldown, then recover
	fmt.Println("=== Phase 2: Wait for cooldown ===")
	time.Sleep(3 * time.Second)
	err := cb.Call(unreliableService) // callCount=5, will succeed
	fmt.Printf("recovery call: state=%s err=%v\n\n", cb.state, err)

	// Phase 3: Retry with backoff
	fmt.Println("=== Phase 3: Retry with backoff ===")
	callCount = 0 // reset, first 4 fail again
	err = CallWithRetry(cb, unreliableService, 5, 500*time.Millisecond)
	fmt.Printf("final result: %v\n", err)
}

/*
=== Phase 1: Failing requests ===
call 1: state=CLOSED err=service error
call 2: state=CLOSED err=service error
[CB] CLOSED → OPEN: 3 failures
call 3: state=OPEN err=service error
call 4: state=OPEN err=circuit is OPEN, request rejected
call 5: state=OPEN err=circuit is OPEN, request rejected

=== Phase 2: Wait for cooldown ===
[CB] OPEN → HALF-OPEN: testing...
service: OK!
[CB] HALF-OPEN → CLOSED: recovered!
recovery call: state=CLOSED err=<nil>

=== Phase 3: Retry with backoff ===
attempt 1 failed: service error
attempt 2 failed: service error
[CB] CLOSED → OPEN: 3 failures
attempt 3 failed: service error
attempt 4 failed: circuit is OPEN, request rejected
attempt 5 failed: circuit is OPEN, request rejected
attempt 6 failed: circuit is OPEN, request rejected
final result: all 6 retries failed: circuit is OPEN, request rejected
```

---

## Mental model

Think of it like a **home electrical breaker**:
```
Normal:     electricity flows          → CLOSED
Overload:   breaker trips, cuts power  → OPEN
Test:       you flip it back on        → HALF-OPEN
if it trips again          → OPEN
if it holds                → CLOSED ✓


*/
