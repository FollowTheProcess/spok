package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/FollowTheProcess/spok/shell"
	"github.com/FollowTheProcess/spok/task"
	"github.com/google/go-cmp/cmp"
)

func TestFind(t *testing.T) {
	t.Parallel()
	testdata := getTestdata()

	t.Run("found spokfile", func(t *testing.T) {
		start := filepath.Join(testdata, "nested", "deep", "down", "here") // Start deep inside testdata
		stop := testdata                                                   // Stop at testdata

		want, err := filepath.Abs(filepath.Join(testdata, "nested", "spokfile"))
		if err != nil {
			t.Fatal("could not resolve want")
		}

		path, err := Find(start, stop)
		if err != nil {
			t.Fatalf("find returned an error: %v", err)
		}

		if path != want {
			t.Errorf("got %s, wanted %s", path, want)
		}
	})

	t.Run("missing spokfile", func(t *testing.T) {
		start := filepath.Join(testdata, "missing", "deep", "down", "here")
		stop := testdata

		_, err := Find(start, stop)
		if err == nil {
			t.Fatal("expected ErrNoSpokfile, got nil")
		}

		if !errors.Is(err, errNoSpokfile) {
			t.Errorf("wrong error, got %v, wanted %v", err, errNoSpokfile)
		}
	})
}

