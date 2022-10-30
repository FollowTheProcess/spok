package cache_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/FollowTheProcess/spok/cache"
)

const cacheText string = `
[
	{
		"name": "testtask",
		"digest": "02f15ca4e81f467b84267f82eef52277b4cc29ee71d2f5b9f8b3ada6711b2537"
	},
	{
		"name": "another",
		"digest": "3703972e88411fdc03c96659d3943fa45b363562cbd909ebbfe9f305e4ba572b"
	}
]
`

func TestLoad(t *testing.T) {
	file, cleanup := makeCache(t)
	defer cleanup()

	cached, err := cache.Load(file.Name())
	if err != nil {
		t.Fatalf("cache.Load returned an error: %v", err)
	}

	want := cache.Cache{
		{
			Name:   "testtask",
			Digest: "02f15ca4e81f467b84267f82eef52277b4cc29ee71d2f5b9f8b3ada6711b2537",
		},
		{
			Name:   "another",
			Digest: "3703972e88411fdc03c96659d3943fa45b363562cbd909ebbfe9f305e4ba572b",
		},
	}

	if !reflect.DeepEqual(cached, want) {
		t.Errorf("got %#v, wanted %#v", cached, want)
	}
}

func TestDump(t *testing.T) {
	cached := cache.Cache{
		{
			Name:   "testtask",
			Digest: "02f15ca4e81f467b84267f82eef52277b4cc29ee71d2f5b9f8b3ada6711b2537",
		},
		{
			Name:   "another",
			Digest: "3703972e88411fdc03c96659d3943fa45b363562cbd909ebbfe9f305e4ba572b",
		},
	}

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

// makeCache writes a cache JSON to a temporary file, returning it
// and a cleanup function to be deferred.
func makeCache(t *testing.T) (*os.File, func()) {
	file, err := os.CreateTemp("", ".cache.json")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}

	_, err = file.WriteString(cacheText)
	if err != nil {
		t.Fatalf("Could not write to cache file: %v", err)
	}

	cleanup := func() { _ = os.RemoveAll(file.Name()) }

	return file, cleanup
}
