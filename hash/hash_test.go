package hash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/FollowTheProcess/spok/hash"
)

// Test that the final hash digest is repeatable.
func TestHashFilesIsDeterministic(t *testing.T) {
	files, cleanup := makeFiles(t)
	defer cleanup()

	hasher := hash.New()
	// Run the hasher a number of times and see if the output varies
	runs := 100

	digests := make([]string, 0, runs)
	for i := 0; i < runs; i++ {
		digest, err := hasher.Hash(files)
		if err != nil {
			t.Fatalf("Hash returned an error: %v", err)
		}
		digests = append(digests, digest)
	}

	// Ensure that we don't get any drift across runs
	first := digests[0]
	if first == "" {
		t.Fatal("First generated digest was empty")
	}

	for i, digest := range digests {
		if digest != first {
			t.Errorf("Digest at index %d not correct. Got %q, wanted %q", i, digest, first)
		}
	}
}

// Test that the final digest responds to different content in a single file.
func TestHashDifferentContents(t *testing.T) {
	files, cleanup := makeFilesDifferentContent(t)
	defer cleanup()

	hasher := hash.New()

	digest, err := hasher.Hash(files)
	if err != nil {
		t.Fatalf("Hash returned an error: %v", err)
	}

	// Because the filepath is used in the hash, and the /tmp dir is different on different
	// platforms, the hashes will be different for each, but repeatable on each
	var original string
	switch runtime.GOOS {
	case "darwin":
		original = "0a9ccbd9e6c1db74e78c4c7a7b77c2d0c7853f3b046db24fdc164f4d589cd5cd"
	case "linux":
		original = "a2a890074f4edea78c7f6cb0dd2d129410e4cf9bf9897e475cbecdf6be72936c"
	case "windows":
		original = "replace me"
	default:
		t.Skipf("Unsupported platform: %s", runtime.GOOS)
	}

	if digest == original {
		t.Error("Digest did not respond to different file contents")
	}
}

// Test that the final digest responds to a change in filename.
func TestHashDifferentName(t *testing.T) {
	files, cleanup := makeFilesDifferentName(t)
	defer cleanup()

	hasher := hash.New()

	digest, err := hasher.Hash(files)
	if err != nil {
		t.Fatalf("Hash returned an error: %v", err)
	}

	// Because the filepath is used in the hash, and the /tmp dir is different on different
	// platforms, the hashes will be different for each, but repeatable on each
	var original string
	switch runtime.GOOS {
	case "darwin":
		original = "0a9ccbd9e6c1db74e78c4c7a7b77c2d0c7853f3b046db24fdc164f4d589cd5cd"
	case "linux":
		original = "a2a890074f4edea78c7f6cb0dd2d129410e4cf9bf9897e475cbecdf6be72936c"
	case "windows":
		original = "replace me"
	default:
		t.Skipf("Unsupported platform: %s", runtime.GOOS)
	}

	if digest == original {
		t.Error("Digest did not respond to different file names")
	}
}

// makeFiles makes a load of fake files under /tmp/hashfiles returning
// a slice of their filepaths and a function to clean up the entire dir.
func makeFiles(t *testing.T) ([]string, func()) {
	t.Helper()

	tmp := os.TempDir()
	path := filepath.Join(tmp, "hashfiles")
	err := os.Mkdir(path, 0755)
	if err != nil {
		t.Fatalf("Could not create hashfiles dir under /tmp: %v", err)
	}

	contents := []string{
		"hello",
		"there",
		"general",
		"kenobi",
		"I'm",
		"some files",
		"hash me baby",
		"what's my hash",
		"some slightly longer content akshdbakhsdviaysvdiqhwvd8723t8127t3871t2e",
	}

	var files []string

	for i, content := range contents {
		files = append(files, makeFile(t, path, i, content))
	}

	cleanup := func() { _ = os.RemoveAll(path) }

	return files, cleanup
}

// makeFilesDifferentContent is like makeFiles but one file has slightly
// different contents.
func makeFilesDifferentContent(t *testing.T) ([]string, func()) {
	t.Helper()

	tmp := os.TempDir()
	path := filepath.Join(tmp, "hashfiles")
	err := os.Mkdir(path, 0755)
	if err != nil {
		t.Fatalf("Could not create hashfiles dir under /tmp: %v", err)
	}

	contents := []string{
		"hello",
		"there",
		"general",
		"kenobi",
		"I'm",
		"some slightly different files", // The different one!
		"hash me baby",
		"what's my hash",
		"some slightly longer content akshdbakhsdviaysvdiqhwvd8723t8127t3871t2e",
	}

	var files []string

	for i, content := range contents {
		files = append(files, makeFile(t, path, i, content))
	}

	cleanup := func() { _ = os.RemoveAll(path) }

	return files, cleanup
}

// makeFilesDifferentName is like makeFiles but one of the files will have
// a different name.
func makeFilesDifferentName(t *testing.T) ([]string, func()) {
	t.Helper()

	tmp := os.TempDir()
	path := filepath.Join(tmp, "hashfiles")
	err := os.Mkdir(path, 0755)
	if err != nil {
		t.Fatalf("Could not create hashfiles dir under /tmp: %v", err)
	}

	contents := []string{
		"hello",
		"there",
		"general",
		"kenobi", // This file will have a different name
		"I'm",
		"some slightly different files",
		"hash me baby",
		"what's my hash",
		"some slightly longer content akshdbakhsdviaysvdiqhwvd8723t8127t3871t2e",
	}

	var files []string

	for i, content := range contents {
		key := i
		// Change the name on the 4th file
		if i == 3 {
			key = 82625
		}
		files = append(files, makeFile(t, path, key, content))
	}

	cleanup := func() { _ = os.RemoveAll(path) }

	return files, cleanup
}

// makeFile is a helper that creates a single temporary file with a key
// with some content written to it returning it's path.
func makeFile(t *testing.T, dir string, key int, content string) string {
	path := filepath.Join(dir, fmt.Sprintf("%d.txt", key))
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Could not create tmp file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("Could not write to tmp file: %v", err)
	}
	return file.Name()
}
