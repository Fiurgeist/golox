package interpreter

import (
	"fmt"

	"github.com/fiurgeist/golox/internal/token"
)

type Environment struct {
	enclosing *Environment
	values    map[string]interface{}
}

func NewEnvironment() Environment {
	return Environment{values: map[string]interface{}{}, enclosing: nil}
}

func NewEnclosedEnvironment(enclosing Environment) Environment {
	return Environment{values: map[string]interface{}{}, enclosing: &enclosing}
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
