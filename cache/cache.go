// Package cache implements spok's cache for file dependency hashes.
//
// Each entry is a key value pair of task name to it's dependency hash sum,
// the file itself is a plain text, line separated file where each pair is delimited with
// a single tab character e.g. "test\t<hash>"
package cache

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	dir  = ".spok" // Name of spok's cache dir
	file = "cache" // Name of the cache file itself
)

// Cache represents the entire cache of task name to hash sum.
type Cache struct {
	inner map[string]string
}

// New builds and returns a new cache.
func New() *Cache {
	return &Cache{inner: make(map[string]string)}
}

// Init creates the .spok directory relative to root and writes an empty cache file.
func Init(root string) error {
	path, err := filepath.Abs(filepath.Join(root, dir, file))
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	_, err = os.Create(path)
	if err != nil {
		return err
	}
	return nil
}

// IsEmpty returns whether or not the cache file is empty.
func IsEmpty(root string) bool {
	path, err := filepath.Abs(filepath.Join(root, dir, file))
	if err != nil {
		return false
	}
	file, err := os.Stat(path)
	if err != nil {
		return false
	}

	return file.Size() == 0
}

// Exists returns whether or not the cache file exists.
func Exists(root string) bool {
	path, err := filepath.Abs(filepath.Join(root, dir, file))
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// String implements Stringer for a Cache.
func (c Cache) String() string {
	lines := make([]string, 0, len(c.inner))
	for task, hash := range c.inner {
		line := task + "\t" + hash
		lines = append(lines, line)
	}
	sort.Strings(lines)
	content := strings.Join(lines, "\n")
	return content
}

// Bytes returns the contents of the cache as a raw byte slice
// ready to be written to a file.
func (c Cache) Bytes() []byte {
	return []byte(c.String())
}

// Write writes the state of the cache to a file, if the
// file already exists it will be truncated (contents replaced).
func (c Cache) Write(file string) error {
	if err := os.WriteFile(file, c.Bytes(), 0666); err != nil {
		return fmt.Errorf("Could not write to cache file %s: %w", file, err)
	}
	return nil
}

// Put puts a task name and hash sum into the cache to
// be later written.
func (c *Cache) Put(task, digest string) {
	c.inner[task] = digest
}

// Get gets the digest for a task.
func (c *Cache) Get(task string) (string, bool) {
	digest, ok := c.inner[task]
	return digest, ok
}

// Load reads the cache file into the caller.
func (c *Cache) Load(file string) error {
	contents, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Could not read from cache file %s: %w", file, err)
	}

	lines := bytes.Split(contents, []byte("\n"))
	for _, line := range lines {
		parts := bytes.Split(line, []byte("\t"))
		if len(parts) != 2 {
			return fmt.Errorf("Malformed cache line: %q", string(line))
		}
		task, digest := parts[0], parts[1]
		c.Put(string(task), string(digest))
	}
	return nil
}
