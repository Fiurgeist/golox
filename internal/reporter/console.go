package reporter

import "fmt"

var _ ErrorReporter = (*ConsoleReporter)(nil)

type ConsoleReporter struct {
}

func (r *ConsoleReporter) Error(line int, message string) {
	r.Report(line, "", message)
}

func (r *ConsoleReporter) Report(line int, where, message string) {
	fmt.Printf("[line %d] Error%s: %s\n", line, where, message)
}
