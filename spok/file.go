// Package spok implements the core functionality to do with the spokfile.
package spok

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/FollowTheProcess/spok/ast"
)

// Canonical spokfile filename.
const spokfile = "spokfile"

// errNoSpokfile is what happens when spok can't find a spokfile.
var errNoSpokfile = errors.New("No spokfile found")

// File represents a concrete spokfile.
type File struct {
	Path  string            // The absolute path to the spokfile
	Vars  map[string]string // Global variables in IDENT: value form (functions already evaluated)
	Tasks []Task            // Defined tasks
}

// find climbs the file tree from 'start' to 'stop' looking for a spokfile,
// if it hits 'stop' before finding one, an ErrNoSpokfile will be returned
// If a spokfile is found, it's absolute path will be returned
// typical usage will make start = $CWD and stop = $HOME.
func find(start, stop string) (string, error) {
	for {
		entries, err := os.ReadDir(start)
		if err != nil {
			return "", fmt.Errorf("could not read directory '%s': %w", start, err)
		}

		for _, e := range entries {
			if !e.IsDir() && e.Name() == spokfile {
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

// fromTree converts a parsed spok AST into a concrete File object,
// root is the absolute path to the directory to use as root for glob
// expansion, typically the path to the directory the spokfile sits in.
func fromTree(tree ast.Tree, root string) (File, error) {
	var file File
	file.Path = root
	file.Vars = make(map[string]string)
	for _, node := range tree.Nodes {
		switch {
		case node.Type() == ast.NodeAssign:
			assign, ok := node.(ast.Assign)
			if !ok {
				return File{}, fmt.Errorf("AST node has ast.NodeAssign type but could not be converted to an ast.Assign: %s", node)
			}
			switch {
			case assign.Value.Type() == ast.NodeString:
				file.Vars[assign.Name.Name] = assign.Value.String()

			case assign.Value.Type() == ast.NodeFunction:
				// TODO: This
				panic("TODO: Builtins")

			default:
				return File{}, fmt.Errorf("Unexpected node in assignment %s: %s", assign.Value.Type(), assign.Value)
			}

		case node.Type() == ast.NodeTask:
			taskNode, ok := node.(ast.Task)
			if !ok {
				return File{}, fmt.Errorf("AST node has ast.NodeTask type but could not be converted to an ast.Task: %s", node)
			}

			task, err := newTask(taskNode, root)
			if err != nil {
				return File{}, err
			}

			file.Tasks = append(file.Tasks, task)
		}
	}
	return file, nil
}
