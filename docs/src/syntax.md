# Syntax

When spok runs, it parses the syntax in your spokfile, extracts tasks, variables, and dependencies, and then runs the tasks you specify. In this
section we'll take a look at the syntax and how to use it

## Global Variables

Let's start simple, Spok lets you define variables in the global scope. These variables can be used in any task. For example:

```python
VERSION := "0.3.0"

# Show the version
task version() {
    echo {{.VERSION}}
}
```

!!! tip
    Global variables are also exported as environment variables to the tasks, so if your tasks invoke other scripts that depend on
    environment variables you can just declare them globally in spok.

## Builtin Functions

Sometimes you need to do more than just run a test, or you need to do something that is not supported by the shell. Spok has a few builtin functions that you can use in your tasks. These functions are:

- `join` - Joins a list of path parts with an OS specific path separator (relative to the spokfile)
- `exec` - Executes a shell command and captures the output (stripped of leading/trailing whitespace) in the variable it's assigned to

You use them like this:

```python
DOCS_SRC := join("docs", "src")

GIT_COMMIT := exec("git rev-parse HEAD")
```

## Tasks

Tasks are the main point of Spok and are most likely where you'll spend most of your time. Tasks are preceded with the `task` keyword followed by
the task definition.

For example, a simple task that runs unit tests might look like:

```python
task test() {
    go test ./...
}
```

!!! warning

    Just like functions in most programming languages, tasks must have opening and closing parentheses, omitting them is
    grounds for a syntax error.

Now that you have a task defined, you can run it with `spok test` and your tests will run, how cool is that! 🎉

### Task Documentation

If you want to document your tasks, you can do so by adding a comment above the task definition. For example:

```python
# Run all project unit tests
task test() {
    go test ./...
}
```

Spok will parse this as the task's docstring and it will be output when the tasks are listed, either by the default action
or the `--show` flag. But we'll get to that later in the [CLI](cli.md) section 👍

### Tasks that Depend on Files

This is fine, and might be enough for you if your test suite is fast and/or the language tooling you're using caches results (like Go!). But what
if you have a large test suite and only want to re-run the tests if the source code has changed? Or what if you have a task that depends on another?

Luckily, Spok supports both of these types of dependencies!

Let's say you're working in a very large python project and the tests take a while to run. You don't want to run the tests if the source code hasn't
changed since you last ran them. You can do this by adding a dependency to the `test` task:

```python
task test("**/*.py") {
    python -m pytest
}
```

By including the `"**/*.py"`, Spok will now know that the `test` task depends on all python files in the project, and that (after the initial run)
it should only be *re-run* if any of those files have changed.

So now if you run `spok test`, your tests will run as before. But try running it again! Spok will notice that none of the python files have changed
and you will see:

<div class="termy">

```console
$ spok test
- Task "test" skipped as none of its dependencies have changed
```

</div>

You can do this with as many files or glob patterns as you want, if any of them have changed, the task will be re-run, e.g. this is
completely allowed:

```python
task lots("**/*.go", "**/*.py", "some/specific/file.txt") {
    ...
}
```

