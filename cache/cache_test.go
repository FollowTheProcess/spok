package cache_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"go.followtheprocess.codes/spok/cache"
)

const cacheText string = `
{
	"testtask": "02f15ca4e81f467b84267f82eef52277b4cc29ee71d2f5b9f8b3ada6711b2537",
	"another": "3703972e88411fdc03c96659d3943fa45b363562cbd909ebbfe9f305e4ba572b"
}`

func TestLoad(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		file, cleanup := makeCache(t, cacheText)
		defer cleanup()

		cached, err := cache.Load(file.Name())
		if err != nil {
			t.Fatalf("cache.Load returned an error: %v", err)
		}

		want := cache.New()
		want.Set("testtask", "02f15ca4e81f467b84267f82eef52277b4cc29ee71d2f5b9f8b3ada6711b2537")
		want.Set("another", "3703972e88411fdc03c96659d3943fa45b363562cbd909ebbfe9f305e4ba572b")

		if !reflect.DeepEqual(cached, want) {
			t.Errorf("got %#v, wanted %#v", cached, want)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := cache.Load("missing.json")
		if err == nil {
			t.Fatal("Expected an error but got nil")
		}
	})

	t.Run("bad json", func(t *testing.T) {
		file, cleanup := makeCache(t, "I'm not JSON")
		defer cleanup()

		_, err := cache.Load(file.Name())
		if err == nil {
			t.Fatal("Expected an error but got nil")
		}
	})
}

func TestDump(t *testing.T) {
	cached := cache.New()
	cached.Set("testtask", "02f15ca4e81f467b84267f82eef52277b4cc29ee71d2f5b9f8b3ada6711b2537")
	cached.Set("another", "3703972e88411fdc03c96659d3943fa45b363562cbd909ebbfe9f305e4ba572b")

	file, err := os.CreateTemp("", ".cache.json")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	defer os.RemoveAll(file.Name())

	if err = cached.Dump(file.Name()); err != nil {
		t.Fatalf("cache.Dump returned an error: %v", err)
	}

	loaded, err := cache.Load(file.Name())
	if err != nil {
		t.Fatalf("cache.Load return an error: %v", err)
	}

	if !reflect.DeepEqual(cached, loaded) {
		t.Errorf("got %#v, wanted %#v", loaded, cached)
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "missing",
			path: "/not/here.txt",
			want: false,
		},
		{
			name: "present",
			path: "cache.go",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cache.Exists(tt.path); got != tt.want {
				t.Errorf("got %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestSet(t *testing.T) {
	cached := cache.New()
	cached.Set("testtask", "02f15ca4e81f467b84267f82eef52277b4cc29ee71d2f5b9f8b3ada6711b2537")
	cached.Set("another", "3703972e88411fdc03c96659d3943fa45b363562cbd909ebbfe9f305e4ba572b")

	cached.Set("another", "something else")
	cached.Set("something", "I'm new here")

	another, ok := cached.Get("another")
	if !ok {
		t.Fatal("another was not in the cache")
	}
	something, ok := cached.Get("something")
	if !ok {
		t.Fatal("something was not in the cache")
	}

	if another != "something else" {
		t.Error("another was not 'something else'")
	}
	if something != "I'm new here" {
		t.Error("something was not 'Im new here'")
	}
}

func TestGet(t *testing.T) {
	cached := cache.New()
	cached.Set("testtask", "02f15ca4e81f467b84267f82eef52277b4cc29ee71d2f5b9f8b3ada6711b2537")
	cached.Set("another", "3703972e88411fdc03c96659d3943fa45b363562cbd909ebbfe9f305e4ba572b")

	got, ok := cached.Get("testtask")
	if !ok {
		t.Fatalf("cache.Get unexpected 'ok' value %v, expected %v", ok, true)
	}
	if got != "02f15ca4e81f467b84267f82eef52277b4cc29ee71d2f5b9f8b3ada6711b2537" {
		t.Errorf("Retrieved value for 'testtask' wrong: got %s, expected %s", got, "02f15ca4e81f467b84267f82eef52277b4cc29ee71d2f5b9f8b3ada6711b2537")
	}
}

func TestInit(t *testing.T) {
	tmp, err := os.MkdirTemp("", "spoktemp")
	if err != nil {
		t.Fatalf("Could not create temp dir: %v", err)
	}
	defer os.RemoveAll(tmp)

	if err = cache.Init(filepath.Join(tmp, cache.Dir, cache.File), "one", "two", "three"); err != nil {
		t.Fatalf("cache.Init returned an error: %v", err)
	}

	gitIgnorePath := filepath.Join(tmp, cache.Dir, ".gitignore")
	if !exists(gitIgnorePath) {
		t.Errorf(".gitignore not found at %s", gitIgnorePath)
	}

	cachePath := filepath.Join(tmp, cache.Dir, cache.File)
	if !exists(cachePath) {
		t.Errorf("cache.json not found at %s", cachePath)
	}

	dirTagPath := filepath.Join(tmp, cache.Dir, "CACHEDIR.TAG")
	if !exists(dirTagPath) {
		t.Errorf("CACHEDIR.TAG not found at %s", dirTagPath)
	}

	loaded, err := cache.Load(cachePath)
	if err != nil {
		t.Fatalf("Could not load cache: %v", err)
	}

	want := cache.New()
	want.Set("one", "")
	want.Set("two", "")
	want.Set("three", "")

	if !reflect.DeepEqual(loaded, want) {
		t.Errorf("Got %#v, wanted %#v", loaded, want)
	}
}

// makeCache writes a cache JSON to a temporary file, returning it
// and a cleanup function to be deferred.
func makeCache(t *testing.T, text string) (*os.File, func()) {
	t.Helper()
	file, err := os.CreateTemp("", "cache.json")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}

	_, err = file.WriteString(text)
	if err != nil {
		t.Fatalf("Could not write to cache file: %v", err)
	}

	cleanup := func() { _ = os.RemoveAll(file.Name()) }

	return file, cleanup
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
