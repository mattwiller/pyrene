package system

import (
	"github.com/shopspring/decimal"
)

type Decimal struct {
	inner decimal.Decimal
}

func NewDecimal(d float64) Decimal {
	return Decimal{inner: decimal.NewFromFloatWithExponent(d, -10)}
}

var _ Value = Decimal{inner: decimal.NewFromFloat(3.141592653)}

func (Decimal) Type() ValueType {
	return DecimalType
}

func (d Decimal) String() string {
	return d.inner.String()
}
