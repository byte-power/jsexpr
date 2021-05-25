package utility

import (
	"fmt"
	"strconv"
	"unicode"
)

func IntOutofAny(in interface{}) int {
	switch v := in.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)

	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)

	case float32:
		return int(v)
	case float64:
		return int(v)

	case string:
		runes := []rune(v)
		startIndex := 0
		endIndex := len(runes)
		var started bool
		for i := 0; i < len(runes); i++ {
			r := runes[i]
			if !unicode.IsDigit(r) {
				if started {
					endIndex = i
					break
				}
				if unicode.IsSpace(r) && !started {
					continue
				}
				if r == '+' || r == '-' {
					started = true
					startIndex = i
					continue
				}
				panic(fmt.Sprintf("cannot trim an integer out of string \"%s\"", v))
			}
			if !started {
				started = true
				startIndex = i
			}
		}
		number, err := strconv.Atoi(string(runes[startIndex:endIndex]))
		if err != nil {
			panic(fmt.Sprintf("cannot trim an integer out of string \"%s\"", v))
		}
		return number
	default:
		panic(fmt.Sprintf("doesn't know how to convert %s of type %T to number", v, v))
	}
}
