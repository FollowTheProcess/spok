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

If you've read the [syntax guide](syntax.md) you'll know that Spok calculates the state of the dependency graph by hashing the contents of
all the declared files in your task declaration. This avoids unnecessary work by only running tasks who's dependencies have changed.

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
