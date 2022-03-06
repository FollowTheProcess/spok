package builtins

import (
	"path/filepath"
	"testing"
)

func TestBuiltins(t *testing.T) {
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
