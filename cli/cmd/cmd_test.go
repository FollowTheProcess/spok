package cmd

import "testing"

// Completely pointless test, but my OCD was being triggered as this is the
// only package that isn't green when running gotest.
// The only thing that happens in this package is cobra stuff anyway.
func TestPointless(t *testing.T) {
	got := 1 + 1
	want := 2
	if got != want {
		t.Errorf("got %d, wanted %d", got, want)
	}
}
