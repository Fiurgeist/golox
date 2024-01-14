package interpreter

import (
	"fmt"

	"github.com/fiurgeist/golox/internal/ast/stmt"
)

var _ Callable = (*Function)(nil)

type Function struct {
	declaration   *stmt.Function
	closure       *Environment
	isInitializer bool
}

func NewFunction(declaration *stmt.Function, closure *Environment, isInitializer bool) *Function {
	return &Function{declaration: declaration, closure: closure, isInitializer: isInitializer}
}

func (c *Function) Call(interpreter *Interpreter, arguments []interface{}) interface{} {
	environment := NewFunctionEnvironment(c.closure)
	for i, param := range c.declaration.Params {
		environment.Define(param.Lexeme, arguments[i])
	}

	interpreter.executeBlock(c.declaration.Body, environment)

	if c.isInitializer {
		return c.closure.ReadAt(0, "this")
	}

	return environment.ReadReturn()
}

func (c *Function) Arity() int {
	return len(c.declaration.Params)
}

func (c *Function) bind(instance *Instance) *Function {
	environment := NewFunctionEnvironment(c.closure)
	environment.Define("this", instance)
	return NewFunction(c.declaration, environment, c.declaration.Name.Lexeme == "init")
}

func (c *Function) String() string {
	return fmt.Sprintf("<fn %s>", c.declaration.Name.Lexeme)
}
