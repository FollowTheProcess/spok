// Package file implements the core functionality to do with the spokfile.
package file

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/FollowTheProcess/collections/dag"
	"github.com/FollowTheProcess/spok/ast"
	"github.com/FollowTheProcess/spok/builtins"
	"github.com/FollowTheProcess/spok/cache"
	"github.com/FollowTheProcess/spok/hash"
	"github.com/FollowTheProcess/spok/iostream"
	"github.com/FollowTheProcess/spok/logger"
	"github.com/FollowTheProcess/spok/shell"
	"github.com/FollowTheProcess/spok/task"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/exp/maps"
)

// NAME is the canonical spok file name.
const NAME = "spokfile"

// SpokFile represents a concrete spokfile.
type SpokFile struct {
	logger logger.Logger        // Shared logger
	Vars   map[string]string    // Global variables in IDENT: value form (functions already evaluated)
	Tasks  map[string]task.Task // Map of task name to the task itself
	Globs  map[string][]string  // Map of glob pattern to their concrete filepaths (avoids recalculating)
	Path   string               // The absolute path to the spokfile
	Dir    string               // The directory under which the spokfile sits
}

// HasTask returns whether or not the SpokFile has a task with the given name.
func (s *SpokFile) HasTask(name string) bool {
	_, ok := s.Tasks[name]
	return ok
}

// hasGlob returns whether or not the SpokFile has already expanded a glob pattern.
func (s *SpokFile) hasGlob(pattern string) bool {
	expanded, ok := s.Globs[pattern]
	if !ok {
		return false
	}
	if len(expanded) == 0 {
		return false
	}
	return true
}

// Env returns the spokfile Vars as a string slice of KEY=VALUE format
// so it may be passed to running task commands.
func (s *SpokFile) Env() []string {
	results := make([]string, 0, len(s.Vars))
	for key, val := range s.Vars {
		results = append(results, key+"="+val)
	}
	return results
}

// expandGlobs gathers up all the glob patterns in every task in the spokfile and expands them
// saving the results to the Globs map as e.g. {"**/*.go": ["file1.go", "file2.go"]}.
func (s *SpokFile) expandGlobs() error {
	start := time.Now()
	count := 0
	var toExpand []string
	for _, task := range s.Tasks {
		toExpand = append(toExpand, task.GlobDependencies...)
		toExpand = append(toExpand, task.GlobOutputs...)
	}
	for _, pattern := range toExpand {
		if !s.hasGlob(pattern) {
			matches, err := expandGlob(s.Dir, pattern)
			if err != nil {
				return err
			}
			count += len(matches)
			s.Globs[pattern] = matches
		}
	}
	s.logger.Debug("Expanded globs to %d unique filepaths in %v", count, time.Since(start))
	return nil
}

// buildGraph takes in a list of requested tasks, examines their dependencies, constructs
// and returns the dependency graph.
func (s *SpokFile) buildGraph(requested ...string) (*dag.Graph[string, task.Task], error) {
	start := time.Now()
	// DAG of tasks using the name as the unique id
	graph := dag.New[string, task.Task]()

	// TODO: Make this recursive so it will go through dependencies of dependencies
	for _, name := range requested {
		requestedTask, ok := s.Tasks[name]
		if !ok {
			closest := s.findClosestMatch(name)
			err := fmt.Errorf("Spokfile has no task %q", name)
			if closest != "" {
				// We have a close enough match to do a "did you mean X?"
				err = fmt.Errorf("Spokfile has no task %q. Did you mean %q?", name, closest)
			}
			return nil, err
		}
		// Add the task as a vertex to the graph if it doesn't already exist
		if !graph.ContainsVertex(name) {
			err := graph.AddVertex(name, requestedTask)
			if err != nil {
				return nil, fmt.Errorf("could not add vertex for task %s: %w", name, err)
			}
		}

		// For all of this tasks dependencies, do the same
		for _, dep := range requestedTask.TaskDependencies {
			depTask, ok := s.Tasks[dep]
			if !ok {
				closest := s.findClosestMatch(dep)
				err := fmt.Errorf("Task %q declares a dependency on task %q, which does not exist", requestedTask.Name, dep)
				if closest != "" {
					// We have a close enough match to do a "did you mean X?"
					err = fmt.Errorf("Task %q declares a dependency on task %q, which does not exist. Did you mean %q?", requestedTask.Name, dep, closest)
				}
				return nil, err
			}
			s.logger.Debug("Task %s depends on task %s", requestedTask.Name, depTask.Name)
			if !graph.ContainsVertex(dep) {
				err := graph.AddVertex(dep, depTask)
				if err != nil {
					return nil, fmt.Errorf("could not add vertex for task %s: %w", dep, err)
				}
			}

			// Now create the dependency connection between the parent task and this one
			// dep is the parent here because it must be run before the task we're
			// currently in
			err := graph.AddEdge(dep, name)
			if err != nil {
				return nil, fmt.Errorf("could not add edge %s -> %s: %w", dep, name, err)
			}
		}
	}

	s.logger.Debug("Built dependency graph for requested tasks: %v in %v", requested, time.Since(start))

	return graph, nil
}

