package fhirpath

import (
	"fmt"
	"strings"

	"github.com/mattwiller/pyrene/system"
)

type BuiltinFunc func(v *astVisitor, input Collection, params *Atom) (Collection, error)

var BuiltinFunctions = map[string]BuiltinFunc{
	"replace": func(v *astVisitor, input Collection, params *Atom) (Collection, error) {
		args, err := v.visitAtom(params)
		if err != nil {
			return nil, fmt.Errorf("error evaluating parameters: %w", err)
		}
		inputVal, err := input.Singleton(system.StringType)
		if err != nil {
			return nil, err
		}
		if len(input) == 0 || len(args) < 2 {
			return nil, nil
		}

		pattern, ok := args[0].PrimitiveValue().(system.String)
		if !ok {
			return nil, fmt.Errorf("expected String value for pattern parameter, got %s", args[0].Type())
		}
		substitution, ok := args[1].PrimitiveValue().(system.String)
		if !ok {
			return nil, fmt.Errorf("expected String value for substitution parameter, got %s", args[1].Type())
		}

		output := strings.ReplaceAll(string(inputVal.(system.String)), string(pattern), string(substitution))
		return Collection{Primitive(system.String(output))}, nil
	},
}