func TestExpandGlobs(t *testing.T) {
	t.Parallel()
	testdata := getTestdata()
	tests := []struct {
		file *SpokFile
		want *SpokFile
		name string
	}{
		{
			name: "dependencies",
			file: &SpokFile{
				Path:  filepath.Join(testdata, "spokfile"),
				Vars:  make(map[string]string),
				Globs: make(map[string][]string),
				Tasks: map[string]task.Task{
					"test": {
						Doc:              "A simple test task",
						Name:             "test",
						GlobDependencies: []string{"**/*.txt"},
					},
				},
			},
			want: &SpokFile{
				Path: filepath.Join(testdata, "spokfile"),
				Vars: make(map[string]string),
				Globs: map[string][]string{
					"**/*.txt": {
						mustAbs(testdata, "top.txt"),
						mustAbs(testdata, "deps/sub1/sub2/blah.txt"),
						mustAbs(testdata, "deps/sub1/sub2/sub3/hello.txt"),
						mustAbs(testdata, "deps/suba/subb/stuff.txt"),
						mustAbs(testdata, "deps/suba/subb/subc/something.txt"),
					},
				},
				Tasks: map[string]task.Task{
					"test": {
						Doc:              "A simple test task",
						Name:             "test",
						GlobDependencies: []string{"**/*.txt"},
					},
				},
			},
		},
		{
			name: "outputs",
			file: &SpokFile{
				Path:  filepath.Join(testdata, "spokfile"),
				Vars:  make(map[string]string),
				Globs: make(map[string][]string),
				Tasks: map[string]task.Task{
					"test": {
						Doc:         "A simple test task",
						Name:        "test",
						GlobOutputs: []string{"**/*.json"},
					},
				},
			},
			want: &SpokFile{
				Path: filepath.Join(testdata, "spokfile"),
				Vars: make(map[string]string),
				Globs: map[string][]string{
					"**/*.json": {
						mustAbs(testdata, "top.json"),
						mustAbs(testdata, "outputs/sub1/sub2/blah.json"),
						mustAbs(testdata, "outputs/sub1/sub2/sub3/hello.json"),
						mustAbs(testdata, "outputs/suba/subb/stuff.json"),
						mustAbs(testdata, "outputs/suba/subb/subc/something.json"),
					},
				},
				Tasks: map[string]task.Task{
					"test": {
						Doc:         "A simple test task",
						Name:        "test",
						GlobOutputs: []string{"**/*.json"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.file.ExpandGlobs(); err != nil {
				t.Fatalf("ExpandGlobs returned an error: %v", err)
			}

			if diff := cmp.Diff(tt.want, tt.file); diff != "" {
				t.Errorf("File mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFromAST(t *testing.T) {
	t.Parallel()
	testdata := getTestdata()
	tests := []struct {
		want    *SpokFile
		name    string
		tree    ast.Tree
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
			want: &SpokFile{
				Path:  filepath.Join(testdata, "spokfile"),
				Vars:  make(map[string]string),
				Globs: make(map[string][]string),
				Tasks: map[string]task.Task{
					"test": {
						Doc:      "A simple test task",
						Name:     "test",
						Commands: []string{"go test ./..."},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "task with glob dependencies",
			tree: ast.Tree{
				Nodes: []ast.Node{
					ast.Task{
						Name:         ast.Ident{Name: "test", NodeType: ast.NodeIdent},
						Docstring:    ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
						Commands:     []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
						NodeType:     ast.NodeTask,
						Dependencies: []ast.Node{ast.String{Text: "**/*.go", NodeType: ast.NodeString}},
					},
				},
			},
			want: &SpokFile{
				Path:  filepath.Join(testdata, "spokfile"),
				Vars:  make(map[string]string),
				Globs: map[string][]string{"**/*.go": nil},
				Tasks: map[string]task.Task{
					"test": {
						Doc:              "A simple test task",
						Name:             "test",
						Commands:         []string{"go test ./..."},
						GlobDependencies: []string{"**/*.go"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "task with glob output",
			tree: ast.Tree{
				Nodes: []ast.Node{
					ast.Task{
						Name:      ast.Ident{Name: "test", NodeType: ast.NodeIdent},
						Docstring: ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
						Commands:  []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
						NodeType:  ast.NodeTask,
						Outputs:   []ast.Node{ast.String{Text: "**/*.go", NodeType: ast.NodeString}},
					},
				},
			},
			want: &SpokFile{
				Path:  filepath.Join(testdata, "spokfile"),
				Vars:  make(map[string]string),
				Globs: map[string][]string{"**/*.go": nil},
				Tasks: map[string]task.Task{
					"test": {
						Doc:         "A simple test task",
						Name:        "test",
						Commands:    []string{"go test ./..."},
						GlobOutputs: []string{"**/*.go"},
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
			want: &SpokFile{
				Path:  filepath.Join(testdata, "spokfile"),
				Vars:  make(map[string]string),
				Globs: make(map[string][]string),
				Tasks: map[string]task.Task{
					"test": {
						Name:     "test",
						Commands: []string{"go test ./..."},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "duplicate tasks",
			tree: ast.Tree{
				Nodes: []ast.Node{
					ast.Task{
						Name:      ast.Ident{Name: "test", NodeType: ast.NodeIdent},
						Docstring: ast.Comment{Text: " A simple test task", NodeType: ast.NodeComment},
						Commands:  []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
						NodeType:  ast.NodeTask,
					},
					ast.Task{
						Name:      ast.Ident{Name: "test", NodeType: ast.NodeIdent},
						Docstring: ast.Comment{Text: " A duplicate test task", NodeType: ast.NodeComment},
						Commands:  []ast.Command{{Command: "go test ./...", NodeType: ast.NodeCommand}},
						NodeType:  ast.NodeTask,
					},
				},
			},
			want:    nil,
			wantErr: true,
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
			want: &SpokFile{
				Path:  filepath.Join(testdata, "spokfile"),
				Vars:  map[string]string{"global1": "hello", "global2": "hello again"},
				Globs: make(map[string][]string),
				Tasks: make(map[string]task.Task),
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
			want: &SpokFile{
				Path:  filepath.Join(testdata, "spokfile"),
				Vars:  map[string]string{"global1": filepath.Join("path", "parts", "more")},
				Globs: make(map[string][]string),
				Tasks: make(map[string]task.Task),
			},
			wantErr: false,
		},
		{
			name: "globals with exec builtin",
			tree: ast.Tree{
				Nodes: []ast.Node{
					ast.Assign{
						Value: ast.Function{
							Name: ast.Ident{Name: "exec", NodeType: ast.NodeIdent},
							Arguments: []ast.Node{
								ast.String{Text: "echo hello", NodeType: ast.NodeString},
							},
							NodeType: ast.NodeFunction,
						},
						Name:     ast.Ident{Name: "global1", NodeType: ast.NodeIdent},
						NodeType: ast.NodeAssign,
					},
				},
			},
			want: &SpokFile{
				Path:  filepath.Join(testdata, "spokfile"),
				Vars:  map[string]string{"global1": "hello"},
				Globs: make(map[string][]string),
				Tasks: make(map[string]task.Task),
			},
			wantErr: false,
		},
		{
			name: "globals with exec builtin single arg",
			tree: ast.Tree{
				Nodes: []ast.Node{
					ast.Assign{
						Value: ast.Function{
							Name: ast.Ident{Name: "exec", NodeType: ast.NodeIdent},
							Arguments: []ast.Node{
								ast.String{Text: "echo", NodeType: ast.NodeString},
							},
							NodeType: ast.NodeFunction,
						},
						Name:     ast.Ident{Name: "global1", NodeType: ast.NodeIdent},
						NodeType: ast.NodeAssign,
					},
				},
			},
			want: &SpokFile{
				Path:  filepath.Join(testdata, "spokfile"),
				Vars:  map[string]string{"global1": ""},
				Globs: make(map[string][]string),
				Tasks: make(map[string]task.Task),
			},
			wantErr: false,
		},
		{
			name: "globals with exec builtin no arg",
			tree: ast.Tree{
				Nodes: []ast.Node{
					ast.Assign{
						Value: ast.Function{
							Name:      ast.Ident{Name: "exec", NodeType: ast.NodeIdent},
							Arguments: []ast.Node{},
							NodeType:  ast.NodeFunction,
						},
						Name:     ast.Ident{Name: "global1", NodeType: ast.NodeIdent},
						NodeType: ast.NodeAssign,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "globals with exec builtin error",
			tree: ast.Tree{
				Nodes: []ast.Node{
					ast.Assign{
						Value: ast.Function{
							Name: ast.Ident{Name: "exec", NodeType: ast.NodeIdent},
							Arguments: []ast.Node{
								ast.String{Text: "exit 1", NodeType: ast.NodeString},
							},
							NodeType: ast.NodeFunction,
						},
						Name:     ast.Ident{Name: "global1", NodeType: ast.NodeIdent},
						NodeType: ast.NodeAssign,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "globals with undefined builtin",
			tree: ast.Tree{
				Nodes: []ast.Node{
					ast.Assign{
						Value: ast.Function{
							Name: ast.Ident{Name: "undefined", NodeType: ast.NodeIdent},
							Arguments: []ast.Node{
								ast.String{Text: "hello", NodeType: ast.NodeString},
							},
							NodeType: ast.NodeFunction,
						},
						Name:     ast.Ident{Name: "global1", NodeType: ast.NodeIdent},
						NodeType: ast.NodeAssign,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.tree, testdata)
			if (err != nil) != tt.wantErr {
				t.Fatalf("fromTree() err = %v, wantErr = %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("File mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		spokfile *SpokFile
		name     string
		tasks    []string
		want     []task.Result
		sync     bool
		force    bool
	}{
		{
			name: "simple",
			spokfile: &SpokFile{
				Tasks: map[string]task.Task{
					"test": {
						Name: "test",
						Commands: []string{
							"echo hello",
						},
					},
				},
			},
			sync:  false,
			force: false,
			tasks: []string{"test"},
			want: []task.Result{
				{
					CommandResults: []shell.Result{
						{
							Cmd:    "echo hello",
							Stdout: "hello\n",
							Stderr: "",
							Status: 0,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.spokfile.Run(tt.sync, tt.force, tt.tasks...)
			if err != nil {
				t.Fatalf("Run returned an error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

//
// INTEGRATION TESTS START HERE
//
// Thar be larger, integration tests below.

// The tests below will only run if SPOK_INTEGRATION_TEST is set, making it easy to run only isolated
// tests while developing to limit potentially distracting failing integration test output until ready.
//
// The benchmarks below will run when invoking go test with the -bench flag, there is no concept of unit
// or integration benchmarks here.
//

// fullSpokfileAST is an example AST of a real spokfile, with the exception
// that it's designed to hit the testdata directory so all globs point to .txt files
// and commands that would otherwise return different things each time e.g. git rev-parse HEAD
// have been changed to be consistent over time e.g. echo "hello".
var testdata = getTestdata()
var fullSpokfileAST = ast.Tree{
	Nodes: []ast.Node{
		ast.Comment{
			Text:     " This is a top level comment",
			NodeType: ast.NodeComment,
		},
		ast.Comment{
			Text:     " This variable is presumably important later",
			NodeType: ast.NodeComment,
		},
		ast.Assign{
			Name: ast.Ident{
				Name:     "GLOBAL",
				NodeType: ast.NodeIdent,
			},
			Value: ast.String{
				Text:     "very important stuff here",
				NodeType: ast.NodeString,
			},
			NodeType: ast.NodeAssign,
		},
		ast.Assign{
			Value: ast.Function{
				Name: ast.Ident{
					Name:     "exec",
					NodeType: ast.NodeIdent,
				},
				Arguments: []ast.Node{
					ast.String{
						Text:     "echo hello",
						NodeType: ast.NodeString,
					},
				}, NodeType: ast.NodeFunction,
			},
			Name: ast.Ident{
				Name:     "GIT_COMMIT",
				NodeType: ast.NodeIdent,
			},
			NodeType: ast.NodeAssign,
		},
		ast.Task{
			Name: ast.Ident{
				Name:     "test",
				NodeType: ast.NodeIdent,
			},
			Docstring: ast.Comment{
				Text:     " Run the project unit tests",
				NodeType: ast.NodeComment,
			},
			Dependencies: []ast.Node{
				ast.Ident{
					Name:     "fmt",
					NodeType: ast.NodeIdent,
				},
			},
			Outputs: []ast.Node{},
			Commands: []ast.Command{
				{
					Command:  "go test -race ./...",
					NodeType: ast.NodeCommand,
				},
			},
			NodeType: ast.NodeTask,
		},
		ast.Task{
			Name: ast.Ident{
				Name:     "fmt",
				NodeType: ast.NodeIdent,
			},
			Docstring: ast.Comment{
				Text:     " Format the project source",
				NodeType: ast.NodeComment,
			},
			Dependencies: []ast.Node{
				ast.String{
					Text:     "**/*.txt",
					NodeType: ast.NodeString,
				},
			},
			Outputs: []ast.Node{},
			Commands: []ast.Command{
				{
					Command:  "go fmt ./...",
					NodeType: ast.NodeCommand,
				},
			},
			NodeType: ast.NodeTask,
		},
		ast.Task{
			Name: ast.Ident{
				Name:     "many",
				NodeType: ast.NodeIdent,
			},
			Docstring: ast.Comment{
				Text:     " Do many things",
				NodeType: ast.NodeComment,
			},
			Dependencies: []ast.Node{},
			Outputs:      []ast.Node{},
			Commands: []ast.Command{
				{
					Command:  "line 1",
					NodeType: ast.NodeCommand,
				},
				{
					Command:  "line 2",
					NodeType: ast.NodeCommand,
				},
				{
					Command:  "line 3",
					NodeType: ast.NodeCommand,
				},
				{
					Command:  "line 4",
					NodeType: ast.NodeCommand,
				},
			},
			NodeType: ast.NodeTask,
		},
		ast.Task{
			Name: ast.Ident{
				Name:     "build",
				NodeType: ast.NodeIdent,
			},
			Docstring: ast.Comment{
				Text:     " Compile the project",
				NodeType: ast.NodeComment,
			},
			Dependencies: []ast.Node{
				ast.String{
					Text:     "**/*.txt",
					NodeType: ast.NodeString,
				},
			},
			Outputs: []ast.Node{
				ast.String{
					Text:     "./bin/main",
					NodeType: ast.NodeString,
				},
			},
			Commands: []ast.Command{
				{
					Command:  `go build -ldflags="-X main.version=test -X main.commit=7cb0ec5609efb5fe0"`,
					NodeType: ast.NodeCommand,
				},
			},
			NodeType: ast.NodeTask,
		},
		ast.Task{
			Name: ast.Ident{
				Name:     "show",
				NodeType: ast.NodeIdent,
			},
			Docstring: ast.Comment{
				Text:     " Show the global variables",
				NodeType: ast.NodeComment,
			},
			Dependencies: []ast.Node{},
			Outputs:      []ast.Node{},
			Commands: []ast.Command{
				{
					Command:  "echo {{.GLOBAL}}",
					NodeType: ast.NodeCommand,
				},
			},
			NodeType: ast.NodeTask,
		},
		ast.Task{
			Name: ast.Ident{
				Name:     "moar_things",
				NodeType: ast.NodeIdent,
			},
			Docstring: ast.Comment{
				Text:     " Generate multiple outputs",
				NodeType: ast.NodeComment,
			},
			Dependencies: []ast.Node{},
			Outputs: []ast.Node{
				ast.String{
					Text:     "output1.go",
					NodeType: ast.NodeString,
				},
				ast.String{
					Text:     "output2.go",
					NodeType: ast.NodeString,
				},
			},
			Commands: []ast.Command{
				{
					Command:  "do some stuff here",
					NodeType: ast.NodeCommand,
				},
			},
			NodeType: ast.NodeTask,
		},
	},
}

// spokFileWant is the expected concrete spok.File object when the above AST is concretised.
var spokFileWant = &SpokFile{
	Path: filepath.Join(getTestdata(), "spokfile"),
	Vars: map[string]string{
		"GLOBAL":     "very important stuff here",
		"GIT_COMMIT": "hello",
	},
	Globs: map[string][]string{
		"**/*.txt": nil,
	},
	Tasks: map[string]task.Task{
		"test": {
			Doc:              "Run the project unit tests",
			Name:             "test",
			TaskDependencies: []string{"fmt"},
			FileDependencies: nil,
			Commands:         []string{"go test -race ./..."},
			NamedOutputs:     nil,
			FileOutputs:      nil,
		},
		"fmt": {
			Doc:              "Format the project source",
			Name:             "fmt",
			TaskDependencies: nil,
			FileDependencies: nil,
			GlobDependencies: []string{"**/*.txt"},
			Commands:         []string{"go fmt ./..."},
			NamedOutputs:     nil,
			FileOutputs:      nil,
		},
		"many": {
			Doc:              "Do many things",
			Name:             "many",
			TaskDependencies: nil,
			FileDependencies: nil,
			Commands: []string{
				"line 1",
				"line 2",
				"line 3",
				"line 4",
			},
			NamedOutputs: nil,
			FileOutputs:  nil,
		},
		"build": {
			Doc:              "Compile the project",
			Name:             "build",
			TaskDependencies: nil,
			FileDependencies: nil,
			GlobDependencies: []string{"**/*.txt"},
			Commands:         []string{`go build -ldflags="-X main.version=test -X main.commit=7cb0ec5609efb5fe0"`},
			NamedOutputs:     nil,
			FileOutputs:      []string{mustAbs(testdata, "./bin/main")},
		},
		"show": {
			Doc:              "Show the global variables",
			Name:             "show",
			TaskDependencies: nil,
			FileDependencies: nil,
			Commands:         []string{"echo very important stuff here"},
			NamedOutputs:     nil,
			FileOutputs:      nil,
		},
		"moar_things": {
			Doc:              "Generate multiple outputs",
			Name:             "moar_things",
			TaskDependencies: nil,
			FileDependencies: nil,
			Commands:         []string{"do some stuff here"},
			NamedOutputs:     nil,
			FileOutputs: []string{
				mustAbs(testdata, "output1.go"),
				mustAbs(testdata, "output2.go"),
			},
		},
	},
}

// TestBuildFullSpokfile tests spok's ability to take an entire representative spok AST
// and convert it to a concrete File object with expanded globs, run shell commands etc.
func TestBuildFullSpokfile(t *testing.T) {
	if os.Getenv("SPOK_INTEGRATION_TEST") == "" {
		t.Skip("Set SPOK_INTEGRATION_TEST to run this test.")
	}
	t.Parallel()

	got, err := New(fullSpokfileAST, getTestdata())
	if err != nil {
		t.Fatalf("fromAST returned an error: %v", err)
	}

	if diff := cmp.Diff(spokFileWant, got); diff != "" {
		t.Errorf("File mismatch (-want +got):\n%s", diff)
	}
}

func getTestdata() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic("could not get cwd")
	}
	testdata := filepath.Join(cwd, "testdata")
	return testdata
}

// mustAbs returns the resolved 'path' or panics if it cannot.
func mustAbs(root, path string) string {
	abs, err := filepath.Abs(filepath.Join(root, path))
	if err != nil {
		panic(fmt.Sprintf("mustAbs could not resolve '%s'", filepath.Join(root, path)))
	}
	return abs
}
