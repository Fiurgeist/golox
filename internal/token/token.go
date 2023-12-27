package token

import "fmt"

type TokenType int

const (
	// Single-character tokens
	LEFT_PAREN TokenType = iota
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	COMMA
	DOT
	MINUS
	PLUS
	SEMICOLON
	SLASH
	STAR

	// One or two character tokens
	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL

	// Literals
	IDENTIFIER
	STRING
	NUMBER

	// Keywords
	AND
	CLASS
	ELSE
	FALSE
	FUN
	FOR
	IF
	NIL
	OR
	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	VAR
	WHILE

	EOF
)

type Token struct {
	tokenType TokenType
	lexeme    string
	literal   []byte // not sure what type yet
	line      int32
}

func NewToken(tokenType TokenType,
	lexeme string,
	literal []byte,
	line int32) Token {
	return Token{tokenType: tokenType, lexeme: lexeme, literal: literal, line: line}
}

func (t *Token) String() string {
	return fmt.Sprintf("%d %s %s", t.tokenType, t.lexeme, t.literal)
}
