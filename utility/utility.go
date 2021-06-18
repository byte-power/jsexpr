package utility

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cast"
)

func IntOutofAny(in interface{}) int {
	if in == nil {
		panic("cannot convert nil to number")
	}
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
				if unicode.IsSpace(r) {
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

func FloatOutofAny(in interface{}) float64 {
	if in == nil {
		panic("cannot convert nil to number")
	}
	switch v := in.(type) {
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case string:
		runes := []rune(v)
		startIndex, endIndex := 0, len(runes)
		var started, dotted bool
		for i := 0; i < len(runes); i++ {
			r := runes[i]
			if !unicode.IsDigit(r) {
				if started {
					if r == '.' && !dotted {
						dotted = true
						continue
					}
					endIndex = i
					break
				}
				if unicode.IsSpace(r) {
					continue
				}
				if r == '.' {
					startIndex = i
					started = true
					dotted = true
					continue
				}
				if r == '+' || r == '-' {
					startIndex = i
					started = true
					continue
				}
				panic(fmt.Sprintf("cannot trim a float out of string \"%s\"", v))
			}
			if !started {
				startIndex = i
				started = true
			}
		}
		number, err := strconv.ParseFloat(string(runes[startIndex:endIndex]), 64)
		if err != nil {
			panic(fmt.Sprintf("cannot trim a float out of string \"%s\"", v))
		}
		return number
	default:
		panic(fmt.Sprintf("doesn't know how to convert %s of type %T to number", v, v))
	}
}

func StrToLowerCamel(str string) string {
	return strcase.ToLowerCamel(str)
}

func GetFieldTagName(field reflect.StructField) string {
	tag := field.Tag.Get(StructTagKey)

	if tag == "-" {
		return ""
	}

	if tag == "" {
		tag = StrToLowerCamel(field.Name)
	}
	return tag
}

func MakeVariadicFuncInput(elementKind reflect.Kind, inputArray []reflect.Value, variadicArgumentIndex int) reflect.Value {
	variadicArgumentSliceLength := len(inputArray) - variadicArgumentIndex
	switch elementKind {
	case reflect.Int:
		input := make([]int, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToInt(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)
	case reflect.Int64:
		input := make([]int64, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToInt64(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)
	case reflect.Int16:
		input := make([]int16, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToInt16(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)
	case reflect.Int32:
		input := make([]int32, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToInt32(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)
	case reflect.Int8:
		input := make([]int8, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToInt8(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)

	case reflect.Uint:
		input := make([]uint, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToUint(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)
	case reflect.Uint64:
		input := make([]uint64, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToUint64(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)
	case reflect.Uint16:
		input := make([]uint16, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToUint16(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)
	case reflect.Uint32:
		input := make([]uint32, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToUint32(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)
	case reflect.Uint8:
		input := make([]uint8, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToUint8(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)

	case reflect.Float32:
		input := make([]float32, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToFloat32(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)
	case reflect.Float64:
		input := make([]float64, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToFloat64(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)

	case reflect.Interface:
		input := make([]interface{}, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = inputArray[i].Interface()
			j++
		}
		return reflect.ValueOf(input)

	case reflect.String:
		input := make([]string, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToString(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)
	case reflect.Bool:
		input := make([]bool, variadicArgumentSliceLength)
		j := 0
		for i := variadicArgumentIndex; i < len(inputArray); i++ {
			input[j] = cast.ToBool(inputArray[i].Interface())
			j++
		}
		return reflect.ValueOf(input)
	default:
		panic(fmt.Sprintf("cannot make variadic argument of kind: %s", elementKind.String()))
	}
}

func ReflectCast(kind reflect.Kind, value reflect.Value) reflect.Value {
	switch kind {
	case reflect.Int:
		v := cast.ToInt(value.Interface())
		return reflect.ValueOf(v)
	case reflect.Int64:
		v := cast.ToInt64(value.Interface())
		return reflect.ValueOf(v)
	case reflect.Int32:
		v := cast.ToInt32(value.Interface())
		return reflect.ValueOf(v)
	case reflect.Int16:
		v := cast.ToInt16(value.Interface())
		return reflect.ValueOf(v)
	case reflect.Int8:
		v := cast.ToInt8(value.Interface())
		return reflect.ValueOf(v)

	case reflect.Uint:
		v := cast.ToUint(value.Interface())
		return reflect.ValueOf(v)
	case reflect.Uint64:
		v := cast.ToUint64(value.Interface())
		return reflect.ValueOf(v)
	case reflect.Uint32:
		v := cast.ToUint32(value.Interface())
		return reflect.ValueOf(v)
	case reflect.Uint16:
		v := cast.ToUint16(value.Interface())
		return reflect.ValueOf(v)
	case reflect.Uint8:
		v := cast.ToUint8(value.Interface())
		return reflect.ValueOf(v)

	case reflect.Float64:
		v := cast.ToFloat64(value.Interface())
		return reflect.ValueOf(v)
	case reflect.Float32:
		v := cast.ToFloat32(value.Interface())
		return reflect.ValueOf(v)
	case reflect.Interface:
		v := value.Interface()
		return reflect.ValueOf(v)
	case reflect.String:
		v := cast.ToString(value.Interface())
		return reflect.ValueOf(v)
	case reflect.Bool:
		v := cast.ToBool(value.Interface())
		return reflect.ValueOf(v)
	default:
		return value
	}
}
