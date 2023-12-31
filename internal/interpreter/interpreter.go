package interpreter

import (
	"errors"
	"fmt"

	"github.com/fiurgeist/golox/internal/ast"
	"github.com/fiurgeist/golox/internal/reporter"
	"github.com/fiurgeist/golox/internal/token"
)

var ErrRuntime = errors.New("RuntimeError")

type RuntimeError struct {
	Token   token.Token
	Message string
}

func Interpret(expr ast.Expr, reporter reporter.ErrorReporter) (err error) {
	defer func() { // TODO: replace panic with error returns
		if p := recover(); p != nil {
			if re, ok := p.(RuntimeError); ok {
				reporter.RuntimeError(re.Token, re.Message)
				err = ErrRuntime
			} else {
				panic(p)
			}
		}
	}()

	value := interpret(expr)
	fmt.Println(stringify(value))

	return err
}

func interpret(expr ast.Expr) interface{} {
	switch e := expr.(type) {
	case ast.Binary:
		left := interpret(e.Left)
		right := interpret(e.Right)

		switch e.Operator.Type {
		case token.GREATER:
			left, right, _ := numberOperands(e.Operator, left, right)
			return left > right
		case token.GREATER_EQUAL:
			left, right, _ := numberOperands(e.Operator, left, right)
			return left >= right
		case token.LESS:
			left, right, _ := numberOperands(e.Operator, left, right)
			return left < right
		case token.LESS_EQUAL:
			left, right, _ := numberOperands(e.Operator, left, right)
			return left <= right
		case token.MINUS:
			left, right, _ := numberOperands(e.Operator, left, right)
			return left - right
		case token.PLUS:
			switch left.(type) {
			case float64:
				if fRight, ok := right.(float64); ok {
					return left.(float64) + fRight
				}
			case string:
				if sRight, ok := right.(string); ok {
					return left.(string) + sRight
				}
			}
			panic(RuntimeError{Token: e.Operator, Message: fmt.Sprintf("Operands must be two numbers or two strings, got '%T', '%T'", left, right)})
		case token.SLASH:
			left, right, _ := numberOperands(e.Operator, left, right)
			return left / right
		case token.STAR:
			left, right, _ := numberOperands(e.Operator, left, right)
			return left * right
		case token.BANG_EQUAL:
			return left != right
		case token.EQUAL_EQUAL:
			return left == right
		}

		return nil
	case ast.Grouping:
		return interpret(e.Expression)
	case ast.Unary:
		right := interpret(e.Right)

		switch e.Operator.Type {
		case token.MINUS:
			if fRight, ok := right.(float64); ok {
				return -fRight
			}
			panic(RuntimeError{Token: e.Operator, Message: fmt.Sprintf("Operand must be a number, got '%T'", right)})
		case token.BANG:
			return !isTruthy(right)
		}

		return nil
	case ast.Literal:
		return e.Value
	default:
		return fmt.Sprintf("Unhandled expr %v", expr)
	}
}

func isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}

	if b, ok := value.(bool); ok {
		return b
	}

	return true
}

func numberOperands(operand token.Token, left, right interface{}) (float64, float64, error) {
	l, okL := left.(float64)
	r, okR := right.(float64)

	if okL && okR {
		return l, r, nil
	}

	panic(RuntimeError{Token: operand, Message: fmt.Sprintf("Operands must be numbers, got '%T', '%T'", left, right)})
	// return 0, 0, fmt.Errorf("Operands must be numbers, got '%T', '%T'", left, right)
}

func stringify(value interface{}) string {
	if value == nil {
		return "nil"
	}

	return fmt.Sprintf("%v", value)
}
