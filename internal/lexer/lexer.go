package lexer

import (
	"errors"
	"fmt"

	"github.com/fiurgeist/golox/internal/reporter"
	"github.com/fiurgeist/golox/internal/token"
)

var ErrLexer = errors.New("LexerError")

type Lexer struct {
	source   []byte
	start    int
	current  int
	line     int
	hasError bool
	tokens   []token.Token
	reporter reporter.ErrorReporter
}

func NewLexer(source []byte, reporter reporter.ErrorReporter) Lexer {
	return Lexer{source: source, start: 0, current: 0, line: 1, tokens: []token.Token{}, reporter: reporter}
}

func (l *Lexer) ScanTokens() ([]token.Token, error) {
	for !l.isAtEnd() {
		l.start = l.current
		l.scanToken()
	}

	if l.hasError {
		return l.tokens, ErrLexer
	}

	return l.tokens, nil
}

func (l *Lexer) isAtEnd() bool {
	return l.current+1 >= len(l.source)
}

func (l *Lexer) scanToken() {
	c := l.advance()
	switch c {
	case '(':
		l.addToken(token.LEFT_PAREN)
	case ')':
		l.addToken(token.RIGHT_PAREN)
	case '{':
		l.addToken(token.LEFT_BRACE)
	case '}':
		l.addToken(token.RIGHT_BRACE)
	case ',':
		l.addToken(token.COMMA)
	case '.':
		l.addToken(token.DOT)
	case '-':
		l.addToken(token.MINUS)
	case '+':
		l.addToken(token.PLUS)
	case ';':
		l.addToken(token.SEMICOLON)
	case '*':
		l.addToken(token.STAR)
	default:
		l.hasError = true
		l.reporter.Error(l.line, fmt.Sprintf("Unexpected character '%s' / b'%b'", string(c), c))
	}
}

func (l *Lexer) advance() byte {
	l.current++
	return l.source[l.current]
}

func (l *Lexer) addToken(tokenType token.TokenType) {
	text := string(l.source[l.start:l.current])
	literal := []byte{}
	l.tokens = append(l.tokens, token.NewToken(tokenType, text, literal, l.line))
}
