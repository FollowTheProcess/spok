package task

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/FollowTheProcess/spok/iostream"
	"github.com/FollowTheProcess/spok/shell"
	"github.com/google/go-cmp/cmp"
)

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

func TestTaskRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		task    Task
		want    shell.Results
		wantErr bool
	}{
		{
			name:    "empty",
			task:    Task{Name: "empty"},
			want:    nil,
			wantErr: true,
		},
		{
			name: "simple",
			task: Task{Name: "simple", Commands: []string{
				"echo hello",
			}},
			want: shell.Results{{
				Stdout: "hello\n",
				Stderr: "",
				Status: 0,
				Cmd:    "echo hello",
			}},
			wantErr: false,
		},
		{
			name: "stderr",
			task: Task{Name: "stderr", Commands: []string{
				"echo hello stderr >&2",
			}},
			want: shell.Results{{
				Stdout: "",
				Stderr: "hello stderr\n",
				Status: 0,
				Cmd:    "echo hello stderr >&2",
			}},
			wantErr: false,
		},
		{
			name: "multiple",
			task: Task{Name: "multiple", Commands: []string{
				"echo hello",
				"echo hello stderr >&2",
				"true",
				"false",
			}},
			want: shell.Results{
				{Stdout: "hello\n", Stderr: "", Status: 0, Cmd: "echo hello"},
				{Stdout: "", Stderr: "hello stderr\n", Status: 0, Cmd: "echo hello stderr >&2"},
				{Stdout: "", Stderr: "", Status: 0, Cmd: "true"},
				{Stdout: "", Stderr: "", Status: 1, Cmd: "false"},
			},
			wantErr: false,
		},
		{
			name: "error in the middle",
			task: Task{Name: "multiple", Commands: []string{
				"echo hello",
				"false",                 // 1 status code here
				"echo hello stderr >&2", // Should still see these
				"true",
			}},
			want: shell.Results{
				{Stdout: "hello\n", Stderr: "", Status: 0, Cmd: "echo hello"},
				{Stdout: "", Stderr: "", Status: 1, Cmd: "false"},
				{Stdout: "", Stderr: "hello stderr\n", Status: 0, Cmd: "echo hello stderr >&2"},
				{Stdout: "", Stderr: "", Status: 0, Cmd: "true"},
			},
			wantErr: false,
		},
		{
			name: "bad syntax",
			task: Task{Name: "bad", Commands: []string{
				"(*^$$",
			}},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := shell.NewIntegratedRunner()
			got, err := tt.task.Run(runner, iostream.Null(), nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Run() err = %v, wantErr = %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResultOk(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		result Result
		want   bool
	}{
		{
			name: "one success",
			result: Result{
				CommandResults: shell.Results{
					{Stdout: "Hello", Stderr: "", Status: 0},
				},
			},
			want: true,
		},
		{
			name: "one failure",
			result: Result{
				CommandResults: shell.Results{
					{Stdout: "", Stderr: "Hello", Status: 1},
				},
			},
			want: false,
		},
		{
			name: "multiple successes",
			result: Result{
				CommandResults: shell.Results{
					{Stdout: "Hello", Stderr: "", Status: 0},
					{Stdout: "There", Stderr: "", Status: 0},
					{Stdout: "General", Stderr: "", Status: 0},
					{Stdout: "Kenobi", Stderr: "", Status: 0},
				},
			},
			want: true,
		},
		{
			name: "multiple failures",
			result: Result{
				CommandResults: shell.Results{
					{Stdout: "Hello", Stderr: "", Status: 1},
					{Stdout: "There", Stderr: "", Status: 1},
					{Stdout: "General", Stderr: "", Status: 1},
					{Stdout: "Kenobi", Stderr: "", Status: 1},
				},
			},
			want: false,
		},
		{
			name: "failure in the middle",
			result: Result{
				CommandResults: shell.Results{
					{Stdout: "Hello", Stderr: "", Status: 0},
					{Stdout: "There", Stderr: "", Status: 1},
					{Stdout: "General", Stderr: "", Status: 0},
					{Stdout: "Kenobi", Stderr: "", Status: 0},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Ok(); got != tt.want {
				t.Errorf("got %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestResultsOk(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		results Results
		want    bool
	}{
		{
			name: "one success",
			results: Results{
				{
					Task: "test",
					CommandResults: shell.Results{
						{Status: 0},
					},
					Skipped: false,
				},
			},
			want: true,
		},
		{
			name: "one failure",
			results: Results{
				{
					Task: "test",
					CommandResults: shell.Results{
						{Status: 1},
					},
					Skipped: false,
				},
			},
			want: false,
		},
		{
			name: "multiple successes",
			results: Results{
				{
					Task: "test",
					CommandResults: shell.Results{
						{Status: 0},
					},
					Skipped: false,
				},
				{
					Task: "test2",
					CommandResults: shell.Results{
						{Status: 0},
					},
					Skipped: false,
				},
				{
					Task: "test3",
					CommandResults: shell.Results{
						{Status: 0},
					},
					Skipped: false,
				},
			},
			want: true,
		},
		{
			name: "multiple failures",
			results: Results{
				{
					Task: "test",
					CommandResults: shell.Results{
						{Status: 1},
					},
					Skipped: false,
				},
				{
					Task: "test2",
					CommandResults: shell.Results{
						{Status: 3},
					},
					Skipped: false,
				},
				{
					Task: "test3",
					CommandResults: shell.Results{
						{Status: 1},
					},
					Skipped: false,
				},
			},
			want: false,
		},
		{
			name: "multiple both",
			results: Results{
				{
					Task: "test",
					CommandResults: shell.Results{
						{Status: 0},
					},
					Skipped: false,
				},
				{
					Task: "test2",
					CommandResults: shell.Results{
						{Status: 3},
					},
					Skipped: false,
				},
				{
					Task: "test3",
					CommandResults: shell.Results{
						{Status: 0},
					},
					Skipped: false,
				},
				{
					Task: "test4",
					CommandResults: shell.Results{
						{Status: 1},
					},
					Skipped: false,
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.results.Ok(); got != tt.want {
				t.Errorf("got %v, wanted %v", got, tt.want)
			}
		})
	}
}

func TestResultsJSON(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		results Results
	}{
		{
			name: "single result",
			results: Results{
				{
					Task: "test",
					CommandResults: shell.Results{
						{
							Cmd:    "echo hello",
							Stdout: "hello\n",
							Stderr: "",
							Status: 0,
						},
					},
					Skipped: false,
				},
			},
			want: `[{"task":"test","results":[{"cmd":"echo hello","stdout":"hello\n","stderr":"","status":0}],"skipped":false}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.results.JSON()
			if err != nil {
				t.Fatalf("Could not marshal results to JSON: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("JSON mismatch (-want +got):\n%s", diff)
			}
		})
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
