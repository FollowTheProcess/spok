// Package file implements the core functionality to do with the spokfile.
package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/FollowTheProcess/spok/builtins"
	"github.com/FollowTheProcess/spok/task"
)

// Name is the canonical spok file name.
const Name = "spokfile"

// errNoSpokfile is what happens when spok can't find a spokfile.
var errNoSpokfile = errors.New("No spokfile found")

// SpokFile represents a concrete spokfile.
type SpokFile struct {
	Vars  map[string]string    // Global variables in IDENT: value form (functions already evaluated)
	Tasks map[string]task.Task // Map of task name to the task itself
	Globs map[string][]string  // Map of glob pattern to their concrete filepaths (avoids recalculating)
	Path  string               // The absolute path to the spokfile
}

// hasTask returns whether or not the SpokFile has a task with the given name.
func (s SpokFile) hasTask(name string) bool {
	_, ok := s.Tasks[name]
	return ok
}

// Find climbs the file tree from 'start' to 'stop' looking for a spokfile,
// if it hits 'stop' before finding one, an errNoSpokfile will be returned
// If a spokfile is found, it's absolute path will be returned
// typical usage will make start = $CWD and stop = $HOME.
func Find(start, stop string) (string, error) {
	for {
		entries, err := os.ReadDir(start)
		if err != nil {
			return "", fmt.Errorf("could not read directory '%s': %w", start, err)
		}

		for _, e := range entries {
			if !e.IsDir() && e.Name() == Name {
				// We've found it
				abs, err := filepath.Abs(filepath.Join(start, e.Name()))
				if err != nil {
					return "", fmt.Errorf("could not resolve '%s': %w", e.Name(), err)
				}
				return abs, nil
			} else if start == stop {
				return "", errNoSpokfile
			}
		}
		start = filepath.Dir(start)
	}
}

// New converts a parsed spok AST into a concrete File object,
// root is the absolute path to the directory to use as root for glob
// expansion, typically the path to the directory the spokfile sits in.
func New(tree ast.Tree, root string) (SpokFile, error) {
	var file SpokFile
	file.Path = filepath.Join(root, Name)
	file.Vars = make(map[string]string)
	file.Tasks = make(map[string]task.Task)
	file.Globs = make(map[string][]string)

	for _, node := range tree.Nodes {
		switch {
		case node.Type() == ast.NodeAssign:
			assign, ok := node.(ast.Assign)
			if !ok {
				return SpokFile{}, fmt.Errorf("AST node has ast.NodeAssign type but could not be converted to an ast.Assign: %s", node)
			}
			switch {
			case assign.Value.Type() == ast.NodeString:
				file.Vars[assign.Name.Name] = assign.Value.Literal()

			case assign.Value.Type() == ast.NodeFunction:
				function, ok := assign.Value.(ast.Function)
				if !ok {
					return SpokFile{}, fmt.Errorf("AST node has ast.NodeFunction type but could not be converted to an ast.Function: %s", assign.Value)
				}
				args := make([]string, 0, len(function.Arguments))
				for _, arg := range function.Arguments {
					if arg.Type() != ast.NodeString {
						return SpokFile{}, fmt.Errorf("Spok builtin functions take only string arguments, got %s", arg.Type())
					}
					args = append(args, arg.Literal())
				}
				fn, ok := builtins.Get(function.Name.Name)
				if !ok {
					return SpokFile{}, fmt.Errorf("Builtin function undefined: %s", function.Name.Name)
				}
				val, err := fn(args...)
				if err != nil {
					return SpokFile{}, fmt.Errorf("Builtin function %s returned an error: %s", function.Name.Name, err)
				}
				// Assign the value to the variable
				file.Vars[assign.Name.Name] = val

			default:
				return SpokFile{}, fmt.Errorf("Unexpected node in assignment %s: %s", assign.Value.Type(), assign.Value)
			}

		case node.Type() == ast.NodeTask:
			taskNode, ok := node.(ast.Task)
			if !ok {
				return SpokFile{}, fmt.Errorf("AST node has ast.NodeTask type but could not be converted to an ast.Task: %s", node)
			}

			task, err := task.New(taskNode, root, file.Vars)
			if err != nil {
				return SpokFile{}, err
			}

			if file.hasTask(task.Name) {
				return SpokFile{}, fmt.Errorf("Duplicate task: spokfile already contains task named %q, duplicate tasks not allowed", task.Name)
			}

			// Add the glob patterns from the tasks to the files' map of globs
			// this enables us to only calculate the glob expansion once if multiple
			// tasks share the same pattern, since glob expansion does a lot of ReadDir
			// it is relatively expensive
			for _, pattern := range task.GlobDependencies {
				var emptySlice []string
				file.Globs[pattern] = emptySlice
			}
			for _, pattern := range task.GlobOutputs {
				var emptySlice []string
				file.Globs[pattern] = emptySlice
			}

			// Add the task to the file
			file.Tasks[task.Name] = task
		}
	}
	return file, nil
}
