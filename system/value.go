package system

import "fmt"

type ValueType string

const (
	BooleanType  ValueType = "System.Boolean"
	StringType   ValueType = "System.String"
	IntegerType  ValueType = "System.Integer"
	DecimalType  ValueType = "System.Decimal"
	DateType     ValueType = "System.Date"
	DateTimeType ValueType = "System.DateTime"
	TimeType     ValueType = "System.Time"
	QuantityType ValueType = "System.Quantity"
	AnyType      ValueType = "System.Any"
)

type Value interface {
	Type() ValueType
	fmt.Stringer
}
