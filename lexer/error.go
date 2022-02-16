package lexer

import "fmt"

// syntaxError represents a lexical syntax error.
type syntaxError struct {
	message string
	line    int
}

func (s syntaxError) Error() string {
	return fmt.Sprintf("SyntaxError: %s (Line %d)", s.message, s.line)
}
