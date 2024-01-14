package interpreter

import (
	"fmt"
	"time"

	"github.com/fiurgeist/golox/internal/stmt"
	"github.com/fiurgeist/golox/internal/token"
)

type Callable interface {
	Call(interpreter *Interpreter, arguments []interface{}) interface{}
	Arity() int
	String() string
}

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

type Class struct {
	name    string
	methods map[string]*Function
}

func NewClass(name string, methods map[string]*Function) *Class {
	return &Class{name: name, methods: methods}
}

func (c *Class) Call(interpreter *Interpreter, arguments []interface{}) interface{} {
	instance := &Instance{class: c, fields: map[string]interface{}{}}

	if initializer := c.findMethod("init"); initializer != nil {
		initializer.bind(instance).Call(interpreter, arguments)
	}

	return instance
}

func (c *Class) Arity() int {
	if initializer := c.findMethod("init"); initializer != nil {
		return initializer.Arity()
	}

	return 0
}

func (c *Class) findMethod(name string) *Function {
	return c.methods[name]
}

func (c *Class) String() string {
	return c.name
}

type Instance struct {
	class  *Class
	fields map[string]interface{}
}

func (i *Instance) Get(name token.Token) interface{} {
	if value, ok := i.fields[name.Lexeme]; ok {
		return value
	}

	if method := i.class.findMethod(name.Lexeme); method != nil {
		return method.bind(i)
	}

	panic(NewRuntimeError(name, fmt.Sprintf("Undefined property '%s'", name.Lexeme)))
}

func (i *Instance) Set(name token.Token, value interface{}) {
	i.fields[name.Lexeme] = value
}

func (i *Instance) String() string {
	return fmt.Sprintf("%s instance", i.class.name)
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
