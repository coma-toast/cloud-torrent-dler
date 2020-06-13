package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"gitlab.jasondale.me/jdale/cloud-torrent-dler/pkg/jsonIo"
)

// Cache is the cache for storing episode lists
type Cache struct {
	Name  string
	path  string
	state map[string]DownloadItem
	mutex *sync.Mutex
}

// Initialize the cache so it doesn't panic when trying to assign to the map when the map is nil
func (c *Cache) Initialize(path string) error {
	c.mutex = &sync.Mutex{}
	c.path = path
	cache.state = make(map[string]DownloadItem)
	err := jsonIo.ReadFile(path, &cache.state)
	if err != nil {
		return err
	}

	return nil
}

// Set sets the cache
func (c *Cache) Set(key string, value DownloadItem) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	key = c.sanitizeString(key)
	c.state[key] = value
	err := jsonIo.WriteFile(c.path, c.state)
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves data from the cache
func (c *Cache) Get(key string) DownloadItem {
	key = c.sanitizeString(key)
	fmt.Printf("Getting %s from cache\n", key)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.state[key]
}

// IsSet determines if the item exists already
func (c *Cache) IsSet(key string) bool {
	key = c.sanitizeString(key)
	_, ok := c.state[key]

	return ok
}

func (c *Cache) sanitizeString(input string) string {
	input = strings.ToLower(input)
	regex, err := regexp.Compile("[^a-z0-9\\s]+")
	if err != nil {
		log.Println(err)
	}

	output := regex.ReplaceAllString(input, "")
	fmt.Println("input", input)
	fmt.Println("output", output)

	return output
}
