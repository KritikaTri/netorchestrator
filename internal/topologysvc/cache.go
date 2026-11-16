package topologysvc

import (
	"container/list"
	"sync"
)

// lru is a small fixed-capacity, concurrency-safe LRU of response bodies.
type lru struct {
	mu  sync.Mutex
	cap int
	ll  *list.List
	m   map[string]*list.Element
}

type entry struct {
	key  string
	body []byte
}

func newLRU(capacity int) *lru {
	return &lru{cap: capacity, ll: list.New(), m: map[string]*list.Element{}}
}

func (c *lru) Get(k string) ([]byte, bool) {
	if c.cap <= 0 {
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.m[k]; ok {
		c.ll.MoveToFront(e)
		return e.Value.(*entry).body, true
	}
	return nil, false
}

func (c *lru) Put(k string, body []byte) {
	if c.cap <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.m[k]; ok {
		e.Value.(*entry).body = body
		c.ll.MoveToFront(e)
		return
	}
	c.m[k] = c.ll.PushFront(&entry{k, body})
	if c.ll.Len() > c.cap {
		last := c.ll.Back()
		c.ll.Remove(last)
		delete(c.m, last.Value.(*entry).key)
	}
}

func (c *lru) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}
