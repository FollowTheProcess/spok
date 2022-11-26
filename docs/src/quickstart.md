# Quickstart

Getting started with spok is easy! All you need to do is run:

<div class="termy">

```shell
$ spok --init

✅ Done
```

</div>

This will create a spokfile in your current directory that looks like this:

```python
# This is a spokfile example

VERSION := "0.3.0"

# Run the unit tests
task test("**/*.go") {
    go test ./...
}

# Which version am I
task version() {
    echo {{.VERSION}}
}
```

And add the following to your `.gitignore`:

```gitignore
# Ignore the spok cache directory
.spok/
```

So now you're ready to go, if you run `spok` you'll see the following:

<div class="termy">

```console
$ spok
Tasks defined in /Users/you/yourproject/spokfile:
Name        Description
test        Run the unit tests
version     Which version am I
```

</div>

If you run `spok version` you'll see:

<div class="termy">

```console
$ spok version
0.3.0
```

</div>

Next let's take a look at the spokfile [syntax](syntax.md) and how to use it.
