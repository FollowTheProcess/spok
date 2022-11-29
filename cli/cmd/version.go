package cmd

import "fmt"

var versionTemplate = fmt.Sprintf(
	`{{printf "%s %s\n%s %s\n%s %s\n%s %s\n"}}`,
	headerStyle.Sprint("Version:"),
	version,
	headerStyle.Sprint("Commit:"),
	commit,
	headerStyle.Sprint("Build Date:"),
	buildDate,
	headerStyle.Sprint("Built By:"),
	builtBy,
)
