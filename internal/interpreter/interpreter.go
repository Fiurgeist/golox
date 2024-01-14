package interpreter

import (
	"errors"
	"fmt"

	"github.com/fiurgeist/golox/internal/ast/expr"
	"github.com/fiurgeist/golox/internal/ast/stmt"
	"github.com/fiurgeist/golox/internal/reporter"
	"github.com/fiurgeist/golox/internal/token"
)

var ErrRuntime = errors.New("RuntimeError")

type RuntimeError struct {
	Token   token.Token
	Message string
}

func NewRuntimeError(token token.Token, message string) RuntimeError {
	return RuntimeError{Token: token, Message: message}
}

type Interpreter struct {
	globals       *Environment
	environment   *Environment
	locals        map[expr.Expr]int
	reporter      reporter.ErrorReporter
	breakOccurred bool
}

func NewInterpreter(environment *Environment, reporter reporter.ErrorReporter) Interpreter {
	environment.Define("clock", &Clock{})

	return Interpreter{environment: environment, globals: environment, reporter: reporter, locals: map[expr.Expr]int{}}
}

func (i *Interpreter) Interpret(statements []stmt.Stmt) (err error) {
	defer func() {
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

func (i *Interpreter) Resolve(expression expr.Expr, depth int) {
	i.locals[expression] = depth
}

func (i *Interpreter) execute(statement stmt.Stmt) {
	switch s := statement.(type) {
	case *stmt.Print:
		value := i.evaluate(s.Expression)
		fmt.Println(stringify(value))
	case *stmt.Var:
		var value interface{}
		if s.Initializer != nil {
			value = i.evaluate(s.Initializer)
		}

		i.environment.Define(s.Name.Lexeme, value)
	case *stmt.Expression:
		i.evaluate(s.Expression)
	case *stmt.Block:
		environment := NewEnclosedEnvironment(i.environment)
		i.executeBlock(s.Statements, environment)
	case *stmt.If:
		if isTruthy(i.evaluate(s.Condition)) {
			i.execute(s.ThenBranch)
		} else if s.ElseBranch != nil {
			i.execute(s.ElseBranch)
		}
	case *stmt.While:
		for isTruthy(i.evaluate(s.Condition)) {
			i.execute(s.Body)
			if i.breakOccurred || i.environment.ReturnOccurred() {
				i.breakOccurred = false
				break
			}
		}
	case *stmt.Break:
		i.breakOccurred = true
	case *stmt.Function:
		i.environment.Define(s.Name.Lexeme, NewFunction(s, i.environment, false))
	case *stmt.Return:
		var value interface{}
		if s.Value != nil {
			value = i.evaluate(s.Value)
		}
		i.environment.StoreReturn(s.Keyword, value)
	case *stmt.Class:
		i.environment.Define(s.Name.Lexeme, nil)

		methods := map[string]*Function{}
		for _, method := range s.Methods {
			methods[method.Name.Lexeme] = NewFunction(method, i.environment, method.Name.Lexeme == "init")
		}

		class := NewClass(s.Name.Lexeme, methods)
		i.environment.Assign(s.Name, class)
	default:
		panic(fmt.Sprintf("Unhandled statement %#v", statement))
	}
}

func (i *Interpreter) executeBlock(statements []stmt.Stmt, environment *Environment) {
	previous := i.environment
	defer func() {
		i.environment = previous
	}()

	i.environment = environment

	for _, statement := range statements {
		i.execute(statement)
		if i.breakOccurred || i.environment.ReturnOccurred() {
			break
		}
	}
}

func (i *Interpreter) evaluate(expression expr.Expr) interface{} {
	switch e := expression.(type) {
	case *expr.Binary:
		left := i.evaluate(e.Left)
		right := i.evaluate(e.Right)

		switch e.Operator.Type {
		case token.GREATER:
			left, right := numberOperands(e.Operator, left, right)
			return left > right
		case token.GREATER_EQUAL:
			left, right := numberOperands(e.Operator, left, right)
			return left >= right
		case token.LESS:
			left, right := numberOperands(e.Operator, left, right)
			return left < right
		case token.LESS_EQUAL:
			left, right := numberOperands(e.Operator, left, right)
			return left <= right
		case token.MINUS:
			left, right := numberOperands(e.Operator, left, right)
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
			panic(NewRuntimeError(
				e.Operator,
				fmt.Sprintf("Operands must be two numbers or two strings, got '%s' and '%s'", loxTxpe(left), loxTxpe(right)),
			))
		case token.SLASH:
			left, right := numberOperands(e.Operator, left, right)
			return left / right
		case token.STAR:
			left, right := numberOperands(e.Operator, left, right)
			return left * right
		case token.BANG_EQUAL:
			return left != right
		case token.EQUAL_EQUAL:
			return left == right
		}

		return nil
	case *expr.Logical:
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
	case *expr.Grouping:
		return i.evaluate(e.Expression)
	case *expr.Unary:
		right := i.evaluate(e.Right)

		switch e.Operator.Type {
		case token.MINUS:
			if fRight, ok := right.(float64); ok {
				return -fRight
			}
			panic(NewRuntimeError(
				e.Operator,
				fmt.Sprintf("Operand must be a number, got '%s'", loxTxpe(right)),
			))
		case token.BANG:
			return !isTruthy(right)
		}

		return nil
	case *expr.Literal:
		return e.Value
	case *expr.Variable:
		return i.lookUpVariable(e.Name, e)
	case *expr.Assign:
		value := i.evaluate(e.Value)
		if distance, ok := i.locals[expression]; ok {
			i.environment.AssignAt(distance, e.Name, value)
		} else {
			i.globals.Assign(e.Name, value)
		}

		return value
	case *expr.Call:
		callee := i.evaluate(e.Callee)

		var arguments []interface{}
		for _, arg := range e.Arguments {
			arguments = append(arguments, i.evaluate(arg))
		}

		function, ok := callee.(Callable)
		if !ok {
			panic(NewRuntimeError(e.ClosingParen, fmt.Sprintf("'%s' is not a function", e.Callee)))
		}

		if len(arguments) != function.Arity() {
			panic(NewRuntimeError(
				e.ClosingParen,
				fmt.Sprintf("Expected %d arguments but got %d", function.Arity(), len(arguments)),
			))
		}

		return function.Call(i, arguments)
	case *expr.Get:
		object := i.evaluate(e.Object)
		if i, ok := object.(*Instance); ok {
			return i.Get(e.Name)
		}

		panic(NewRuntimeError(e.Name, fmt.Sprintf("'%s' is not an instance", e.Object)))
	case *expr.Set:
		object := i.evaluate(e.Object)

		inst, ok := object.(*Instance)
		if !ok {
			panic(NewRuntimeError(e.Name, fmt.Sprintf("'%s' is not an instance", e.Object)))
		}

		value := i.evaluate(e.Value)
		inst.Set(e.Name, value)

		return value
	case *expr.This:
		return i.lookUpVariable(e.Keyword, e)
	default:
		panic(fmt.Sprintf("Unhandled expr %#v", expression))
	}
}

func (i *Interpreter) lookUpVariable(name token.Token, expression expr.Expr) interface{} {
	if distance, ok := i.locals[expression]; ok {
		return i.environment.ReadAt(distance, name.Lexeme)
	}
	return i.globals.Read(name)
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

func numberOperands(operand token.Token, left, right interface{}) (float64, float64) {
	l, okL := left.(float64)
	r, okR := right.(float64)

	if okL && okR {
		return l, r
	}

	panic(NewRuntimeError(
		operand,
		fmt.Sprintf("Operands must be numbers, got '%s' and '%s'", loxTxpe(left), loxTxpe(right)),
	))
}

func stringify(value interface{}) string {
	if value == nil {
		return "nil"
	}
	if c, ok := value.(interface{ String() string }); ok {
		return c.String()
	}

	return fmt.Sprintf("%v", value)
}

func loxTxpe(value interface{}) string {
	if value == nil {
		return "nil"
	}

	switch value.(type) {
	case float64:
		return "number"
	case string:
		return "string"
	case bool:
		return "Boolean"
	default:
		panic(fmt.Sprintf("Unhandled type '%T'", value))
	}
}
