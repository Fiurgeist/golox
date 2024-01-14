package parser

import (
	"errors"
	"fmt"

	"github.com/fiurgeist/golox/internal/ast/expr"
	"github.com/fiurgeist/golox/internal/ast/stmt"
	"github.com/fiurgeist/golox/internal/reporter"
	"github.com/fiurgeist/golox/internal/token"
)

/*
program        → declaration* EOF ;

declaration    → varDecl
               | classDecl
               | funDecl
               | statement ;
classDecl      → "class" IDENTIFIER "{" function* "}" ;
varDecl        → "var" IDENTIFIER ( "=" expression )? ";" ;
funDecl        → "fun" function ;
function       → IDENTIFIER "(" parameters? ")" block ;
parameters     → IDENTIFIER ( "," IDENTIFIER )* ;

statement      → exprStmt
               | printStmt
               | ifStmt
               | whileStmt
               | forStmt
               | breakStmt
               | returnStmt
               | block ;

exprStmt       → expression ";" ;
printStmt      → "print" expression ";" ;
ifStmt         → "if" "(" expression ")" statement
               ( "else" statement )? ;
whileStmt      → "while" "(" expression ")" statement ;
forStmt        → "for" "(" ( varDecl | exprStmt | ";" )
                 expression? ";"
                 expression? ")" statement ;
breakStmt      → "break" ";" ;
returnStmt     → "return" expression? ";" ;
block          → "{" declaration* "}" ;

expression     → assignment ;
assignment     → ( call "." )? IDENTIFIER "=" assignment
               | logic_or ;
logic_or       → logic_and ( "or" logic_and )* ;
logic_and      → equality ( "and" equality )* ;
equality       → comparison ( ( "!=" | "==" ) comparison )* ;
comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
term           → factor ( ( "-" | "+" ) factor )* ;
factor         → unary ( ( "/" | "*" ) unary )* ;
unary          → ( "!" | "-" ) unary | call ;
call           → primary ( "(" arguments? ")" | "." IDENTIFIER )* ;
arguments      → expression ( "," expression )* ;
primary        → NUMBER | STRING | "true" | "false" | "nil"
               | "(" expression ")" | IDENTIFIER ;
*/

var ErrParser = errors.New("ParseError")

type Parser struct {
	current  int
	tokens   []token.Token
	inLoop   bool
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

	if p.match(token.FUN) {
		return p.function("function"), err
	}

	if p.match(token.CLASS) {
		return p.class(), err
	}

	return p.statement(), err
}

func (p *Parser) varDeclaration() stmt.Stmt {
	name, _ := p.consume(token.IDENTIFIER, "Expect variable name")

	var initializer expr.Expr
	if p.match(token.EQUAL) {
		initializer = p.expression()
	}

	p.consume(token.SEMICOLON, "Expect ';' after variable declaration")

	return stmt.NewVar(name, initializer)
}

func (p *Parser) function(kind string) *stmt.Function {
	name, _ := p.consume(token.IDENTIFIER, fmt.Sprintf("Expect %s name", kind))

	p.consume(token.LEFT_PAREN, fmt.Sprintf("Expect '(' after %s name", kind))

	var params []token.Token
	if !p.check(token.RIGHT_PAREN) {
		param, _ := p.consume(token.IDENTIFIER, fmt.Sprintf("Expect %s parameter", kind))
		params = []token.Token{param}

		for p.match(token.COMMA) {
			if len(params) >= 255 {
				p.reporter.ParseError(p.peek(), "Can't have more than 255 parameters")
			}

			param, _ := p.consume(token.IDENTIFIER, fmt.Sprintf("Expect %s parameter", kind))
			params = append(params, param)
		}
	}

	p.consume(token.RIGHT_PAREN, fmt.Sprintf("Expect ')' after %s parameters", kind))

	p.consume(token.LEFT_BRACE, fmt.Sprintf("Expect '{' before %s body", kind))
	body := p.block()

	return stmt.NewFunction(name, params, body)
}

func (p *Parser) class() stmt.Stmt {
	name, _ := p.consume(token.IDENTIFIER, "Expect class name")

	p.consume(token.LEFT_BRACE, "Expect '{' before class body")

	var methods []*stmt.Function
	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		methods = append(methods, p.function("method"))
	}

	p.consume(token.RIGHT_BRACE, "Expect '}' after class body")

	return stmt.NewClass(name, methods)
}

