package main

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

func checkErrorGroupWithParallelWork() {
	G, ctx := errgroup.WithContext(context.Background())
	G.SetLimit(3)

	G.Go(func() error {
		time.Sleep(200 * time.Millisecond)
		fmt.Println("task1 done")
		return nil
	})

	// Task 2 (fails)
	G.Go(func() error {
		time.Sleep(300 * time.Millisecond)
		return fmt.Errorf("task2 failed")
	})

	// Task 3 (checks cancellation)
	G.Go(func() error {
		select {
		case <-time.After(1 * time.Second):
			fmt.Println("task3 done")
			return nil
		case <-ctx.Done():
			fmt.Println("task3 canceled:", ctx.Err())
			return ctx.Err()
		}
	})

	if err := G.Wait(); err != nil {
		fmt.Println("errgroup returned:", err)
	}
}
