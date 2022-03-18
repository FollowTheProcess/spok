package task

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/google/go-cmp/cmp"
)

func TestExpandGlob(t *testing.T) {
	t.Parallel()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get cwd: %v", err)
	}

	root, err := filepath.Abs(filepath.Join(cwd, "testdata"))
	if err != nil {
		t.Fatalf("could not resolve root: %v", err)
	}

	got, err := expandGlob(root, "**/*.txt")
	if err != nil {
		t.Fatalf("expandGlob returned an error: %v", err)
	}

	want := []string{"top.txt", "sub1/sub2/blah.txt", "sub1/sub2/sub3/hello.txt", "suba/subb/stuff.txt", "suba/subb/subc/something.txt"}
	var wantAbs []string
	for _, w := range want {
		wantAbs = append(wantAbs, mustAbs(root, w))
	}
	if !reflect.DeepEqual(got, wantAbs) {
		t.Errorf("got %#v, wanted %#v", got, wantAbs)
	}
}

func TestExpandVars(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		vars    map[string]string
		command string
		want    string
	}{
		{
			name:    "test",
			vars:    map[string]string{"REPO": "https://github.com/FollowTheProcess/spok.git"},
			command: "git clone REPO",
			want:    "git clone https://github.com/FollowTheProcess/spok.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expandVars(tt.command, tt.vars); got != tt.want {
				t.Errorf("got %q, wanted %q", got, tt.want)
			}
		})
	}
}

