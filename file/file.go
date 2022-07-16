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
	"github.com/bmatcuk/doublestar/v4"
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
func (s *SpokFile) hasTask(name string) bool {
	_, ok := s.Tasks[name]
	return ok
}

// hasGlob returns whether or not the SpokFile has already expanded a glob pattern.
func (s *SpokFile) hasGlob(pattern string) bool {
	_, ok := s.Globs[pattern]
	return ok
}

// ExpandGlobs gathers up all the glob patterns in the spokfile and expands them
// saving the results to the Globs map as e.g. {"**/*.go": [file1.go, file2.go]}.
func (s *SpokFile) ExpandGlobs() error {
	for _, task := range s.Tasks {
		for _, pattern := range task.GlobDependencies {
			if !s.hasGlob(pattern) {
				matches, err := expandGlob(filepath.Dir(s.Path), pattern)
				if err != nil {
					return err
				}
				s.Globs[pattern] = matches
			}
		}

		for _, pattern := range task.GlobOutputs {
			if !s.hasGlob(pattern) {
				matches, err := expandGlob(filepath.Dir(s.Path), pattern)
				if err != nil {
					return err
				}
				s.Globs[pattern] = matches
			}
		}
	}
	return nil
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
func New(tree ast.Tree, root string) (*SpokFile, error) {
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
				return nil, fmt.Errorf("AST node has ast.NodeAssign type but could not be converted to an ast.Assign: %s", node)
			}
			switch {
			case assign.Value.Type() == ast.NodeString:
				file.Vars[assign.Name.Name] = assign.Value.Literal()

			case assign.Value.Type() == ast.NodeFunction:
				function, ok := assign.Value.(ast.Function)
				if !ok {
					return nil, fmt.Errorf("AST node has ast.NodeFunction type but could not be converted to an ast.Function: %s", assign.Value)
				}
				args := make([]string, 0, len(function.Arguments))
				for _, arg := range function.Arguments {
					if arg.Type() != ast.NodeString {
						return nil, fmt.Errorf("Spok builtin functions take only string arguments, got %s", arg.Type())
					}
					args = append(args, arg.Literal())
				}
				fn, ok := builtins.Get(function.Name.Name)
				if !ok {
					return nil, fmt.Errorf("Builtin function undefined: %s", function.Name.Name)
				}
				val, err := fn(args...)
				if err != nil {
					return nil, fmt.Errorf("Builtin function %s returned an error: %s", function.Name.Name, err)
				}
				// Assign the value to the variable
				file.Vars[assign.Name.Name] = val

			default:
				return nil, fmt.Errorf("Unexpected node in assignment %s: %s", assign.Value.Type(), assign.Value)
			}

		case node.Type() == ast.NodeTask:
			taskNode, ok := node.(ast.Task)
			if !ok {
				return nil, fmt.Errorf("AST node has ast.NodeTask type but could not be converted to an ast.Task: %s", node)
			}

			task, err := task.New(taskNode, root, file.Vars)
			if err != nil {
				return nil, err
			}

			if file.hasTask(task.Name) {
				return nil, fmt.Errorf("Duplicate task: spokfile already contains task named %q, duplicate tasks not allowed", task.Name)
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
	return &file, nil
}

// expandGlob expands out the glob pattern from root and returns all the matches,
// the matches are made absolute before returning, root should be absolute.
func expandGlob(root, pattern string) ([]string, error) {
	matches, err := doublestar.FilepathGlob(filepath.Join(root, pattern))
	if err != nil {
		return nil, fmt.Errorf("could not expand glob pattern '%s': %w", filepath.Join(root, pattern), err)
	}
	absMatches := make([]string, 0, len(matches))
	for _, match := range matches {
		abs, err := filepath.Abs(match)
		if err != nil {
			return nil, fmt.Errorf("could not resolve path '%s' to absolute: %w", match, err)
		}
		absMatches = append(absMatches, abs)
	}

	return absMatches, nil
}
