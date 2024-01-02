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

type Interpreter struct {
	environment Environment
	reporter    reporter.ErrorReporter
}

func NewInterpreter(environment Environment, reporter reporter.ErrorReporter) Interpreter {
	return Interpreter{environment: environment, reporter: reporter}
}

func (i *Interpreter) Interpret(statements []ast.Stmt) (err error) {
	defer func() { // TODO: replace panic with error returns
		if p := recover(); p != nil {
			if re, ok := p.(RuntimeError); ok {
				i.reporter.RuntimeError(re.Token, re.Message)
				err = ErrRuntime
			} else {
				panic(p)
			}
		}
	}()

	for _, statement := range statements {
		i.execute(statement)
	}

	return err
}

func (i *Interpreter) execute(statement ast.Stmt) {
	switch s := statement.(type) {
	case ast.PrintStmt:
		value := i.evaluate(s.Expression)
		fmt.Println(stringify(value))
	case ast.VarStmt:
		var value interface{}
		if s.Initializer != nil {
			value = i.evaluate(s.Initializer)
		}

		i.environment.Define(s.Name.Lexeme, value)
	case ast.ExpressionStmt:
		i.evaluate(s.Expression)
	case ast.BlockStmt:
		previous := i.environment
		defer func() {
			i.environment = previous
		}()

		environment := NewEnclosedEnvironment(previous)
		i.environment = environment

		for _, statement := range s.Statements {
			i.execute(statement)
		}
	default:
		panic(fmt.Sprintf("Unhandled statement %v", statement))
	}
}

func (i *Interpreter) evaluate(expr ast.Expr) interface{} {
	switch e := expr.(type) {
	case ast.Binary:
		left := i.evaluate(e.Left)
		right := i.evaluate(e.Right)

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
		return i.evaluate(e.Expression)
	case ast.Unary:
		right := i.evaluate(e.Right)

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
	case ast.Variable:
		return i.environment.Read(e.Name)
	case ast.Assign:
		value := i.evaluate(e.Value)
		i.environment.Assign(e.Name, value)
		return value
	default:
		panic(fmt.Sprintf("Unhandled expr %v", expr))
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
