package spok

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/google/go-cmp/cmp"
)

func TestFind(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get cwd: %v", err)
	}

	t.Run("found spokfile", func(t *testing.T) {
		start := filepath.Join(cwd, "testdata", "suba", "subb", "subc") // Start deep inside testdata
		stop := filepath.Join(cwd, "testdata")                          // Stop at testdata

		want, err := filepath.Abs(filepath.Join(cwd, "testdata", "suba", "spokfile"))
		if err != nil {
			t.Fatal("could not resolve want")
		}

		path, err := find(start, stop)
		if err != nil {
			t.Fatalf("find returned an error: %v", err)
		}

		if path != want {
			t.Errorf("got %s, wanted %s", path, want)
		}
	})

	t.Run("missing spokfile", func(t *testing.T) {
		start := filepath.Join(cwd, "testdata", "sub1", "sub2", "sub3")
		stop := filepath.Join(cwd, "testdata")

		_, err := find(start, stop)
		if err == nil {
			t.Fatal("expected ErrNoSpokfile, got nil")
		}

		if !errors.Is(err, errNoSpokfile) {
			t.Errorf("wrong error, got %v, wanted %v", err, errNoSpokfile)
		}
	})
}

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

	want := []string{"sub1/sub2/blah.txt", "sub1/sub2/sub3/hello.txt", "suba/subb/stuff.txt", "suba/subb/subc/something.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, wanted %#v", got, want)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newTask(tt.in, testdata)
			if (err != nil) != tt.wantErr {
				t.Fatalf("newTask() err = %v, wanted %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Function mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
