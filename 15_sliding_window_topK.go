package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

func getAlignment(now time.Time, bucketDuration time.Duration, bucketsCount int) (int, int64) {
	tick := now.UnixNano() / bucketDuration.Nanoseconds()
	index := int(tick % int64(bucketsCount))
	return index, tick
}

type Bucket struct {
	items    map[string]int
	lastTick int64 // to know if items belongs to current window or not
}

type SlidingWindowTopK struct {
	mx             sync.RWMutex
	buckets        []Bucket
	bucketCount    int
	bucketDuration time.Duration
}

func NewSlidingWindowTopK(bucketCount int, bucketDuration time.Duration) *SlidingWindowTopK {
	sw := &SlidingWindowTopK{
		bucketDuration: bucketDuration,
		bucketCount:    bucketCount,
		buckets:        make([]Bucket, bucketCount),
	}
	for i := 0; i < bucketCount; i++ {
		sw.buckets[i].items = make(map[string]int)
	}
	return sw
}

func (s *SlidingWindowTopK) Add(key string) {
	now := time.Now()
	index, tick := getAlignment(now, s.bucketDuration, s.bucketCount)

	s.mx.Lock()
	defer s.mx.Unlock()
	if s.buckets[index].lastTick != tick {
		s.buckets[index].items = make(map[string]int)
		s.buckets[index].lastTick = tick
	}

	s.buckets[index].items[key]++
}

func (s *SlidingWindowTopK) TopK(k int) []string {
	s.mx.RLock()
	defer s.mx.RUnlock()
	now := time.Now()
	_, currentTick := getAlignment(now, s.bucketDuration, s.bucketCount)
	resultMap := make(map[string]int)
	for _, bucket := range s.buckets {
		if currentTick-bucket.lastTick < int64(s.bucketCount) {
			for key, count := range bucket.items {
				resultMap[key] = resultMap[key] + count
			}
		}
	}
	return topK(resultMap, k)
}

// Note: Heap/PQ can also be used here
func topK(resultMap map[string]int, k int) []string {
	type kv struct {
		key string
		val int
	}
	pairs := make([]kv, 0, len(resultMap))
	for k, v := range resultMap {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].val > pairs[j].val
	})
	var results []string
	for idx, pair := range pairs {
		if idx < k {
			results = append(results, pair.key)
		}
	}
	return results
}

func checkSlidingWindowTopK() {
	sw := NewSlidingWindowTopK(5, time.Second) // 5-sec sliding window
	sw.Add("jerseys")
	sw.Add("shoes")
	sw.Add("shoes")
	sw.Add("hats")
	sw.Add("hats")

	fmt.Printf("Top-3 items: %v \n", sw.TopK(3))

}
