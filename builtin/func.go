package builtin

import (
	"strconv"

	"github.com/byte-power/jsexpr/utility"
)

var funcs = map[string]JSFunc{
	"parseInt": jsParseInt,
}

func Funcs() map[string]JSFunc {
	return funcs
}

type JSFunc func(inputs ...interface{}) interface{}

func jsParseInt(inputs ...interface{}) interface{} {
	if len(inputs) == 0 {
		return nil
	}
	if len(inputs) == 1 {
		return parseInt(utility.IntOutofAny(inputs[0]), 10)
	}
	return parseInt(utility.IntOutofAny(inputs[0]), utility.IntOutofAny(inputs[1]))
}

func parseInt(number, radix int) int {
	out, err := strconv.ParseInt(strconv.Itoa(number), radix, 32)
	if err != nil {
		panic(err)
	}
	return int(out)
}
