// Package iostream provides convenient wrappers around things like stdout, stderr
// and enables spok to easily talk to a variety of readers and writers.
package iostream

import (
	"bytes"
	"io"
	"os"
)

// IOStream is an object containing io.Writers for spok to talk to.
type IOStream struct {
	Stdout io.Writer
	Stderr io.Writer
}

// OS returns an IOStream configured to talk to the OS streams.
func OS() IOStream {
	return IOStream{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// Test returns an IOStream configured to talk to temporary buffers
// that can then be read from to verify output.
func Test() IOStream {
	return IOStream{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
}

// Null returns an IOStream configured to discard all output.
func Null() IOStream {
	return IOStream{
		Stdout: io.Discard,
		Stderr: io.Discard,
	}
}
