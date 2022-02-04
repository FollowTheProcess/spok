<p align="center">
<img src="https://github.com/FollowTheProcess/spok/raw/main/img/logo.png" alt="logo" width=75%>
</p>

# Spok

[![License](https://img.shields.io/github/license/FollowTheProcess/spok)](https://github.com/FollowTheProcess/spok)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/spok)](https://goreportcard.com/report/github.com/FollowTheProcess/spok)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/spok?logo=github&sort=semver)](https://github.com/FollowTheProcess/spok)
[![CI](https://github.com/FollowTheProcess/spok/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/spok/actions?query=workflow%3ACI)
<!-- [![codecov](https://codecov.io/gh/FollowTheProcess/spok/branch/main/graph/badge.svg?token=Q8Y5KFA9ZK)](https://codecov.io/gh/FollowTheProcess/spok) -->

***It's a build system Jim, but not as we know it! 🖖🏻***

* Free software: Apache Software License 2.0

## Project Description

:warning: **Spok is in early development and is not ready for use yet!**

`spok` is a lightweight build system and command runner inspired by things like [make], [just] and others.

On top of this, `spok` provides:

* A cleaner, more "developer friendly" syntax
* Fully cross compatible (tested on Windows, Linux and Mac)
* Incremental runs based on file hashing and sum checks (not dates like e.g. [make]), so nothing runs if nothing's changed!
* Parallel execution by default (unless task dependencies preclude this)
* More features TBC

## Installation

## Quickstart

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
# Comments are preceeded by a hash

# You can store global variables like this (caps are optional)
GLOBAL_VARIABLE := "hello
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
    echo GLOBAL_VARIABLE
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
task build(fmt) -> "./bin/main" {
    go build
}

# Can also use global variables as outputs
task build2(fmt) -> BIN {
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

### Credits

This package was created with [cookiecutter] and the [FollowTheProcess/go_cookie] project template.

[cookiecutter]: https://github.com/cookiecutter/cookiecutter
[FollowTheProcess/go_cookie]: https://github.com/FollowTheProcess/go_cookie
[make]: https://www.gnu.org/software/make/
[just]: https://github.com/casey/just
