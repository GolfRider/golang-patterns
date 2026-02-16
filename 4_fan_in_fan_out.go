package main

import (
	"context"
	"fmt"
	"sync"
)

func fanOut(ctx context.Context, in <-chan Task, n int) []<-chan Result {
	outChannels := make([]<-chan Result, n)
	for i := 0; i < n; i++ {
		outChannels[i] = transform(ctx, in)
	}
	return outChannels
}

func fanIn(ctx context.Context, in []<-chan Result) <-chan Result {
	out := make(chan Result)
	var wg sync.WaitGroup
	for _, ch := range in {
		wg.Add(1)
		go func(ch <-chan Result) {
			defer wg.Done()
			for v := range ch {
				select {
				case <-ctx.Done():
					return
				default:
					out <- v
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func checkFanInFanOut() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := make(chan Task)
	outChannels := fanOut(ctx, input, 4)
	finalChannel := fanIn(ctx, outChannels)

	go func() {
		input <- Task{1, "AA"}
		input <- Task{2, "BB"}
		input <- Task{3, "CC"}
		input <- Task{4, "DD"}
		input <- Task{5, "EE"}
		close(input)
	}()

	for v := range finalChannel {
		fmt.Println(v.Output, " || ")
	}
}
