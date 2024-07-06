package cmd

import (
	"fmt"

	"github.com/fatih/color"
)

var headerStyle = color.New(color.Bold, color.Underline) // Setting header style to use in usage message

// Custom usage template with the header style applied, here by itself because it looks kind of messy.
var usageTemplate = fmt.Sprintf(`%s:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}} {{.CommandPath}} [COMMAND]{{end}}{{if gt (len .Aliases) 0}}

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
	headerStyle.Sprint("Usage"), headerStyle.Sprint("Aliases"), headerStyle.Sprint("Examples"), headerStyle.Sprint("Commands"),
	headerStyle.Sprint("Options"), headerStyle.Sprint("Global Options"))
