package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Periodic sync
	defer ticker.Stop()

	debounceTimer := time.NewTimer(500 * time.Millisecond)
	debounceTimer.Stop() // Initially stopped

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			//c.Reconcile()
		case <-debounceTimer.C:
			// Only run after 500ms of no new events
			//c.Reconcile()

		case <-eventUpdate():
			// Reset the timer every time a new event arrives
			debounceTimer.Reset(500 * time.Millisecond)

		}
	}
}

// dummy method
func eventUpdate() chan struct{} {
	return nil
}

///// Another way

type Debouncer struct {
	mu    sync.Mutex
	timer *time.Timer
	delay time.Duration
}

func NewDebouncer(delay time.Duration) *Debouncer {
	return &Debouncer{delay: delay}
}

// Do schedules the function f to run after the delay.
// If Do is called again before the delay, the previous timer is cancelled.
func (d *Debouncer) Do(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 1. If a timer is already running, stop it
	if d.timer != nil {
		d.timer.Stop()
	}

	// 2. Schedule the function to run after the delay
	d.timer = time.AfterFunc(d.delay, f)
}

func problem15() {
	d := NewDebouncer(500 * time.Millisecond)

	fmt.Println("Events starting...")

	// Simulate rapid events
	for i := 0; i < 5; i++ {
		d.Do(func() {
			fmt.Println("!!! Action Executed (Debounced) !!!")
		})
		time.Sleep(100 * time.Millisecond) // Events closer than 500ms
	}

	// Wait long enough for the final timer to fire
	time.Sleep(1 * time.Second)
}
