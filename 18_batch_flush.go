package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Batcher struct {
	maxWaitTime  time.Duration
	maxBatchSize int
	inputChan    chan Record
	wg           sync.WaitGroup
}

func NewBatcher(maxWaitTime time.Duration, maxBatchSize int) *Batcher {
	return &Batcher{
		maxWaitTime:  maxWaitTime,
		maxBatchSize: maxBatchSize,
		inputChan:    make(chan Record, maxBatchSize*2),
	}
}

// Add is concurrent safe, can be called from other go-routines
func (b *Batcher) Add(ctx context.Context, r Record) error {
	select {
	case b.inputChan <- r:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *Batcher) Start(ctx context.Context) {
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()
		batch := make([]Record, 0, b.maxBatchSize) // Note: ideally get it from sync.Pool to reduce allocations!

		// Use a Ticker for simpler periodic flushing logic
		ticker := time.NewTicker(b.maxWaitTime)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				if len(batch) > 0 {
					flush(batch)
				}
				return
			case r, ok := <-b.inputChan:
				if !ok {
					if len(batch) > 0 {
						flush(batch)
					}
					return
				}
				batch = append(batch, r)
				if len(batch) >= b.maxBatchSize {
					flush(batch)                              // flush
					batch = make([]Record, 0, b.maxBatchSize) // new-batch
					ticker.Reset(b.maxWaitTime)
				}
			case <-ticker.C:
				if len(batch) > 0 {
					flush(batch)                              // flush
					batch = make([]Record, 0, b.maxBatchSize) // new-batch
				}
			}
		}
	}()
}

func (b *Batcher) Stop() {
	close(b.inputChan) // Signal no more data
	b.wg.Wait()        // Wait for the background loop to finish flushing
}

// can be its own go-routine & async
func flush(records []Record) {
	for _, r := range records {
		fmt.Println("flushing: ", r)
		atomic.AddInt64(&counter, 1)
	}
}

var counter int64

func checkBatchFlush() {
	batcher := NewBatcher(250*time.Millisecond, 20)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batcher.Start(ctx)
	for i := 0; i < 10_000; i++ {
		record := Record{i, fmt.Sprintf("event-%d", i)}
		if err := batcher.Add(ctx, record); err != nil {
			log.Fatal(err)
		}
	}
	batcher.Stop()
	fmt.Println("processed items-count: ", counter)
}
