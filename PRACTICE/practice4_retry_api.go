package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

func FindResourceWithRetry(ctx context.Context, region string) (string, error) {
	maxAttempts := 3
	backoff := 1 * time.Second
	var lastErr error // Track the last error for better reporting

	for i := 0; i < maxAttempts; i++ {
		fmt.Printf("Attempt %d for %s...\n", i+1, region) // i+1 is more readable for logs

		res, err := callAPI(ctx, region)
		if err == nil {
			return res, nil
		}
		lastErr = err

		// Immediate exit for terminal context errors
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return "", err
		}

		// Calculate wait with Jitter
		wait := backoff + time.Duration(rand.Intn(100))*time.Millisecond

		// PRINCIPAL FIX: Use NewTimer to prevent memory leaks
		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop() // Clean up timer immediately
			return "", ctx.Err()
		case <-t.C:
			backoff *= 2 // Prepare next backoff
		}
	}

	return "", fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

func callAPI(ctx context.Context, region string) (string, error) {
	if rand.Float64() < 0.5 {
		return "", errors.New("network timeout")
	}
	return "gpu-123", nil
}

func problem4() {
	// Total time we are willing to wait for the whole process
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := FindResourceWithRetry(ctx, "us-east-1")
	if err != nil {
		fmt.Printf("Final Error: %v\n", err)
		return
	}
	fmt.Printf("Success: %s\n", result)
}
