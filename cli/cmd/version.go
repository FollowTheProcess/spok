package cmd

import "fmt"

var versionTemplate = fmt.Sprintf(`{{printf "%s %s\n%s %s\n"}}`, headerStyle.Sprint("Version:"), version, headerStyle.Sprint("Commit:"), commit)
