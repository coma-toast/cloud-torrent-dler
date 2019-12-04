package main

import (
	"sync"

	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/jsonIo"
)

// Cache is the cache for storing episode lists
type Cache struct {
	Name  string
	path  string
	state map[int]string
	mutex *sync.Mutex
}

// Initialize the cache so it doesn't panic when trying to assign to the map when the map is nil
func (c *Cache) Initialize(path string) error {
	c.mutex = &sync.Mutex{}
	c.path = path
	cache.state = make(map[int]string)
	err := jsonIo.ReadFile(path, &cache.state)
	if err != nil {
		return err
	}

	return nil
}

// Set sets the cache
func (c *Cache) Set(key int, value string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.state[key] = value
	err := jsonIo.WriteFile(c.path, c.state)
	if err != nil {
		return err
	}

	return nil
}

// IsSet determines if the item exists already
func (c *Cache) IsSet(key int) bool {
	_, ok := c.state[key]

	return ok
}
