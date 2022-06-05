// Package task handles core spok functionality related to the processing of declared
// tasks e.g. expanding glob patterns, parsing from an ast node etc.
package task

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/bmatcuk/doublestar/v4"
)

// Task represents a spok Task.
type Task struct {
	Doc               string   // The task docstring
	Name              string   // Task name
	NamedDependencies []string // Other tasks or idents this task depends on (by name)
	FileDependencies  []string // Filepaths this task depends on (globs expanded and made absolute)
	Commands          []string // Shell commands to run
	NamedOutputs      []string // Other outputs by ident
	FileOutputs       []string // Filepaths this task outputs
}

// New parses a task AST node into a concrete task,
// root is the absolute path of the directory to use as the root for
// glob expansion, typically the path to the spokfile.
func New(t ast.Task, root string, vars map[string]string) (Task, error) {
	var (
		fileDeps     []string
		namedDeps    []string
		commands     []string
		fileOutputs  []string
		namedOutputs []string
	)

	for _, dep := range t.Dependencies {
		switch {
		case dep.Type() == ast.NodeString:
			// String means file dependency, in which case Literal is the go representation of the string
			if strings.Contains(dep.Literal(), "*") {
				// We have a glob pattern
				matches, err := expandGlob(root, dep.Literal())
				if err != nil {
					return Task{}, err
				}
				fileDeps = append(fileDeps, matches...)
			} else {
				// We have something like "file.go"
				fileDeps = append(fileDeps, filepath.Join(root, dep.Literal()))
			}
		case dep.Type() == ast.NodeIdent:
			// Ident means it depends on another task
			namedDeps = append(namedDeps, dep.Literal())
		default:
			return Task{}, fmt.Errorf("unknown dependency: %s", dep)
		}
	}

	for _, cmd := range t.Commands {
		expanded, err := expandVars(cmd.Command, vars)
		if err != nil {
			return Task{}, err
		}
		commands = append(commands, expanded)
	}

	for _, out := range t.Outputs {
		switch {
		case out.Type() == ast.NodeString:
			// String means file
			if strings.Contains(out.Literal(), "*") {
				// We have a glob pattern
				matches, err := expandGlob(root, out.Literal())
				if err != nil {
					return Task{}, err
				}
				fileOutputs = append(fileOutputs, matches...)
			} else {
				// We have something like "file.go"
				fileOutputs = append(fileOutputs, filepath.Join(root, out.Literal()))
			}
		case out.Type() == ast.NodeIdent:
			// Ident means it outputs something named by global scope
			namedOutputs = append(namedOutputs, out.Literal())
		default:
			return Task{}, fmt.Errorf("unknown dependency: %s", out)
		}
	}

	task := Task{
		Doc:               strings.TrimSpace(t.Docstring.Text),
		Name:              t.Name.Name,
		NamedDependencies: namedDeps,
		FileDependencies:  fileDeps,
		Commands:          commands,
		NamedOutputs:      namedOutputs,
		FileOutputs:       fileOutputs,
	}
	return task, nil
}

// expandVars performs a find and replace on any templated variables in
// a command, using the provided variables map.
func expandVars(command string, vars map[string]string) (string, error) {
	tmp := template.New("tmp")
	parsed, err := tmp.Parse(command)
	if err != nil {
		return "", err
	}
	out := &bytes.Buffer{}
	if err := parsed.Execute(out, vars); err != nil {
		return "", err
	}

	return out.String(), nil
}

// HashFiles takes a list of absolute filepaths e.g. a task's file dependencies,
// calculates the SHA1 hash of the contents of each one, sums them up and returns the sum.
func HashFiles(files []string) (string, error) {
	open := func(file string) (io.ReadCloser, error) {
		return os.Open(file)
	}
	return hashFiles(files, open)
}

// hashFiles takes a list of absolute filepaths e.g. a task's file dependencies as well as a
// function that opens the content of each file, it opens, reads, hashes and closes each
// file and returns the overall hash sum.
func hashFiles(files []string, open func(string) (io.ReadCloser, error)) (string, error) {
	// We aren't actually after a cryptographically secure hash
	// we just need to see if files have changed and in testing benchmarks SHA1 was fastest
	hash := sha1.New()

	for _, file := range files {
		readCloser, err := open(file)
		if err != nil {
			return "", err
		}

		h := sha1.New()
		_, err = io.Copy(h, readCloser)
		readCloser.Close()
		if err != nil {
			return "", err
		}

		hash.Write(h.Sum(nil))
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// expandGlob expands out the glob pattern from root and returns all the matches,
// the matches are made absolute before returning, root should be absolute.
func expandGlob(root, pattern string) ([]string, error) {
	matches, err := doublestar.Glob(os.DirFS(root), pattern)
	if err != nil {
		return nil, fmt.Errorf("could not expand glob pattern '%s': %w", filepath.Join(root, pattern), err)
	}

	absMatches := make([]string, 0, len(matches))
	for _, match := range matches {
		joined := filepath.Join(root, match)
		abs, err := filepath.Abs(joined)
		if err != nil {
			return nil, fmt.Errorf("could not resolve path '%s' to absolute: %w", joined, err)
		}
		absMatches = append(absMatches, abs)
	}

	return absMatches, nil
}

// ByName allows a slice of Task to be sorted alphabetically by name.
type ByName []Task

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
