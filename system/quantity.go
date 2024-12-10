package system

import (
	"fmt"
)

type Quantity struct {
	value Decimal
	unit  string
}

func NewQuantity(value float64, unit string) Quantity {
	return Quantity{value: NewDecimal(value), unit: unit}
}

var _ Value = Quantity{value: NewDecimal(3.14), unit: "rad"}

func (Quantity) Type() ValueType {
	return QuantityType
}

func (q Quantity) String() string {
	return fmt.Sprintf(`%s '%s'`, q.value, q.unit)
}
