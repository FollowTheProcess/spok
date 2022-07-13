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
	t.Parallel()
	testdata := mustGetTestData()

	got, err := expandGlob(testdata, "**/*.txt")
	if err != nil {
		t.Fatalf("expandGlob returned an error: %v", err)
	}

	want := []string{"top.txt", "sub1/sub2/blah.txt", "sub1/sub2/sub3/hello.txt", "suba/subb/stuff.txt", "suba/subb/subc/something.txt"}
	var wantAbs []string
	for _, w := range want {
		wantAbs = append(wantAbs, mustAbs(testdata, w))
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
			name:    "simple",
			vars:    map[string]string{"REPO": "https://github.com/FollowTheProcess/spok.git"},
			command: "git clone {{.REPO}}",
			want:    "git clone https://github.com/FollowTheProcess/spok.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandVars(tt.command, tt.vars)
			if err != nil {
				t.Errorf("expandVars returned an error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, wanted %q", got, tt.want)
			}
		})
	}
}

func TestNewTask(t *testing.T) {
	t.Parallel()
	testdata := mustGetTestData()
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
			name: "simple with vars",
			want: Task{
				Doc:              "A simple test task with global variables",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: nil,
				Commands:         []string{"go test hello"},
			},
			in: ast.Task{
				Name:         ast.Ident{Name: "simple", NodeType: ast.NodeIdent},
				Docstring:    ast.Comment{Text: " A simple test task with global variables", NodeType: ast.NodeComment},
				Dependencies: []ast.Node{},
				Outputs:      []ast.Node{},
				Commands:     []ast.Command{{Command: "go test {{.GLOBAL}}", NodeType: ast.NodeCommand}},
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
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: []string{mustAbs(testdata, "file.go")},
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
			name: "task with a named dependency",
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
			name: "task with multi file dependency",
			want: Task{
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: []string{
					mustAbs(testdata, "file1.go"),
					mustAbs(testdata, "file2.go"),
				},
				Commands: []string{"go test ./..."},
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
			name: "task with double glob dependency",
			want: Task{
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: nil,
				GlobDependencies: []string{"**/*.txt"},
				Commands:         []string{"go test ./..."},
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
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: nil,
				GlobDependencies: []string{"*.txt"},
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
		{
			name: "task with single file output",
			want: Task{
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: nil,
				Commands:         []string{"go test ./..."},
				NamedOutputs:     nil,
				FileOutputs:      []string{mustAbs(testdata, "file.go")},
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
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: nil,
				Commands:         []string{"go test ./..."},
				NamedOutputs:     nil,
				FileOutputs:      nil,
				GlobOutputs:      []string{"**/*.txt"},
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
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: nil,
				Commands:         []string{"go test ./..."},
				NamedOutputs:     nil,
				FileOutputs: []string{
					mustAbs(testdata, "file1.go"),
					mustAbs(testdata, "file2.go"),
				},
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
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: nil,
				Commands:         []string{"go test ./..."},
				NamedOutputs:     []string{"OUT"},
				FileOutputs:      nil,
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
				Doc:              "A simple test task",
				Name:             "simple",
				TaskDependencies: nil,
				FileDependencies: nil,
				Commands:         []string{"go test ./..."},
				NamedOutputs:     []string{"OUT", "OTHER"},
				FileOutputs:      nil,
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
				Doc:              "Very complex things here",
				Name:             "complex",
				TaskDependencies: nil,
				FileDependencies: nil,
				GlobDependencies: []string{"**/*.txt"},
				Commands:         []string{"go build ."},
				NamedOutputs:     nil,
				FileOutputs:      []string{mustAbs(testdata, "./bin/main")},
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

func BenchmarkExpandGlob(b *testing.B) {
	testdata := mustGetTestData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := expandGlob(testdata, "**/*.txt")
		if err != nil {
			b.Fatalf("expandGlob returned an error: %v", err)
		}
	}
}

func BenchmarkNewTask(b *testing.B) {
	testdata := mustGetTestData()

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
		_, err := New(input, testdata, make(map[string]string))
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

// mustGetTestData returns the absolute path to this packages testdata folder
// and panics if it cannot.
func mustGetTestData() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path, err := filepath.Abs(filepath.Join(cwd, "testdata"))
	if err != nil {
		panic(err)
	}

	return path
}
