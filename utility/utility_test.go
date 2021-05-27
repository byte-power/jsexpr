package utility

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntOutofAny(t *testing.T) {

	type test struct {
		input       interface{}
		expect      int
		expectPanic bool
	}
	tests := []test{
		{
			input:  "11",
			expect: 11,
		},
		{
			input:  " +11 ",
			expect: 11,
		},
		{
			input:  12.21,
			expect: 12,
		},
		{
			input:  int16(1),
			expect: 1,
		},
		{
			input:  "+11",
			expect: 11,
		},
		{
			input:       "+ ",
			expectPanic: true,
		},
		{
			input:  " -12 hahaha",
			expect: -12,
		},
		{
			input:       "nothing her e 123",
			expectPanic: true,
		},
		{
			input:  "12 34 56",
			expect: 12,
		},
	}
	for _, test := range tests {
		if test.expectPanic {
			assert.Panics(t, func() {
				IntOutofAny(test.input)
			})
		} else {
			actual := IntOutofAny(test.input)
			assert.Equal(t, test.expect, actual)
		}
	}
}

func TestFloatOutofAny(t *testing.T) {
	type test struct {
		input       interface{}
		expected    float64
		expectPanic bool
	}

	tests := []test{
		{
			input:    12,
			expected: 12,
		},
		{
			input:    `-12.1`,
			expected: -12.1,
		},
		{
			input:    `   .5 ababa`,
			expected: 0.5,
		},
		{
			input:       `- 0.5`,
			expectPanic: true,
		},
		{
			input:       ` no number -1`,
			expectPanic: true,
		},
		{
			input:    ` 1.1.1`,
			expected: 1.1,
		},
		{
			input:       ``,
			expectPanic: true,
		},
		{
			input:    `  +.5`,
			expected: 0.5,
		},
	}

	for _, test := range tests {
		if test.expectPanic {
			assert.Panics(t, func() {
				FloatOutofAny(test.input)
			})
		} else {
			actual := FloatOutofAny(test.input)
			assert.Equal(t, test.expected, actual)
		}
	}
}
