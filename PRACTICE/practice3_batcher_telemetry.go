package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Batcher
/*
   Note: The General "Order of Operations"OrderActionPurpose
   1. Signal => cancel() | Stop the Producers from trying to add more work.
   2. Close  => close(ch) |   Signal the Workers that no more work is coming on the conveyor belt.
   3. Wait   => wg.Wait() | Block the main thread until the workers have finished the items already in the belt and exited.

   // cancel(), close(), wait()
    [struct + state] > pipeline > safe-loop > graceful-shutdown
*/

type Batcher struct {
	// pool config
	jobQueue chan string
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	once     sync.Once

	// batcher-specific
	flushInterval time.Duration
	maxBatchSize  int
}

func NewBatcher(flushInterval time.Duration, maxBatchSize int) *Batcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &Batcher{
		flushInterval: flushInterval,
		jobQueue:      make(chan string, maxBatchSize),
		maxBatchSize:  maxBatchSize,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (batcher *Batcher) Start() {
	batcher.wg.Add(1)
	go func() {
		defer batcher.wg.Done()

		batch := make([]string, 0, batcher.maxBatchSize)
		ticker := time.NewTicker(batcher.flushInterval)
		defer ticker.Stop()

		for {
			select {
			case job, ok := <-batcher.jobQueue:
				if !ok {
					if len(batch) > 0 {
						batcher.flush(batch)
					}
					return
				}
				batch = append(batch, job)
				if len(batch) >= batcher.maxBatchSize {
					batcher.flush(batch)
					batch = make([]string, 0, batcher.maxBatchSize)
					ticker.Reset(batcher.flushInterval)
				}
			case <-ticker.C:
				// flush items
				if len(batch) > 0 {
					batcher.flush(batch)
					batch = make([]string, 0, batcher.maxBatchSize)
				}
			case <-batcher.ctx.Done():
				if len(batch) > 0 {
					batcher.flush(batch)
				}
				return
			}
		}
	}()
}

func (batcher *Batcher) Stop() {
	// PHASE 1: Stop the input
	// We signal the context first. This tells all 'AddItem' callers
	// to stop trying to send immediately.
	batcher.cancel()
	time.Sleep(10 * time.Millisecond) // grace period

	// PHASE 2: Drain the buffer
	// Now that we've signaled producers to stop, we close the channel
	// to let the workers know they should finish the remaining items
	// and then exit.
	batcher.once.Do(func() {
		close(batcher.jobQueue)
	})

	// PHASE 3: Synchronize
	// We wait for the 'Done' signal from all workers.
	batcher.wg.Wait()

}

func (batcher *Batcher) flush(payload []string) {
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Flushing: ", len(payload))
}

// can be called by many go-routines
func (batcher *Batcher) Submit(item string) error {
	select {
	case <-batcher.ctx.Done():
		return fmt.Errorf("batcher shutting down")
	default:
	}

	select {
	case <-batcher.ctx.Done():
		return fmt.Errorf("batcher shutting down")
	case batcher.jobQueue <- item:
		return nil
	case <-time.After(1 * time.Second): // Backpressure!
		return fmt.Errorf("batcher queue full")
	}
}

func problem3() {

	batcher := NewBatcher(time.Duration(3)*time.Second, 10)
	batcher.Start()
	for i := 0; i < 100; i++ {
		batcher.Submit(fmt.Sprintf("job: %d", i))
	}
	batcher.Stop()
}
