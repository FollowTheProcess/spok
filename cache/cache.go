// Package cache implements spok's mechanism for storing and retrieving the
// cached SHA256 digest for a spok task.
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var (
	Dir  = ".spok"                  // Dir is the directory under which the spok cache is kept
	File = "cache.json"             // File is filename of the spok cache file
	Path = filepath.Join(Dir, File) // Path is the whole filepath to the spok cache file
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

// Init populates the entire .spok cache directory and writes a placeholder
// cache file containing the names of all the tasks but no digests.
func Init(path string, names ...string) error {
	cache := make(Cache, 0, len(names))
	for _, name := range names {
		cache = append(cache, Entry{Name: name})
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if err := cache.Dump(path); err != nil {
		return err
	}

	if err := makeGitIgnore(filepath.Dir(path)); err != nil {
		return err
	}

	return nil
}

// Dump saves the cache to disk.
func (c Cache) Dump(path string) error {
	contents, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(path, contents, 0666)
	if err != nil {
		return fmt.Errorf("Could not write spok cache at %q: %s", path, err)
	}
	return nil
}

// Get retrieves the digest value for a given name as well as
// a bool `ok` for whether or not it was found.
func (c Cache) Get(name string) (string, bool) {
	// In general not a huge fan of this because we have to loop over the whole
	// list of entries in order to find one but we don't know what the user's tasks
	// will be called ahead of time and using a map here is clunky and error prone

	// Range is okay here as we only want to read the variable, not modify it
	for _, entry := range c {
		if entry.Name == name {
			return entry.Digest, true
		}
	}
	return "", false
}

// Set sets the digest value for a given name.
func (c Cache) Set(name, digest string) {
	// Same comment as above re. O(n) time required to lookup/set a value

	// Can't use range here as values are copied into a range statement preventing
	// modification in-place
	for i := 0; i < len(c); i++ {
		if c[i].Name == name {
			c[i].Digest = digest
		}
	}
}

// makeGitIgnore puts a .gitignore file in the .spok directory.
func makeGitIgnore(dir string) error {
	err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*\n"), 0666)
	if err != nil {
		return err
	}
	return nil
}
