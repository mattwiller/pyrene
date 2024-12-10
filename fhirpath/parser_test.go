package fhirpath_test

import (
	"testing"

	"github.com/mattwiller/pyrene/fhirpath"
	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	assert := assert.New(t)
	ast, err := fhirpath.Parse([]byte("Patient.name.family.replace('er', 'iams')"))
	assert.NoError(err)
	expected := fhirpath.InvocationExpression(
		fhirpath.InvocationExpression(
			fhirpath.InvocationExpression(
				fhirpath.Identifier([]byte("Patient")),
				fhirpath.Identifier([]byte("name")),
			),
			fhirpath.Identifier([]byte("family")),
		),
		fhirpath.Function(
			fhirpath.Identifier([]byte("replace")),
			fhirpath.ParamList(fhirpath.Union(
				fhirpath.String([]byte("er")),
				fhirpath.String([]byte("iams")),
			)),
		),
	)
	assert.True(ast.Equals(expected), "AST does not have expected shape")
}

var ParserBenchmarkResult *fhirpath.Atom

func BenchmarkParser(b *testing.B) {
	input := []byte("Patient.name.family.replace('er', 'iams')")
	for i := 0; i < b.N; i++ {
		ParserBenchmarkResult, _ = fhirpath.Parse(input)
	}
}
