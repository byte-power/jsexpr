package vm

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/byte-power/jsexpr/builtin"
	"github.com/byte-power/jsexpr/file"
	"github.com/byte-power/jsexpr/utility"
)

var (
	MemoryBudget int = 1e6
)

func Run(program *Program, env interface{}) (interface{}, error) {
	if program == nil {
		return nil, fmt.Errorf("program is nil")
	}

	vm := VM{}
	vm.Init(program, env)
	return vm.Run(program, env)
}

type VM struct {
	stack     []interface{}
	constants []interface{}
	bytecode  []byte
	ip        int
	pp        int
	scopes    []Scope
	debug     bool
	step      chan struct{}
	curr      chan int
	memory    int
	limit     int

	builtinObjs  map[string]interface{}
	builtinFuncs map[string]builtin.JSFunc

	structCallIndex int
	structCallCache []string
	envStructMap    map[string]fieldReflection
}

func Debug() *VM {
	vm := &VM{
		debug: true,
		step:  make(chan struct{}, 0),
		curr:  make(chan int, 0),
	}
	return vm
}

func (vm *VM) initEnv(env interface{}) {
	if env == nil {
		return
	}

	v := reflect.ValueOf(env)
	kind := v.Kind()
	if kind == reflect.Ptr && reflect.Indirect(v).Kind() == reflect.Struct {
		v = reflect.Indirect(v)
		kind = v.Kind()
	}

	if kind != reflect.Struct {
		return
	}

	vm.structCallCache = make([]string, 1<<10)
	vm.envStructMap = structReflection(v)
}

func (vm *VM) Init(program *Program, env interface{}) {
	vm.reset(program)

	vm.limit = MemoryBudget
	vm.builtinFuncs = builtin.Funcs()
	vm.builtinObjs = builtin.Objs()
	vm.initEnv(env)
}

func (vm *VM) reset(program *Program) {
	vm.ip = 0
	vm.pp = 0
	vm.memory = 0

	if vm.stack == nil {
		vm.stack = make([]interface{}, 0, 2)
	} else {
		vm.stack = vm.stack[0:0]
	}

	if vm.scopes != nil {
		vm.scopes = vm.scopes[0:0]
	}

	vm.bytecode = program.Bytecode
	vm.constants = program.Constants
	vm.structCallIndex = 0
}

func (vm *VM) getCurrentStructMap() map[string]fieldReflection {
	calls := vm.structCallCache[:vm.structCallIndex]
	// first lookup from injected env using current cached calls
	current := vm.envStructMap
	for _, fieldName := range calls {
		current = current[fieldName].nestedFieldReflections
	}
	return current
}

func (vm *VM) getFieldFromStruct(v reflect.Value, f string) (interface{}, bool) {
	fType := v.Type()
	for i := 0; i < fType.NumField(); i++ {
		structField := fType.Field(i)
		if structField.Name == f || structField.Tag.Get(utility.StructTagKey) == f {
			return v.Field(i).Interface(), true
		}
	}
	return nil, false
}

func (vm *VM) fetch(from interface{}, i interface{}) interface{} {
	if from != nil {
		v := reflect.ValueOf(from)
		kind := v.Kind()

		// Structures can be access through a pointer or through a value, when they
		// are accessed through a pointer we don't want to copy them to a value.
		if kind == reflect.Ptr && reflect.Indirect(v).Kind() == reflect.Struct {
			v = reflect.Indirect(v)
			kind = v.Kind()
		}

		switch kind {

		case reflect.Array, reflect.Slice, reflect.String:
			value := v.Index(toInt(i))
			if value.IsValid() && value.CanInterface() {
				return value.Interface()
			}

		case reflect.Map:
			value := v.MapIndex(reflect.ValueOf(i))
			if value.IsValid() && value.CanInterface() {
				return value.Interface()
			}

		case reflect.Struct:
			if provider, ok := from.(PropertyProvider); ok {
				return provider.FetchProperty(reflect.ValueOf(i).String())
			}

			m := vm.getCurrentStructMap()
			if v, ok := m[i.(string)]; ok {
				vm.structCallCache[vm.structCallIndex] = i.(string)
				vm.structCallIndex++
				return v.value.Interface()
			}

			if value, ok := vm.getFieldFromStruct(v, i.(string)); ok {
				return value
			}
		}
	}

	// not found in passed-in env, so fetch from vm's env
	if value, ok := vm.builtinObjs[i.(string)]; ok {
		return value
	}

	// still not found
	panic(fmt.Sprintf("cannot fetch %v from %T", i, from))
}

