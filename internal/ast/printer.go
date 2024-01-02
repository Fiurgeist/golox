package ast

import (
	"fmt"
	"strings"
)

func PrintExpression(expr Expr) string {
	switch e := expr.(type) {
	case Binary:
		return parenthesize(e.Operator.Lexeme, e.Left, e.Right)
	case Grouping:
		return parenthesize("group", e.Expression)
	case Unary:
		return parenthesize(e.Operator.Lexeme, e.Right)
	case Literal:
		if e.Value == nil {
			return "nil"
		}
		return fmt.Sprintf("%v", e.Value)
	default:
		return fmt.Sprintf("Unhandled expr %v", expr)
	}
}

func parenthesize(name string, exprs ...Expr) string {
	var sb strings.Builder
	sb.WriteString("(")
	sb.WriteString(name)

	for _, expr := range exprs {
		sb.WriteString(" ")
		sb.WriteString(PrintExpression(expr))
	}

	sb.WriteString(")")

	return sb.String()
}
