package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func checkProducerConsumerBackpressure() {
	queue := make(chan int, 5)
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// note: you can have multiple producers/consumers
	var wg sync.WaitGroup
	wg.Add(1)

	// producer
	go func() {
		defer wg.Done()
		defer close(queue)
		for _, item := range items {
			select {
			case <-ctx.Done():
				fmt.Println("context cancelled")
				return
			case queue <- item: // blocks if the buffer is full (backpressure)
				fmt.Println("producer is sending: ", item)
			}
		}
	}()

	// consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("context cancelled")
				return

			case item, ok := <-queue:
				if !ok {
					return
				}
				fmt.Println("consumer is processing:", item)
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			}
		}
	}()

	wg.Wait()
	fmt.Println("Done")
}
