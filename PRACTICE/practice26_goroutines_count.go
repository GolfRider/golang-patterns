package main

import (
	"fmt"
	"runtime"
	"time"
)

func practice26() {

	fmt.Println("CPU# : ", runtime.NumCPU())
	go func() {
		time.Sleep(1 * time.Second)
	}()
	go func() {
		time.Sleep(1 * time.Second)
	}()
	fmt.Println("G# ", runtime.NumGoroutine())
}
