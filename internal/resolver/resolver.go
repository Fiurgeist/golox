package resolver

import (
	"fmt"

	"github.com/fiurgeist/golox/internal/expr"
	"github.com/fiurgeist/golox/internal/interpreter"
	"github.com/fiurgeist/golox/internal/reporter"
	"github.com/fiurgeist/golox/internal/stmt"
	"github.com/fiurgeist/golox/internal/token"
)

type FunctionType int

const (
	NONE FunctionType = iota
	FUNCTION
)

type Resolver struct {
	interpreter     interpreter.Interpreter
	reporter        reporter.ErrorReporter
	scopes          []map[string]bool
	currentFunction FunctionType
}

func NewResolver(interpreter interpreter.Interpreter, reporter reporter.ErrorReporter) Resolver {
	return Resolver{interpreter: interpreter, reporter: reporter, scopes: []map[string]bool{}}
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
		r.resolveFunction(s, FUNCTION)
	case *stmt.Return:
		if r.currentFunction == NONE {
			r.reporter.ParseError(s.Keyword, "Can't return from top-level code")
		}
		if s.Value != nil {
			r.resolveExpr(s.Value)
		}
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
			if def, ok := r.scopes[0][e.Name.Lexeme]; ok && !def {
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
	default:
		panic(fmt.Sprintf("Unhandled expr %#v", expression))
	}
}

func (r *Resolver) resolveLocal(expression expr.Expr, name token.Token) {
	for i, scope := range r.scopes {
		if scope[name.Lexeme] {
			r.interpreter.Resolve(expression, i)
			return
		}
	}
}

func (r *Resolver) resolveFunction(function *stmt.Function, functionType FunctionType) {
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
	r.scopes = append([]map[string]bool{{}}, r.scopes...)
}

func (r *Resolver) endScope() {
	r.scopes = r.scopes[1:]
}

func (r *Resolver) declare(name token.Token) {
	if len(r.scopes) == 0 {
		return
	}

	scope := r.scopes[0]
	if scope[name.Lexeme] {
		r.reporter.ParseError(name, "Already a variable with this name in this scope")
	}

	scope[name.Lexeme] = false
}

func (r *Resolver) define(name token.Token) {
	if len(r.scopes) == 0 {
		return
	}

	r.scopes[0][name.Lexeme] = true
}
