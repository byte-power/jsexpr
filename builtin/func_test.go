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

func TestParseFloat(t *testing.T) {
	type test struct {
		numbers     []interface{}
		expected    float64
		expectPanic bool
	}

	tests := []test{
		{
			numbers: []interface{}{
				"-12.12",
				1,
				2,
				"whatsoever",
			},
			expected: -12.12,
		},
		{
			numbers: []interface{}{
				"not a number 12.456",
				"doesn't matter",
				999999,
			},
			expectPanic: true,
		},
		{
			numbers: []interface{}{
				"   -.5 not parsed, not matter",
			},
			expected: -0.5,
		},
	}
	for _, test := range tests {
		if test.expectPanic {
			assert.Panics(t, func() {
				jsParseFloat(test.numbers...)
			})
		} else {
			actual := jsParseFloat(test.numbers...)
			assert.Equal(t, test.expected, actual)
		}
	}
}
