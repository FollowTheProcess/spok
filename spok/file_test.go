package spok

import (
	"errors"
	"os"
	"path/filepath"
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

func TestFromTree(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get cwd: %v", err)
	}
	testdata := filepath.Join(cwd, "testdata")
	tests := []struct {
		name    string
		tree    ast.Tree
		want    File
		wantErr bool
	}{
		{
			name: "just a task",
			tree: ast.Tree{
				Nodes: []ast.Node{
					ast.Task{
						Name:      ast.Ident{Name: "test", NodeType: ast.NodeIdent},
						Docstring: ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
						Commands:  []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
						NodeType:  ast.NodeTask,
					},
				},
			},
			want: File{
				Path: testdata,
				Vars: make(map[string]string),
				Tasks: []Task{
					{
						Doc:      "A simple test task",
						Name:     "test",
						Commands: []string{"go test ./..."},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "just a task no docstring",
			tree: ast.Tree{
				Nodes: []ast.Node{
					ast.Task{
						Name:     ast.Ident{Name: "test", NodeType: ast.NodeIdent},
						Commands: []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
						NodeType: ast.NodeTask,
					},
				},
			},
			want: File{
				Path: testdata,
				Vars: make(map[string]string),
				Tasks: []Task{
					{
						Name:     "test",
						Commands: []string{"go test ./..."},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "just some globals",
			tree: ast.Tree{
				Nodes: []ast.Node{
					ast.Assign{
						Value:    ast.String{Text: "hello", NodeType: ast.NodeString},
						Name:     ast.Ident{Name: "global1", NodeType: ast.NodeIdent},
						NodeType: ast.NodeAssign,
					},
					ast.Assign{
						Value:    ast.String{Text: "hello again", NodeType: ast.NodeString},
						Name:     ast.Ident{Name: "global2", NodeType: ast.NodeIdent},
						NodeType: ast.NodeAssign,
					},
				},
			},
			want: File{
				Path: testdata,
				Vars: map[string]string{"global1": "hello", "global2": "hello again"},
			},
			wantErr: false,
		},
		{
			name: "globals with join builtin",
			tree: ast.Tree{
				Nodes: []ast.Node{
					ast.Assign{
						Value: ast.Function{
							Name: ast.Ident{Name: "join", NodeType: ast.NodeIdent},
							Arguments: []ast.Node{
								ast.String{Text: "path", NodeType: ast.NodeString},
								ast.String{Text: "parts", NodeType: ast.NodeString},
								ast.String{Text: "more", NodeType: ast.NodeString},
							},
							NodeType: ast.NodeFunction,
						},
						Name:     ast.Ident{Name: "global1", NodeType: ast.NodeIdent},
						NodeType: ast.NodeAssign,
					},
				},
			},
			want: File{
				Path: testdata,
				Vars: map[string]string{"global1": filepath.Join("path", "parts", "more")},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fromTree(tt.tree, testdata)
			if (err != nil) != tt.wantErr {
				t.Fatalf("fromTree() err = %v, wantErr = %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("File mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
