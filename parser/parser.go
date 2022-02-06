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
	for p.peek().Type != token.EOF {
		switch tok := p.next(); {
		case tok.Is(token.HASH):
			tree.Append(p.parseComment())
		case tok.Is(token.IDENT):
			switch {
			case p.next().Is(token.DECLARE):
				tree.Append(p.parseAssign(tok))
			default:
				tree.Append(p.parseIdent(tok))
			}
		case tok.Is(token.ERROR):
			return nil, fmt.Errorf("Parser error: %s", tok.Value)
		}
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
	return p.buffer[p.peekCount]
}

// peek returns, but does not consume, the next token from the lexer.
func (p *Parser) peek() token.Token {
	if p.peekCount > 0 {
		return p.buffer[p.peekCount-1]
	}
	p.peekCount = 1
	p.buffer[0] = p.lexer.NextToken()
	return p.buffer[0]
}

// backup backs up by one token.
func (p *Parser) backup() { // nolint: unused
	// TODO: Unused for now but we know we'll need it
	p.peekCount++
}

// parseComment parses a comment token into a comment ast node,
// the # has already been consumed.
func (p *Parser) parseComment() *ast.CommentNode {
	token := p.next()
	return &ast.CommentNode{
		Text:     token.Value,
		NodeType: ast.NodeComment,
	}
}

// parseIdent parses an ident token into an ident ast node.
func (p *Parser) parseIdent(token token.Token) *ast.IdentNode {
	return &ast.IdentNode{
		Name:     token.Value,
		NodeType: ast.NodeIdent,
	}
}

// parseAssign parses a global variable assignment into an assign ast node.
// the ':=' is known to exist and has already been consumed, the encountered ident token is passed in.
func (p *Parser) parseAssign(ident token.Token) *ast.AssignNode {
	name := p.parseIdent(ident)

	var rhs ast.Node

	switch next := p.next(); {
	case next.Is(token.STRING):
		rhs = &ast.StringNode{
			Text:     next.Value,
			NodeType: ast.NodeString,
		}
	case next.Is(token.IDENT):
		// Only other thing is a built in function
		rhs = &ast.IdentNode{
			Name:     next.Value,
			NodeType: ast.NodeIdent,
		}
	}

	return &ast.AssignNode{
		Name:     name,
		Value:    rhs,
		NodeType: ast.NodeAssign,
	}
}
