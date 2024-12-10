package fhirpath_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/mattwiller/pyrene/fhirpath"
	"github.com/mattwiller/pyrene/system"
	"github.com/stretchr/testify/assert"
)

func TestResourceFromJSON(t *testing.T) {
	assert := assert.New(t)

	value := fhirpath.ParseJSON(bytes.NewReader(testResource), "FHIR.Patient")
	assert.Equal(system.ValueType("FHIR.Patient"), value.Type())
	assert.Equal(system.String("unknown"), value.Get("gender")[0].PrimitiveValue())
	assert.Equal(system.String("Justice"), value.Get("name[0].given")[1].PrimitiveValue())

	name := value.Get("name[0]")[0]
	assert.Equal(system.String("Chalmers"), name.Get("family")[0].PrimitiveValue())

	identifiers := value.Get("identifier")
	assert.Len(identifiers, 2, "Expected two identifiers")
	assert.Equal(system.String("1"), identifiers[0].Get("value")[0].PrimitiveValue())
	assert.Equal(system.String("999-99-9999"), identifiers[1].Get("value")[0].PrimitiveValue())
}

var ResourceBenchmarkResult *fhirpath.Value

func BenchmarkResourceFromJSON(b *testing.B) {
	input := bytes.NewReader(testResource)
	for i := 0; i < b.N; i++ {
		ResourceBenchmarkResult = fhirpath.ParseJSON(input, "FHIR.Patient")
		b.SetBytes(int64(len(testResource)))
		input.Seek(0, io.SeekStart)
	}
}

func BenchmarkResourceGetProperty(b *testing.B) {
	input := bytes.NewReader(testResource)
	resource := fhirpath.ParseJSON(input, "FHIR.Patient")

	b.Run("single level", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := resource.Get("identifier")
			ResourceBenchmarkResult = &result[0]
		}
	})
	b.Run("nested", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := resource.Get("name[0].given")
			ResourceBenchmarkResult = &result[0]
		}
	})
}
