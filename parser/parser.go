// Package parser implements spok's parser.
//
// Spok is a very simple mostly declarative language and as such, the parser is incredibly simple. It is
// a simple top-down parser with a very small initial state space and requires only 1 token of lookahead.
//
// It switches on the token it encounters to process the appropriate ast.Node (parseXXX methods), appending
// each to the list of nodes as it goes.
//
// The parser also keeps a stack of errors and adds to this if it encounters an error from the lexer
// or if it encounters an error during it's own operation. The errors are checked before the AST is returned
// and if any are present it will return the tree it's managed to parse so far and the error encountered.
package parser

import (
	"fmt"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/FollowTheProcess/spok/lexer"
	"github.com/FollowTheProcess/spok/token"
)

// Parser is spok's AST parser.
type Parser struct {
	lexer     lexer.Tokeniser // The lexer
	errors    []error         // Stack of errors collected during the parse
	buffer    [3]token.Token  // 3 token buffer, allows us to peek and backup in the token stream
	peekCount int             // How far we've peeked into our buffer
}

// New creates and returns a new Parser for an input string.
func New(text string) *Parser {
	return &Parser{
		lexer: lexer.New(text),
	}
}

// Parse is the top level parse method. It parses the entire
// input text to EOF or Error and returns the full AST.
func (p *Parser) Parse() (ast.Tree, error) {
	tree := ast.Tree{}

	for next := p.next(); !next.Is(token.EOF); {
		switch {
		case next.Is(token.HASH):
			comment := p.parseComment()
			switch {
			case p.next().Is(token.TASK):
				// The comment was a tasks' docstring
				tree.Append(p.parseTask(comment))
			default:
				// Just a normal comment
				p.backup()
				tree.Append(comment)
			}
		case next.Is(token.IDENT):
			tree.Append(p.parseAssign(next))
		case next.Is(token.TASK):
			// Pass an empty comment in if it doesn't have one
			tree.Append(p.parseTask(ast.Comment{NodeType: ast.NodeComment}))

		default:
			// Illegal top level token that slipped through the lexer somehow
			// unlikely but let's catch it anyway
			p.errorf("Illegal token (Line %d, Position %d): %s", next.Line, next.Pos, next.String())

		}
		next = p.next()
	}

	// Check for errors and pop the first one
	if p.hasErrors() {
		return tree, p.popError()
	}

	return tree, nil
}

// next returns, and consumes, the next token from the lexer
// if it encounters an error token emitted by the lexer, it adds it to
// the stack of parser errors.
func (p *Parser) next() token.Token {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.buffer[0] = p.lexer.NextToken()
	}
	tok := p.buffer[p.peekCount]
	if tok.Is(token.ERROR) {
		p.errorf(tok.Value)
	}
	return tok
}

// backups backs up in the input stream by one token.
func (p *Parser) backup() {
	p.peekCount++
}

// expect checks if the next token is of the expected type, consuming it in the process
// if not it will add an error to the parser error stack.
func (p *Parser) expect(token token.Type) {
	tok := p.next()
	if !tok.Is(token) {
		p.errorf("Unexpected token (Line %d, Position %d): got %s, expected %s", tok.Line, tok.Pos, tok.String(), token.String())
	}
}

// hasErrors returns whether or not the parser has encountered errors on it's travels.
func (p *Parser) hasErrors() bool {
	return len(p.errors) != 0
}

// errorf adds a formatted error message to the parser's error stack.
func (p *Parser) errorf(format string, args ...interface{}) {
	p.errors = append(p.errors, fmt.Errorf(format, args...))
}

// popError returns the first error in the stack of errors.
func (p *Parser) popError() error {
	return p.errors[0]
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
func (p *Parser) parseFunction(ident token.Token) ast.Function {
	args := []ast.Node{}
	p.expect(token.LPAREN) // '('

	for next := p.next(); !next.Is(token.RPAREN); {
		switch {
		case next.Is(token.STRING):
			args = append(args, p.parseString(next))
		case next.Is(token.IDENT):
			args = append(args, p.parseIdent(next))
		default:
			p.errorf("Illegal token (Line %d, Position %d): %s", next.Line, next.Pos, next.String())
		}
		next = p.next()
	}

	fn := ast.Function{
		Name:      p.parseIdent(ident),
		Arguments: args,
		NodeType:  ast.NodeFunction,
	}
	return fn
}

// parseAssign parses a global variable assignment into an assign ast node.
// the ':=' is known to exist but has yet to be consumed, the encountered ident token is passed in.
func (p *Parser) parseAssign(ident token.Token) ast.Assign {
	name := p.parseIdent(ident)
	p.expect(token.DECLARE) // ':='

	var rhs ast.Node

	switch next := p.next(); {
	case next.Is(token.STRING):
		rhs = p.parseString(next)
	case next.Is(token.IDENT):
		// Only other thing is a built in function or assigning to another ident
		if p.next().Is(token.LPAREN) {
			p.backup()
			rhs = p.parseFunction(next)
		} else {
			p.backup()
			rhs = p.parseIdent(next)
		}
	default:
		p.errorf("Illegal token (Line %d, Position %d): %s", next.Line, next.Pos, next.String())
	}

	return ast.Assign{
		Name:     name,
		Value:    rhs,
		NodeType: ast.NodeAssign,
	}
}

// parseTask parses and returns a task ast node, the task keyword has already
// been encountered and consumed, the docstring comment is passed in if present
// and will be empty if there is no comment.
func (p *Parser) parseTask(doc ast.Comment) ast.Task {
	name := p.parseIdent(p.next())

	p.expect(token.LPAREN) // '('

	dependencies := []ast.Node{}
	for next := p.next(); !next.Is(token.RPAREN); {
		switch {
		case next.Is(token.STRING):
			dependencies = append(dependencies, p.parseString(next))
		case next.Is(token.IDENT):
			dependencies = append(dependencies, p.parseIdent(next))
		default:
			p.errorf("Illegal token (Line %d, Position %d): %s", next.Line, next.Pos, next.String())
		}
		next = p.next()
	}

	outputs := []ast.Node{}
	if p.next().Is(token.OUTPUT) {
		switch next := p.next(); {
		case next.Is(token.STRING):
			outputs = append(outputs, p.parseString(next))
		case next.Is(token.IDENT):
			outputs = append(outputs, p.parseIdent(next))
		case next.Is(token.LPAREN):
			for tok := p.next(); !tok.Is(token.RPAREN); {
				switch {
				case tok.Is(token.STRING):
					outputs = append(outputs, p.parseString(tok))
				case tok.Is(token.IDENT):
					outputs = append(outputs, p.parseIdent(tok))
				default:
					p.errorf("Illegal token (Line %d, Position %d): %s", tok.Line, tok.Pos, tok.String())
				}
				tok = p.next()
			}
		default:
			p.errorf("Illegal token (Line %d, Position %d): %s", next.Line, next.Pos, next.String())
		}
	}

	commands := []ast.Command{}
	for {
		next := p.next()
		if next.Is(token.RBRACE) {
			break
		}
		if next.Is(token.COMMAND) {
			commands = append(commands, p.parseCommand(next))
		}
	}

	return ast.Task{
		Name:         name,
		Docstring:    doc,
		Dependencies: dependencies,
		Outputs:      outputs,
		Commands:     commands,
		NodeType:     ast.NodeTask,
	}
}

// parseCommand parses task commands into ast command nodes.
func (p *Parser) parseCommand(command token.Token) ast.Command {
	return ast.Command{
		Command:  command.Value,
		NodeType: ast.NodeCommand,
	}
}
