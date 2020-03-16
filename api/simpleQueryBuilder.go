package api

import (
	"fmt"

	"github.com/supplyon/gremcos/interfaces"
)

type simpleQB struct {
	value string
}

func NewSimpleQB(format string, a ...interface{}) interfaces.QueryBuilder {
	return &simpleQB{
		value: fmt.Sprintf(format, a...),
	}
}

func (sqb *simpleQB) String() string {
	return sqb.value
}
