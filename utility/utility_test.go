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
