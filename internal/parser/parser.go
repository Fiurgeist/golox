package parser

import (
	"errors"

	"github.com/fiurgeist/golox/internal/expr"
	"github.com/fiurgeist/golox/internal/reporter"
	"github.com/fiurgeist/golox/internal/stmt"
	"github.com/fiurgeist/golox/internal/token"
)

/*
program        → declaration* EOF ;

declaration    → varDecl
               | statement ;
varDecl        → "var" IDENTIFIER ( "=" expression )? ";" ;

statement      → exprStmt
               | printStmt
               | block ;

block          → "{" declaration* "}" ;

exprStmt       → expression ";" ;
printStmt      → "print" expression ";" ;

expression     → assignment ;
assignment     → IDENTIFIER "=" assignment
               | equality ;
equality       → comparison ( ( "!=" | "==" ) comparison )* ;
comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
term           → factor ( ( "-" | "+" ) factor )* ;
factor         → unary ( ( "/" | "*" ) unary )* ;
unary          → ( "!" | "-" ) unary
               | primary ;
primary        → NUMBER | STRING | "true" | "false" | "nil"
               | "(" expression ")" | IDENTIFIER ;
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

func (p *Parser) Parse() (statements []stmt.Stmt, anyErr error) {
	for !p.isAtEnd() {
		declaration, err := p.declaration()
		if err != nil {
			anyErr = err
			continue
		}

		statements = append(statements, declaration)
	}

	return statements, anyErr
}

func (p *Parser) declaration() (statement stmt.Stmt, err error) {
	defer func() { // TODO: remove, panic until stmts are implemented
		if err := recover(); err != nil {
			p.synchronize()
			err = ErrParser
		}
	}()

	if p.match(token.VAR) {
		return p.varDeclaration(), err
	}

	return p.statement(), err
}

func (p *Parser) varDeclaration() stmt.Stmt {
	name, _ := p.consume(token.IDENTIFIER, "Expect variable name")

	var Initializer expr.Expr
	if p.match(token.EQUAL) {
		Initializer = p.expression()
	}

	p.consume(token.SEMICOLON, "Expect ';' after variable declaration")

	return stmt.NewVar(name, Initializer)
}

func (p *Parser) statement() stmt.Stmt {
	if p.match(token.PRINT) {
		return p.printStatement()
	}

	if p.match(token.LEFT_BRACE) {
		return stmt.NewBlock(p.block())
	}

	return p.expressionStatement()
}

func (p *Parser) printStatement() stmt.Stmt {
	value := p.expression()
	p.consume(token.SEMICOLON, "Expect ';' after value")
	return stmt.NewPrint(value)
}

func (p *Parser) block() []stmt.Stmt {
	var statements []stmt.Stmt

	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		decl, _ := p.declaration()
		statements = append(statements, decl)
	}

	p.consume(token.RIGHT_BRACE, "Expect '}' after block")
	return statements
}

func (p *Parser) expressionStatement() stmt.Stmt {
	expression := p.expression()
	p.consume(token.SEMICOLON, "Expect ';' after expression")
	return stmt.NewExpression(expression)
}

func (p *Parser) expression() expr.Expr {
	return p.assignment()
}

func (p *Parser) assignment() expr.Expr {
	expression := p.equality()

	if p.match(token.EQUAL) {
		equals := p.previous()
		value := p.assignment()

		if varExpr, ok := expression.(expr.Variable); ok {
			return expr.NewAssign(varExpr.Name, value)
		}

		p.reporter.ParseError(equals, "Invalid assignment target") // report error, but continue
	}

	return expression
}

func (p *Parser) equality() expr.Expr {
	expression := p.comparison()

	for p.match(token.BANG_EQUAL, token.EQUAL_EQUAL) {
		operator := p.previous()
		right := p.comparison()
		expression = expr.NewBinary(expression, operator, right)
	}

	return expression
}

func (p *Parser) comparison() expr.Expr {
	expression := p.term()

	for p.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.previous()
		right := p.term()
		expression = expr.NewBinary(expression, operator, right)
	}

	return expression
}

func (p *Parser) term() expr.Expr {
	expression := p.factor()

	for p.match(token.PLUS, token.MINUS) {
		operator := p.previous()
		right := p.factor()
		expression = expr.NewBinary(expression, operator, right)
	}

	return expression
}

func (p *Parser) factor() expr.Expr {
	expression := p.unary()

	for p.match(token.SLASH, token.STAR) {
		operator := p.previous()
		right := p.unary()
		expression = expr.NewBinary(expression, operator, right)
	}

	return expression
}

func (p *Parser) unary() expr.Expr {
	if p.match(token.BANG, token.MINUS) {
		operator := p.previous()
		right := p.unary()
		return expr.NewUnary(operator, right)
	}

	return p.primary()
}

func (p *Parser) primary() expr.Expr {
	if p.match(token.FALSE) {
		return expr.NewLiteral(false)
	}

	if p.match(token.TRUE) {
		return expr.NewLiteral(true)
	}

	if p.match(token.NIL) {
		return expr.NewLiteral(nil)
	}

	if p.match(token.NUMBER, token.STRING) {
		return expr.NewLiteral(p.previous().Literal)
	}

	if p.match(token.IDENTIFIER) {
		return expr.NewVariable(p.previous())
	}

	if p.match(token.LEFT_PAREN) {
		expression := p.expression()

		_, err := p.consume(token.RIGHT_PAREN, "Expected ')' after expression")
		if err != nil {
			panic("Parse Error") // TODO remove
		}

		return expr.NewGrouping(expression)
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

	//return p.advance(), ErrParser
	p.advance()
	panic("Parse Error") // TODO remove
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
