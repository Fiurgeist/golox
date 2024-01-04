package interpreter

import (
	"fmt"

	"github.com/fiurgeist/golox/internal/token"
)

type Environment struct {
	enclosing           *Environment
	values              map[string]interface{}
	functionEnvironment *functionEnvironment
}

type functionEnvironment struct {
	returnValue *returnValue
}

type returnValue struct {
	value interface{}
}

func NewEnvironment() *Environment {
	return &Environment{values: map[string]interface{}{}}
}

func NewEnclosedEnvironment(enclosing *Environment) *Environment {
	return &Environment{values: map[string]interface{}{}, enclosing: enclosing}
}

func NewFunctionEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		values:              map[string]interface{}{},
		enclosing:           enclosing,
		functionEnvironment: &functionEnvironment{},
	}
}

func (e *Environment) Define(name string, value interface{}) {
	e.values[name] = value
}

func (e *Environment) Read(name token.Token) interface{} {
	value, ok := e.values[name.Lexeme]
	if ok {
		return value
	}

	if e.enclosing != nil {
		return e.enclosing.Read(name)
	}

	panic(RuntimeError{Token: name, Message: fmt.Sprintf("Undefined variable '%s'", name.Lexeme)})
}

func (e *Environment) Assign(name token.Token, value interface{}) {
	if _, ok := e.values[name.Lexeme]; ok {
		e.values[name.Lexeme] = value
		return
	}

	if e.enclosing != nil {
		e.enclosing.Assign(name, value)
		return
	}

	panic(RuntimeError{Token: name, Message: fmt.Sprintf("Undefined variable '%s'", name.Lexeme)})
}

func (e *Environment) StoreReturn(token token.Token, value interface{}) {
	if e.functionEnvironment != nil {
		e.functionEnvironment.returnValue = &returnValue{value: value}
		return
	}

	if e.enclosing != nil {
		e.enclosing.StoreReturn(token, value)
		return
	}

	panic(RuntimeError{Token: token, Message: "Return outside of function"})
}

func (e *Environment) ReadReturn() interface{} {
	// null pointer exception if called outside of a function
	if e.functionEnvironment.returnValue != nil {
		return e.functionEnvironment.returnValue.value
	}

	return nil
}

func (e *Environment) ReturnOccurred() bool {
	if e.functionEnvironment != nil {
		return e.functionEnvironment.returnValue != nil
	}

	if e.enclosing != nil {
		return e.enclosing.ReturnOccurred()
	}

	return false
}
