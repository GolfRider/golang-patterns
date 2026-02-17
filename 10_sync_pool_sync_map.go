package main

import (
	"fmt"
	"sync"
	"time"
)

type Record struct {
	Id   int
	Name string
}

func (rec *Record) reset() {
	rec.Id = 0
	rec.Name = ""
}

var pool = sync.Pool{
	New: func() interface{} {
		fmt.Println("creating new record from pool")
		return &Record{}
	},
}

func getStruct() {
	rec := pool.Get().(*Record)

	rec.Id = 245
	rec.Name = "test_245"
	fmt.Printf("rec.Id:%d\n", rec.Id)
	time.Sleep(100 * time.Millisecond)

	rec.reset()
	pool.Put(rec)
}

func checkSyncPool() {
	for i := 0; i < 10; i++ {
		getStruct() // object gets reused on subsequent iterations
	}
}

func checkSyncMap() {
	var syncMap sync.Map // concurrent hash-map, tuned for read-heavy workload
	syncMap.Store("key1", "1")
	syncMap.Store("key2", "2")

	if val, ok := syncMap.Load("key1"); ok {
		fmt.Printf("found key1-val: %s\n", val)
	}
	actual, loaded := syncMap.LoadOrStore("key100", "Tahoe")
	fmt.Printf("Key100: Value=%v; IsPresent=%v\n", actual, loaded)
}