func (vm *VM) fetchFn(from interface{}, name string, envCall bool) reflect.Value {
	if from != nil {
		v := reflect.ValueOf(from)
		t := reflect.TypeOf(from)
		for i := 0; i < v.NumMethod(); i++ {
			methodSig := t.Method(i)
			if utility.StrToLowerCamel(methodSig.Name) == name {
				return v.Method(i)
			}
		}

		d := v
		if v.Kind() == reflect.Ptr {
			d = v.Elem()
		}

		switch d.Kind() {
		case reflect.Map:
			value := d.MapIndex(reflect.ValueOf(name))
			if value.IsValid() && value.CanInterface() {
				return value.Elem()
			}
		case reflect.Struct:
			t = d.Type()
			for i := 0; i < t.NumField(); i++ {
				if tag := t.Field(i).Tag.Get(utility.StructTagKey); tag == name {
					return d.Field(i)
				}
			}
		}
	}

	// no luck from passed-in env, so fetch from vm's env
	// vmFuncs := reflect.ValueOf(vm.builtinFuncs)
	// value := vmFuncs.MapIndex(reflect.ValueOf(name))
	value := reflect.ValueOf(vm.builtinFuncs[name])
	if value.IsValid() && value.CanInterface() {
		return value
	}

	// also not in vm env, so panic
	panic(fmt.Sprintf(`cannot get "%v" from %T, also not found in vm's environment`, name, from))
}

func (vm *VM) getFuncParamsFromStack(call Call) []reflect.Value {
	in := make([]reflect.Value, call.Size)
	for i := call.Size - 1; i >= 0; i-- {
		param := vm.popThroughValueFetcher()
		if param == nil && reflect.TypeOf(param) == nil {
			// In case of nil value and nil type use this hack,
			// otherwise reflect.Call will panic on zero value.
			in[i] = reflect.ValueOf(&param).Elem()
		} else {
			in[i] = reflect.ValueOf(param)
		}
	}
	return in
}

func (vm *VM) callFunc(f reflect.Value, call Call, in []reflect.Value) []reflect.Value {
	fType := f.Type()
	numIn := fType.NumIn()
	var hasVariadic bool
	if call.Size > numIn && fType.IsVariadic() {
		input := utility.MakeVariadicFuncInput(fType.In(numIn-1).Elem().Kind(), in, numIn-1)
		in[numIn-1] = input
		hasVariadic = true
	}
	validParams := func() int {
		if call.Size > numIn {
			return numIn
		} else {
			return call.Size
		}
	}()
	return vm.call(f, in[:validParams], hasVariadic)
}

func (vm *VM) call(fn reflect.Value, input []reflect.Value, callVariadic bool) []reflect.Value {
	fType := fn.Type()

	if !callVariadic {
		castedInput := make([]reflect.Value, len(input))
		for i := 0; i < len(input); i++ {
			castedInput[i] = utility.ReflectCast(fType.In(i).Kind(), input[i])
		}
		return fn.Call(castedInput)
	} else {
		castedInput := make([]reflect.Value, fType.NumIn())
		for i := 0; i < fType.NumIn(); i++ {
			castedInput[i] = utility.ReflectCast(fType.In(i).Kind(), input[i])
		}
		return fn.CallSlice(castedInput)
	}

}

