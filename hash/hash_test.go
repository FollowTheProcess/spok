package hash_test

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/FollowTheProcess/spok/hash"
)

const (
	darwin       = "darwin"
	darwinDigest = "0a9ccbd9e6c1db74e78c4c7a7b77c2d0c7853f3b046db24fdc164f4d589cd5cd"
	linux        = "linux"
	linuxDigest  = "a2a890074f4edea78c7f6cb0dd2d129410e4cf9bf9897e475cbecdf6be72936c"
	windows      = "windows"
)

func TestAlwaysHasher(t *testing.T) {
	t.Parallel()
	hasher := hash.AlwaysRun{}
	got, _ := hasher.Hash([]string{"doesn't", "matter"}) //nolint: errcheck // We don't care about the error here
	if got != "DIFFERENT" {
		t.Errorf("got %s, wanted %s", got, "DIFFERENT")
	}
}

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
	case darwin:
		original = darwinDigest
	case linux:
		original = linuxDigest
	case windows:
		// Some weirdness where the test would seemingly randomly fail despite the hash
		// being correct
		t.Skip("Skipped on Windows")
	default:
		t.Skipf("Unsupported platform: %s", runtime.GOOS)
	}

	if digest == original {
		t.Error("Digest did not respond to different file contents")
	}
}

func TestHashSkipsDirectories(t *testing.T) {
	files, cleanup := makeFilesWithDirectory(t)
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
	case darwin:
		original = darwinDigest
	case linux:
		original = linuxDigest
	case windows:
		// Some weirdness where the test would seemingly randomly fail despite the hash
		// being correct
		t.Skip("Skipped on Windows")
	default:
		t.Skipf("Unsupported platform: %s", runtime.GOOS)
	}

	// Here we've added a directory, so this test is identical to the one above
	// but if we try and open the directory we will get an error so if this test
	// fails with an "is a directory" error, we know we haven't ignored dirs
	// which is the behaviour we're testing for here

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
	case darwin:
		original = darwinDigest
	case linux:
		original = linuxDigest
	case windows:
		// Some weirdness where the test would seemingly randomly fail despite the hash
		// being correct
		t.Skip("Skipped on Windows")
	default:
		t.Skipf("Unsupported platform: %s", runtime.GOOS)
	}

	if digest == original {
		t.Error("Digest did not respond to different file names")
	}
}

func TestMin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{
			name: "a",
			a:    2,
			b:    10,
			want: 2,
		},
		{
			name: "b",
			a:    10,
			b:    2,
			want: 2,
		},
		{
			name: "equal",
			a:    10,
			b:    10,
			want: 10,
		},
		{
			name: "negative",
			a:    -2,
			b:    10,
			want: -2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := min(tt.a, tt.b); got != tt.want {
				t.Errorf("got %d, wanted %d", got, tt.want)
			}
		})
	}
}

// makeFiles makes a load of fake files under /tmp/hashfiles returning
// a slice of their filepaths and a function to clean up the entire dir.
func makeFiles(t *testing.T) ([]string, func()) {
	t.Helper()

	tmp := os.TempDir()
	path := filepath.Join(tmp, fmt.Sprintf("hashfiles-%s", randSeq()))
	err := os.Mkdir(path, 0o755)
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
	path := filepath.Join(tmp, fmt.Sprintf("hashfiles-%s", randSeq()))
	err := os.Mkdir(path, 0o755)
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

func makeFilesWithDirectory(t *testing.T) ([]string, func()) {
	t.Helper()

	tmp := os.TempDir()
	path := filepath.Join(tmp, fmt.Sprintf("hashfiles-%s", randSeq()))
	err := os.Mkdir(path, 0o755)
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

	tmpdir, err := os.MkdirTemp(path, "dir")
	if err != nil {
		t.Fatalf("Could not create temp dir: %v", err)
	}

	files = append(files, tmpdir)

	cleanup := func() { _ = os.RemoveAll(path) }

	return files, cleanup
}

// makeFilesDifferentName is like makeFiles but one of the files will have
// a different name.
func makeFilesDifferentName(t *testing.T) ([]string, func()) {
	t.Helper()

	tmp := os.TempDir()
	path := filepath.Join(tmp, fmt.Sprintf("hashfiles-%s", randSeq()))
	err := os.Mkdir(path, 0o755)
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
	t.Helper()
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

// randSeq generates a random string of length n.
func randSeq() string {
	rand.NewSource(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
