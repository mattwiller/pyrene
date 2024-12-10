package ucum

import (
	"bytes"
	_ "embed"

	xml "github.com/dgrr/quickxml"
	"github.com/mattwiller/pyrene/system"

	"github.com/shopspring/decimal"
)

//go:embed ucum-essence.xml
var definitionsData []byte

type Prefix struct {
	Code    string
	Name    string
	Display string
	Factor  decimal.Decimal
}

type BaseUnit struct {
	Code     string
	Name     string
	Display  string
	Property string
}

type Unit struct {
	Code     string
	Name     string
	Display  string
	Property string
	Unit     string
	Value    decimal.Decimal
	IsMetric bool
}

type UnitConverter struct {
	Prefixes  map[string]Prefix
	BaseUnits []BaseUnit
	Units     map[string]Unit
}

func (converter UnitConverter) Convert(quantity system.Quantity, newUnits string) (system.Quantity, error) {
	//
	return system.Quantity{}, nil
}

const (
	prefixEl   = "prefix"
	baseUnitEl = "base-unit"
	unitEl     = "unit"
)

func NewUnitConverter(data []byte) UnitConverter {
	converter := UnitConverter{
		Prefixes:  make(map[string]Prefix, 24),
		BaseUnits: make([]BaseUnit, 0, 8),
		Units:     make(map[string]Unit, 256),
	}

	r := xml.NewReader(bytes.NewReader(definitionsData))
	for r.Next() {
		switch el := r.Element().(type) {
		case *xml.StartElement:
			switch el.Name() {
			case prefixEl:
				prefix := parsePrefix(el, r)
				converter.Prefixes[prefix.Code] = prefix
			case baseUnitEl:
				baseUnit := parseBaseUnit(el, r)
				converter.BaseUnits = append(converter.BaseUnits, baseUnit)
				converter.Units[baseUnit.Code] = Unit{
					Code:     baseUnit.Code,
					Name:     baseUnit.Name,
					Display:  baseUnit.Display,
					Property: baseUnit.Property,
					Value:    decimal.NewFromInt(1),
					IsMetric: true,
				}
			case unitEl:
				if isSpecial := el.Attrs().Get("isSpecial"); isSpecial != nil {
					continue
				} else if isArbitrary := el.Attrs().Get("isArbitrary"); isArbitrary != nil {
					continue
				}
				unit := parseUnit(el, r)
				converter.Units[unit.Code] = unit
			}
		case *xml.EndElement:

		}
	}

	return converter
}

func parsePrefix(start *xml.StartElement, r *xml.Reader) Prefix {
	prefix := Prefix{
		Code: start.Attrs().Get("Code").Value(),
	}
loop:
	for r.Next() {
		switch el := r.Element().(type) {
		case *xml.StartElement:
			switch el.Name() {
			case "name":
				r.AssignNext(&prefix.Name)
			case "printSymbol":
				r.AssignNext(&prefix.Display)
			case "value":
				value := el.Attrs().Get("value").Value()
				prefix.Factor = decimal.RequireFromString(value)
			}
		case *xml.EndElement:
			if el.Name() == prefixEl {
				break loop
			}
		}
	}
	return prefix
}

func parseBaseUnit(start *xml.StartElement, r *xml.Reader) BaseUnit {
	baseUnit := BaseUnit{
		Code: start.Attrs().Get("Code").Value(),
	}
loop:
	for r.Next() {
		switch el := r.Element().(type) {
		case *xml.StartElement:
			switch el.Name() {
			case "name":
				r.AssignNext(&baseUnit.Name)
			case "printSymbol":
				r.AssignNext(&baseUnit.Display)
			case "property":
				r.AssignNext(&baseUnit.Property)
			}
		case *xml.EndElement:
			if el.Name() == baseUnitEl {
				break loop
			}
		}
	}
	return baseUnit
}

func parseUnit(start *xml.StartElement, r *xml.Reader) Unit {
	unit := Unit{
		Code:     start.Attrs().Get("Code").Value(),
		IsMetric: start.Attrs().Get("isMetric").Value() == "yes",
	}
loop:
	for r.Next() {
		switch el := r.Element().(type) {
		case *xml.StartElement:
			switch el.Name() {
			case "name":
				r.AssignNext(&unit.Name)
			case "printSymbol":
				r.AssignNext(&unit.Display)
			case "property":
				r.AssignNext(&unit.Property)
			case "value":
				unitExpr := el.Attrs().Get("Unit").Value()
				value := el.Attrs().Get("value").Value()

				unit.Unit = unitExpr
				unit.Value = decimal.RequireFromString(value)
			}
		case *xml.EndElement:
			if el.Name() == unitEl {
				break loop
			}
		}
	}
	return unit
}

var UCUM = NewUnitConverter(definitionsData)
