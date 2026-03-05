package main

import (
	"fmt"
	"sync"
	"time"
)

type entry struct {
	val       interface{}
	expiresAt time.Time
}

type TTLCache struct {
	mu    sync.RWMutex
	ttl   time.Duration
	items map[string]entry

	// singleflight: one in-flight request per key
	sfMu  sync.Mutex
	calls map[string]*call
}

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

func NewTTLCache(ttl time.Duration) *TTLCache {
	return &TTLCache{
		items: make(map[string]entry),
		calls: make(map[string]*call),
		ttl:   ttl,
	}
}

// Get returns cached value if present and not expired.
func (c *TTLCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.val, true
}

// Set stores a value with TTL.
func (c *TTLCache) Set(key string, val interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry{val: val, expiresAt: time.Now().Add(c.ttl)}
}

// GetOrFetch checks cache first, otherwise uses singleflight to call fn once
// per key even if many goroutines request simultaneously.
func (c *TTLCache) GetOrFetch(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 1. Check cache
	if v, ok := c.Get(key); ok {
		return v, nil
	}

	// 2. Singleflight: deduplicate concurrent calls for the same key
	c.sfMu.Lock()
	if cl, ok := c.calls[key]; ok {
		// another goroutine is already fetching — just wait
		c.sfMu.Unlock()
		cl.wg.Wait()
		return cl.val, cl.err
	}
	cl := &call{}
	cl.wg.Add(1)
	c.calls[key] = cl
	c.sfMu.Unlock()

	// 3. We're the leader — do the actual fetch
	cl.val, cl.err = fn()
	if cl.err == nil {
		c.Set(key, cl.val)
	}
	cl.wg.Done()

	// 4. Cleanup so future requests after TTL expiry can fetch again
	c.sfMu.Lock()
	delete(c.calls, key)
	c.sfMu.Unlock()

	return cl.val, cl.err
}

// --- Demo ---

func problem6() {
	cache := NewTTLCache(2 * time.Second)
	fetchCount := 0

	fetch := func() (interface{}, error) {
		fetchCount++
		fmt.Println("  actual fetch called")
		time.Sleep(100 * time.Millisecond) // simulate slow work
		return "data", nil
	}

	// Multiple goroutines request same key — only one fetch fires
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			val, _ := cache.GetOrFetch("key1", fetch)
			fmt.Printf("  goroutine %d got: %v\n", id, val)
		}(i)
	}
	wg.Wait()
	fmt.Printf("fetch was called %d time(s)\n\n", fetchCount)

	// Cached hit
	val, ok := cache.Get("key1")
	fmt.Printf("cached: %v (hit=%v)\n", val, ok)

	// Wait for TTL to expire, then fetch again
	time.Sleep(3 * time.Second)
	val, ok = cache.Get("key1")
	fmt.Printf("after TTL: hit=%v\n", ok)

	val, _ = cache.GetOrFetch("key1", fetch)
	fmt.Printf("re-fetched: %v, total fetches: %d\n", val, fetchCount)
}

// Your career does not need heroic effort every day. It needs a non zero minimum standard that runs regardless of mood.
