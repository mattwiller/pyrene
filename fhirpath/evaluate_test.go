package fhirpath_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/mattwiller/pyrene/fhirpath"
	"github.com/mattwiller/pyrene/system"
	"github.com/stretchr/testify/assert"
)

func TestSimpleEvaluate(t *testing.T) {
	assert := assert.New(t)
	ast, err := fhirpath.Parse([]byte(`'Willer'.replace('er', 'iams')`))
	assert.NoError(err, "Error parsing expression")

	result, err := fhirpath.Evaluate(ast, nil)
	assert.NoError(err, "Error evaluating expression")
	assert.Len(result, 1)
	assert.Equal(fhirpath.StringType, result[0].Type())
	assert.Equal(system.String("Williams"), result[0].PrimitiveValue())
}

var testResource = []byte(`{
	"resourceType": "Patient",
	"meta": {
		"profile": ["http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient"],
		"lastUpdated": "2023-12-10T2059:27-08:00"
	},
	"extension": [{
		"url": "http://hl7.org/fhir/us/core/StructureDefinition/us-core-sex",
		"valueCode": "asked-declined"
	},{
		"url": "http://hl7.org/fhir/us/core/StructureDefinition/us-core-genderIdentity",
		"valueCodeableConcept": {
			"coding": [{
				"system": "http://snomed.info/sct",
				"code": "446131000124102",
				"display": "Identifies as non-conforming gender (finding)"
			},{
				"system": "http://terminology.hl7.org/CodeSystem/v3-NullFlavor",
				"code": "OTH",
				"display": "other"
			}],
			"text": "Genderqueer"
		}
	}],
	"active": true,
	"gender": "unknown",
	"birthDate": "2000-01-01",
	"identifier": [{
		"system": "http://example.com/mrn",
		"value": "1"
	},{
		"system": "http://hl7.org/fhir/sid/us-ssn",
		"value": "999-99-9999"
	}],
	"name": [{
		"use": "official",
		"given": ["Alex", "Justice"],
		"family": "Chalmers"
	}],
	"generalPractitioner": [{
		"reference": "Practitioner/7575c453-7ee6-4e18-b502-309a1e5a9d28",
		"display": "Dr. Amber Johnson"
	}]
}`)

func TestResourceEvaluate(t *testing.T) {
	assert := assert.New(t)
	resource := fhirpath.ParseJSON(bytes.NewReader(testResource), "FHIR.Patient")
	assert.Equal(system.ValueType("FHIR.Patient"), resource.Type())
	ast, err := fhirpath.Parse([]byte(`name.family.replace('er', 'iams')`))
	assert.NoError(err, "Error parsing expression")

	result, err := fhirpath.Evaluate(ast, resource)
	assert.NoError(err, "Error evaluating expression")
	assert.Len(result, 1)
	assert.Equal(fhirpath.StringType, result[0].Type())
	assert.Equal(system.String("Chalmiamss"), result[0].PrimitiveValue())
}

var EvalBenchmarkResult fhirpath.Collection

func BenchmarkSimpleEvaluate(b *testing.B) {
	ast, _ := fhirpath.Parse([]byte(`'Willer'.replace('er', 'iams')`))
	for i := 0; i < b.N; i++ {
		EvalBenchmarkResult, _ = fhirpath.Evaluate(ast, nil)
	}
}

func BenchmarkFullParseAndEvaluate(b *testing.B) {
	input := bytes.NewReader(testResource)
	for i := 0; i < b.N; i++ {
		input.Seek(0, io.SeekStart)
		resource := fhirpath.ParseJSON(input, "FHIR.Patient")
		ast, _ := fhirpath.Parse([]byte(`name.family.replace('er', 'iams')`))
		EvalBenchmarkResult, _ = fhirpath.Evaluate(ast, resource)
		b.SetBytes(int64(len(testResource)))
	}
}
