package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInt(t *testing.T) {
	type test struct {
		number int
		radix  int
		expect int
	}
	tests := []test{
		{
			10,
			16,
			16,
		},
		{
			11,
			8,
			9,
		},
		{
			111,
			2,
			7,
		},
	}
	for _, test := range tests {
		actual := parseInt(test.number, test.radix)
		assert.Equal(t, test.expect, actual)
	}
}
