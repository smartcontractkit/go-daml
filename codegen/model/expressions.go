package model

import (
	"fmt"
)

type DamlExpression interface {
	GoType() string
	GoImports() []ExternalPackage
	GoDef() string
}

type TextLiteral struct {
	Text
	Value string
}

func (s TextLiteral) GoDef() string {
	return fmt.Sprintf("%s(\"%s\")", s.Text.GoType(), s.Value)
}

type Int64Literal struct {
	Int64
	Value int64
}

func (i Int64Literal) GoDef() string {
	return fmt.Sprintf("%s(%d)", i.Int64.GoType(), i.Value)
}

type NumericLiteral struct {
	Numeric
	Value string
}

func (n NumericLiteral) GoDef() string {
	return fmt.Sprintf("%s(\"%s\")", n.Numeric.GoType(), n.Value)
}
