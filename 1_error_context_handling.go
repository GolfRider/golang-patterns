package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// 1 - Error Handling
func checkError() {
	err := doWork()
	var ae *AppErrror
	if err != nil && errors.As(err, &ae) {
		fmt.Println("Error=>>", ae.Msg, err.Error())
	}
}

func doWork() error {
	time.Sleep(1 * time.Second)
	return fmt.Errorf("oops some error: %w", &AppErrror{ID: 101, Msg: "This is an error"})
}

type AppErrror struct {
	ID  int
	Msg string
}

func (ap *AppErrror) Error() string {
	return ap.Msg
}

func checkContext() {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go func(cx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-cx.Done():
				fmt.Println("Context Cancelled/Timeout: ", cx.Err())
				return

			default:
				// do work
				fmt.Println("Working... (context-example)")
				time.Sleep(200 * time.Millisecond)
			}
		}

	}(ctx)

	wg.Wait()
}

func checkContextAndError() {
	checkError()
	checkContext()
}
