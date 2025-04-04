// Package cache implements spok's mechanism for storing and retrieving the
// cached SHA256 digest for a spok task.
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	Dir  string = ".spok"      // Dir is the directory under which the spok cache is kept
	File string = "cache.json" // File is filename of the spok cache file

	filePerms = 0o666 // filePerms is the file permissions for the spok cache file
	dirPerms  = 0o755 // dirPerms is the directory permissions for the spok cache directory
)

var Path = filepath.Join(Dir, File) // Path is the whole filepath to the spok cache file

// Cache represents the entire spok cache.
type Cache struct {
	inner map[string]string
}

// New creates and returns an empty cache.
func New() *Cache {
	return &Cache{inner: make(map[string]string)}
}

// Load reads in the current cache state from file.
func Load(path string) (*Cache, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cache := &Cache{}
	err = json.Unmarshal(contents, &cache.inner)
	if err != nil {
		return nil, err
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
	cache := New()
	for _, name := range names {
		cache.inner[name] = ""
	}

	if err := os.MkdirAll(filepath.Dir(path), dirPerms); err != nil {
		return err
	}

	if err := cache.Dump(path); err != nil {
		return err
	}

	if err := makeGitIgnore(filepath.Dir(path)); err != nil {
		return err
	}

	return makeCacheDirTag(filepath.Dir(path))
}

// Dump saves the cache to disk.
func (c *Cache) Dump(path string) error {
	contents, err := json.Marshal(c.inner)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, contents, filePerms)
	if err != nil {
		return fmt.Errorf("could not write spok cache at %q: %s", path, err)
	}
	return nil
}

// Get retrieves the digest value for a given name as well as
// a bool `ok` for whether or not it was found.
func (c *Cache) Get(name string) (string, bool) {
	digest, ok := c.inner[name]
	return digest, ok
}

// Set sets the digest value for a given name.
func (c *Cache) Set(name, digest string) {
	c.inner[name] = digest
}

// makeGitIgnore puts a .gitignore file in the .spok directory.
func makeGitIgnore(dir string) error {
	err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*\n"), filePerms)
	if err != nil {
		return err
	}
	return nil
}

// makeCacheDirTag creates the CACHEDIR.TAG file in the .spok directory.
func makeCacheDirTag(dir string) error {
	contents := []byte("Signature: 8a477f597d28d172789f06886806bc55")
	err := os.WriteFile(filepath.Join(dir, "CACHEDIR.TAG"), contents, filePerms)
	if err != nil {
		return err
	}
	return nil
}
