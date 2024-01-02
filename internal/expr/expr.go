package expr

import (
	"github.com/fiurgeist/golox/internal/token"
)

type Expr interface {
	isExpr()
}

type Binary struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

func NewBinary(left Expr, operator token.Token, right Expr) Binary {
	return Binary{Left: left, Operator: operator, Right: right}
}

func (e Binary) isExpr() {}

type Grouping struct {
	Expression Expr
}

func NewGrouping(expression Expr) Grouping {
	return Grouping{Expression: expression}
}

func (e Grouping) isExpr() {}

type Unary struct {
	Operator token.Token
	Right    Expr
}

func NewUnary(operator token.Token, right Expr) Unary {
	return Unary{Operator: operator, Right: right}
}

func (e Unary) isExpr() {}

type Literal struct {
	Value interface{}
}

func NewLiteral(value interface{}) Literal {
	return Literal{Value: value}
}

func (e Literal) isExpr() {}

type Variable struct {
	Name token.Token
}

func NewVariable(name token.Token) Variable {
	return Variable{Name: name}
}

func (e Variable) isExpr() {}

type Assign struct {
	Name  token.Token
	Value Expr
}

func NewAssign(name token.Token, value Expr) Assign {
	return Assign{Name: name, Value: value}
}

func (e Assign) isExpr() {}
