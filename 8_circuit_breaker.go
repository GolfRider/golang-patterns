package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type CircuitBreakerState int

const (
	CBClosed CircuitBreakerState = iota
	CBOpen
	CBHalfOpen
)

type CircuitBreaker struct {
	mx           sync.Mutex
	lastErrorTS  time.Time
	resetTimeout time.Duration
	errorCount   int
	maxErrors    int
	state        CircuitBreakerState
}

func NewCircuitBreaker(maxErrors int) *CircuitBreaker {
	return &CircuitBreaker{
		maxErrors:    maxErrors,
		resetTimeout: time.Second * 5,
		state:        CBClosed,
	}
}

func (cb *CircuitBreaker) Call(f func() error) error {
	// 1. Quick check of state (Acquire lock briefly)
	cb.mx.Lock()
	if cb.state == CBOpen {
		if time.Since(cb.lastErrorTS) > cb.resetTimeout {
			cb.state = CBHalfOpen
		} else {
			cb.mx.Unlock()
			return errors.New("circuit breaker is open")
		}
	}
	cb.mx.Unlock()

	// 2. Execute the function (OUTSIDE the lock)
	err := f()

	// 3. Update state based on result (Acquire lock again)
	cb.mx.Lock()
	defer cb.mx.Unlock()

	if err != nil {
		cb.errorCount++
		cb.lastErrorTS = time.Now()
		if cb.errorCount >= cb.maxErrors || cb.state == CBHalfOpen {
			cb.state = CBOpen
		}
		return err
	}

	// Success logic
	cb.state = CBClosed
	cb.errorCount = 0
	return nil
}

func checkCircuitBreaker() {
	cb := NewCircuitBreaker(3)
	for i := 0; i < 5; i++ {
		err := cb.Call(func() error {
			fmt.Println("check-circuit breaker")
			return errors.New("circuit breaker is open")
		})
		if err != nil {
			fmt.Println("check-circuit breaker: ", err.Error())
		}
	}
}
