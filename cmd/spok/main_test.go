package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var binName = "spok"

func TestMain(m *testing.M) {
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	build := exec.Command("go", "build", "-o", binName)
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not compile spok: %s", err)
		os.Exit(1)
	}

	result := m.Run()

	os.Remove(binName)

	os.Exit(result)
}

// TestCLISmoke just tests a few core things on the CLI to ensure it's not
// totally broken e.g. does --help work, --version etc.
func TestCLISmoke(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cmdPath := filepath.Join(dir, binName)

	t.Run("--help", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "--help")

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})
}
