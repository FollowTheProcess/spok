package cmd

import (
	"fmt"

	"github.com/fatih/color"
)

var headerStyle = color.New(color.FgWhite, color.Bold) // Setting header style to use in usage message (usage.go)

// Custom usage template with the header style applied, here by itself because it looks kind of messy.
var usageTemplate = fmt.Sprintf(`%s:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

%s:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

%s:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

%s:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`,
	headerStyle.Sprint("USAGE"), headerStyle.Sprint("ALIASES"), headerStyle.Sprint("EXAMPLES"), headerStyle.Sprint("AVAILABLE COMMANDS"),
	headerStyle.Sprint("FLAGS"), headerStyle.Sprint("GLOBAL FLAGS"))
