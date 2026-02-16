package main

import (
	"context"
	"fmt"
)

func generate(ctx context.Context, tasks []Task) <-chan Task {
	out := make(chan Task)
	go func() {
		defer close(out)
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return
			default:
				out <- task
			}
		}
	}()
	return out
}

func transform(ctx context.Context, in <-chan Task) <-chan Result {
	out := make(chan Result)
	go func() {
		defer close(out)
		for task := range in {
			select {
			case <-ctx.Done():
				out <- Result{TaskID: task.ID, Err: ctx.Err()}
				return
			default:
				out <- Result{TaskID: task.ID, Output: fmt.Sprintf("processed: %d", task.ID)}
			}
		}
	}()
	return out
}

func checkPipeline() {
	ctx := context.Background()
	tasks := []Task{{1, "AA"}, {2, "BB"}, {3, "CC"}, {4, "DD"}}
	taskCh := generate(ctx, tasks)
	resultCh := transform(ctx, taskCh)

	for result := range resultCh {
		fmt.Println(result.TaskID, result.Output)
	}
}
