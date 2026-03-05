package main

import (
	"fmt"
	"sync"
	"time"
)

// TokenBucket: tokens refill at a steady rate, each request costs 1 token.
// Think of it like a bucket that fills with tokens over time.
//
//	capacity=5, refillRate=1/sec
//	[•••••] full  → Allow 5 burst requests
//	[••___] 2 left → 3 were consumed
//	... 1 sec later → [•••__] 1 token refilled
type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity, // start full
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now

	// Try to consume 1 token
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

func problem7() {
	// 5 burst capacity, refills 2 tokens/sec
	rl := NewTokenBucket(5, 2)

	// Burn through the burst
	for i := 0; i < 7; i++ {
		fmt.Printf("req %d: allowed=%v\n", i, rl.Allow())
	}

	// Wait and try again
	fmt.Println("sleeping 2s...")
	time.Sleep(2 * time.Second) // should refill ~4 tokens

	for i := 0; i < 5; i++ {
		fmt.Printf("req %d: allowed=%v\n", i, rl.Allow())
	}
}
