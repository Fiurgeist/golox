package interpreter

import (
	"time"
)

var _ Callable = (*Clock)(nil)

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
