package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

/*  Bounded Memory RateLimiter: LRU + TTL + Fixed Window RateLimiter
    This uses a mix of various concepts:
          - lru-cache (space-based eviction) + ttl (time-based eviction)
          - fixed window-size rate-limiter
*/

type RateLimiterEntry struct {
	ip              string
	count           int
	windowStartTime time.Time
	element         *list.Element
}

type BoundedRateLimiterFixedWindow struct {
	mx sync.Mutex

	// configs
	maxAllowed int
	windowSize time.Duration
	maxEntries int

	// state
	entriesMap map[string]*RateLimiterEntry
	lru        *list.List

	// lifecycle
	stopChan chan struct{}
}

func NewBoundedRateLimiterFixedWindow(maxAllowed int, windowSize time.Duration, maxEntries int) *BoundedRateLimiterFixedWindow {
	rl := &BoundedRateLimiterFixedWindow{
		maxAllowed: maxAllowed,
		windowSize: windowSize,
		maxEntries: maxEntries,
		entriesMap: make(map[string]*RateLimiterEntry, maxEntries),
		lru:        list.New(),
		stopChan:   make(chan struct{}),
	}
	go rl.ttlCleanupLoop()
	return rl
}

func (b *BoundedRateLimiterFixedWindow) Allow(ip string) bool {
	b.mx.Lock()
	defer b.mx.Unlock()
	now := time.Now()

	if entry, ok := b.entriesMap[ip]; ok {
		b.lru.MoveToFront(entry.element)

		if now.Sub(entry.windowStartTime) > b.windowSize { // new window
			entry.count = 1
			entry.windowStartTime = now
			return true
		}
		if entry.count < b.maxAllowed {
			entry.count++
			return true
		}
		return false
	}
	if len(b.entriesMap) >= b.maxEntries {
		b.evictLRU()
	}
	entry := &RateLimiterEntry{
		ip:              ip,
		count:           1,
		windowStartTime: now,
	}
	entry.element = b.lru.PushFront(entry)
	b.entriesMap[ip] = entry
	return true
}

func (b *BoundedRateLimiterFixedWindow) evictLRU() {
	tail := b.lru.Back()
	if tail == nil {
		return
	}
	entry := tail.Value.(*RateLimiterEntry)
	b.lru.Remove(tail)
	delete(b.entriesMap, entry.ip)
}

func (b *BoundedRateLimiterFixedWindow) Stop() {
	close(b.stopChan)
}

func (b *BoundedRateLimiterFixedWindow) ttlCleanupLoop() {
	ticker := time.NewTicker(b.windowSize)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopChan:
			return

		case <-ticker.C:
			b.sweep()
		}
	}
}

func (b *BoundedRateLimiterFixedWindow) sweep() {
	b.mx.Lock()
	defer b.mx.Unlock()
	now := time.Now()

	for {
		tail := b.lru.Back()
		if tail == nil {
			break
		}
		entry := tail.Value.(*RateLimiterEntry)

		if now.Sub(entry.windowStartTime) <= b.windowSize {
			break
		}
		b.lru.Remove(tail)
		delete(b.entriesMap, entry.ip)
	}
}

func (b *BoundedRateLimiterFixedWindow) Len() int {
	b.mx.Lock()
	defer b.mx.Unlock()
	return len(b.entriesMap)
}

// Test Bounded RateLimiter
func checkBoundedRateLimiter() {
	checkConcurrentAccess()
	testAllowWithinLimit()
}

func checkConcurrentAccess() {
	rl := NewBoundedRateLimiterFixedWindow(1000, 1*time.Second, 10000)
	defer rl.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ip := fmt.Sprintf("192.0.2.%d", i)
			for j := 0; j < 50; j++ {
				rl.Allow(ip)
			}
		}()
	}
	wg.Wait()
	if rl.Len() != 100 {
		fmt.Printf("\n expected 100 tracked IPs, got %d", rl.Len())
	} else {
		fmt.Printf("\n all-good !!")
	}
}

func testAllowWithinLimit() {
	rl := NewBoundedRateLimiterFixedWindow(5, 1*time.Second, 100)
	defer rl.Stop()

	for i := 0; i < 5; i++ {
		if !rl.Allow("1.2.3.4") {
			fmt.Println("request %d should be allowed", i+1)
		}
	}

	// 6th should be rejected.
	if rl.Allow("1.2.3.4") {
		fmt.Println("request past limit should be denied")
	} else {
		fmt.Println("request past limit should be denied")
	}
}
