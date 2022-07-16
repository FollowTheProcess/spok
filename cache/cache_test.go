package cache

import (
	"bytes"
	"os"
	"testing"
)

func TestString(t *testing.T) {
	tests := []struct {
		name  string
		cache Cache
		want  string
	}{
		{
			name:  "empty",
			cache: map[string]string{},
			want:  "",
		},
		{
			name: "single",
			cache: map[string]string{
				"test": "4440044af910a451502908de30e986fb97cdacf6",
			},
			want: "test\t4440044af910a451502908de30e986fb97cdacf6",
		},
		{
			name: "multiple",
			cache: map[string]string{
				"test": "4440044af910a451502908de30e986fb97cdacf6",
				"docs": "254ac0c1e553c18be4c7baa82eba5a6293cec2c8",
				"lint": "106dfe5fcfa96f4b70036d0e3f9ac1c126f03175",
			},
			want: "docs\t254ac0c1e553c18be4c7baa82eba5a6293cec2c8\nlint\t106dfe5fcfa96f4b70036d0e3f9ac1c126f03175\ntest\t4440044af910a451502908de30e986fb97cdacf6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cache.String(); got != tt.want {
				t.Errorf("Got %q\nWanted %q", got, tt.want)
			}
		})
	}
}

func TestBytes(t *testing.T) {
	tests := []struct {
		name  string
		cache Cache
		want  []byte
	}{
		{
			name:  "empty",
			cache: map[string]string{},
			want:  []byte(""),
		},
		{
			name: "single",
			cache: map[string]string{
				"test": "4440044af910a451502908de30e986fb97cdacf6",
			},
			want: []byte("test\t4440044af910a451502908de30e986fb97cdacf6"),
		},
		{
			name: "multiple",
			cache: map[string]string{
				"test": "4440044af910a451502908de30e986fb97cdacf6",
				"docs": "254ac0c1e553c18be4c7baa82eba5a6293cec2c8",
				"lint": "106dfe5fcfa96f4b70036d0e3f9ac1c126f03175",
			},
			want: []byte("docs\t254ac0c1e553c18be4c7baa82eba5a6293cec2c8\nlint\t106dfe5fcfa96f4b70036d0e3f9ac1c126f03175\ntest\t4440044af910a451502908de30e986fb97cdacf6"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cache.Bytes(); !bytes.Equal(got, tt.want) {
				t.Errorf("Got %q\nWanted %q", string(got), string(tt.want))
			}
		})
	}
}

func TestGetPut(t *testing.T) {
	cache := New()
	cache.Put("test", "4440044af910a451502908de30e986fb97cdacf6")
	cache.Put("docs", "254ac0c1e553c18be4c7baa82eba5a6293cec2c8")

	test, ok := cache.Get("test")
	if !ok {
		t.Error("test not found in cache")
	}
	if test != "4440044af910a451502908de30e986fb97cdacf6" {
		t.Errorf("Wrong value for test.\nGot %q\nWant %q", test, "4440044af910a451502908de30e986fb97cdacf6")
	}

	docs, ok := cache.Get("docs")
	if !ok {
		t.Error("docs not found in cache")
	}
	if docs != "254ac0c1e553c18be4c7baa82eba5a6293cec2c8" {
		t.Errorf("Wrong value for docs.\nGot %q\nWant %q", test, "254ac0c1e553c18be4c7baa82eba5a6293cec2c8")
	}
}

func TestWrite(t *testing.T) {
	tmp, err := os.CreateTemp("", "cache")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	tmp.Close()
	defer os.RemoveAll(tmp.Name())

	cache := New()
	cache.Put("test", "4440044af910a451502908de30e986fb97cdacf6")
	cache.Put("docs", "254ac0c1e553c18be4c7baa82eba5a6293cec2c8")

	if err = cache.Write(tmp.Name()); err != nil {
		t.Fatalf("Write returned an error: %v", err)
	}

	contents, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatalf("Could not read temp file: %v", err)
	}

	want := "docs\t254ac0c1e553c18be4c7baa82eba5a6293cec2c8\ntest\t4440044af910a451502908de30e986fb97cdacf6"

	if string(contents) != want {
		t.Errorf("Wrong file contents\nGot %q, Want %q", string(contents), want)
	}
}

func TestRead(t *testing.T) {
	tmp, err := os.CreateTemp("", "cache")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	_, err = tmp.WriteString("docs\t254ac0c1e553c18be4c7baa82eba5a6293cec2c8\ntest\t4440044af910a451502908de30e986fb97cdacf6")
	if err != nil {
		t.Fatalf("Could not write to temp file: %v", err)
	}
	tmp.Close()
	defer os.RemoveAll(tmp.Name())

	cache := New()
	if err = cache.Load(tmp.Name()); err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}

	if len(cache) != 2 {
		t.Errorf("Wrong number of entries in the cache\nGot %d, Want %d", len(cache), 2)
	}

	test, ok := cache.Get("test")
	if !ok {
		t.Error("test not found in cache")
	}
	if test != "4440044af910a451502908de30e986fb97cdacf6" {
		t.Errorf("Wrong value for test.\nGot %q\nWant %q", test, "4440044af910a451502908de30e986fb97cdacf6")
	}

	docs, ok := cache.Get("docs")
	if !ok {
		t.Error("docs not found in cache")
	}
	if docs != "254ac0c1e553c18be4c7baa82eba5a6293cec2c8" {
		t.Errorf("Wrong value for docs.\nGot %q\nWant %q", test, "254ac0c1e553c18be4c7baa82eba5a6293cec2c8")
	}
}
