package lexer

import "github.com/fiurgeist/golox/internal/token"

type Lexer struct {
	source string
}

func NewLexer(source string) Lexer {
	return Lexer{source: source}
}

func (l *Lexer) ScanTokens() ([]token.Token, error) {
	return []token.Token{}, nil
}
