package builtins

import (
	"path/filepath"
	"testing"
)

func TestBuiltins(t *testing.T) {
	t.Parallel()
	tests := []struct {
		fn      Builtin
		name    string
		want    string
		args    []string
		wantErr bool
	}{
		{
			fn:      join,
			name:    "join",
			want:    filepath.Join("hello", "filepath", "parts"),
			args:    []string{"hello", "filepath", "parts"},
			wantErr: false,
		},
		{
			fn:      execute,
			name:    "exec",
			want:    "hello",
			args:    []string{"echo hello"}, // exec takes a single string
			wantErr: false,
		},
		{
			fn:      execute,
			name:    "exec more than 1 arg",
			want:    "",
			args:    []string{"echo hello", "uh oh"},
			wantErr: true,
		},
		{
			fn:      execute,
			name:    "exec non-zero exit code",
			want:    "",
			args:    []string{"exit 1"},
			wantErr: true,
		},
		{
			fn:      execute,
			name:    "exec single arg",
			want:    "",
			args:    []string{"echo"},
			wantErr: false,
		},
		{
			fn:      execute,
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
	_, ok := Get("exec")
	if !ok {
		t.Fatal("Get failed to retrieve 'exec' which is known to exist")
	}

	_, ok = Get("dingle")
	if ok {
		t.Fatal("Get returned true for getting 'dingle' which doesn't exist")
	}
}
