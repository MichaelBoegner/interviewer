package jdsummary

import (
	"sync"
)

type Cache struct {
	store map[int]string
	mu    sync.RWMutex
}

var JDCache *Cache

func Init() {
	JDCache = &Cache{
		store: make(map[int]string),
	}
}

func (c *Cache) Set(interviewID int, summary string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[interviewID] = summary
}

func (c *Cache) Get(interviewID int) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	summary, ok := c.store[interviewID]
	return summary, ok
}

func (c *Cache) Delete(interviewID int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, interviewID)
}
