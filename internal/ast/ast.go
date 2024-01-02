package ast

import (
	"github.com/fiurgeist/golox/internal/token"
)

type Expr interface {
}

type Binary struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

func NewBinary(left Expr, operator token.Token, right Expr) Binary {
	return Binary{Left: left, Operator: operator, Right: right}
}

type Grouping struct {
	Expression Expr
}

func NewGrouping(expression Expr) Grouping {
	return Grouping{Expression: expression}
}

type Unary struct {
	Operator token.Token
	Right    Expr
}

func NewUnary(operator token.Token, right Expr) Unary {
	return Unary{Operator: operator, Right: right}
}

type Literal struct {
	Value interface{}
}

func NewLiteral(value interface{}) Literal {
	return Literal{Value: value}
}

type Variable struct {
	Name token.Token
}

func NewVariable(name token.Token) Variable {
	return Variable{Name: name}
}

type Assign struct {
	Name  token.Token
	Value Expr
}

func NewAssign(name token.Token, value Expr) Assign {
	return Assign{Name: name, Value: value}
}

type Stmt interface {
}

type ExpressionStmt struct {
	Expression Expr
}

func NewExpressionStmt(expression Expr) ExpressionStmt {
	return ExpressionStmt{Expression: expression}
}

type PrintStmt struct {
	Expression Expr
}

func NewPrintStmt(expression Expr) PrintStmt {
	return PrintStmt{Expression: expression}
}

type VarStmt struct {
	Name        token.Token
	Initializer Expr
}

func NewVarStmt(name token.Token, Initializer Expr) VarStmt {
	return VarStmt{Name: name, Initializer: Initializer}
}

type BlockStmt struct {
	Statements []Stmt
}

func NewBlockStmt(statements []Stmt) BlockStmt {
	return BlockStmt{Statements: statements}
}
