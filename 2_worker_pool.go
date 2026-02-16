package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Task represents a unit of work
type Task struct {
	ID   int
	Data string
}

// Result represents the outcome of processing a task
type Result struct {
	TaskID int
	Output string
	Err    error
}

func workerPool(ctx context.Context, tasks []Task, numWorkers int) []Result {
	taskCh := make(chan Task, len(tasks))
	resultCh := make(chan Result, len(tasks))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range taskCh {
				select {
				case <-ctx.Done():
					resultCh <- Result{TaskID: task.ID, Err: ctx.Err()}
					return
				default:
					output := processTask(task)
					resultCh <- Result{TaskID: task.ID, Output: output}
				}
			}
		}(i)
	}

	// Send tasks
	for _, t := range tasks {
		taskCh <- t
	}
	close(taskCh)

	// Wait for completion, then close results
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	var results []Result
	for r := range resultCh {
		results = append(results, r)
	}
	return results
}

func processTask(t Task) string {
	time.Sleep(100 * time.Millisecond) // simulate work
	return fmt.Sprintf("processed: %s", t.Data)
}

func checkWorkerPool() {
	tasks := []Task{{1, "AA"}, {2, "BB"}, {3, "CC"}, {4, "DD"}}
	results := workerPool(context.Background(), tasks, 4)
	for _, result := range results {
		fmt.Println(result.TaskID, result.Output)
	}
}
