<p align="center">
<img src="https://github.com/FollowTheProcess/spok/raw/main/docs/img/logo.png" alt="logo" width=75%>
</p>

# Spok

[![License](https://img.shields.io/github/license/FollowTheProcess/spok)](https://github.com/FollowTheProcess/spok)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/spok)](https://goreportcard.com/report/github.com/FollowTheProcess/spok)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/spok?logo=github&sort=semver)](https://github.com/FollowTheProcess/spok)
[![CI](https://github.com/FollowTheProcess/spok/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/spok/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/FollowTheProcess/spok/branch/main/graph/badge.svg?token=Q8Y5KFA9ZK)](https://codecov.io/gh/FollowTheProcess/spok)
[![pre-commit.ci status](https://results.pre-commit.ci/badge/github/FollowTheProcess/spok/main.svg)](https://results.pre-commit.ci/latest/github/FollowTheProcess/spok/main)

***It's a build system Jim, but not as we know it! 🖖🏻***

* Free software: Apache Software License 2.0

> [!NOTE]
> Spok was somewhat of an educational project for me and whilst it works and I use it, I have a large number of improvements and new ideas in mind that are best suited to a fresh start (which I will do soon 👀). In the meantime Spok will continue to be supported but no large changes are likely to be made as my effort goes to the new thing (whatever that is)

## Project Description

`spok` is a lightweight build system and command runner inspired by things like [make], [just] and others.

On top of this, `spok` provides:

* A cleaner, more "developer friendly" syntax
* Fully cross compatible (tested on Windows, Linux and Mac)
* Ships with it's own shell interpreter so no dependency on `sh`, `bash`, or `zsh`!
* Incremental runs based on file hashing and sum checks (not timestamps like e.g. [make]), so nothing runs if nothing's changed!
* Incredibly fast execution! Expensive operations are cached, only does *any* work when it's absolutely necessary
* Auto loading of `.env` files
* Debug info with the `--debug` flag
* An auto `spokfile` formatter
* More features TBC

## Installation

There are binaries for Mac, Linux and Windows in the [GitHub releases] section, just download the correct one for your OS and architecture and place it somewhere on `$PATH`.

For Mac and Linux there is also a [homebrew] tap:

```shell
brew install FollowTheProcess/homebrew-tap/spok
```

## Quickstart

To get started with spok, simply create a `spokfile` in your project and add a task:

```python
# Run the go tests
task test() {
    go test ./...
}
```

We also recommend you add the following to your `.gitignore`:

```gitignore
.spok/
```

Now on the command line you can run:

```shell
spok test
```

And your tests will be run!

If you want spok to help you out by initialising a demo spokfile (and adding the `.spok` entry to `.gitignore`) you can run:

```shell
spok --init
```

## The Spokfile

Spok is driven by a single file (usually placed in the root of your project) called `spokfile`.

The syntax for a `spokfile` is inspired by a few different things and is intended to be expressive yet very simple:

#### Makefiles

* The general structure of a `spokfile` will be broadly familiar to those who have used [make] or [just] before.
* Tasks (make's targets or just's recipes) are independent declarations. A `spokfile` can have any number of tasks declared but each must have a unique name.
* Global variable definitions look similar.

#### Go

* Spok borrows quite a bit of Go syntax!
* Global variables use the `:=` declaration operator.
* Tasks look kind of like Go functions.
* Tasks that output multiple things use Go's multiple return syntax.
* Task bodies are bounded with curly braces.

#### Python

* Although in general whitespace is not significant in a `spokfile`, you'll notice there are no semicolons!
* Spok looks for line breaks in certain contexts to delimit things.
* Task outputs make use of the `->` operator similar to declaring function return types in Python.

### Example

A `spokfile` looks like this...

<!-- Ignore the python syntax highlighting, it's obviously not python. It's just what looked best when rendered on GitHub -->
```python
# Comments are preceded by a hash

# You can store global variables like this (caps are optional)
# these will also be exported as environment variables available for use
# in any tasks commands
GLOBAL_VARIABLE := "hello"
BIN := "./bin/main"

# You can store the output of a shell command as a variable
# leading and trailing whitespace will always be trimmed off when doing this
GIT_COMMIT := exec("git rev-parse HEAD")

# The core concept in spok is a task (think make target)
# they are sort of based on go functions except arguments are dependencies
# A dependency can be filepaths (including globs) or names of other tasks

# Tasks have optional outputs (if they generate things)
# This enables `spok --clean` to restore everything to it's original state

# Generally, a task is structured like this...

# A line comment above a task is it's docstring
# task <name>(<deps>?...) -> [(]<outputs>?...[)] {
#     command(s) to run
# }

# Some simple examples below

# Use a global variable like this
task hello() {
    echo {{.GLOBAL_VARIABLE}}
}

# Run the go tests (depends on all go source files)
task test("**/*.go") {
    go test ./...
}

# Format the project source code (depends on all go source files)
# if the go source files have not changed, this becomes a no op
task fmt("**/*.go") {
    go fmt ./...
}

# Compile the program (depends on fmt, fmt will run first)
# also outputs a build binary
task build(fmt, "**/*.go") -> "./bin/main" {
    go build
}

# Can also use global variables as outputs
task build2(fmt, "**/*.go") -> BIN {
    go build
}

# Tasks can generate multiple things
task many("**/*.go") -> ("output1.go", "output2.go") {
    go do many things
}

# Can also do glob outputs
# e.g. tasks that populate entire directories like building documentation
task glob("docs/src/*.md") -> "docs/build/*.html" {
    build docs
}

# Can register a default task (by default spok will list all tasks)
task default() {
    echo "default"
}

# Can register a custom clean task
# By default `spok --clean` will remove all declared outputs
# if a task called "clean" is present in the spokfile
# this task will be run instead when `--clean` is used
task clean() {
    rm -rf somedir
}
```

## Benchmarks

Although still in early development, I've benchmarked spok against some very large repos and it performs very well!

For example on the [golang/go] repo itself with 8872 `.go` files (at the time of writing) and the following benchmark task:

```python
# Benchmark hashing all go files
task test("**/*.go") {
    echo "I depend on all go files"
}
```

![go_files](https://github.com/FollowTheProcess/spok/raw/main/docs/img/go_files.png)

Spok is able to hash all 8872 files in just 300ms!

![benchmark](https://github.com/FollowTheProcess/spok/raw/main/docs/img/benchmark.png)

Does that mean I can call spok **"Blazingly Fast!"**? 🤔

## Editor Support

There is a [VSCode Extension] available that provides basic syntax highlighting for spokfiles. It's still in active development so more features TBC!

[make]: https://www.gnu.org/software/make/
[just]: https://github.com/casey/just
[GitHub releases]: https://github.com/FollowTheProcess/spok/releases
[homebrew]: https://brew.sh
[VSCode Extension]: https://marketplace.visualstudio.com/items?itemName=FollowTheProcess.spok
[golang/go]: https://github.com/golang/go
