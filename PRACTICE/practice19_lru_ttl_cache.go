package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// Entry holds the data and its expiration metadata
type Entry struct {
	Key    string
	Value  interface{}
	Expiry time.Time
}

type LRUTTLCache struct {
	capacity  int
	ttl       time.Duration
	mu        sync.Mutex
	items     map[string]*list.Element
	evictList *list.List // Doubly linked list for LRU
}

func NewLRUTTLCache(capacity int, ttl time.Duration) *LRUTTLCache {
	c := &LRUTTLCache{
		capacity:  capacity,
		ttl:       ttl,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
	// Start background cleanup
	go c.janitor()
	return c
}

// Get retrieves an item and updates its LRU position
func (c *LRUTTLCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, exists := c.items[key]
	if !exists {
		return nil, false
	}

	entry := element.Value.(*Entry)
	// Check if expired
	if time.Now().After(entry.Expiry) {
		c.removeElement(element)
		return nil, false
	}

	// Move to front (Most Recently Used)
	c.evictList.MoveToFront(element)
	return entry.Value, true
}

// Put adds or updates an item, enforcing capacity
func (c *LRUTTLCache) Put(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiry := time.Now().Add(c.ttl)

	// Update existing item
	if element, exists := c.items[key]; exists {
		entry := element.Value.(*Entry)
		entry.Value = value
		entry.Expiry = expiry
		c.evictList.MoveToFront(element)
		return
	}

	// Add new item
	entry := &Entry{Key: key, Value: value, Expiry: expiry}
	element := c.evictList.PushFront(entry)
	c.items[key] = element

	// Evict if over capacity
	if c.evictList.Len() > c.capacity {
		c.removeOldest()
	}
}

func (c *LRUTTLCache) removeOldest() {
	element := c.evictList.Back()
	if element != nil {
		c.removeElement(element)
	}
}

func (c *LRUTTLCache) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	entry := e.Value.(*Entry)
	delete(c.items, entry.Key)
}

// janitor periodically removes expired items to prevent memory leaks
func (c *LRUTTLCache) janitor() {
	ticker := time.NewTicker(c.ttl / 2) // Sweep twice per TTL period
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for _, element := range c.items {
			entry := element.Value.(*Entry)
			if time.Now().After(entry.Expiry) {
				c.removeElement(element)
			}
		}
		c.mu.Unlock()
	}
}

func problem19() {
	cache := NewLRUTTLCache(2, 2*time.Second)

	cache.Put("node-1", "gpu-cluster-a")
	cache.Put("node-2", "gpu-cluster-b")

	val, _ := cache.Get("node-1")
	fmt.Println("Found:", val)

	// This will evict "node-2" because "node-1" was just accessed (MRU)
	cache.Put("node-3", "gpu-cluster-c")

	_, exists := cache.Get("node-2")
	fmt.Println("Node-2 exists:", exists) // false

	time.Sleep(3 * time.Second)
	_, exists = cache.Get("node-1")
	fmt.Println("Node-1 exists after TTL:", exists) // false
}
