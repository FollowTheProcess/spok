// Package cache implements spok's mechanism for storing and retrieving the
// cached SHA256 digest for a spok task.
package cache

import (
	"encoding/json"
	"os"
)

// Entry represents a single cache entry.
type Entry struct {
	Name   string `json:"name"`   // Name of the entry (e.g. task name)
	Digest string `json:"digest"` // The SHA256 digest
}

// Cache represents the entire spok cache.
type Cache []Entry

// Load reads in the current cache state from file.
func Load(path string) (Cache, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Cache{}, err
	}

	cache := Cache{}
	err = json.Unmarshal(contents, &cache)
	if err != nil {
		return Cache{}, err
	}

	return cache, nil
}

// Exists returns whether or not the spok cache exists at all, e.g.
// if spok has not been run before.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Dump saves the cache to disk.
func (c Cache) Dump(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	contents, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(path, contents, 0666)
	if err != nil {
		return err
	}
	return nil
}
