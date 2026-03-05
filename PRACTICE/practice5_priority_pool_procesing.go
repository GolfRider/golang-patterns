package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type TaskType int

const (
	Critical TaskType = iota
	Background
)

type Task struct {
	ID      int
	Type    TaskType
	Payload string
}

type WorkerPool struct {
	criticalChan   chan Task
	backgroundChan chan Task
	wg             sync.WaitGroup
	workerCount    int
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewWorkerPool(workerCount int, bufferSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		criticalChan:   make(chan Task, bufferSize),
		backgroundChan: make(chan Task, bufferSize),
		workerCount:    workerCount,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	fmt.Printf("Worker %d started\n", id)

	for {
		select {
		case <-wp.ctx.Done():
			return
		// The "Priority Trick": Check critical first explicitly
		case task := <-wp.criticalChan:
			wp.process(id, task)
		default:
			// If no critical task, then we select between both.
			// This allows background tasks to progress but checks critical first.
			select {
			case <-wp.ctx.Done():
				return
			case task := <-wp.criticalChan:
				wp.process(id, task)
			case task := <-wp.backgroundChan:
				wp.process(id, task)
			}
		}
	}
}

func (wp *WorkerPool) process(workerID int, t Task) {
	// In production, you'd have telemetry here (e.g., prometheus.Inc())
	fmt.Printf("[Worker %d] Processing %v Task ID: %d\n", workerID, t.Type, t.ID)
	time.Sleep(100 * time.Millisecond) // Simulate work
}

func (wp *WorkerPool) Submit(t Task) error {
	select {
	case <-wp.ctx.Done():
		return errors.New("pool is closed")
	default:
		if t.Type == Critical {
			wp.criticalChan <- t
		} else {
			wp.backgroundChan <- t
		}
		return nil
	}
}

func (wp *WorkerPool) Shutdown() {
	wp.cancel()
	wp.wg.Wait()
	fmt.Println("All workers shut down gracefully.")
}
