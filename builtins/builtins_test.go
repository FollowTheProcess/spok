package builtins_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"go.followtheprocess.codes/spok/builtins"
)

func TestBuiltins(t *testing.T) {
	t.Parallel()
	tests := []struct {
		fn      builtins.Builtin
		name    string
		want    string
		args    []string
		wantErr bool
	}{
		{
			fn:      mustGet("join"),
			name:    "join",
			want:    abs(filepath.Join("hello", "filepath", "parts")),
			args:    []string{"hello", "filepath", "parts"},
			wantErr: false,
		},
		{
			fn:      mustGet("exec"),
			name:    "exec",
			want:    "hello",
			args:    []string{"echo hello"}, // exec takes a single string
			wantErr: false,
		},
		{
			fn:      mustGet("exec"),
			name:    "exec more than 1 arg",
			want:    "",
			args:    []string{"echo hello", "uh oh"},
			wantErr: true,
		},
		{
			fn:      mustGet("exec"),
			name:    "exec non-zero exit code",
			want:    "",
			args:    []string{"exit 1"},
			wantErr: true,
		},
		{
			fn:      mustGet("exec"),
			name:    "exec single arg",
			want:    "",
			args:    []string{"echo"},
			wantErr: false,
		},
		{
			fn:      mustGet("exec"),
			name:    "bad syntax",
			want:    "",
			args:    []string{"(*^$$"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fn(tt.args...)

			if (err != nil) != tt.wantErr {
				t.Fatalf("%s returned an error: %v", tt.name, err)
			}

			if got != tt.want {
				t.Errorf("got %s, wanted %s", got, tt.want)
			}
		})
	}
}

func TestGet(t *testing.T) {
	t.Parallel()
	_, ok := builtins.Get("exec")
	if !ok {
		t.Fatal("Get failed to retrieve 'exec' which is known to exist")
	}

	_, ok = builtins.Get("dingle")
	if ok {
		t.Fatal("Get returned true for getting 'dingle' which doesn't exist")
	}
}

// abs returns the absolute path of the given path, panicking
// if it cannot do so for whatever reason.
func abs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return abs
}

// Gets a builtin and panics if it's not there.
func mustGet(fn string) builtins.Builtin {
	f, ok := builtins.Get(fn)
	if !ok {
		panic(fmt.Sprintf("builtin %s not found", fn))
	}
	return f
}