!!! info

    When you declare file dependencies like this, behind the scenes Spok will expand the glob patterns to their concrete, absolute filepaths,
    open and read the contents of each one, and then generate a SHA256 hash of the contents, summing it all up into a final digest and
    caching this digest against the name of the task.

    When you run the task again, Spok will do the same procedure, and compare the newly calculated digest against the cached one to determine
    if the task should be re run. This type of content checking is more accurate than e.g. [make](https://www.gnu.org/software/make/) which
    looks at file modification timestamps.

#### A Note on Performance

> "But if you open and read the contents of every single file every time you run the task, isn't this really slow?"

In a word... no! Spok is designed to be *fast*:

- The expansion of glob patterns happens once, when the spokfile is parsed, and the results are cached in memory for re-use
- The opening, reading and hashing of file contents happens concurrently across all your cores
- It's written in Go so it's naturally pretty fast anyway!

All this means that, even on very large projects, Spok can perform this check in a few hundred milliseconds 🚀

For example on the [golang/go](https://github.com/golang/go) repo itself with 8872 `.go` files (at the time of writing) and the following benchmark task:

```python
# Benchmark hashing all go files
task test("**/*.go") {
    echo "I depend on all go files"
}
```

![go_files](https://github.com/FollowTheProcess/spok/raw/main/docs/img/go_files.png)

![benchmark](https://github.com/FollowTheProcess/spok/raw/main/docs/img/benchmark.png)

Spok is able to detect that nothing has changed in any of the 8872 files in just 300ms on my laptop! Don't forget, this also includes the time it takes to:

- Launch the program itself
- Read and parse the spokfile
- Expand the glob pattern `"**/*.go"` and collect the results

So hopefully it's plenty fast enough!

### Tasks that Depend on Other Tasks

Not only can you depend on files, you can also depend on other tasks, or a mix of both! If you put the name of another task in the task arguments,
Spok will recognise this as a task dependency and will ensure that the declared task will always run before the one you want.

For example, let's say you want a Spok task to compile your project, but before that you have to run some sort of code generation, or
you want to run the linter or formatter first. You can do this by declaring a dependency in your build task on whatever you want
to run before it:

```python
# Run the formatter
task fmt() {
    go fmt ./...
}

# Run the linter
task lint(fmt) {
    golangci-lint run
}

# Compile the project
task build(lint) {
    go build ./...
}
```

In this example, the `build` task depends on the `lint` task, which in-turn depends on the `fmt` task. So when you run `spok build`, Spok will
construct the dependency graph of the requested tasks, and then run them in the correct order, so you should see something like:

<div class="termy">

```console
$ spok build

✅ Task "fmt" completed successfully
✅ Task "lint" completed successfully
✅ Task "build" completed successfully
```

</div>

You can also mix and match tasks depending on files and each other, as in the following example:

```python
# Generate the API schema from swagger
task generate("api/swagger.yaml") {
    swagger generate spec -o api/schema.json
}

# Compile the project
task build(generate, "**/*.go") {
    go build ./...
}
```

Here we want to generate an API schema from a swagger file, but only if the swagger file has changed. We also want to compile the project
which may include or embed this swagger schema, so `generate` will always run first, and then `build` will run, but only if any of the
go files have changed since the last run.

Hopefully you can see how powerful this is! Using a very simple and expressive syntax, you can build up complex dependency graphs of tasks
that will only run when they need to, so no time is wasted doing unnecessary work.

!!! note

    At the moment, spok will not recurse down the dependency graph and so will not look at dependencies of dependencies. For example if you
    had the following spokfile:

    ```python
    # Run the linter
    task lint("**/*.go") {
        golangci-lint run
    }

    # Run the tests
    task test(lint) {
        go test ./...
    }
    ```

    Task `test` would always run after `lint`, but it would run **even if no go files had changed.** To declare that `test` should also only
    run if any of the go files have changed, you would need to add the dependency to `test` as well:

    ```python
    # Run the tests
    task test(lint, "**/*.go") {
        go test ./...
    }
    ```

### Task Outputs

Some tasks generate external artifacts, such as compiled binaries, or generated code. In Spok, you can explicitly declare this by using
the output operator `->`. For example, here's a task that compiles a Go binary and saves it under the `bin` directory:

```python
# Build the project
task build("**/*.go") -> "bin/myproject" {
    go build -o bin/myproject ./...
}
```

Declaring an output like this is optional, but it has a pretty cool benefit!

<div class="termy">

```console
$ spok --clean

Removed /Users/you/myproject/bin/myproject
```

</div>

Yep, Spok can automatically clean up after you! 🎉

This is useful if you generate a lot of stuff and want to easily clean up after yourself e.g. if your tasks generate
profiles, coverage reports, flamegraphs or other artifacts that you don't want to commit to your repo.

You can also declare multiple outputs by separating them with commas inside parentheses, like so:

```python
# Output lots of stuff
task build("**/*.go") -> ("bin/myproject", "bin/myproject2") {
    go build -o bin/myproject ./...
    go build -o bin/myproject2 ./...
}
```

Or even glob patterns of outputs! For example if your task is to convert markdown to html docs pages, you could do something like:

```python
# Build the docs html from source
task docs("docs/src/*.md") -> "docs/build/*.html" {
    mkdocs build
}
```

Just like with file dependencies, these globs will be expanded to their concrete filepaths and each one would be deleted
by `spok --clean`

That's really it for the syntax! Let's move on and talk about what you can do with the [CLI](cli.md)