func (vm *VM) Run(program *Program, env interface{}) (out interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			f := &file.Error{
				Location: program.Locations[vm.pp],
				Message:  fmt.Sprintf("%v", r),
			}
			err = f.Bind(program.Source)
		}
	}()

	vm.reset(program)

	for vm.ip < len(vm.bytecode) {

		if vm.debug {
			<-vm.step
		}

		vm.pp = vm.ip
		vm.ip++
		op := vm.bytecode[vm.pp]

		switch op {

		case OpPush:
			vm.push(vm.constant())

		case OpPop:
			vm.pop()

		case OpRot:
			b := vm.pop()
			a := vm.pop()
			vm.push(b)
			vm.push(a)

		case OpFetch:
			vm.structCallIndex = 0
			vm.push(vm.fetch(env, vm.constant()))

		case OpFetchMap:
			vm.push(env.(map[string]interface{})[vm.constant().(string)])

		case OpTrue:
			vm.push(true)

		case OpFalse:
			vm.push(false)

		case OpNil:
			vm.push(nil)

		case OpNegate:
			v := negate(vm.popThroughValueFetcher())
			vm.push(v)

		case OpNot:
			v := vm.popThroughValueFetcher().(bool)
			vm.push(!v)

		case OpEqual:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(equal(a, b))

		case OpEqualInt:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(a.(int) == b.(int))

		case OpEqualString:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(a.(string) == b.(string))

		case OpJump:
			offset := vm.arg()
			vm.ip += int(offset)

		case OpJumpIfTrue:
			offset := vm.arg()
			if vm.current().(bool) {
				vm.ip += int(offset)
			}

		case OpJumpIfFalse:
			offset := vm.arg()
			if !vm.current().(bool) {
				vm.ip += int(offset)
			}

		case OpJumpBackward:
			offset := vm.arg()
			vm.ip -= int(offset)

		case OpIn:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(in(a, b))

		case OpLess:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(less(a, b))

		case OpMore:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(more(a, b))

		case OpLessOrEqual:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(lessOrEqual(a, b))

		case OpMoreOrEqual:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(moreOrEqual(a, b))

		case OpAdd:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(add(a, b))

		case OpSubtract:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(subtract(a, b))

		case OpMultiply:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(multiply(a, b))

		case OpDivide:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(divide(a, b))

		case OpModulo:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(modulo(a, b))

		case OpExponent:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(exponent(a, b))

		case OpRange:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			min := toInt(a)
			max := toInt(b)
			size := max - min + 1
			if vm.memory+size >= vm.limit {
				panic("memory budget exceeded")
			}
			vm.push(makeRange(min, max))
			vm.memory += size

		case OpMatches:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			match, err := regexp.MatchString(b.(string), a.(string))
			if err != nil {
				panic(err)
			}

			vm.push(match)

		case OpMatchesConst:
			a := vm.popThroughValueFetcher()
			r := vm.constant().(*regexp.Regexp)
			vm.push(r.MatchString(a.(string)))

		case OpContains:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(strings.Contains(a.(string), b.(string)))

		case OpStartsWith:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(strings.HasPrefix(a.(string), b.(string)))

		case OpEndsWith:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(strings.HasSuffix(a.(string), b.(string)))

		case OpIndex:
			b := vm.popThroughValueFetcher()
			a := vm.popThroughValueFetcher()
			vm.push(vm.fetch(a, b))

		case OpSlice:
			from := vm.popThroughValueFetcher()
			to := vm.popThroughValueFetcher()
			node := vm.popThroughValueFetcher()
			vm.push(slice(node, from, to))

		case OpProperty:
			a := vm.pop()
			b := vm.constant()
			vm.push(vm.fetch(a, b))

		case OpCall:
			call := vm.getCall()
			in := vm.getFuncParamsFromStack(call)
			f := vm.fetchFn(env, call.Name, true)
			out := vm.callFunc(f, call, in)
			vm.push(out[0].Interface())

		case OpCallFast:
			call := vm.getCall()
			in := make([]interface{}, call.Size)
			for i := call.Size - 1; i >= 0; i-- {
				in[i] = vm.popThroughValueFetcher()
			}
			fn := vm.fetchFn(env, call.Name, true).Interface()
			vm.push(fn.(func(...interface{}) interface{})(in...))

		case OpMethod:
			call := vm.getCall()
			in := vm.getFuncParamsFromStack(call)
			f := vm.fetchFn(vm.pop(), call.Name, false)
			out := vm.callFunc(f, call, in)
			vm.push(out[0].Interface())

		case OpArray:
			size := vm.pop().(int)
			array := make([]interface{}, size)
			for i := size - 1; i >= 0; i-- {
				array[i] = vm.popThroughValueFetcher()
			}
			vm.push(array)
			vm.memory += size
			if vm.memory >= vm.limit {
				panic("memory budget exceeded")
			}

		case OpMap:
			size := vm.pop().(int)
			m := make(map[string]interface{})
			for i := size - 1; i >= 0; i-- {
				value := vm.popThroughValueFetcher()
				key := vm.popThroughValueFetcher()
				m[key.(string)] = value
			}
			vm.push(m)
			vm.memory += size
			if vm.memory >= vm.limit {
				panic("memory budget exceeded")
			}

		case OpLen:
			vm.push(length(vm.current()))

		case OpCast:
			t := vm.arg()
			switch t {
			case 0:
				vm.push(toInt64(vm.popThroughValueFetcher()))
			case 1:
				vm.push(toFloat64(vm.popThroughValueFetcher()))
			}

		case OpStore:
			scope := vm.Scope()
			key := vm.constant().(string)
			value := vm.popThroughValueFetcher()
			scope[key] = value

		case OpLoad:
			scope := vm.Scope()
			key := vm.constant().(string)
			vm.push(scope[key])

		case OpInc:
			scope := vm.Scope()
			key := vm.constant().(string)
			i := scope[key].(int)
			i++
			scope[key] = i

		case OpBegin:
			scope := make(Scope)
			vm.scopes = append(vm.scopes, scope)

		case OpEnd:
			vm.scopes = vm.scopes[:len(vm.scopes)-1]

		default:
			panic(fmt.Sprintf("unknown bytecode %#x", op))
		}

		if vm.debug {
			vm.curr <- vm.ip
		}
	}

	if vm.debug {
		close(vm.curr)
		close(vm.step)
	}

	if len(vm.stack) > 0 {
		return vm.pop(), nil
	}

	return nil, nil
}