func (p *Parser) statement() stmt.Stmt {
	if p.match(token.PRINT) {
		return p.printStatement()
	}

	if p.match(token.IF) {
		return p.ifStatement()
	}

	if p.match(token.WHILE) {
		return p.whileStatement()
	}

	if p.match(token.FOR) {
		return p.forStatement()
	}

	if p.match(token.BREAK) {
		return p.breakStatement()
	}

	if p.match(token.RETURN) {
		return p.returnStatement()
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

func (p *Parser) ifStatement() stmt.Stmt {
	p.consume(token.LEFT_PAREN, "Expect '(' after if")
	condition := p.expression()
	p.consume(token.RIGHT_PAREN, "Expect ')' after if condition")

	thenBranch := p.statement()

	var elseBranch stmt.Stmt
	if p.match(token.ELSE) {
		elseBranch = p.statement()
	}

	return stmt.NewIf(condition, thenBranch, elseBranch)
}

func (p *Parser) whileStatement() stmt.Stmt {
	previousInLoop := p.inLoop
	p.inLoop = true
	defer func() {
		p.inLoop = previousInLoop
	}()

	p.consume(token.LEFT_PAREN, "Expect '(' after while")
	condition := p.expression()
	p.consume(token.RIGHT_PAREN, "Expect ')' after while condition")

	body := p.statement()
	return stmt.NewWhile(condition, body)
}

func (p *Parser) forStatement() stmt.Stmt {
	previousInLoop := p.inLoop
	p.inLoop = true
	defer func() {
		p.inLoop = previousInLoop
	}()

	p.consume(token.LEFT_PAREN, "Expect '(' after for")

	var initializer stmt.Stmt
	if !p.match(token.SEMICOLON) {
		if p.match(token.VAR) {
			initializer = p.varDeclaration()
		} else {
			initializer = p.expressionStatement()
		}
	}

	var condition expr.Expr
	if !p.check(token.SEMICOLON) {
		condition = p.expression()
	}
	p.consume(token.SEMICOLON, "Expect ';' after for condition")

	var increment expr.Expr
	if !p.check(token.RIGHT_PAREN) {
		increment = p.expression()
	}

	p.consume(token.RIGHT_PAREN, "Expect ')' after for condition")

	body := p.statement()
	if increment != nil {
		body = stmt.NewBlock([]stmt.Stmt{body, stmt.NewExpression(increment)})
	}

	if condition == nil {
		condition = expr.NewLiteral(true)
	}

	var desugaredFor stmt.Stmt = stmt.NewWhile(condition, body)
	if initializer != nil {
		desugaredFor = stmt.NewBlock([]stmt.Stmt{initializer, desugaredFor})
	}

	return desugaredFor
}

func (p *Parser) breakStatement() stmt.Stmt {
	if !p.inLoop {
		p.reporter.ParseError(p.previous(), "Outside of a loop")
	}

	p.consume(token.SEMICOLON, "Expect ';' after break")
	return stmt.NewBreak()
}

func (p *Parser) returnStatement() stmt.Stmt {
	keyword := p.previous()

	var value expr.Expr
	if !p.check(token.SEMICOLON) {
		value = p.expression()
	}

	p.consume(token.SEMICOLON, "Expect ';' after return value")
	return stmt.NewReturn(keyword, value)
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
	expression := p.or()

	if p.match(token.EQUAL) {
		equals := p.previous()
		value := p.assignment()

		switch e := expression.(type) {
		case *expr.Variable:
			return expr.NewAssign(e.Name, value)
		case *expr.Get:
			return expr.NewSet(e.Object, e.Name, value)
		}

		p.reporter.ParseError(equals, "Invalid assignment target") // report error, but continue
	}

	return expression
}

func (p *Parser) or() expr.Expr {
	expression := p.and()

	for p.match(token.OR) {
		operator := p.previous()
		right := p.and()

		expression = expr.NewLogical(expression, operator, right)
	}

	return expression
}

func (p *Parser) and() expr.Expr {
	expression := p.equality()

	if p.match(token.AND) {
		operator := p.previous()
		right := p.equality()

		expression = expr.NewLogical(expression, operator, right)
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

	return p.call()
}

func (p *Parser) call() expr.Expr {
	expression := p.primary()
	for {
		if p.match(token.LEFT_PAREN) {
			var arguments []expr.Expr
			if !p.check(token.RIGHT_PAREN) {
				arguments = p.arguments()
			}

			paren, _ := p.consume(token.RIGHT_PAREN, "Expected ')' after arguments")
			expression = expr.NewCall(expression, arguments, paren)
		} else if p.match(token.DOT) {
			name, _ := p.consume(token.IDENTIFIER, "Expected property name after '.'")
			expression = expr.NewGet(expression, name)
		} else {
			break
		}
	}

	return expression
}

func (p *Parser) arguments() []expr.Expr {
	expressions := []expr.Expr{p.expression()}
	for p.match(token.COMMA) {
		if len(expressions) >= 255 {
			p.reporter.ParseError(p.peek(), "Can't have more than 255 arguments")
		}
		expressions = append(expressions, p.expression())
	}

	return expressions
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

	if p.match(token.THIS) {
		return expr.NewThis(p.previous())
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
