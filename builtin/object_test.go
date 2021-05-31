package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSHypotenuse(t *testing.T) {
	out := jsHypotenuse()
	assert.Equal(t, float64(0), out)

	out = jsHypotenuse(1)
	assert.Equal(t, float64(1), out)

	out = jsHypotenuse(3, 4)
	assert.Equal(t, float64(5), out)

	out = jsHypotenuse(1, 2, 2)
	assert.Equal(t, float64(3), out)

	out = jsHypotenuse(1, 1, 1, 1)
	assert.Equal(t, float64(2), out)
}