func TestNewTask(t *testing.T) {
	t.Parallel()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get cwd: %v", err)
	}
	testdata := filepath.Join(cwd, "testdata")
	tests := []struct {
		name    string
		want    Task
		vars    map[string]string
		in      ast.Task
		wantErr bool
	}{
		{
			name: "simple",
			want: Task{
				Doc:               "A simple test task",
				Name:              "simple",
				NamedDependencies: nil,
				FileDependencies:  nil,
				Commands:          []string{"go test ./..."},
			},
			in: ast.Task{
				Name:         ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring:    ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{},
				Outputs:      []ast.Node{},
				Commands:     []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
				NodeType:     ast.NodeTask,
			},
			wantErr: false,
		},
		{
			name: "simple with vars",
			want: Task{
				Doc:               "A simple test task with global variables",
				Name:              "simple",
				NamedDependencies: nil,
				FileDependencies:  nil,
				Commands:          []string{"go test hello"},
			},
			in: ast.Task{
				Name:         ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring:    ast.Comment{Text: " A simple test task with global variables", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{},
				Outputs:      []ast.Node{},
				Commands:     []ast.Command{{Command: "go test GLOBAL", NodeType: ast.NodeCommand}},
				NodeType:     ast.NodeTask,
			},
			vars: map[string]string{
				"GLOBAL": "hello",
			},
			wantErr: false,
		},
		{
			name: "task with a file dependency",
			want: Task{
				Doc:               "A simple test task",
				Name:              "simple",
				NamedDependencies: nil,
				FileDependencies:  []string{"file.go"},
				Commands:          []string{"go test ./..."},
			},
			in: ast.Task{
				Name:         ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring:    ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{ast.String{Text: "file.go", NodeType: ast.NodeString}},
				Outputs:      []ast.Node{},
				Commands:     []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
				NodeType:     ast.NodeTask,
			},
			wantErr: false,
		},
		{
			name: "task with a named dependency",
			want: Task{
				Doc:               "A simple test task",
				Name:              "simple",
				NamedDependencies: []string{"fmt"},
				FileDependencies:  nil,
				Commands:          []string{"go test ./..."},
			},
			in: ast.Task{
				Name:         ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring:    ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{ast.Ident{Name: "fmt", NodeType: ast.NodeIdent}},
				Outputs:      []ast.Node{},
				Commands:     []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
				NodeType:     ast.NodeTask,
			},
			wantErr: false,
		},
		{
			name: "task with multi file dependency",
			want: Task{
				Doc:               "A simple test task",
				Name:              "simple",
				NamedDependencies: nil,
				FileDependencies:  []string{"file1.go", "file2.go"},
				Commands:          []string{"go test ./..."},
			},
			in: ast.Task{
				Name:      ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring: ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					ast.String{Text: "file1.go", NodeType: ast.NodeString},
					ast.String{Text: "file2.go", NodeType: ast.NodeString},
				},
				Outputs:  []ast.Node{},
				Commands: []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
				NodeType: ast.NodeTask,
			},
			wantErr: false,
		},
		{
			name: "task with multi task dependency",
			want: Task{
				Doc:               "A simple test task",
				Name:              "simple",
				NamedDependencies: []string{"fmt", "lint"},
				FileDependencies:  nil,
				Commands:          []string{"go test ./..."},
			},
			in: ast.Task{
				Name:      ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring: ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{
					ast.Ident{Name: "fmt", NodeType: ast.NodeIdent},
					ast.Ident{Name: "lint", NodeType: ast.NodeIdent},
				},
				Outputs:  []ast.Node{},
				Commands: []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
				NodeType: ast.NodeTask,
			},
			wantErr: false,
		},
		{
			name: "task with double glob dependency",
			want: Task{
				Doc:               "A simple test task",
				Name:              "simple",
				NamedDependencies: nil,
				FileDependencies: []string{
					mustAbs(testdata, "top.txt"),
					mustAbs(testdata, "sub1/sub2/blah.txt"),
					mustAbs(testdata, "sub1/sub2/sub3/hello.txt"),
					mustAbs(testdata, "suba/subb/stuff.txt"),
					mustAbs(testdata, "suba/subb/subc/something.txt"),
				},
				Commands: []string{"go test ./..."},
			},
			in: ast.Task{
				Name:         ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring:    ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{ast.String{Text: "**/*.txt", NodeType: ast.NodeString}},
				Outputs:      []ast.Node{},
				Commands:     []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
				NodeType:     ast.NodeTask,
			},
			wantErr: false,
		},
		{
			name: "task with single glob dependency",
			want: Task{
				Doc:               "A simple test task",
				Name:              "simple",
				NamedDependencies: nil,
				FileDependencies:  []string{mustAbs(testdata, "top.txt")},
				Commands:          []string{"go test ./..."},
			},
			in: ast.Task{
				Name:         ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring:    ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{ast.String{Text: "*.txt", NodeType: ast.NodeString}},
				Outputs:      []ast.Node{},
				Commands:     []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
				NodeType:     ast.NodeTask,
			},
			wantErr: false,
		},
		{
			name: "task with single file output",
			want: Task{
				Doc:               "A simple test task",
				Name:              "simple",
				NamedDependencies: nil,
				FileDependencies:  nil,
				Commands:          []string{"go test ./..."},
				NamedOutputs:      nil,
				FileOutputs:       []string{"file.go"},
			},
			in: ast.Task{
				Name:         ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring:    ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{},
				Outputs:      []ast.Node{ast.String{Text: "file.go", NodeType: ast.NodeString}},
				Commands:     []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
				NodeType:     ast.NodeTask,
			},
			wantErr: false,
		},
		{
			name: "task with glob output",
			want: Task{
				Doc:               "A simple test task",
				Name:              "simple",
				NamedDependencies: nil,
				FileDependencies:  nil,
				Commands:          []string{"go test ./..."},
				NamedOutputs:      nil,
				FileOutputs: []string{
					mustAbs(testdata, "top.txt"),
					mustAbs(testdata, "sub1/sub2/blah.txt"),
					mustAbs(testdata, "sub1/sub2/sub3/hello.txt"),
					mustAbs(testdata, "suba/subb/stuff.txt"),
					mustAbs(testdata, "suba/subb/subc/something.txt"),
				},
			},
			in: ast.Task{
				Name:         ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring:    ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{},
				Outputs:      []ast.Node{ast.String{Text: "**/*.txt", NodeType: ast.NodeString}},
				Commands:     []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
				NodeType:     ast.NodeTask,
			},
			wantErr: false,
		},
		{
			name: "task with multi file output",
			want: Task{
				Doc:               "A simple test task",
				Name:              "simple",
				NamedDependencies: nil,
				FileDependencies:  nil,
				Commands:          []string{"go test ./..."},
				NamedOutputs:      nil,
				FileOutputs:       []string{"file1.go", "file2.go"},
			},
			in: ast.Task{
				Name:         ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring:    ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{},
				Outputs: []ast.Node{
					ast.String{Text: "file1.go", NodeType: ast.NodeString},
					ast.String{Text: "file2.go", NodeType: ast.NodeString},
				},
				Commands: []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
				NodeType: ast.NodeTask,
			},
			wantErr: false,
		},
		{
			name: "task with single named output",
			want: Task{
				Doc:               "A simple test task",
				Name:              "simple",
				NamedDependencies: nil,
				FileDependencies:  nil,
				Commands:          []string{"go test ./..."},
				NamedOutputs:      []string{"OUT"},
				FileOutputs:       nil,
			},
			in: ast.Task{
				Name:         ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring:    ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{},
				Outputs:      []ast.Node{ast.Ident{Name: "OUT", NodeType: ast.NodeIdent}},
				Commands:     []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
				NodeType:     ast.NodeTask,
			},
			wantErr: false,
		},
		{
			name: "task with multi named output",
			want: Task{
				Doc:               "A simple test task",
				Name:              "simple",
				NamedDependencies: nil,
				FileDependencies:  nil,
				Commands:          []string{"go test ./..."},
				NamedOutputs:      []string{"OUT", "OTHER"},
				FileOutputs:       nil,
			},
			in: ast.Task{
				Name:         ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring:    ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{},
				Outputs: []ast.Node{
					ast.Ident{Name: "OUT", NodeType: ast.NodeIdent},
					ast.Ident{Name: "OTHER", NodeType: ast.NodeIdent},
				},
				Commands: []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
				NodeType: ast.NodeTask,
			},
			wantErr: false,
		},
		{
			name: "complex task with everything",
			want: Task{
				Doc:               "Very complex things here",
				Name:              "complex",
				NamedDependencies: nil,
				FileDependencies: []string{
					mustAbs(testdata, "top.txt"),
					mustAbs(testdata, "sub1/sub2/blah.txt"),
					mustAbs(testdata, "sub1/sub2/sub3/hello.txt"),
					mustAbs(testdata, "suba/subb/stuff.txt"),
					mustAbs(testdata, "suba/subb/subc/something.txt"),
				},
				Commands:     []string{"go build ."},
				NamedOutputs: nil,
				FileOutputs:  []string{"./bin/main"},
			},
			in: ast.Task{
				Name:         ast.Ident{Name: "complex", NodeType: ast.NodeIdent},
				Docstring:    ast.Comment{Text: " Very complex things here", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{ast.String{Text: "**/*.txt", NodeType: ast.NodeString}},
				Outputs:      []ast.Node{ast.String{Text: "./bin/main", NodeType: ast.NodeString}},
				Commands:     []ast.Command{{Command: "go build .", NodeType: ast.NodeCommand}},
				NodeType:     ast.NodeTask,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.in, testdata, tt.vars) // Initialise root at testdata
			if (err != nil) != tt.wantErr {
				t.Fatalf("newTask() err = %v, wanted %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Task mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHashFiles(t *testing.T) {
	files := []string{"test", "file", "me", "too"}
	open := func(file string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("I'm some content for " + file)), nil
	}
	want := "99602d7b885eb92aef160ad5933c12e426534314"
	got, err := hashFiles(files, open)
	if err != nil {
		t.Fatalf("hashFiles returned an error: %v", err)
	}

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestHashFileDeps(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "TestHashFileDeps")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err = os.WriteFile(filepath.Join(dir, "test"), []byte("I'm some content for test"), 0666); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "file"), []byte("I'm some content for file"), 0666); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "me"), []byte("I'm some content for me"), 0666); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "too"), []byte("I'm some content for too"), 0666); err != nil {
		t.Fatal(err)
	}

	files := []string{filepath.Join(dir, "test"), filepath.Join(dir, "file"), filepath.Join(dir, "me"), filepath.Join(dir, "too")}

	want := "99602d7b885eb92aef160ad5933c12e426534314"
	got, err := HashFiles(files)
	if err != nil {
		t.Fatalf("hashFiles returned an error: %v", err)
	}

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func BenchmarkHashFileDeps(b *testing.B) {
	dir, err := os.MkdirTemp(os.TempDir(), "BenchmarkHashFileDeps")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "test"), []byte("I'm some content for test"), 0666); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file"), []byte("I'm some content for file"), 0666); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "me"), []byte("I'm some content for me"), 0666); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "too"), []byte("I'm some content for too"), 0666); err != nil {
		b.Fatal(err)
	}

	files := []string{filepath.Join(dir, "test"), filepath.Join(dir, "file"), filepath.Join(dir, "me"), filepath.Join(dir, "too")}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := HashFiles(files)
		if err != nil {
			b.Fatalf("hashFiles returned an error: %v", err)
		}
	}
}

