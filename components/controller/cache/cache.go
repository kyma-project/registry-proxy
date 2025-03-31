package cache

import (
	"sync"
)

type BoolCache interface {
	Set(value bool)
	Get() bool
}
type inMemoryBoolCache struct {
	value bool
	sync.Mutex
}

func NewInMemoryBoolCache() BoolCache {
	return &inMemoryBoolCache{}
}

func (c *inMemoryBoolCache) Set(value bool) {
	c.Lock()
	defer c.Unlock()

	c.value = value
}

func (c *inMemoryBoolCache) Get() bool {
	c.Lock()
	defer c.Unlock()

	return c.value
}
