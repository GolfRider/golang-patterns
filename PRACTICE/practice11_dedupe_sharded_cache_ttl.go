package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

type Shard struct {
	mu   sync.Mutex
	seen map[string]time.Time
}

type Deduplicator struct {
	shards     []*Shard
	shardCount uint32
	ttl        time.Duration
}

func NewDeduplicator(shardCount uint32, ttl time.Duration) *Deduplicator {
	d := &Deduplicator{
		shards:     make([]*Shard, shardCount),
		shardCount: shardCount,
		ttl:        ttl,
	}
	for i := uint32(0); i < shardCount; i++ {
		d.shards[i] = &Shard{seen: make(map[string]time.Time)}
	}
	return d
}

func (d *Deduplicator) getShard(id string) *Shard {
	h := fnv.New32a()
	h.Write([]byte(id))
	return d.shards[h.Sum32()%d.shardCount] // simple modulo
}

func (d *Deduplicator) IsDuplicate(id string) bool {
	shard := d.getShard(id)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	lastSeen, exists := shard.seen[id]
	if exists && time.Since(lastSeen) < d.ttl {
		return true
	}

	shard.seen[id] = time.Now() // NEW ENTRY
	return false
}

func (d *Deduplicator) StartJanitor(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.sweep()
		}
	}
}

func (d *Deduplicator) sweep() {
	now := time.Now()
	for _, shard := range d.shards {
		shard.mu.Lock() // shard-locks
		for id, lastSeen := range shard.seen {
			if now.Sub(lastSeen) > d.ttl {
				delete(shard.seen, id)
			}
		}
		shard.mu.Unlock()
	}
}

func problem11() {
	dedup := NewDeduplicator(4, 2*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go dedup.StartJanitor(ctx, 1*time.Second)

	fmt.Println("First:      ", dedup.IsDuplicate("msg_123")) // false
	fmt.Println("Duplicate:  ", dedup.IsDuplicate("msg_123")) // true
	fmt.Println("Different:  ", dedup.IsDuplicate("msg_456")) // false

	fmt.Println("\nWaiting for TTL...")
	time.Sleep(3 * time.Second)
	fmt.Println("After TTL:  ", dedup.IsDuplicate("msg_123")) // false

	// Concurrent: exactly 1 goroutine sees false
	fmt.Println("\n=== 10 goroutines, same ID ===")
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("  goroutine %d: dup=%v\n", id, dedup.IsDuplicate("job_789"))
		}(i)
	}
	wg.Wait()
}
