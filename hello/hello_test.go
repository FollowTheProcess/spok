package hello

import "testing"

func TestHello(t *testing.T) {
	want := "Hello thing"
	got := Say("thing")

	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}
