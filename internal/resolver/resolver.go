package resolver

import (
	"fmt"

	"github.com/fiurgeist/golox/internal/ast/class"
	"github.com/fiurgeist/golox/internal/ast/expr"
	"github.com/fiurgeist/golox/internal/ast/function"
	"github.com/fiurgeist/golox/internal/ast/stmt"
	"github.com/fiurgeist/golox/internal/interpreter"
	"github.com/fiurgeist/golox/internal/reporter"
	"github.com/fiurgeist/golox/internal/token"
)

type Resolver struct {
	interpreter     interpreter.Interpreter
	reporter        reporter.ErrorReporter
	scopes          []map[string]*variableStatus
	currentFunction function.Type
	currentClass    class.Type
}

type variableStatus struct {
	name    token.Token
	defined bool
	used    bool
}

func NewResolver(interpreter interpreter.Interpreter, reporter reporter.ErrorReporter) Resolver {
	return Resolver{
		interpreter: interpreter,
		reporter:    reporter,
		scopes:      []map[string]*variableStatus{},
	}
}

func (r *Resolver) Resolve(statements []stmt.Stmt) {
	for _, statement := range statements {
		r.resolveStmt(statement)
	}
}

func (r *Resolver) resolveStmt(statement stmt.Stmt) {
	switch s := statement.(type) {
	case *stmt.Print:
		r.resolveExpr(s.Expression)
	case *stmt.Var:
		r.declare(s.Name)
		if s.Initializer != nil {
			r.resolveExpr(s.Initializer)
		}
		r.define(s.Name)
	case *stmt.Expression:
		r.resolveExpr(s.Expression)
	case *stmt.Block:
		r.beginScope()
		r.Resolve(s.Statements)
		r.endScope()
	case *stmt.If:
		r.resolveExpr(s.Condition)
		r.resolveStmt(s.ThenBranch)
		if s.ElseBranch != nil {
			r.resolveStmt(s.ThenBranch)
		}
	case *stmt.While:
		r.resolveExpr(s.Condition)
		r.resolveStmt(s.Body)
	case *stmt.Break:
		break
	case *stmt.Function:
		r.declare(s.Name)
		r.define(s.Name)
		r.resolveFunction(s, function.FUNCTION)
	case *stmt.Return:
		if r.currentFunction == function.NONE {
			r.reporter.ParseError(s.Keyword, "Can't return from top-level code")
		}
		if s.Value != nil {
			if r.currentFunction == function.INITIALIZER {
				r.reporter.ParseError(s.Keyword, "Can't return a value from an initializer")
			}
			r.resolveExpr(s.Value)
		}
	case *stmt.Class:
		enclosingClass := r.currentClass
		r.currentClass = class.CLASS

		r.declare(s.Name)
		r.define(s.Name)

		r.beginScope()
		r.scopes[0]["this"] = &variableStatus{defined: true, used: true}

		for _, method := range s.Methods {
			declaration := function.METHOD
			if method.Name.Lexeme == "init" {
				declaration = function.INITIALIZER
			}

			r.resolveFunction(method, declaration)
		}
		r.endScope()

		r.currentClass = enclosingClass
	default:
		panic(fmt.Sprintf("Unhandled statement %#v", statement))
	}
}

func (r *Resolver) resolveExpr(expression expr.Expr) {
	switch e := expression.(type) {
	case *expr.Binary:
		r.resolveExpr(e.Left)
		r.resolveExpr(e.Right)
	case *expr.Logical:
		r.resolveExpr(e.Left)
		r.resolveExpr(e.Right)
	case *expr.Grouping:
		r.resolveExpr(e.Expression)
	case *expr.Unary:
		r.resolveExpr(e.Right)
	case *expr.Literal:
		break
	case *expr.Variable:
		if len(r.scopes) != 0 {
			if val, ok := r.scopes[0][e.Name.Lexeme]; ok && !val.defined {
				r.reporter.ParseError(e.Name, "Can't read local variable in its own initializer")
			}
		}
		r.resolveLocal(e, e.Name)
	case *expr.Assign:
		r.resolveExpr(e.Value)
		r.resolveLocal(e, e.Name)
	case *expr.Call:
		r.resolveExpr(e.Callee)

		for _, arg := range e.Arguments {
			r.resolveExpr(arg)
		}
	case *expr.Get:
		r.resolveExpr(e.Object)
	case *expr.Set:
		r.resolveExpr(e.Object)
		r.resolveExpr(e.Value)
	case *expr.This:
		if r.currentClass == class.NONE {
			r.reporter.ParseError(e.Keyword, "Can't use 'this' outside of a class")
		}
		r.resolveLocal(e, e.Keyword)
	default:
		panic(fmt.Sprintf("Unhandled expr %#v", expression))
	}
}

func (r *Resolver) resolveLocal(expression expr.Expr, name token.Token) {
	for i, scope := range r.scopes {
		if val, ok := scope[name.Lexeme]; ok && val.defined {
			r.interpreter.Resolve(expression, i)
			val.used = true
			return
		}
	}
}

func (r *Resolver) resolveFunction(function *stmt.Function, functionType function.Type) {
	enclosingType := r.currentFunction
	r.currentFunction = functionType

	r.beginScope()
	for _, param := range function.Params {
		r.declare(param)
		r.define(param)
	}
	r.Resolve(function.Body)
	r.endScope()

	r.currentFunction = enclosingType
}

func (r *Resolver) beginScope() {
	r.scopes = append([]map[string]*variableStatus{{}}, r.scopes...)
}

func (r *Resolver) endScope() {
	for _, stat := range r.scopes[0] {
		if !stat.used {
			r.reporter.ParseError(stat.name, "Local variable is unused")
		}
	}
	r.scopes = r.scopes[1:]
}

func (r *Resolver) declare(name token.Token) {
	if len(r.scopes) == 0 {
		return
	}

	scope := r.scopes[0]
	if val, ok := scope[name.Lexeme]; ok && val.defined {
		r.reporter.ParseError(name, "Already a variable with this name in this scope")
	}

	scope[name.Lexeme] = &variableStatus{name: name}
}

func (r *Resolver) define(name token.Token) {
	if len(r.scopes) == 0 {
		return
	}

	r.scopes[0][name.Lexeme].defined = true
}
