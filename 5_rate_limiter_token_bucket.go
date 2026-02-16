package main

import (
	"log"
	"sync"
	"time"
)

type TokenBucket struct {
	mx         sync.Mutex
	maxTokens  float64
	tokens     float64
	rate       float64
	lastRefill time.Time
}

func NewTokenBucket(rate, maxTokens float64) *TokenBucket {
	if rate <= 0 || maxTokens <= 0 {
		log.Fatal("Invalid rate or max tokens")
		return nil
	}
	return &TokenBucket{
		rate:       rate,
		maxTokens:  maxTokens,
		tokens:     maxTokens,
		lastRefill: time.Now(),
	}
}

func (t *TokenBucket) Allow() bool {
	t.mx.Lock()
	defer t.mx.Unlock()

	now := time.Now()
	elapsed := now.Sub(t.lastRefill).Seconds()
	t.tokens = t.tokens + t.rate*elapsed
	t.lastRefill = now

	if t.tokens > t.maxTokens {
		t.tokens = t.maxTokens
	}

	if t.tokens >= 1.0 {
		t.tokens = t.tokens - 1.0
		return true
	}
	return false
}
