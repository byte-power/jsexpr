package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchFn(t *testing.T) {
	env := map[string]interface{}{
		"Fn": func() string { return "hh" },
	}
	// env := envStruct{Fn: yoyo}
	v := FetchFn(env, "Fn")
	assert.Equal(t, "", v)
}

type envStruct struct {
	Fn func() (string, string)
}

func yoyo() (string, string) {
	return "kk", "ll"
}
