// Package parser implements spok's parser.
//
// Spok is a very simple mostly declarative language and as such, the parser is incredibly simple. It is
// a simple top-down parser with a very small initial state space and requires only 1 token of lookahead.
//
// It switches on the token it encounters to process the appropriate ast.Node (parseXXX methods), appending
// each to the list of nodes as it goes.
//
// If the parser encounters an error, either an ERROR token from the lexer, or an error of it's own making
// it will immediately return with the AST it's managed to parse so far and the error.
package parser

import (
	"errors"
	"strings"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/FollowTheProcess/spok/lexer"
	"github.com/FollowTheProcess/spok/token"
)

// Parser is spok's AST parser.
type Parser struct {
	lexer     lexer.Tokeniser // The lexer
	input     string          // The raw input, used for showing error context based on token line and position
	buffer    [3]token.Token  // 3 token buffer, allows us to peek and backup in the token stream
	peekCount int             // How far we've peeked into our buffer
}

// New creates and returns a new Parser for an input string.
func New(input string) *Parser {
	return &Parser{
		lexer: lexer.New(input),
		input: input,
	}
}

// Parse is the top level parse method. It parses the entire
// input text to EOF or Error and returns the full AST.
func (p *Parser) Parse() (ast.Tree, error) {
	tree := ast.Tree{}

	for next := p.next(); !next.Is(token.EOF); {
		switch {
		case next.Is(token.ERROR):
			return tree, errors.New(next.Value)

		case next.Is(token.HASH):
			comment := p.parseComment()
			switch {
			case p.next().Is(token.TASK):
				// The comment was a tasks' docstring
				task, err := p.parseTask(comment)
				if err != nil {
					return tree, err
				}
				tree.Append(task)
			default:
				// Just a normal comment
				p.backup()
				tree.Append(comment)
			}

		case next.Is(token.IDENT):
			assign, err := p.parseAssign(next)
			if err != nil {
				return tree, err
			}
			tree.Append(assign)

		case next.Is(token.TASK):
			// Pass an empty comment in if it doesn't have one
			task, err := p.parseTask(ast.Comment{NodeType: ast.NodeComment})
			if err != nil {
				return tree, err
			}
			tree.Append(task)

		default:
			// Illegal top level token that slipped through the lexer somehow
			// unlikely but let's catch it anyway
			return tree, illegalToken{
				expected:    []token.Type{token.HASH, token.IDENT, token.TASK},
				encountered: next,
				line:        p.getLine(next),
			}
		}
		next = p.next()
	}

	return tree, nil
}

// next returns, and consumes, the next token from the lexer.
func (p *Parser) next() token.Token {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.buffer[0] = p.lexer.NextToken()
	}
	tok := p.buffer[p.peekCount]
	return tok
}

// backup backs up in the input stream by one token.
func (p *Parser) backup() {
	p.peekCount++
}

// expect checks if the next token is of the expected type, consuming it in the process
// if not it will return an unexpected token error.
func (p *Parser) expect(expected token.Type) error {
	switch got := p.next(); {
	case got.Is(token.ERROR):
		// If it's already an error, just return it as is
		// goes without saying that we don't expect an error
		return errors.New(got.Value)
	case !got.Is(expected):
		return illegalToken{
			expected:    []token.Type{expected},
			encountered: got,
			line:        p.getLine(got),
		}
	default:
		return nil
	}
}

// parseComment parses a comment token into a comment ast node,
// the # has already been consumed.
func (p *Parser) parseComment() ast.Comment {
	return ast.Comment{
		Text:     p.next().Value,
		NodeType: ast.NodeComment,
	}
}

// parseIdent parses an ident token into an ident ast node.
func (p *Parser) parseIdent(ident token.Token) ast.Ident {
	return ast.Ident{
		Name:     ident.Value,
		NodeType: ast.NodeIdent,
	}
}

// parseString parses a string token into a string ast node.
func (p *Parser) parseString(s token.Token) ast.String {
	return ast.String{
		Text:     s.Value,
		NodeType: ast.NodeString,
	}
}

// parseFunction parses an ident token into a function ast node.
func (p *Parser) parseFunction(ident token.Token) (ast.Function, error) {
	args := []ast.Node{}

	// If next is not '(', we have a problem
	if err := p.expect(token.LPAREN); err != nil {
		return ast.Function{}, err
	}

	for next := p.next(); !next.Is(token.RPAREN); {
		switch {
		case next.Is(token.STRING):
			args = append(args, p.parseString(next))
		case next.Is(token.IDENT):
			args = append(args, p.parseIdent(next))
		case next.Is(token.COMMA):
			// Absorb a comma
		case next.Is(token.ERROR):
			return ast.Function{}, errors.New(next.Value)
		default:
			return ast.Function{}, illegalToken{
				expected:    []token.Type{token.STRING, token.IDENT, token.RPAREN},
				encountered: next,
				line:        p.getLine(next),
			}
		}
		next = p.next()
	}

	fn := ast.Function{
		Name:      p.parseIdent(ident),
		Arguments: args,
		NodeType:  ast.NodeFunction,
	}
	return fn, nil
}

