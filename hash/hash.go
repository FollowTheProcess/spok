// Package hash implements a concurrent file hasher used by spok
// to detect when task dependencies have changed.
//
// The hasher opens, reads, hashes contents and filepath, and closes each file with a sha256
// digest, these are then collected in an order independent way and summed into an overall
// sha256 hash sum representing the state of all the hashed files.
//
// The fully qualified filepath is included in each files hash so if any part of this
// changes this will count as a change as well as any change to file contents.
package hash

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"sync"
)

// ALWAYS is a constant string that is different to the string returned from
// the AlwaysRun hasher to be used as the cache comparison value so that a task
// is always re-run regardless of it's dependencies.
const ALWAYS = "ALWAYS"

// Hasher is the interface representing something capable of hashing
// a number of files into a final digest.
type Hasher interface {
	// Hash takes a list of filepaths and returns a hash digest.
	Hash(files []string) (string, error)
}

// AlwaysRun is a Hasher that always returns a constant value for the final digest: "DIFFERENT",
// it is used primarily when forcing tasks to re-run regardless of the state
// of their dependencies e.g. (spok <tasks...> --force).
type AlwaysRun struct{}

// Hash implements Hasher for AlwaysRun and returns the constant string "DIFFERENT"
// in place of an actual hash digest. Comparing the return of this method
// to the "ALWAYS" constant will result in the digests never matching and therefore
// tasks that use this will always run regardless of the state of their dependencies.
func (a AlwaysRun) Hash(_ []string) (string, error) {
	return "DIFFERENT", nil
}

// result encodes the result of a concurrent hashing operation on
// a single file, to be passed around on channels.
type result struct {
	err  error  // Any error encountered while hashing
	file string // The filepath being hashed
	hash []byte // The hash of the contents and filepath
}

// Concurrent is primary file hasher used by spok.
type Concurrent struct{}

// New creates and returns a new file hasher.
func New() Concurrent {
	return Concurrent{}
}

// Hash takes a list of absolute filepaths to hash contents and paths for,
// these are then hashed and combined into an SHA256 digest which is returned
// as a hex encoded string along with any errors encountered.
func (c Concurrent) Hash(files []string) (string, error) {
	jobs := make(chan string)
	results := make(chan result)

	// Keep a wait group so we can track when all the workers are done
	var wg sync.WaitGroup

	// Launch a concurrent worker pool to chew through the queue of files to hash
	// these will all initially block as no files are on the jobs channel yet
	// nWorkers is min of NumCPU and len(files) so we don't start more workers than
	// is necessary (no point kicking off 8 workers to do 3 files for example)
	nWorkers := min(runtime.NumCPU(), len(files))
	for i := 0; i < nWorkers; i++ {
		wg.Add(1)
		go worker(results, jobs, &wg)
	}

	// Put files on the jobs channel, this is a goroutine so
	// it doesn't block the main goroutine as channel cap is 0
	go func() {
		for _, file := range files {
			jobs <- file
		}
		close(jobs)
	}()

	// Wait for all the workers to finish in another goroutine so
	// it doesn't block, and close results channel when done
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(results)
	}(&wg)

	// Finally, range over the results channel until it gets closed
	// by the goroutine above, adding each hash to the accumulator which
	// will then itself be sorted and hashed to form the overall digest
	var accumulator [][]byte
	var errors []error
	for r := range results {
		// Accumulating errors as no matter what we'll need to range over the results
		// channel to drain it
		if r.err != nil {
			errors = append(errors, fmt.Errorf("Could not get hash result for %s: %w", r.file, r.err))
		}

		// Include the filepath in the hash so a rename counts as a change
		hashItem := [][]byte{r.hash, []byte(r.file)}
		joinedHashItem := []byte(bytes.Join(hashItem, []byte(""))) //nolint: unconvert
		accumulator = append(accumulator, joinedHashItem)
	}

	if len(errors) != 0 {
		// Any error here is pretty much a dealbreaker so we just bail out
		// on the first one
		return "", errors[0]
	}

	// Because files may be larger/smaller, take longer to hash etc. the order
	// of results is non-deterministic so to generate a deterministic hash we must
	// sort the accumulator prior to generating the final digest
	sort.Stable(sortByteSlices(accumulator))
	digest := bytes.Join(accumulator, []byte(""))
	hash := sha256.New()
	hash.Write(digest)
	sum := hex.EncodeToString(hash.Sum(nil))

	return sum, nil
}

// worker is a concurrent worker contributing to hashing a number of files,
// it pulls files off the files channel, hashes them, and puts the results
// on the results channel. It takes a reference to a WaitGroup so we can be
// sure all the workers have finished before closing the results channel.
func worker(results chan<- result, files <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for file := range files {
		var res result
		res.file = file
		f, err := os.Open(file)
		if err != nil {
			res.err = err
		}
		info, _ := f.Stat() //nolint: errcheck // The file is already open here so we can ignore the error
		// Skip directories
		if info.IsDir() {
			continue
		}
		hash := sha256.New()
		_, err = io.Copy(hash, f)
		f.Close()
		if err != nil {
			res.err = err
		}
		res.hash = hash.Sum(nil)

		results <- res
	}
}

// sortByteSlices satisfies the sort.Sort interface for nested byte slices.
type sortByteSlices [][]byte

// Len returns the length of the outer slice.
func (b sortByteSlices) Len() int {
	return len(b)
}

// Less returns whether i should be sorted less than j.
func (b sortByteSlices) Less(i, j int) bool {
	return bytes.Compare(b[i], b[j]) == -1
}

// Swap swaps byte slice i with byte slice j in the outer slice.
func (b sortByteSlices) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}
