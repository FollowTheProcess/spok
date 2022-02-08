// Package parser implements spok's parser.
package parser

import (
	"fmt"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/FollowTheProcess/spok/lexer"
	"github.com/FollowTheProcess/spok/token"
)

// Parser is spok's AST parser.
type Parser struct {
	lexer     lexer.Tokeniser // The lexer.
	buffer    [3]token.Token  // 3 token buffer, allows us to peek and backup in the token stream.
	peekCount int             // How far we've "peeked" into our buffer.
}

// New creates and returns a new Parser for an input string.
func New(text string) *Parser {
	return &Parser{
		lexer: lexer.New(text),
	}
}

// Parse is the top level parse method. It parses the entire
// input text to EOF or Error and returns the full AST.
func (p *Parser) Parse() (*ast.Tree, error) {
	tree := &ast.Tree{}
	return tree, nil
}

// next returns, and consumes, the next token from the lexer.
func (p *Parser) next() token.Token {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.buffer[0] = p.lexer.NextToken()
	}
	return p.buffer[p.peekCount]
}

// peek returns, but does not consume, the next token from the lexer.
func (p *Parser) peek() token.Token { // nolint: unused
	if p.peekCount > 0 {
		return p.buffer[p.peekCount-1]
	}
	p.peekCount = 1
	p.buffer[0] = p.lexer.NextToken()
	return p.buffer[0]
}

// expect consumes the given token if present, and returns
// an error if not.
func (p *Parser) expect(expected token.Type) error {
	next := p.next()
	if !next.Is(expected) {
		return fmt.Errorf("Unexpected token: %s", next)
	}
	return nil
}

// parseComment parses a comment token into a comment ast node,
// the # has already been consumed.
func (p *Parser) parseComment(comment token.Token) *ast.Comment {
	return &ast.Comment{
		Text:     comment.Value,
		NodeType: ast.NodeComment,
	}
}

// parseIdent parses an ident token into an ident ast node.
func (p *Parser) parseIdent(ident token.Token) *ast.Ident {
	return &ast.Ident{
		Name:     ident.Value,
		NodeType: ast.NodeIdent,
	}
}

// parseString parses a string token into a string ast node.
func (p *Parser) parseString(s token.Token) *ast.String {
	return &ast.String{
		Text:     s.Value,
		NodeType: ast.NodeString,
	}
}

// parseFunction parses an ident token into a function ast node.
func (p *Parser) parseFunction(ident token.Token) *ast.Function {
	args := []ast.Node{}
	p.next() // '('

	for next := p.next(); !next.Is(token.RPAREN); {
		switch {
		case next.Is(token.STRING):
			args = append(args, p.parseString(next))
		case next.Is(token.IDENT):
			args = append(args, p.parseIdent(next))
		}
		next = p.next()
	}

	fn := &ast.Function{
		Name:      p.parseIdent(ident),
		Arguments: args,
		NodeType:  ast.NodeFunction,
	}
	return fn
}

// parseAssign parses a global variable assignment into an assign ast node.
// the ':=' is known to exist and has already been consumed, the encountered ident token is passed in.
func (p *Parser) parseAssign(ident token.Token) *ast.Assign {
	name := p.parseIdent(ident)

	var rhs ast.Node

	switch next := p.next(); {
	case next.Is(token.STRING):
		rhs = p.parseString(next)
	case next.Is(token.IDENT):
		// Only other thing is a built in function
		rhs = p.parseFunction(next)
	}

	return &ast.Assign{
		Name:     name,
		Value:    rhs,
		NodeType: ast.NodeAssign,
	}
}

// parseTask parses and returns a task ast node, the task keyword has already
// been encountered and consumed, the docstring comment is passed in if present
// and will be nil if no docstring.
func (p *Parser) parseTask(doc *ast.Comment) *ast.Task {
	name := p.parseIdent(p.next())

	docstring := &ast.Comment{NodeType: ast.NodeComment}
	if doc != nil {
		docstring = doc
	}

	p.next() // '('

	dependencies := []ast.Node{}
	for next := p.next(); !next.Is(token.RPAREN); {
		switch {
		case next.Is(token.STRING):
			dependencies = append(dependencies, p.parseString(next))
		case next.Is(token.IDENT):
			dependencies = append(dependencies, p.parseIdent(next))
		}
		next = p.next()
	}

	outputs := []ast.Node{}
	if p.peek().Is(token.OUTPUT) {
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
				}
				tok = p.next()
			}
		}
	}

	p.next() // '{'

	commands := []*ast.Command{}
	for {
		next := p.next()
		if next.Is(token.RBRACE) {
			break
		}
		if next.Is(token.COMMAND) {
			commands = append(commands, p.parseCommand(next))
		}
	}

	return &ast.Task{
		Name:         name,
		Docstring:    docstring,
		Dependencies: dependencies,
		Outputs:      outputs,
		Commands:     commands,
		NodeType:     ast.NodeTask,
	}
}

// parseCommand parses task commands into ast command nodes.
func (p *Parser) parseCommand(command token.Token) *ast.Command {
	return &ast.Command{
		Command:  command.Value,
		NodeType: ast.NodeCommand,
	}
}
