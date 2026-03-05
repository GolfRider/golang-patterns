package main

import (
	"fmt"
	"sync"
	"time"
)

// SlidingWindow: allow N requests per window duration.
// Uses sub-buckets for smoother counting than a fixed window.
//
//	limit=10/min, 6 buckets of 10s each
//
//	time ──────────────────────────►
//	[2][3][1][4][0][2]  ← counts per 10s bucket
//	 ↑ oldest           ↑ current
//	sum=12 > 10 → DENY
type SlidingWindow struct {
	mu    sync.Mutex
	limit int

	window    time.Duration
	bucketDur time.Duration
	buckets   []int // circular buffer of counts / ring-buffer of buckets

	lastBucket int // index of current bucket
	lastTime   time.Time
}

func NewSlidingWindow(limit int, window time.Duration, numBuckets int) *SlidingWindow {
	return &SlidingWindow{
		limit:      limit,
		window:     window,
		buckets:    make([]int, numBuckets),
		bucketDur:  window / time.Duration(numBuckets),
		lastBucket: 0,
		lastTime:   time.Now(),
	}
}

// advance the buckets
func (sw *SlidingWindow) expireAndSlideWindow() {
	now := time.Now()
	elapsed := now.Sub(sw.lastTime)
	bucketsToAdvance := int(elapsed / sw.bucketDur)

	if bucketsToAdvance <= 0 {
		return
	}

	// Clear old buckets as we advance
	n := len(sw.buckets)
	if bucketsToAdvance > n {
		bucketsToAdvance = n
	}
	for i := 0; i < bucketsToAdvance; i++ {
		sw.lastBucket = (sw.lastBucket + 1) % n
		sw.buckets[sw.lastBucket] = 0 // CLEAR old-buckets by making it ZERO
	}
	sw.lastTime = now
}

// Allow : check and update the counters
func (sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.expireAndSlideWindow()

	// Sum all buckets
	total := 0
	for _, c := range sw.buckets {
		total += c
	}

	if total >= sw.limit {
		return false
	}
	sw.buckets[sw.lastBucket]++ // increment the counter
	return true
}

func problem8() {
	// 5 requests per 2 seconds, 4 sub-buckets
	rl := NewSlidingWindow(5, 2*time.Second, 4)

	for i := 0; i < 7; i++ {
		fmt.Printf("req %d: allowed=%v\n", i, rl.Allow())
	}

	fmt.Println("sleeping 2s...")
	time.Sleep(2 * time.Second)

	for i := 0; i < 7; i++ {
		fmt.Printf("req %d: allowed=%v\n", i, rl.Allow())
	}
}
