package attribs

import (
	"reflect"
	"sync"
)

type Cache interface {
	Set(value reflect.Type, attr Attribute) error
	Get(value reflect.Type) (Attribute, bool)
}

func newCache() Cache {
	return &cache{
		attrs: make(map[reflect.Type]Attribute),
	}
}

type cache struct {
	attrs map[reflect.Type]Attribute
	mutex sync.RWMutex
}

func (c *cache) Set(value reflect.Type, attr Attribute) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.attrs[value] = attr
	return nil
}

func (c *cache) Get(value reflect.Type) (Attribute, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	attr, ok := c.attrs[value]
	return attr, ok
}