// Run runs the specified tasks, it takes force which is a boolean flag set by the CLI which
// always reruns tasks and an io.Writer which is used only to echo the commands being run, the command's stdout and stderr
// is stored in the result.
func (s *SpokFile) Run(stream iostream.IOStream, runner shell.Runner, force bool, tasks ...string) (task.Results, error) {
	// Perform glob expansion for every glob pattern in the whole file and save
	// the list of filepaths to the Globs map
	if err := s.expandGlobs(); err != nil {
		return nil, err
	}

	// Build the task dependency graph based on the requested tasks and their dependencies
	dag, err := s.buildGraph(tasks...)
	if err != nil {
		return nil, err
	}

	// Topological sort on the DAG to determine a run order
	sortStart := time.Now()
	runOrder, err := dag.Sort()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(runOrder))
	for _, taskToRun := range runOrder {
		names = append(names, taskToRun.Name)
	}
	s.logger.Debug("Calculated topological sort of dependency graph %v in %v", names, time.Since(sortStart))

	// Submit the run order to be executed and gather up the results
	results, err := s.run(stream, runner, force, runOrder)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// run is the implementation of the public Run method.
func (s *SpokFile) run(stream iostream.IOStream, runner shell.Runner, force bool, runOrder []task.Task) (task.Results, error) {
	results := make(task.Results, 0, len(runOrder))

	cachePath := filepath.Join(s.Dir, cache.Path)
	if !cache.Exists(cachePath) {
		// Spok has not run at all before and the cache does not exist
		// so just dump a placeholder cache in with all the task names and empty digest entries
		s.logger.Debug("Spok cache at %s not found, initialising new cache", cachePath)
		if err := cache.Init(cachePath, maps.Keys(s.Tasks)...); err != nil {
			return nil, err
		}
	}

	cachedState, err := cache.Load(cachePath)
	if err != nil {
		return nil, fmt.Errorf("Could not load spok cache file at %q: %s", cachePath, err)
	}

	// Whether or not we want to update the cache after running e.g.
	// if there were no file dependencies to update or if the task
	// did not succeed
	updateCache := true

	for _, taskToRun := range runOrder {
		// Gather up all the files to be hashed into a single slice
		var toHash []string

		// First, any glob file dependencies need their expanded files retrieving from
		// the s.Globs map of pattern -> slice
		for _, pattern := range taskToRun.GlobDependencies {
			globs := s.Globs[pattern]
			toHash = append(toHash, globs...)
			s.logger.Debug("Task %s glob dependency pattern %q expanded to %d files", taskToRun.Name, pattern, len(globs))
		}

		// Second, any non-glob file dependencies
		toHash = append(toHash, taskToRun.FileDependencies...)

		s.logger.Debug("Task %s depends on %d files", taskToRun.Name, len(toHash))

		// If the task did not declare any file dependencies, let's not
		// update the cache, this way it will always run
		if len(toHash) == 0 {
			updateCache = false
		}

		var hasher hash.Hasher
		if force {
			hasher = hash.AlwaysRun{}
		} else {
			hasher = hash.New()
		}

		hashStart := time.Now()
		currentDigest, err := hasher.Hash(toHash)
		if err != nil {
			return nil, err
		}
		s.logger.Debug("Calculated digest of %d files in %v", len(toHash), time.Since(hashStart))

		// By the time we get here, we know the cache file will exist (even if it has no digests)
		// so we can go ahead and load as normal. If a task is not in the cache, it means it was
		// added to the spokfile since we last ran a cache, so add it to the current cachedState
		cachedDigest, ok := cachedState.Get(taskToRun.Name)
		if !ok {
			cachedState.Set(taskToRun.Name, "")
		}

		s.logger.Debug("Task %s current checksum: %.15s cached checksum: %.15s", taskToRun.Name, currentDigest, cachedDigest)

		var result shell.Results
		skipped := false

		switch {
		case cachedDigest == "" || currentDigest != cachedDigest:
			// The digest is either empty or out of date, in which case the action to be taken is the same
			// update the cache digest and run the task
			if updateCache {
				cachedState.Set(taskToRun.Name, currentDigest)
			}
			result, err = taskToRun.Run(runner, stream, s.Env())
			if err != nil {
				return nil, fmt.Errorf("Task %q encountered an error: %w", taskToRun.Name, err)
			}

		case currentDigest == cachedDigest:
			// This task has been run before and its digest has not changed, therefore
			// we don't need to run it again
			skipped = true
			updateCache = false
		}

		// Gather up all the task results
		results = append(results, task.Result{CommandResults: result, Task: taskToRun.Name, Skipped: skipped})
	}

	// Only update the cache if force was not set, the task declares file dependencies
	// and the task run was successful
	if !force && updateCache && results.Ok() {
		s.logger.Debug("Updating cached state")
		if err := cachedState.Dump(cachePath); err != nil {
			return nil, err
		}
	}

	return results, nil
}

