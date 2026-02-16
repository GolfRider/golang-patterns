package main

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

func doHeavyWork(jobId int) (string, error) {
	time.Sleep(1 * time.Second)
	fmt.Printf("Job #%d finished\n", jobId)
	return fmt.Sprintf("job finished: %v", jobId), nil
}

func checkSingleFlight() {
	var sf singleflight.Group
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err, shared := sf.Do("", func() (any, error) {
				return doHeavyWork(1)
			})
			if err != nil {
				panic(err)
			}
			fmt.Println("FINAL: ", v, err, shared)
		}()
	}
	wg.Wait()
}
