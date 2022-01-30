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
	lexer     *lexer.Lexer   // The lexer.
	text      string         // The raw input text.
	buffer    [3]token.Token // 3 token buffer, allows us to peek and backup in the token stream.
	peekCount int            // How far we've "peeked" into our buffer.
}

// New creates and returns a new Parser for an input string.
func New(text string) *Parser {
	return &Parser{
		lexer: lexer.New(text),
		text:  text,
	}
}

// Parse is the top level parse method. It parses the entire
// input text to EOF or Error and returns the full AST.
func (p *Parser) Parse() (*ast.Tree, error) {
	tree := &ast.Tree{}
	for p.peek().Type != token.EOF {
		switch tok := p.next(); {
		case tok.Is(token.HASH):
			tree.Nodes = append(tree.Nodes, p.parseComment())
		case tok.Is(token.ERROR):
			return nil, fmt.Errorf("Error token: %s", tok)
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
func (p *Parser) parseComment() ast.CommentNode {
	token := p.next()
	return ast.CommentNode{
		Text:     token.Value,
		NodeType: ast.NodeComment,
	}
}