// parseAssign parses a global variable assignment into an assign ast node.
// the ':=' is known to exist but has yet to be consumed, the encountered ident token is passed in.
func (p *Parser) parseAssign(ident token.Token) (ast.Assign, error) {
	name := p.parseIdent(ident)

	// If next is not ':=', we have a problem
	if err := p.expect(token.DECLARE); err != nil {
		return ast.Assign{}, err
	}

	var rhs ast.Node
	var err error

	switch next := p.next(); {
	case next.Is(token.STRING):
		rhs = p.parseString(next)
	case next.Is(token.IDENT):
		// Only other thing is a built in function or assigning to another ident
		if p.next().Is(token.LPAREN) {
			p.backup()
			rhs, err = p.parseFunction(next)
			if err != nil {
				return ast.Assign{}, err
			}
		} else {
			p.backup()
			rhs = p.parseIdent(next)
		}
	case next.Is(token.ERROR):
		return ast.Assign{}, errors.New(next.Value)
	default:
		return ast.Assign{}, illegalToken{
			expected:    []token.Type{token.STRING, token.IDENT},
			encountered: next,
			line:        p.getLine(next),
		}
	}

	assign := ast.Assign{
		Name:     name,
		Value:    rhs,
		NodeType: ast.NodeAssign,
	}
	return assign, nil
}

// parseTask parses and returns a task ast node, the task keyword has already
// been encountered and consumed, the docstring comment is passed in if present
// and will be empty if there is no comment.
func (p *Parser) parseTask(doc ast.Comment) (ast.Task, error) {
	name := p.parseIdent(p.next())

	// If next is not '(' we have a problem
	if err := p.expect(token.LPAREN); err != nil {
		return ast.Task{}, err
	}

	dependencies, err := p.parseTaskDependencies()
	if err != nil {
		return ast.Task{}, err
	}

	outputs, err := p.parseTaskOutputs()
	if err != nil {
		return ast.Task{}, err
	}

	// If next is not '{', we have a problem
	err = p.expect(token.LBRACE)
	if err != nil {
		return ast.Task{}, err
	}

	commands, err := p.parseTaskCommands()
	if err != nil {
		return ast.Task{}, err
	}

	task := ast.Task{
		Name:         name,
		Docstring:    doc,
		Dependencies: dependencies,
		Outputs:      outputs,
		Commands:     commands,
		NodeType:     ast.NodeTask,
	}

	return task, nil
}

// parseTaskDependencies parses any declared dependencies in a task and returns
// the []ast.Node containing them.
func (p *Parser) parseTaskDependencies() ([]ast.Node, error) {
	dependencies := []ast.Node{}
	for next := p.next(); !next.Is(token.RPAREN); {
		switch {
		case next.Is(token.STRING):
			dependencies = append(dependencies, p.parseString(next))
		case next.Is(token.IDENT):
			dependencies = append(dependencies, p.parseIdent(next))
		case next.Is(token.COMMA):
			// Absorb a comma
		case next.Is(token.ERROR):
			return nil, errors.New(next.Value)
		default:
			return nil, illegalToken{
				expected:    []token.Type{token.STRING, token.ERROR},
				encountered: next,
				line:        p.getLine(next),
			}
		}
		next = p.next()
	}

	return dependencies, nil
}

// parseTaskOutputs parses any declared outputs in a task and returns
// the []ast.Node containing them.
// If no output is declared, this method will backup so that the parser
// state is the same as when it was called.
func (p *Parser) parseTaskOutputs() ([]ast.Node, error) {
	outputs := []ast.Node{}
	if p.next().Is(token.OUTPUT) {
		switch next := p.next(); {
		case next.Is(token.STRING):
			outputs = append(outputs, p.parseString(next))
		case next.Is(token.IDENT):
			outputs = append(outputs, p.parseIdent(next))
		case next.Is(token.COMMA):
			// Absorb a comma
		case next.Is(token.LPAREN):
			for tok := p.next(); !tok.Is(token.RPAREN); {
				switch {
				case tok.Is(token.STRING):
					outputs = append(outputs, p.parseString(tok))
				case tok.Is(token.IDENT):
					outputs = append(outputs, p.parseIdent(tok))
				case tok.Is(token.COMMA):
					// Absorb a comma
				case tok.Is(token.ERROR):
					return nil, errors.New(next.Value)
				default:
					return nil, illegalToken{
						expected:    []token.Type{token.STRING, token.IDENT, token.COMMA},
						encountered: tok,
						line:        p.getLine(next),
					}
				}
				tok = p.next()
			}
		case next.Is(token.ERROR):
			return nil, errors.New(next.Value)
		default:
			return nil, illegalToken{
				expected:    []token.Type{token.STRING, token.IDENT, token.LPAREN},
				encountered: next,
				line:        p.getLine(next),
			}
		}
	} else {
		// No outputs declared, undo our call to p.next() in the if branch
		p.backup()
	}

	return outputs, nil
}

// parseTaskCommands parses any number of command tokens in a task body and returns them.
func (p *Parser) parseTaskCommands() ([]ast.Command, error) {
	commands := []ast.Command{}
	for {
		next := p.next()
		if next.Is(token.ERROR) {
			return commands, errors.New(next.Value)
		}
		if next.Is(token.RBRACE) {
			break
		}
		if next.Is(token.COMMAND) {
			commands = append(commands, p.parseCommand(next))
		}
	}

	return commands, nil
}

// parseCommand parses task commands into ast command nodes.
func (p *Parser) parseCommand(command token.Token) ast.Command {
	return ast.Command{
		Command:  command.Value,
		NodeType: ast.NodeCommand,
	}
}

// getLine returns the line of the input on which the given token appears
// primarily used to provide context for parser errors given back to the user.
func (p *Parser) getLine(token token.Token) string {
	rawLines := strings.Split(p.input, "\n")

	var lines []string
	for _, line := range rawLines {
		lines = append(lines, strings.TrimSpace(line))
	}
	if token.Line == 0 {
		return lines[token.Line]
	}
	return lines[token.Line-1]
}
