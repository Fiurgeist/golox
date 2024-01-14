package interpreter

import (
	"fmt"

	"github.com/fiurgeist/golox/internal/token"
)

var _ Callable = (*Class)(nil)

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
