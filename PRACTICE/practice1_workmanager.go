package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type WorkManager struct {
	jobQueue chan func()
	wg       sync.WaitGroup
	ctx      context.Context    // The lifecycle signal ; The "Signal" line
	cancel   context.CancelFunc // The "switch" to stop everything; The "Signal" switch

	done chan struct{}
	once sync.Once
}

func NewWorkManager(size int) *WorkManager {
	// Initialize a context that we can manually cancel
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkManager{
		jobQueue: make(chan func(), size),
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
	}
}

func (wm *WorkManager) Start(workerCount int) {
	wm.wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wm.wg.Done()
			for {
				select {
				case job, ok := <-wm.jobQueue:
					if !ok {
						return // Channel closed, exit worker
					}
					job()
				case <-wm.ctx.Done():
					// drain the queue before exit ??
					for job := range wm.jobQueue {
						job()
					}
					return // Context cancelled, exit worker
				}
			}
		}()
	}
	go func() {
		wm.wg.Wait()
		close(wm.done)
	}()
}

func (wm *WorkManager) Submit(job func()) error {
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	select {
	case <-wm.ctx.Done():
		// If the manager is shutting down, reject the job immediately.
		// This prevents sending to a closed channel.
		return errors.New("work manager is shutting down")
	case wm.jobQueue <- job:
		return nil
	case <-timer.C:
		return fmt.Errorf("system overloaded") // BACK_PRESSURE
	}
}

func (wm *WorkManager) Stop() {
	// Calling cancel() multiple times is safe by design in Go.
	// Signal all Submits to stop immediately via Context
	wm.cancel()

	// sync.Once ensures that even if Stop() is called by 10 different
	// goroutines, the channel is closed exactly once.
	// 2. Close the channel so workers know to finish their current work
	wm.once.Do(func() {
		close(wm.jobQueue)
	})
	<-wm.done
}

func problem1() {
	wm := NewWorkManager(10)
	wm.Start(3)

	// Sample jobs
	for i := 1; i <= 4; i++ { // loop-submit
		jobID := i
		err := wm.Submit(func() {
			fmt.Printf("Executing Job-%d\n", jobID)
			time.Sleep(1 * time.Second)
		})
		checkError(err)
	}

	fmt.Println("Shutting down...")
	wm.Stop()
	fmt.Println("All workers finished.")
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Submit Error:", err)
	}
}
