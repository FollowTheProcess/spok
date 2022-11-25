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
