package fhirpath

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mattwiller/pyrene/system"
)

type Collection []Value

func (c Collection) Singleton(ofType system.ValueType) (system.Value, error) {
	if len(c) == 0 {
		return nil, nil
	} else if len(c) > 1 {
		return nil, fmt.Errorf("expected single %s, got collection of %d items", ofType, len(c))
	} else if !strings.HasPrefix(string(ofType), "System.") {
		return nil, fmt.Errorf("expected singleton argument to be System type, got %s", ofType)
	} else if c[0].PrimitiveValue() == nil {
		return nil, fmt.Errorf("expected singleton value, but found complex value (%s)", c[0].Type())
	}

	value := c[0].PrimitiveValue()
	if value.Type() != ofType {
		return nil, fmt.Errorf("expected %s, got %s", ofType, value.Type())
	}
	return value, nil
}

type astVisitor struct {
	functions map[string]BuiltinFunc
	input     Collection
}

func (v *astVisitor) visitAtom(atom *Atom) (Collection, error) {
	switch atom.atomType {
	case StringAtom:
		return Collection{Primitive(system.String(atom.token))}, nil
	case FunctionAtom:
		identifier, params := atom.left, atom.right
		if identifier.atomType != IdentifierAtom {
			return nil, fmt.Errorf("parse error: expected function identifier, got %s", identifier.atomType)
		} else if params.atomType != ParamListAtom {
			return nil, fmt.Errorf("parse error: expected function params, got %s", identifier.atomType)
		}

		fn, ok := v.functions[string(identifier.token)]
		if !ok {
			return nil, fmt.Errorf("function not found: %s", identifier.token)
		}
		return fn(v, v.input, params)
	case ParamListAtom:
		// Passthrough
		return v.visitAtom(atom.left)
	case UnionAtom:
		left, errL := v.visitAtom(atom.left)
		right, errR := v.visitAtom(atom.right)
		if errL != nil || errR != nil {
			return nil, fmt.Errorf("error evaluating Union: %w", errors.Join(errL, errR))
		}
		left = append(left, right...)
		return left, nil
	case InvocationExpressionAtom:
		left, err := v.visitAtom(atom.left)
		if err != nil {
			return nil, fmt.Errorf("error evaluating invocation: %w", err)
		}
		v.input = left

		result, err := v.visitAtom(atom.right)
		if err != nil {
			return nil, fmt.Errorf("error evaluating invocation: %w", err)
		}
		v.input = result
		return result, nil
	case IdentifierAtom:
		// Member invocation
		output := Collection{}
		for _, value := range v.input {
			result := value.Get(string(atom.token))
			output = append(output, result...)
		}
		return output, nil
	}
	return nil, fmt.Errorf("unhandled atom type: %s", atom.atomType)
}

func Evaluate(ast *Atom, input *Value) (Collection, error) {
	v := astVisitor{
		functions: BuiltinFunctions,
		input:     make(Collection, 0, 8),
	}
	if input != nil {
		v.input = append(v.input, *input)
	}
	return v.visitAtom(ast)
}
