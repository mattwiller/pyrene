package fhirpath

import (
	"fmt"

	"github.com/mattwiller/pyrene/system"
)

var implicitConversions = map[string]func(system.Value) (system.Value, error){
	string(system.IntegerType) + "|" + string(system.DecimalType): func(v system.Value) (system.Value, error) {
		return system.NewDecimal(float64(v.(system.Integer))), nil
	},
	string(system.IntegerType) + "|" + string(system.QuantityType): func(v system.Value) (system.Value, error) {
		return system.NewQuantity(float64(v.(system.Integer)), "1"), nil
	},
}

var explicitConversions = map[string]func(system.Value) (system.Value, error){
	string(system.IntegerType) + "|" + string(system.DecimalType): func(v system.Value) (system.Value, error) {
		return system.NewDecimal(float64(v.(system.Integer))), nil
	},
	string(system.IntegerType) + "|" + string(system.QuantityType): func(v system.Value) (system.Value, error) {
		return system.NewQuantity(float64(v.(system.Integer)), "1"), nil
	},
}

func Cast(value system.Value, targetType system.ValueType) (system.Value, error) {
	if value.Type() == targetType {
		return value, nil
	}
	castFn, isConvertible := implicitConversions[string(value.Type())+"|"+string(targetType)]
	if !isConvertible {
		return nil, fmt.Errorf("cannot use %s value as %s", value.Type(), targetType)
	}
	return castFn(value)
}

func Convert(value system.Value, targetType system.ValueType) (system.Value, error) {
	if value.Type() == targetType {
		return value, nil
	}
	convertFn, isConvertible := explicitConversions[string(value.Type())+"|"+string(targetType)]
	if !isConvertible {
		return nil, fmt.Errorf("cannot convert %s value to %s", value.Type(), targetType)
	}
	return convertFn(value)
}
