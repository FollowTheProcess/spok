package lexer

import "fmt"

// syntaxError represents a lexical syntax error.
type syntaxError struct {
	message string
	context string
	line    int
	pos     int
}

func (s syntaxError) Error() string {
	return fmt.Sprintf("SyntaxError: %s (Line %d). \n\n%d |\t%s", s.message, s.line, s.line, s.context)
}
