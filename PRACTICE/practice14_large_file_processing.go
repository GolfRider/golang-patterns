package main

import (
	"bufio"
	"context"
	"os"
	"sync"
)

func ProcessLargeFile(filePath string, workerCount int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use a context for graceful cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lines := make(chan string, 100) // Backpressure: don't read faster than we can process
	var wg sync.WaitGroup

	// 1. Start Workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for line := range lines {
				// Process line (e.g., regex, aggregation)
				_ = line
			}
		}(i)
	}

	// 2. Stream File
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			break
		case lines <- scanner.Text():
		}
	}

	close(lines) // Signal workers to finish
	wg.Wait()
	return scanner.Err()
}
