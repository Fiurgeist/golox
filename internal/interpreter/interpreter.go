package interpreter

import (
	"errors"
	"fmt"

	"github.com/fiurgeist/golox/internal/expr"
	"github.com/fiurgeist/golox/internal/reporter"
	"github.com/fiurgeist/golox/internal/stmt"
	"github.com/fiurgeist/golox/internal/token"
)

var ErrRuntime = errors.New("RuntimeError")

type RuntimeError struct {
	Token   token.Token
	Message string
}

type Interpreter struct {
	environment   Environment
	reporter      reporter.ErrorReporter
	breakOccurred bool
}

func NewInterpreter(environment Environment, reporter reporter.ErrorReporter) Interpreter {
	return Interpreter{environment: environment, reporter: reporter}
}

func (i *Interpreter) Interpret(statements []stmt.Stmt) (err error) {
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

func (i *Interpreter) execute(statement stmt.Stmt) {
	switch s := statement.(type) {
	case stmt.Print:
		value := i.evaluate(s.Expression)
		fmt.Println(stringify(value))
	case stmt.Var:
		var value interface{}
		if s.Initializer != nil {
			value = i.evaluate(s.Initializer)
		}

		i.environment.Define(s.Name.Lexeme, value)
	case stmt.Expression:
		i.evaluate(s.Expression)
	case stmt.Block:
		previous := i.environment
		defer func() {
			i.environment = previous
		}()

		environment := NewEnclosedEnvironment(previous)
		i.environment = environment

		for _, statement := range s.Statements {
			i.execute(statement)
			if i.breakOccurred {
				break
			}
		}
	case stmt.If:
		if isTruthy(i.evaluate(s.Condition)) {
			i.execute(s.ThenBranch)
		} else if s.ElseBranch != nil {
			i.execute(s.ElseBranch)
		}
	case stmt.While:
		for isTruthy(i.evaluate(s.Condition)) {
			i.execute(s.Body)
			if i.breakOccurred {
				i.breakOccurred = false
				break
			}
		}
	case stmt.Break:
		i.breakOccurred = true
	default:
		panic(fmt.Sprintf("Unhandled statement %v", statement))
	}
}

func (i *Interpreter) evaluate(expression expr.Expr) interface{} {
	switch e := expression.(type) {
	case expr.Binary:
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
	case expr.Logical:
		left := i.evaluate(e.Left)

		if e.Operator.Type == token.OR {
			if isTruthy(left) {
				return left
			}
		} else {
			if !isTruthy(left) {
				return left
			}
		}

		return i.evaluate(e.Right)
	case expr.Grouping:
		return i.evaluate(e.Expression)
	case expr.Unary:
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
	case expr.Literal:
		return e.Value
	case expr.Variable:
		return i.environment.Read(e.Name)
	case expr.Assign:
		value := i.evaluate(e.Value)
		i.environment.Assign(e.Name, value)
		return value
	default:
		panic(fmt.Sprintf("Unhandled expr %v", expression))
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