func BenchmarkExpandGlob(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("could not get cwd: %v", err)
	}

	root, err := filepath.Abs(filepath.Join(cwd, "testdata"))
	if err != nil {
		b.Fatalf("could not resolve root: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = expandGlob(root, "**/*.txt")
		if err != nil {
			b.Fatalf("expandGlob returned an error: %v", err)
		}
	}
}

func BenchmarkNewTask(b *testing.B) {
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("could not get cwd: %v", err)
	}

	root, err := filepath.Abs(filepath.Join(cwd, "testdata"))
	if err != nil {
		b.Fatalf("could not resolve root: %v", err)
	}

	input := ast.Task{
		Name:         ast.Ident{Name: "complex", NodeType: ast.NodeIdent},
		Docstring:    ast.Comment{Text: " Very complex things here", NodeType: ast.NodeComment},
		Dependencies: []ast.Node{ast.String{Text: "**/*.txt", NodeType: ast.NodeString}},
		Outputs:      []ast.Node{ast.String{Text: "./bin/main", NodeType: ast.NodeString}},
		Commands:     []ast.Command{{Command: "go build .", NodeType: ast.NodeCommand}},
		NodeType:     ast.NodeTask,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := New(input, root, make(map[string]string))
		if err != nil {
			b.Fatalf("newTask returned an error: %v", err)
		}
	}
}

// mustAbs returns the resolved 'path' or panics if it cannot.
func mustAbs(root, path string) string {
	abs, err := filepath.Abs(filepath.Join(root, path))
	if err != nil {
		panic(fmt.Sprintf("mustAbs could not resolve '%s'", filepath.Join(root, path)))
	}
	return abs
}
