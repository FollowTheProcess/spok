# CLI

Aside from what you can with the spokfile syntax itself, there's also a bunch of stuff you can
do from the CLI.

## Usage

Let's start by showing the help:

<div class="termy">

```console

$ spok --help

It's a build system Jim, but not as we know it!

Spok is a lightweight build system and command runner, inspired by things like
make, just etc.

However, spok offers a number of additional features such as:

- Cleaner, more developer-friendly syntax
- Full cross compatibility
- No dependency on any form of shell
- Load .env files by default
- Incremental runs based on file hashing and sum checks

USAGE:
  spok [tasks]... [flags]

FLAGS:
  -c, --clean             Remove all build artifacts.
      --fmt               Format the spokfile.
  -f, --force             Bypass file hash checks and force running.
  -h, --help              help for spok
      --init              Initialise a new spokfile in $CWD.
  -j, --json              Output task results as JSON.
  -q, --quiet             Silence all CLI output.
  -s, --show              Show all tasks defined in the spokfile.
      --spokfile string   The path to the spokfile (defaults to '$CWD/spokfile').
  -V, --vars              Show all defined variables in spokfile.
  -v, --verbose           Show verbose logging output.
      --version           version for spok

```

</div>

Some of this stuff we've already talked about, but let's look at some stuff we haven't touched on yet.

## `--fmt`

The `--fmt` flag is used to format the spokfile. Spok comes equipped with an (albeit basic) formatter that parses the spokfile
and then dumps it back in place with the desired formatting, simple really!

!!! note

    Because the spokfile has to be parsed before formatting, it's not possible to format a spokfile that contains syntax errors.

## `--force`

If you've read the [syntax guide](syntax.md) you'll know that Spok calculates the state of the task dependency graph by hashing the contents of
all the declared files in your task definition. This avoids unnecessary work by only running tasks who's dependencies have changed.

However, sometimes you want to force a task to run regardless of whether it's dependencies have changed or not. This is where the `--force` flag comes in.

<div class="termy">

```console
$ spok test
- Task "test" skipped as none of it's dependencies have changed

// Okay fine, let's force it to run
$ spok test --force

ok   github.com/FollowTheProcess/spok/ast (cached)
ok   github.com/FollowTheProcess/spok/builtins (cached)
ok   github.com/FollowTheProcess/spok/cache (cached)
?    github.com/FollowTheProcess/spok/cli/app [no test files]
?    github.com/FollowTheProcess/spok/cli/cmd [no test files]
ok   github.com/FollowTheProcess/spok/cmd/spok (cached)
ok   github.com/FollowTheProcess/spok/file (cached)
ok   github.com/FollowTheProcess/spok/graph (cached)
ok   github.com/FollowTheProcess/spok/hash (cached)
?    github.com/FollowTheProcess/spok/iostream [no test files]
ok   github.com/FollowTheProcess/spok/lexer (cached)
?    github.com/FollowTheProcess/spok/logger [no test files]
ok   github.com/FollowTheProcess/spok/parser (cached)
ok   github.com/FollowTheProcess/spok/shell (cached)
ok   github.com/FollowTheProcess/spok/task (cached)
ok   github.com/FollowTheProcess/spok/token (cached)

✅ Task "test" completed successfully
```

</div>

## `--json`

By default, spok outputs the results of the running tasks in their original format straight to the terminal. This is great for humans, but not so great for machines.

If you want to programmatically access the results of a spok run, you can use the `--json` flag to output the results as JSON, which can then be queried
by external programs e.g. [jq](https://stedolan.github.io/jq/).

For example, let's run a sample task and pipe the output to `jq`:

Here's the spokfile:

```python
# Do some things with JSON
task echo() {
    echo "I succeeded"
}
```

Running `spok echo --json` will get you:

<div class="termy">

```console
$ spok echo --json | jq

[
  {
    "task": "echo",
    "command_results": [
      {
        "cmd": "echo I succeeded",
        "stdout": "I succeeded\n",
        "stderr": "",
        "status": 0
      }
    ],
    "skipped": false
  }
]


```

</div>

Spok's output JSON is a list of objects, each object representing a task. Each task object contains the following fields:

- `task`: The name of the task
- `command_results`: A list of objects, each object representing a command run by the task. Each command object contains the following fields:
  - `cmd`: The command that was run
  - `stdout`: The stdout of the command
  - `stderr`: The stderr of the command
  - `status`: The exit status of the command

You can imagine how this could be useful for things like CI/CD pipelines where tasks are more complicated and you may need
to query or parse the results of a task or a whole run.

## `--quiet`

The `--quiet` flag does exactly what it says on the tin, shuts Spok up!

When using the `--quiet` flag, Spok will not show any of the task results or any other output at all, simply exit with a zero status
code if the run was successful or a non-zero status code if it wasn't.

I'd include an example here, but by definition it would be empty! 🤓

## `--show`

The `--show` flag simply displays all the tasks and their docstrings if present. By default, Spok will do this when it
is invoked with no arguments, unless you have declared a task called `default`, but more on that later! TODO

To show you what this looks like, consider a simple spokfile:

```python
# Run the unit tests
task test() {
    go test ./...
}

# Format the source code
task fmt() {
    go fmt ./...
}

# Run the linter
task lint() {
    golangci-lint run --fix
}

```

Running `spok` with no arguments, or `spok --show` will get you:

<div class="termy">

```console
$ spok --show
Tasks defined in /Users/you/yourproject/spokfile:
Name    Description
fmt     Run go fmt on all project files
lint    Lint the project and auto-fix errors if possible
test    Run all project tests

```

</div>
