package stmt

import (
	"github.com/fiurgeist/golox/internal/ast/expr"
	"github.com/fiurgeist/golox/internal/token"
)

type Stmt interface {
	isStmt()
}

type Expression struct {
	Expression expr.Expr
}

func NewExpression(expression expr.Expr) *Expression {
	return &Expression{Expression: expression}
}

func (s *Expression) isStmt() {}

type Print struct {
	Expression expr.Expr
}

func NewPrint(expression expr.Expr) *Print {
	return &Print{Expression: expression}
}

func (s *Print) isStmt() {}

type Var struct {
	Name        token.Token
	Initializer expr.Expr
}

func NewVar(name token.Token, initializer expr.Expr) *Var {
	return &Var{Name: name, Initializer: initializer}
}

func (s *Var) isStmt() {}

type Block struct {
	Statements []Stmt
}

func NewBlock(statements []Stmt) *Block {
	return &Block{Statements: statements}
}

func (s *Block) isStmt() {}

type If struct {
	Condition  expr.Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func NewIf(condition expr.Expr, thenBranch, elseBranch Stmt) *If {
	return &If{Condition: condition, ThenBranch: thenBranch, ElseBranch: elseBranch}
}

func (s *If) isStmt() {}

type While struct {
	Condition expr.Expr
	Body      Stmt
}

func NewWhile(condition expr.Expr, body Stmt) *While {
	return &While{Condition: condition, Body: body}
}

func (s *While) isStmt() {}

type Break struct{}

func NewBreak() *Break {
	return &Break{}
}

func (s *Break) isStmt() {}

type Function struct {
	Name   token.Token
	Params []token.Token
	Body   []Stmt
}

func NewFunction(name token.Token, params []token.Token, body []Stmt) *Function {
	return &Function{Name: name, Params: params, Body: body}
}

func (s *Function) isStmt() {}

type Return struct {
	Keyword token.Token
	Value   expr.Expr
}

func NewReturn(keyword token.Token, value expr.Expr) *Return {
	return &Return{Keyword: keyword, Value: value}
}

func (s *Return) isStmt() {}

type Class struct {
	Name    token.Token
	Methods []*Function
}

func NewClass(name token.Token, methods []*Function) *Class {
	return &Class{Name: name, Methods: methods}
}

func (s *Class) isStmt() {}
