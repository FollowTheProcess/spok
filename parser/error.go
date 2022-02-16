package parser

import (
	"fmt"
	"strings"

	"github.com/FollowTheProcess/spok/token"
)

// illegalToken is an error that handles a unexpected token encounter during the parse
// it shows a nice message with a list of expected tokens.
type illegalToken struct {
	expected    []token.Type
	encountered token.Token
	line        int
}

func (i illegalToken) Error() string {
	expecteds := []string{}
	for _, exp := range i.expected {
		expecteds = append(expecteds, exp.String())
	}
	switch len(expecteds) {
	case 1:
		return fmt.Sprintf("Illegal Token: %s (Line %d). Expected '%s'", i.encountered, i.line, expecteds[0])
	default:
		return fmt.Sprintf("Illegal Token: %s (Line %d). Expected one of (%s)", i.encountered, i.line, strings.Join(expecteds, ", "))
	}
}
