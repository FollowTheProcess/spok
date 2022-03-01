package spok

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFind(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get cwd: %v", err)
	}

	t.Run("found spokfile", func(t *testing.T) {
		start := filepath.Join(cwd, "testdata", "suba", "subb", "subc") // Start deep inside testdata
		stop := filepath.Join(cwd, "testdata")                          // Stop at testdata

		want, err := filepath.Abs(filepath.Join(cwd, "testdata", "suba", "spokfile"))
		if err != nil {
			t.Fatal("could not resolve want")
		}

		path, err := find(start, stop)
		if err != nil {
			t.Fatalf("find returned an error: %v", err)
		}

		if path != want {
			t.Errorf("got %s, wanted %s", path, want)
		}
	})

	t.Run("missing spokfile", func(t *testing.T) {
		start := filepath.Join(cwd, "testdata", "sub1", "sub2", "sub3")
		stop := filepath.Join(cwd, "testdata")

		_, err := find(start, stop)
		if err == nil {
			t.Fatal("expected ErrNoSpokfile, got nil")
		}

		if !errors.Is(err, errNoSpokfile) {
			t.Errorf("wrong error, got %v, wanted %v", err, errNoSpokfile)
		}
	})
}
