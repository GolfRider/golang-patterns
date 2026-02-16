package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

func retryWithJitter(ctx context.Context, maxRetries int, fn func() error) error {
	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		if attempt == maxRetries {
			break
		}

		// Exponential backoff with full jitter
		base := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
		jitter := time.Duration(rand.N(base))
		backoff := jitter // full jitter: [0, base)

		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(backoff):
		}
	}
	return fmt.Errorf("after %d retries: %w", maxRetries, err)
}

func checkRetry() {
	err := retryWithJitter(context.Background(), 3, func() error {
		fmt.Println("check-1")
		return errors.New("retryable")
	})

	if err != nil {
		fmt.Println("Error:", err)
	}
}