// findClosestMatch takes the name of a task contained in the spokfile
// and finds the closest matching task. If no matches are found, an empty string is returned.
func (s *SpokFile) findClosestMatch(task string) string {
	names := make([]string, 0, len(s.Tasks))
	for _, t := range s.Tasks {
		names = append(names, t.Name)
	}
	matches := fuzzy.RankFindNormalizedFold(task, names)
	sort.Sort(matches)
	if len(matches) != 0 {
		return matches[0].Target
	}

	return ""
}

// Find climbs the file tree from 'start' to 'stop' looking for a spokfile,
// if it hits 'stop' before finding one, an error will be returned
// If a spokfile is found, it's absolute path will be returned
// typical usage will make start = $CWD and stop = $HOME.
func Find(logger logger.Logger, start, stop string) (string, error) {
	for {
		logger.Debug("Looking in %s for spokfile", start)
		entries, err := os.ReadDir(start)
		if err != nil {
			return "", fmt.Errorf("could not read directory '%s': %w", start, err)
		}

		for _, e := range entries {
			if !e.IsDir() && e.Name() == NAME {
				// We've found it
				abs, err := filepath.Abs(filepath.Join(start, e.Name()))
				if err != nil {
					return "", fmt.Errorf("could not resolve '%s': %w", e.Name(), err)
				}
				return abs, nil
			} else if start == stop {
				return "", errors.New("No spokfile found")
			}
		}
		start = filepath.Dir(start)
	}
}

// New converts a parsed spok AST into a concrete File object,
// root is the absolute path to the directory to use as root for glob
// expansion, typically the path to the directory the spokfile sits in.
func New(tree ast.Tree, root string, logger logger.Logger) (*SpokFile, error) {
	file := SpokFile{
		logger: logger,
		Path:   filepath.Join(root, NAME),
		Dir:    root,
		Vars:   make(map[string]string),
		Tasks:  make(map[string]task.Task),
		Globs:  make(map[string][]string),
	}

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

			if file.HasTask(task.Name) {
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
	var matches []string
	ignoreHiddenGlobFn := func(path string, d fs.DirEntry) error {
		if strings.HasPrefix(path, ".") {
			return filepath.SkipDir
		}

		abs, err := filepath.Abs(filepath.Join(root, path))
		if err != nil {
			return fmt.Errorf("could not resolve path '%s': %w", filepath.Join(root, path), err)
		}

		matches = append(matches, abs)
		return nil
	}

	err := doublestar.GlobWalk(os.DirFS(root), pattern, ignoreHiddenGlobFn)
	if err != nil {
		return nil, fmt.Errorf("could not expand glob pattern %q: %w", pattern, err)
	}

	return matches, nil
}
