package interpreter

import (
	"fmt"
	"time"

	"github.com/fiurgeist/golox/internal/stmt"
)

type Callable interface {
	Call(interpreter *Interpreter, arguments []interface{}) interface{}
	Arity() int
	String() string
}

type Function struct {
	declaration stmt.Function
	closure     *Environment
}

func NewFunction(declaration stmt.Function, closure *Environment) *Function {
	return &Function{declaration: declaration, closure: closure}
}

func (c *Function) Call(interpreter *Interpreter, arguments []interface{}) interface{} {
	environment := NewFunctionEnvironment(c.closure)
	for i, param := range c.declaration.Params {
		environment.Define(param.Lexeme, arguments[i])
	}

	interpreter.executeBlock(c.declaration.Body, environment)

	return environment.ReadReturn()
}

func (c *Function) Arity() int {
	return len(c.declaration.Params)
}

func (c *Function) String() string {
	return fmt.Sprintf("<fn %s>", c.declaration.Name.Lexeme)
}

// native functions

type Clock struct{}

func (c *Clock) Call(interpreter *Interpreter, arguments []interface{}) interface{} {
	return float64(time.Now().UnixMilli())
}

func (c *Clock) Arity() int {
	return 0
}

func (c *Clock) String() string {
	return "<native fn>"
}
