package reporter

type ErrorReporter interface {
	Error(line int, message string)
	Report(line int, where, message string)
}
