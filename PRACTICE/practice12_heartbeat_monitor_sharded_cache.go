package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

// Shard holds a subset of node heartbeats to reduce lock contention
type Shard2 struct {
	mu    sync.RWMutex
	nodes map[string]time.Time
}

type HeartbeatMonitor struct {
	shards     []*Shard2
	shardCount int
	ttl        time.Duration
}

func NewMonitor(shardCount int, ttl time.Duration) *HeartbeatMonitor {
	m := &HeartbeatMonitor{
		shards:     make([]*Shard2, shardCount),
		shardCount: shardCount,
		ttl:        ttl,
	}
	for i := 0; i < shardCount; i++ {
		m.shards[i] = &Shard2{nodes: make(map[string]time.Time)}
	}
	return m
}

// getShard determines which bucket a node belongs to using a standard hash modulo
func (m *HeartbeatMonitor) getShard(nodeID string) *Shard2 {
	h := fnv.New32a()
	h.Write([]byte(nodeID))

	// Standard modulo: works with any shardCount, not just powers of 2
	index := int(h.Sum32()) % m.shardCount
	return m.shards[index]
}

// RecordHeartbeat updates the "last seen" timestamp for a node
func (m *HeartbeatMonitor) RecordHeartbeat(nodeID string) {
	shard := m.getShard(nodeID)
	shard.mu.Lock()
	shard.nodes[nodeID] = time.Now()
	shard.mu.Unlock()
}

// StartJanitor runs a single background loop to detect stale nodes
func (m *HeartbeatMonitor) StartJanitor(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop() // Crucial: prevent timer leaks

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.sweep()
		}
	}
}

func (m *HeartbeatMonitor) sweep() {
	now := time.Now()
	for _, shard := range m.shards {
		shard.mu.Lock()
		for id, lastSeen := range shard.nodes {
			if now.Sub(lastSeen) > m.ttl {
				fmt.Printf("[ALERT] Node %s is STALE. Last seen: %v ago\n", id, now.Sub(lastSeen))
				delete(shard.nodes, id) // Prevent memory leaks/OOM
			}
		}
		shard.mu.Unlock()
	}
}

func problem12() {
	// 10 shards, 5s TTL
	monitor := NewMonitor(10, 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Single background loop for all nodes
	go monitor.StartJanitor(ctx, 1*time.Second)

	monitor.RecordHeartbeat("gpu-node-01")
	time.Sleep(2 * time.Second)
	monitor.RecordHeartbeat("gpu-node-01")

	fmt.Println("Monitoring node heartbeats...")
	time.Sleep(10 * time.Second)
}
