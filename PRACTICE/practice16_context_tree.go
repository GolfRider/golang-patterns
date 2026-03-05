package main

import (
	"context"
	"fmt"
	"time"
)

func problem16() {
	// 1. GLOBAL CONTEXT: The "Service Lifecycle"
	// If this is cancelled, the entire service stops.
	rootCtx, stopService := context.WithCancel(context.Background())
	defer stopService()

	fmt.Println("Service Started...")

	// 2. CHILD CONTEXT: Specific Task with Timeout
	// We want to fetch data, but we won't wait more than 500ms.
	taskCtx, cancelTask := context.WithTimeout(rootCtx, 500*time.Millisecond)
	defer cancelTask() // Always clean up child contexts

	go performDatabaseQuery(taskCtx)

	// Simulate the service running for a bit
	time.Sleep(1 * time.Second)
	fmt.Println("Service shutting down...")
}

func performDatabaseQuery(ctx context.Context) {
	select {
	case <-time.After(1 * time.Second): // Simulate a slow 1-second query
		fmt.Println("Query Successful!")
	case <-ctx.Done():
		// This triggers if the 500ms timeout hits OR if rootCtx is cancelled
		fmt.Printf("Query Aborted: %v\n", ctx.Err())
	}
}
