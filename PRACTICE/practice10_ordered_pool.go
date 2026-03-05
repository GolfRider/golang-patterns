package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

type Event struct {
	Key     string // e.g., UserID or Database Table Key
	Payload string
}

type OrderedPool struct {
	workers []*worker
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

type worker struct {
	ch chan Event
}

func NewOrderedPool(workerCount int, bufferSize int) *OrderedPool {
	ctx, cancel := context.WithCancel(context.Background())
	p := &OrderedPool{
		workers: make([]*worker, workerCount),
		ctx:     ctx,
		cancel:  cancel,
	}

	// Initialize each worker with its own dedicated channel
	for i := 0; i < workerCount; i++ {
		p.workers[i] = &worker{
			ch: make(chan Event, bufferSize),
		}
	}
	return p
}

func (p *OrderedPool) Start() {
	for i, w := range p.workers {
		p.wg.Add(1)
		go func(id int, wn *worker) {
			defer p.wg.Done()
			for {
				select {
				case <-p.ctx.Done():
					// Optional: Drain wn.ch here if "At-Least-Once" is required
					return
				case event, ok := <-wn.ch:
					if !ok {
						return
					}
					// All events for the same Key land in this specific goroutine
					fmt.Printf("[Worker %d] Processing Key: %s | Data: %s\n", id, event.Key, event.Payload)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}(i, w)
	}
}

// Submit uses hashing to route the event to the correct worker
func (p *OrderedPool) Submit(e Event) {
	// 1. Hash the key
	h := fnv.New32a()
	h.Write([]byte(e.Key))

	// 2. Route to the specific worker index
	workerIdx := int(h.Sum32()) % len(p.workers)

	select {
	case <-p.ctx.Done():
		return
	case p.workers[workerIdx].ch <- e:
		// Sent to the specific worker's "conveyor belt"
	}
}

func (p *OrderedPool) Stop() {
	p.cancel()
	for _, w := range p.workers {
		close(w.ch)
	}
	p.wg.Wait()
}

func problem10() {
	pool := NewOrderedPool(3, 10)
	pool.Start()

	// Even though we submit these rapidly, "User-1" events
	// will always be processed in order (1, 2, 3) by the same worker.
	pool.Submit(Event{Key: "User-1", Payload: "Update-1"})
	pool.Submit(Event{Key: "User-2", Payload: "Update-A"})
	pool.Submit(Event{Key: "User-1", Payload: "Update-2"})
	pool.Submit(Event{Key: "User-1", Payload: "Update-3"})

	time.Sleep(1 * time.Second)
	pool.Stop()
}
