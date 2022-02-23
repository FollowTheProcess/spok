package parser

import (
	"fmt"
	"strings"

	"github.com/FollowTheProcess/spok/token"
)

// illegalToken is an error that handles an unexpected token encounter during the parse
// it shows a nice message with a list of expected tokens.
type illegalToken struct {
	line        string
	expected    []token.Type
	encountered token.Token
}

func (i illegalToken) Error() string {
	var expecteds []string
	for _, exp := range i.expected {
		expecteds = append(expecteds, fmt.Sprintf("'%s'", exp.String()))
	}
	switch len(expecteds) {
	case 1:
		return fmt.Sprintf("Illegal Token: %s (Line %d). Expected %s\n\n%d |\t%s", i.encountered, i.encountered.Line, expecteds[0], i.encountered.Line, i.line)
	default:
		return fmt.Sprintf("Illegal Token: %s (Line %d). Expected one of [%s]\n\n%d |\t%s", i.encountered, i.encountered.Line, strings.Join(expecteds, ", "), i.encountered.Line, i.line)
	}
}
