package fhir

import "github.com/shopspring/decimal"

type Boolean bool

func (Boolean) Type() string {
	return "System.Boolean"
}

type String string

func (String) Type() string {
	return "System.String"
}

type Integer int32

func (Integer) Type() string {
	return "System.Integer"
}

type Decimal decimal.Decimal

func (Decimal) Type() string {
	return "System.Decimal"
}

type Quantity struct {
	value Decimal
	unit  string
}
