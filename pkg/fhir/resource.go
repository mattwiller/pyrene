package fhir

import "encoding/json"

type Resource interface {
	Value
	Data() json.RawMessage
}
