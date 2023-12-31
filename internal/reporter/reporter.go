package reporter

import "github.com/fiurgeist/golox/internal/token"

type ErrorReporter interface {
	LexingError(line int, message string)
	ParseError(token token.Token, message string)
	RuntimeError(token token.Token, message string)
	Report(line int, where, message string)
}