func (vm *VM) push(value interface{}) {
	vm.stack = append(vm.stack, value)
}

func (vm *VM) current() interface{} {
	return vm.stack[len(vm.stack)-1]
}

func (vm *VM) pop() interface{} {
	value := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return value
}

func (vm *VM) popThroughValueFetcher() interface{} {
	v := vm.pop()
	if provider, ok := v.(ValueProvider); ok {
		return provider.GetValue()
	}
	return v
}

func (vm *VM) arg() uint16 {
	b0, b1 := vm.bytecode[vm.ip], vm.bytecode[vm.ip+1]
	vm.ip += 2
	return uint16(b0) | uint16(b1)<<8
}

func (vm *VM) constant() interface{} {
	return vm.constants[vm.arg()]
}

func (vm *VM) Stack() []interface{} {
	return vm.stack
}

func (vm *VM) Scope() Scope {
	if len(vm.scopes) > 0 {
		return vm.scopes[len(vm.scopes)-1]
	}
	return nil
}

func (vm *VM) Step() {
	if vm.ip < len(vm.bytecode) {
		vm.step <- struct{}{}
	}
}

func (vm *VM) Position() chan int {
	return vm.curr
}

func (vm *VM) getCall() Call {
	c := vm.constants[vm.arg()]
	switch call := c.(type) {
	case Call:
		return call
	case map[string]interface{}:
		return Call{
			Name: call["name"].(string),
			Size: AnyToInt(call["size"]),
		}
	default:
		panic(fmt.Sprintf("no call"))
	}
}

func AnyToInt(value interface{}) int {
	if value == nil {
		return 0
	}
	switch val := value.(type) {
	case int:
		return val
	case int8:
		return int(val)
	case int16:
		return int(val)
	case int32:
		return int(val)
	case int64:
		return int(val)
	case uint:
		return int(val)
	case uint8:
		return int(val)
	case uint16:
		return int(val)
	case uint32:
		return int(val)
	case uint64:
		return int(val)
	}
	return 0
}
