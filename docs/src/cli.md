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
