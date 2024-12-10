package ucum

import (
	"testing"

	"github.com/mattwiller/pyrene/system"
	"github.com/stretchr/testify/assert"
)

func TestUCUM(t *testing.T) {
	assert.Len(t, UCUM.BaseUnits, 7)
	assert.Len(t, UCUM.Prefixes, 24)
	assert.Len(t, UCUM.Units, 248)
}

var UCUMParseBenchmarkResult UnitConverter

func BenchmarkUCUMParsing(b *testing.B) {
	for i := 0; i < b.N; i++ {
		UCUMParseBenchmarkResult = NewUnitConverter(definitionsData)
	}
}

func TestConvert(t *testing.T) {
	t.SkipNow()
	assert := assert.New(t)
	result, err := UCUM.Convert(system.NewQuantity(26.2, "mi"), "km")
	assert.NoError(err)
	assert.Equal(system.NewQuantity(42, "km"), result)
}
