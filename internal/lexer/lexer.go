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
	return l.current >= len(l.source)
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
	case '!':
		if l.match('=') {
			l.addToken(token.BANG_EQUAL)
		} else {
			l.addToken(token.BANG)
		}
	case '=':
		if l.match('=') {
			l.addToken(token.EQUAL_EQUAL)
		} else {
			l.addToken(token.EQUAL)
		}
	case '<':
		if l.match('=') {
			l.addToken(token.LESS_EQUAL)
		} else {
			l.addToken(token.LESS)
		}
	case '>':
		if l.match('=') {
			l.addToken(token.GREATER_EQUAL)
		} else {
			l.addToken(token.GREATER)
		}
	case '/':
		if l.match('/') {
			for l.peek() != '\n' && !l.isAtEnd() {
				l.advance()
			}
		} else if l.match('*') {
			l.blockComment()
		} else {
			l.addToken(token.SLASH)
		}
	case ' ':
	case '\r':
	case '\t':
		break
	case '\n':
		l.line++
	case '"':
		l.string()
	default:
		if l.isDigit(c) {
			l.number()
		} else if l.isAlpha(c) {
			l.identifier()
		} else {
			l.hasError = true
			l.reporter.Error(l.line, fmt.Sprintf("Unexpected character '%s' / b'%b'", string(c), c))
		}
	}
}

func (l *Lexer) advance() byte {
	c := l.source[l.current]
	l.current++
	return c
}

func (l *Lexer) addToken(tokenType token.TokenType) {
	text := string(l.source[l.start:l.current])
	literal := []byte{}
	l.tokens = append(l.tokens, token.NewToken(tokenType, text, literal, l.line))
}

func (l *Lexer) addStringToken() {
	text := string(l.source[l.start:l.current])
	literal := l.source[l.start+1 : l.current-1] // trim quotes
	l.tokens = append(l.tokens, token.NewToken(token.STRING, text, literal, l.line))
}

func (l *Lexer) addNumberToken() {
	literal := l.source[l.start:l.current]
	text := string(literal)
	l.tokens = append(l.tokens, token.NewToken(token.NUMBER, text, literal, l.line))
}

func (l *Lexer) match(expected byte) bool {
	if l.isAtEnd() {
		return false
	}

	if l.source[l.current] != expected {
		return false
	}

	l.current++
	return true
}

func (l *Lexer) peek() byte {
	if l.isAtEnd() {
		return 0 // null terminated string \x00
	}
	return l.source[l.current]
}

func (l *Lexer) nextPeek() byte {
	if l.current+1 >= len(l.source) {
		return 0 // null terminated string \x00
	}
	return l.source[l.current+1]
}

func (l *Lexer) string() {
	for l.peek() != '"' && !l.isAtEnd() {
		if l.peek() == '\n' {
			l.line++
		}
		l.advance()
	}

	if l.isAtEnd() {
		l.hasError = true
		l.reporter.Error(l.line, "Unterminated string")
		return
	}

	l.advance() // the closing "
	l.addStringToken()
}

func (l *Lexer) isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func (l *Lexer) number() {
	for l.isDigit(l.peek()) {
		l.advance()
	}

	if l.peek() == '.' && l.isDigit(l.nextPeek()) {
		l.advance()
		for l.isDigit(l.peek()) {
			l.advance()
		}
	}

	l.addNumberToken()
}

func (l *Lexer) isAlpha(char byte) bool {
	return char == '_' || (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

func (l *Lexer) isAlphaNumeric(char byte) bool {
	return l.isAlpha(char) || l.isDigit(char)
}

func (l *Lexer) identifier() {
	for l.isAlphaNumeric(l.peek()) {
		l.advance()
	}

	tokenType := token.Keywords[string(l.source[l.start:l.current])]
	if tokenType == token.NONE_ {
		tokenType = token.IDENTIFIER
	}
	l.addToken(tokenType)
}

func (l *Lexer) blockComment() {
	for !l.isAtEnd() && (l.peek() != '*' || l.nextPeek() != '/') {
		if l.peek() == '\n' {
			l.line++
		}
		l.advance()
	}

	if l.isAtEnd() || l.peek() != '*' || l.nextPeek() != '/' {
		l.hasError = true
		l.reporter.Error(l.line, "Unterminated comment block")
		return
	}

	l.advance()
	l.advance()
}
