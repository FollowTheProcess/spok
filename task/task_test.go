package task

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/google/go-cmp/cmp"
)

func TestExpandGlob(t *testing.T) {
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

func TestNewTask(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get cwd: %v", err)
	}
	testdata := filepath.Join(cwd, "testdata")
	tests := []struct {
		name    string
		want    Task
		in      ast.Task
		wantErr bool
	}{
		{
			name: "simple",
			want: Task{
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: nil,
				Commands:         []string{"go test ./..."},
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
			name: "simple with a file dependency",
			want: Task{
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: []string{"file.go"},
				Commands:         []string{"go test ./..."},
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
			name: "simple with a task dependency",
			want: Task{
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: []string{"fmt"},
				FileDependencies: nil,
				Commands:         []string{"go test ./..."},
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
			name: "simple with multi file dependency",
			want: Task{
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: []string{"file1.go", "file2.go"},
				Commands:         []string{"go test ./..."},
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
			name: "simple with multi task dependency",
			want: Task{
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: []string{"fmt", "lint"},
				FileDependencies: nil,
				Commands:         []string{"go test ./..."},
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
			name: "simple with double glob dependency",
			want: Task{
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
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
			name: "simple with single glob dependency",
			want: Task{
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: []string{mustAbs(testdata, "top.txt")},
				Commands:         []string{"go test ./..."},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newTask(tt.in, testdata) // Initialise root at testdata
			if (err != nil) != tt.wantErr {
				t.Fatalf("newTask() err = %v, wanted %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Task mismatch (-want +got):\n%s", diff)
			}
		})
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

// mustAbs returns the resolved 'path' or panics if it cannot.
func mustAbs(root, path string) string {
	abs, err := filepath.Abs(filepath.Join(root, path))
	if err != nil {
		panic(fmt.Sprintf("mustAbs could not resolve '%s'", filepath.Join(root, path)))
	}
	return abs
}
