package reporter

import (
	"fmt"

	"github.com/fiurgeist/golox/internal/token"
)

var _ ErrorReporter = (*ConsoleReporter)(nil)

type ConsoleReporter struct {
}

func (r *ConsoleReporter) LexingError(line int, message string) {
	r.Report(line, "", message)
}

func (r *ConsoleReporter) ParseError(parsedToken token.Token, message string) {
	if parsedToken.Type == token.EOF {
		r.Report(parsedToken.Line, " at end", message)
	} else {
		r.Report(parsedToken.Line, fmt.Sprintf(" at '%s'", parsedToken.Lexeme), message)
	}
}

func (r *ConsoleReporter) RuntimeError(interpretedToken token.Token, message string) {
	fmt.Printf("[line %d] RuntimeError: %s\n", interpretedToken.Line, message)
}

func (r *ConsoleReporter) Report(line int, where, message string) {
	fmt.Printf("[line %d] Error%s: %s\n", line, where, message)
}
