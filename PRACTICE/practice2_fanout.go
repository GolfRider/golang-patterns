package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func FindFirstResponse(ctx context.Context, urls []string) (string, error) {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Even with buffer 1, your select-send below prevents leaks.
	// Principal tip: I'd still use len(urls) to be "obviously safe" to reviewers.
	resultChan := make(chan string, len(urls))

	for _, url := range urls {
		// FIX 1: Pass 'url' into the goroutine to avoid closure capture bug
		go func(u string) {
			if res, err := fetchData(childCtx, u); err == nil {
				select {
				case resultChan <- res:
					// Successfully sent!
				case <-childCtx.Done():
					// Context was cancelled by a winner or timeout; exit safely.
					return
				}
			}
		}(url) // Pass it here
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-resultChan:
		return res, nil
	}
}

func fetchData(ctx context.Context, url string) (string, error) {
	latency := time.Duration(rand.Intn(500)) * time.Millisecond

	// FIX 2: Use NewTimer to avoid memory leaks
	timer := time.NewTimer(latency)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-timer.C:
		return "success from " + url, nil
	}
}

func problem2() {
	regions := []string{"us-east", "us-west", "eu-central", "asia-east"}

	// Global timeout for the whole operation
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	res, err := FindFirstResponse(ctx, regions)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Result found:", res)
	}
}
