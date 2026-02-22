package main

import (
	"container/list"
	"fmt"
)

type entry struct {
	key, value string
}

type cache struct {
	mm       map[string]*list.Element
	orderLL  *list.List
	capacity int
}

func NewCache(capacity int) *cache {
	return &cache{
		mm:       make(map[string]*list.Element),
		orderLL:  list.New(),
		capacity: capacity,
	}
}

func (c *cache) Add(key, value string) {
	if val, ok := c.mm[key]; ok {
		val.Value.(*entry).value = value
		c.orderLL.MoveToFront(val)
		return
	}

	c.mm[key] = c.orderLL.PushFront(&entry{key, value})

	if len(c.mm) > c.capacity { // greater than capacity?
		if lru := c.orderLL.Back(); lru != nil {
			delete(c.mm, lru.Value.(*entry).key)
			c.orderLL.Remove(lru)
		}
	}
}

func (c *cache) Get(key string) string {
	if val, ok := c.mm[key]; ok {
		c.orderLL.MoveToFront(val)
		return val.Value.(*entry).value
	}
	return ""
}

func (c *cache) Size() int {
	return len(c.mm)
}

func checkCache() {
	lruCache := NewCache(4)
	lruCache.Add("1", "1")
	lruCache.Add("2", "2")
	lruCache.Add("3", "3")
	lruCache.Add("4", "4")
	lruCache.Add("5", "5")
	lruCache.Add("5", "6")
	lruCache.Add("7", "7")

	fmt.Println(lruCache.Get("1") == "")
	fmt.Println(lruCache.Get("2"))
	fmt.Println(lruCache.Size())
}
