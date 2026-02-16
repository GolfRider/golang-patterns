package main

import (
	"container/heap"
	"fmt"
)

func checkTopK() {

	numbers := []int{1, 1, 1, 2, 2, 3, 4, 5, 5, 5, 5}
	freqMap := make(map[int]int)
	topK := 3

	for _, v := range numbers {
		freqMap[v]++
	}

	h := &MinHeap{}
	heap.Init(h)

	for k, v := range freqMap {
		heap.Push(h, Element{k, v})
		if h.Len() > topK {
			heap.Pop(h)
		}
	}

	result := make([]int, 0, topK)
	for h.Len() > 0 {
		item := heap.Pop(h).(Element).Number
		result = append(result, item)
	}

	for i := len(result) - 1; i >= 0; i-- {
		fmt.Println(result[i])
	}
}

type Element struct {
	Number int
	Count  int
}

type MinHeap []Element

func (h MinHeap) Len() int { return len(h) }

func (h MinHeap) Less(i, j int) bool { return h[i].Count < h[j].Count }
func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
func (h *MinHeap) Push(x any) {
	*h = append(*h, x.(Element))
}
func (h *MinHeap) Pop() any {
	top := (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]
	return top
}
