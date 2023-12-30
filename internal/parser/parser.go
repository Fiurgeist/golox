package parser

import (
	"errors"
	"fmt"

	"github.com/fiurgeist/golox/internal/ast"
	"github.com/fiurgeist/golox/internal/reporter"
	"github.com/fiurgeist/golox/internal/token"
)

/*
expression     → equality ;
equality       → comparison ( ( "!=" | "==" ) comparison )* ;
comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
term           → factor ( ( "-" | "+" ) factor )* ;
factor         → unary ( ( "/" | "*" ) unary )* ;
unary          → ( "!" | "-" ) unary
               | primary ;
primary        → NUMBER | STRING | "true" | "false" | "nil"
               | "(" expression ")" ;
*/

var ErrParser = errors.New("ParseError")

type Parser struct {
	current  int
	tokens   []token.Token
	reporter reporter.ErrorReporter
}

func NewParser(tokens []token.Token, reporter reporter.ErrorReporter) Parser {
	return Parser{tokens: tokens, reporter: reporter}
}

func (p *Parser) Parse() (expr ast.Expr, err error) {
	defer func() { // TODO: remove, panic until stmts are implemented
		if p := recover(); p != nil {
			fmt.Printf("Panic: %v\n", p)
			err = ErrParser
		}
	}()
	return p.expression(), err
}

func (p *Parser) expression() ast.Expr {
	return p.equality()
}

func (p *Parser) equality() ast.Expr {
	expr := p.comparison()

	for p.match(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := p.previous()
		right := p.comparison()
		expr = ast.NewBinary(expr, operator, right)
	}

	return expr
}

func (p *Parser) comparison() ast.Expr {
	expr := p.term()

	for p.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.previous()
		right := p.term()
		expr = ast.NewBinary(expr, operator, right)
	}

	return expr
}

func (p *Parser) term() ast.Expr {
	expr := p.factor()

	for p.match(token.PLUS, token.MINUS) {
		operator := p.previous()
		right := p.factor()
		expr = ast.NewBinary(expr, operator, right)
	}

	return expr
}

func (p *Parser) factor() ast.Expr {
	expr := p.unary()

	for p.match(token.SLASH, token.STAR) {
		operator := p.previous()
		right := p.unary()
		expr = ast.NewBinary(expr, operator, right)
	}

	return expr
}

func (p *Parser) unary() ast.Expr {
	if p.match(token.BANG, token.MINUS) {
		operator := p.previous()
		right := p.unary()
		return ast.NewUnary(operator, right)
	}

	return p.primary()
}

func (p *Parser) primary() ast.Expr {
	if p.match(token.FALSE) {
		return ast.NewLiteral(false)
	}

	if p.match(token.TRUE) {
		return ast.NewLiteral(true)
	}

	if p.match(token.NIL) {
		return ast.NewLiteral(nil)
	}

	if p.match(token.NUMBER, token.STRING) {
		return ast.NewLiteral(p.previous().Literal)
	}

	if p.match(token.LEFT_PAREN) {
		expr := p.expression()

		_, err := p.consume(token.RIGHT_PAREN, "Expected ')' after expression")
		if err != nil {
			panic("Parse Error") // TODO remove
		}

		return ast.NewGrouping(expr)
	}

	p.reporter.ParseError(p.peek(), "Expect expression")
	panic("Parse Error") // TODO remove
}

func (p *Parser) match(types ...token.TokenType) bool {
	for _, tokenType := range types {
		if p.check(tokenType) {
			p.advance()

			return true
		}
	}

	return false
}

func (p *Parser) check(tokenType token.TokenType) bool {
	/* shouldn't be needed
	if p.isAtEnd() {
		return false
	}
	*/
	return p.peek().Type == tokenType
}

func (p *Parser) advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}

	return p.previous()
}

func (p *Parser) previous() token.Token {
	return p.tokens[p.current-1]
}

func (p *Parser) peek() token.Token {
	return p.tokens[p.current]
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == token.EOF
}

func (p *Parser) consume(tokenType token.TokenType, message string) (token.Token, error) {
	if p.check(tokenType) {
		return p.advance(), nil
	}
	p.reporter.ParseError(p.peek(), message)
	return p.advance(), ErrParser
}

func (p *Parser) synchronize() {
	p.advance()
	for !p.isAtEnd() {
		if p.previous().Type == token.SEMICOLON {
			return
		}

		switch p.peek().Type {
		case token.CLASS:
		case token.FUN:
		case token.VAR:
		case token.FOR:
		case token.IF:
		case token.WHILE:
		case token.PRINT:
		case token.RETURN:
			return
		}

		p.advance()
	}
}
