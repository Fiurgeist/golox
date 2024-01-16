package expr

import (
	"fmt"

	"github.com/fiurgeist/golox/internal/token"
)

type Expr interface {
	isExpr()
	String() string
}

type Binary struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

func NewBinary(left Expr, operator token.Token, right Expr) *Binary {
	return &Binary{Left: left, Operator: operator, Right: right}
}

func (e *Binary) isExpr() {}
func (e *Binary) String() string {
	return fmt.Sprintf("%s %s %s", e.Left, e.Operator.Lexeme, e.Right)
}

type Logical struct {
	Left     Expr
	Operator token.Token
	Right    Expr
}

func NewLogical(left Expr, operator token.Token, right Expr) *Logical {
	return &Logical{Left: left, Operator: operator, Right: right}
}

func (e *Logical) isExpr() {}
func (e *Logical) String() string {
	return fmt.Sprintf("%s %s %s", e.Left, e.Operator.Lexeme, e.Right)
}

type Grouping struct {
	Expression Expr
}

func NewGrouping(expression Expr) *Grouping {
	return &Grouping{Expression: expression}
}

func (e *Grouping) isExpr() {}
func (e *Grouping) String() string {
	return fmt.Sprintf("(%s)", e.Expression)
}

type Unary struct {
	Operator token.Token
	Right    Expr
}

func NewUnary(operator token.Token, right Expr) *Unary {
	return &Unary{Operator: operator, Right: right}
}

func (e *Unary) isExpr() {}
func (e *Unary) String() string {
	return fmt.Sprintf("%s%s", e.Operator.Lexeme, e.Right)
}

type Literal struct {
	Value interface{}
}

func NewLiteral(value interface{}) *Literal {
	return &Literal{Value: value}
}

func (e *Literal) isExpr() {}
func (e *Literal) String() string {
	if e.Value == nil {
		return "nil"
	}

	return fmt.Sprintf("%v", e.Value)
}

type Variable struct {
	Name token.Token
}

func NewVariable(name token.Token) *Variable {
	return &Variable{Name: name}
}

func (e *Variable) isExpr() {}
func (e *Variable) String() string {
	return e.Name.Lexeme
}

type Assign struct {
	Name  token.Token
	Value Expr
}

func NewAssign(name token.Token, value Expr) *Assign {
	return &Assign{Name: name, Value: value}
}

func (e *Assign) isExpr() {}
func (e *Assign) String() string {
	return fmt.Sprintf("%s = %s", e.Name.Lexeme, e.Value)
}

type Call struct {
	Callee       Expr
	Arguments    []Expr
	ClosingParen token.Token
}

func NewCall(callee Expr, arguments []Expr, closingParen token.Token) *Call {
	return &Call{Callee: callee, Arguments: arguments, ClosingParen: closingParen}
}

func (e *Call) isExpr() {}
func (e *Call) String() string {
	return e.Callee.String()
}

type Get struct {
	Object Expr
	Name   token.Token
}

func NewGet(object Expr, name token.Token) *Get {
	return &Get{Object: object, Name: name}
}

func (e *Get) isExpr() {}
func (e *Get) String() string {
	return e.Name.Lexeme
}

type Set struct {
	Object Expr
	Name   token.Token
	Value  Expr
}

func NewSet(object Expr, name token.Token, value Expr) *Set {
	return &Set{Object: object, Name: name, Value: value}
}

func (e *Set) isExpr() {}
func (e *Set) String() string {
	return e.Name.Lexeme
}

type This struct {
	Keyword token.Token
}

func NewThis(keyword token.Token) *This {
	return &This{Keyword: keyword}
}

func (e *This) isExpr() {}
func (e *This) String() string {
	return e.Keyword.Lexeme
}

type Super struct {
	Keyword token.Token
	Method  token.Token
}

func NewSuper(keyword token.Token, method token.Token) *Super {
	return &Super{Keyword: keyword, Method: method}
}

func (e *Super) isExpr() {}
func (e *Super) String() string {
	return e.Keyword.Lexeme
}
