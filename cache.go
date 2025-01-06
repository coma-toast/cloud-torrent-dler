package main

import (
	"path/filepath"
	"sync"
)

// Cache is the cache for storing episode lists
type Cache struct {
	Name              string
	path              string
	state             map[string]DownloadItem
	autodownloadItems map[string]AutoDownload
	mutex             *sync.RWMutex
}

type Data struct {
	State             map[string]DownloadItem `json:"state"`
	AutodownloadItems map[string]AutoDownload `json:"auto_download"`
}

// Initialize the cache so it doesn't panic when trying to assign to the map when the map is nil
func (c *Cache) Initialize(path string) error {
	filePath := filepath.Join(path, "cache.json")
	c.mutex = &sync.RWMutex{}
	c.path = filePath
	cache.state = make(map[string]DownloadItem)
	cache.autodownloadItems = make(map[string]AutoDownload)
	data := Data{}
	err := jsonio.ReadFile(filePath, &data)
	if err != nil {
		c.write()
		return err
	}
	c.state = data.State
	c.autodownloadItems = data.AutodownloadItems
	c.write()

	return nil
}

// Set sets the cache
func (c *Cache) Set(key string, value DownloadItem) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	key = helper.SanitizeText(key)
	// fmt.Printf("Setting cache: %s\n", key)
	c.state[key] = value
	err := c.write()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	key = helper.SanitizeText(key)
	delete(c.state, key)
	err := c.write()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.state = make(map[string]DownloadItem)
	c.autodownloadItems = make(map[string]AutoDownload)
	err := c.write()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) write() error {
	cacheData := Data{State: c.state, AutodownloadItems: c.autodownloadItems}
	err := jsonio.WriteFile(c.path, cacheData)
	return err
}

func (c *Cache) SetAutoDownload(id string, autodownload AutoDownload) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.autodownloadItems[id] = autodownload
	err := c.write()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) GetAutoDownload(torrentHash string) AutoDownload {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.autodownloadItems[torrentHash]
}

func (c *Cache) RemoveAutoDownload(torrentHash string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.autodownloadItems, torrentHash)

	err := c.write()
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves data from the cache
func (c *Cache) Get(key string) DownloadItem {
	key = helper.SanitizeText(key)
	// fmt.Printf("Getting cache: %s\n", key)
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.state[key]
}

// GetAll retrieves all data from the cache
func (c *Cache) GetAll() map[string]DownloadItem {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.state
}

// IsSet determines if the item exists already
func (c *Cache) IsSet(key string) bool {
	key = helper.SanitizeText(key)
	_, ok := c.state[key]

	return ok
}
