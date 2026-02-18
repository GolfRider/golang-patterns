package main

import (
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	checkContextAndError()
	checkWorkerPool()
	checkPipeline()
	checkFanInFanOut()
	checkGracefulShutdown()
	checkRetry()
	checkCircuitBreaker()
	startHttpServer()
	checkSingleFlight()
	checkErrorGroupWithParallelWork()
	checkSyncPool()
	checkCache()
	checkSyncMap()
	checkTopK()
	checkSlidingWindowTopK()
	checkProducerConsumerBackpressure()
}

func main1() {
	rateLimiter := rate.NewLimiter(5, 10) // 5 tokens/sec, up to 10 burst

	for i := 1; i <= 15; i++ {
		if rateLimiter.Allow() {
			fmt.Println("request", i, "at", time.Now().Format("15:04:05.000"))
		} else {
			fmt.Println("[rate-limited]: request", i, "at", time.Now().Format("15:04:05.000"))
			time.Sleep(time.Second)
		}
	}
}
