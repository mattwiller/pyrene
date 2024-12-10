package fhirpath

import (
	"fmt"
	"io"
	"strings"

	"github.com/gammazero/radixtree"
	jsoniter "github.com/json-iterator/go"
	"github.com/mattwiller/pyrene/system"
)

const (
	BooleanType system.ValueType = "FHIR.boolean"
	StringType  system.ValueType = "FHIR.string"
	IntegerType system.ValueType = "FHIR.integer"
)

var primitiveTypes = map[system.ValueType]system.ValueType{
	system.StringType:  StringType,
	system.BooleanType: BooleanType,
	system.IntegerType: IntegerType,
}

type Value struct {
	primitive system.Value
	valueType system.ValueType
	data      *radixtree.Tree
}

func Primitive(value system.Value) Value {
	valueType, ok := primitiveTypes[value.Type()]
	if !ok {
		panic(fmt.Errorf("unsupported primitive type: %s", value.Type()))
	}
	return Value{
		valueType: valueType,
		primitive: value,
	}
}

func ParseJSON(json io.Reader, valueType system.ValueType) *Value {
	iter := jsoniter.Parse(jsoniter.ConfigDefault, json, 256)
	if iter.WhatIsNext() != jsoniter.ObjectValue {
		return nil
	}

	tree := radixtree.New()
	collectObject(iter, tree, "")
	return &Value{
		valueType: valueType,
		data:      tree,
	}
}

func collectObject(iter *jsoniter.Iterator, tree *radixtree.Tree, prefix string) {
	for field := iter.ReadObject(); field != ""; field = iter.ReadObject() {
		if field == "resourceType" {
			// Skip; should assign type somehow?
			_ = iter.ReadString()
			continue
		}
		collectValue(iter, tree, prefix+field)
	}
}

func collectValue(iter *jsoniter.Iterator, tree *radixtree.Tree, key string) {
	next := iter.WhatIsNext()
	switch next {
	case jsoniter.NilValue:
		// Skip; null values are omitted
		_ = iter.Read()
	case jsoniter.ObjectValue:
		collectObject(iter, tree, key+".")
	case jsoniter.BoolValue:
		tree.Put(key, system.Boolean(iter.ReadBool()))
	case jsoniter.StringValue:
		tree.Put(key, system.String(iter.ReadStringAsSlice()))
	case jsoniter.NumberValue:
		n := iter.ReadNumber()
		if i, err := n.Int64(); err == nil {
			tree.Put(key, system.Integer(i))
		} else {
			panic(fmt.Errorf("unhandled number type at %s", key))
		}
	case jsoniter.ArrayValue:
		for i := 0; iter.ReadArray(); i++ {
			collectValue(iter, tree, fmt.Sprintf(`%s[%d]`, key, i))
		}
	default:
		panic(fmt.Errorf("unhandled JSON type: %v", next))
	}
}

func (v *Value) Type() system.ValueType {
	return v.valueType
}

func (v *Value) String() string {
	if v.primitive != nil {
		return v.primitive.String()
	}
	props := make([]string, 0, v.data.Len())
	v.data.Walk("", func(key string, value any) bool {
		props = append(props, fmt.Sprintf("%s: %s", key, value.(system.Value)))
		return false
	})
	return fmt.Sprintf("[%s]: {\n%s\n}", v.valueType, strings.Join(props, "\n"))
}

func (v *Value) PrimitiveValue() system.Value {
	return v.primitive
}

func (v *Value) Get(key string) []Value {
	if v.data == nil {
		return nil
	} else if result, ok := v.data.Get(key); ok {
		value := result.(system.Value)
		primitiveWrapperType := primitiveTypes[value.Type()]
		if primitiveWrapperType == "" {
			panic(fmt.Errorf("unsupported primitive type: %s", value.Type()))
		}
		return []Value{{
			valueType: primitiveWrapperType,
			primitive: value,
		}}
	} else if key == "value" && v.primitive != nil && !strings.HasPrefix(string(v.valueType), "System.") {
		return []Value{{
			valueType: v.primitive.Type(),
			primitive: v.primitive,
		}}
	}

	var output []Value
	for i := 0; ; i++ {
		arrayPrefix := fmt.Sprintf(`%s[%d]`, key, i)
		tree := radixtree.New()
		v.data.Walk(arrayPrefix, func(wk string, wv any) bool {
			if subkey := strings.TrimPrefix(wk, arrayPrefix); len(subkey) > 1 && subkey[0] == '.' {
				tree.Put(subkey[1:], wv)
				return false
			} else {
				output = append(output, Primitive(wv.(system.Value)))
				return true
			}
		})
		if tree.Len() > 0 {
			output = append(output, Value{
				valueType: "???",
				data:      tree,
			})
		}
		if i == len(output) {
			break
		}
	}
	if len(output) > 0 {
		return output
	}

	tree := radixtree.New()
	prefix := key + "."
	v.data.Walk(prefix, func(wk string, wv any) bool {
		tree.Put(strings.TrimPrefix(wk, prefix), wv)
		return false
	})

	if tree.Len() == 0 {
		return nil
	}
	return []Value{{
		valueType: "???",
		data:      tree,
	}}
}
