package jsexpr_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/byte-power/jsexpr"
	"github.com/byte-power/jsexpr/ast"
	"github.com/byte-power/jsexpr/compiler"
	"github.com/byte-power/jsexpr/file"
	"github.com/byte-power/jsexpr/parser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleEval() {
	output, err := jsexpr.Eval("greet + name", map[string]interface{}{
		"greet": "Hello, ",
		"name":  "world!",
	})
	if err != nil {
		fmt.Printf("err: %v", err)
		return
	}

	fmt.Printf("%v", output)

	// Output: Hello, world!
}

func ExampleEval_runtime_error() {
	_, err := jsexpr.Eval(`map(1..3, {1 / (# - 3)})`, nil)
	fmt.Print(err)

	// Output: runtime error: integer divide by zero (1:14)
	//  | map(1..3, {1 / (# - 3)})
	//  | .............^
}

func ExampleCompile() {
	env := map[string]interface{}{
		"foo": 1,
		"bar": 99,
	}

	program, err := jsexpr.Compile("foo in 1..99 and bar in 1..99", jsexpr.Env(env))
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	output, err := jsexpr.Run(program, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("%v", output)

	// Output: true
}

func ExampleEnv() {
	type Segment struct {
		Origin string
	}
	type Passengers struct {
		Adults int
	}
	type Meta struct {
		Tags map[string]string
	}
	type Env struct {
		Meta
		Segments   []*Segment
		Passengers *Passengers
		Marker     string
	}

	code := `all(segments, {.origin == "MOW"}) && passengers.adults > 0 && tags["foo"] startsWith "bar"`

	program, err := jsexpr.Compile(code, jsexpr.Env(Env{}))
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	env := Env{
		Meta: Meta{
			Tags: map[string]string{
				"foo": "bar",
			},
		},
		Segments: []*Segment{
			{Origin: "MOW"},
		},
		Passengers: &Passengers{
			Adults: 2,
		},
		Marker: "test",
	}

	output, err := jsexpr.Run(program, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("%v", output)

	// Output: true
}

func ExampleAsBool() {
	env := map[string]int{
		"foo": 0,
	}

	program, err := jsexpr.Compile("foo >= 0", jsexpr.Env(env), jsexpr.AsBool())
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	output, err := jsexpr.Run(program, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("%v", output.(bool))

	// Output: true
}

func ExampleAsBool_error() {
	env := map[string]interface{}{
		"foo": 0,
	}

	_, err := jsexpr.Compile("foo + 42", jsexpr.Env(env), jsexpr.AsBool())

	fmt.Printf("%v", err)

	// Output: expected bool, but got int
}

func ExampleAsFloat64() {
	program, err := jsexpr.Compile("42", jsexpr.AsFloat64())
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	output, err := jsexpr.Run(program, nil)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("%v", output.(float64))

	// Output: 42
}

func ExampleAsFloat64_error() {
	_, err := jsexpr.Compile(`!!true`, jsexpr.AsFloat64())

	fmt.Printf("%v", err)

	// Output: expected float64, but got bool
}

func ExampleAsInt64() {
	env := map[string]interface{}{
		"rating": 5.5,
	}

	program, err := jsexpr.Compile("rating", jsexpr.Env(env), jsexpr.AsInt64())
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	output, err := jsexpr.Run(program, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("%v", output.(int64))

	// Output: 5
}

func ExampleOperator() {
	code := `
		now() > createdAt &&
		(now() - createdAt).hours() > 24
	`

	type Env struct {
		CreatedAt time.Time
		Now       func() time.Time
		Sub       func(a, b time.Time) time.Duration
		After     func(a, b time.Time) bool
	}

	options := []jsexpr.Option{
		jsexpr.Env(Env{}),
		jsexpr.Operator(">", "after"),
		jsexpr.Operator("-", "sub"),
	}

	program, err := jsexpr.Compile(code, options...)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	env := Env{
		CreatedAt: time.Date(2018, 7, 14, 0, 0, 0, 0, time.UTC),
		Now:       func() time.Time { return time.Now() },
		Sub:       func(a, b time.Time) time.Duration { return a.Sub(b) },
		After:     func(a, b time.Time) bool { return a.After(b) },
	}

	output, err := jsexpr.Run(program, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("%v", output)

	// Output: true
}

func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func ExampleConstExpr() {
	code := `[fib(5), fib(3+3), fib(dyn)]`

	env := map[string]interface{}{
		"fib": fib,
		"dyn": 0,
	}

	options := []jsexpr.Option{
		jsexpr.Env(env),
		jsexpr.ConstExpr("fib"), // Mark fib func as constant expression.
	}

	program, err := jsexpr.Compile(code, options...)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	// Only fib(5) and fib(6) calculated on Compile, fib(dyn) can be called at runtime.
	env["dyn"] = 7

	output, err := jsexpr.Run(program, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("%v\n", output)

	// Output: [5 8 13]
}

func ExampleAllowUndefinedVariables() {
	code := `name == nil ? "Hello, world!" : sprintf("Hello, %v!", name)`

	env := map[string]interface{}{
		"sprintf": fmt.Sprintf,
	}

	options := []jsexpr.Option{
		jsexpr.Env(env),
		jsexpr.AllowUndefinedVariables(), // Allow to use undefined variables.
	}

	program, err := jsexpr.Compile(code, options...)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	output, err := jsexpr.Run(program, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	fmt.Printf("%v\n", output)

	env["name"] = "you" // Define variables later on.

	output, err = jsexpr.Run(program, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	fmt.Printf("%v\n", output)

	// Output: Hello, world!
	// Hello, you!
}

func ExampleAllowUndefinedVariables_zero_value() {
	code := `name == "" ? foo + bar : foo + name`

	// If environment has different zero values, then undefined variables
	// will have it as default value.
	env := map[string]string{}

	options := []jsexpr.Option{
		jsexpr.Env(env),
		jsexpr.AllowUndefinedVariables(), // Allow to use undefined variables.
	}

	program, err := jsexpr.Compile(code, options...)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	env = map[string]string{
		"foo": "Hello, ",
		"bar": "world!",
	}

	output, err := jsexpr.Run(program, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	fmt.Printf("%v", output)

	// Output: Hello, world!
}

func ExampleAllowUndefinedVariables_zero_value_functions() {
	code := `words == "" ? split("foo,bar", ",") : split(words, ",")`

	// Env is map[string]string type on which methods are defined.
	env := mockMapStringStringEnv{}

	options := []jsexpr.Option{
		jsexpr.Env(env),
		jsexpr.AllowUndefinedVariables(), // Allow to use undefined variables.
	}

	program, err := jsexpr.Compile(code, options...)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	output, err := jsexpr.Run(program, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	fmt.Printf("%v", output)

	// Output: [foo bar]
}

func ExamplePatch() {
	/*
		type patcher struct{}

		func (p *patcher) Enter(_ *ast.Node) {}
		func (p *patcher) Exit(node *ast.Node) {
			switch n := (*node).(type) {
			case *ast.PropertyNode:
				ast.Patch(node, &ast.FunctionNode{
					Name:      "get",
					Arguments: []ast.Node{n.Node, &ast.StringNode{Value: n.Property}},
				})
			}
		}
	*/

	program, err := jsexpr.Compile(
		`greet.you.world + "!"`,
		jsexpr.Patch(&patcher{}),
	)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	env := map[string]interface{}{
		"greet": "Hello",
		"get": func(a, b string) string {
			return a + ", " + b
		},
	}

	output, err := jsexpr.Run(program, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	fmt.Printf("%v", output)

	// Output : Hello, you, world!
}

func TestOperator_struct(t *testing.T) {
	env := &mockEnv{
		BirthDay: time.Date(2017, time.October, 23, 18, 30, 0, 0, time.UTC),
	}

	code := `birthDay == "2017-10-23"`

	program, err := jsexpr.Compile(code, jsexpr.Env(&mockEnv{}), jsexpr.Operator("==", "dateEqual"))
	require.NoError(t, err)

	output, err := jsexpr.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, true, output)
}

func TestOperator_interface(t *testing.T) {
	env := &mockEnv{
		Ticket: &ticket{Price: 100},
	}

	code := `ticket == "$100" && "$100" == ticket && now != ticket && now == now`

	program, err := jsexpr.Compile(
		code,
		jsexpr.Env(&mockEnv{}),
		jsexpr.Operator("==", "stringerStringEqual", "stringStringerEqual", "stringerStringerEqual"),
		jsexpr.Operator("!=", "notStringerStringEqual", "notStringStringerEqual", "notStringerStringerEqual"),
	)
	require.NoError(t, err)

	output, err := jsexpr.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, true, output)
}

func TestExpr_readme_example(t *testing.T) {
	env := map[string]interface{}{
		"greet":   "Hello, %v!",
		"names":   []string{"world", "you"},
		"sprintf": fmt.Sprintf,
	}

	code := `sprintf(greet, names[0])`

	program, err := jsexpr.Compile(code, jsexpr.Env(env))
	require.NoError(t, err)

	output, err := jsexpr.Run(program, env)
	require.NoError(t, err)

	require.Equal(t, "Hello, world!", output)
}

func TestExpr(t *testing.T) {
	date := time.Date(2017, time.October, 23, 18, 30, 0, 0, time.UTC)
	env := &mockEnv{
		Any:     "any",
		Int:     0,
		Int32:   0,
		Int64:   0,
		Uint64:  0,
		Float64: 0,
		Bool:    true,
		String:  "string",
		Array:   []int{1, 2, 3, 4, 5},
		Ticket: &ticket{
			Price: 100,
		},
		Passengers: &passengers{
			Adults: 1,
		},
		Segments: []*segment{
			{Origin: "MOW", Destination: "LED"},
			{Origin: "LED", Destination: "MOW"},
		},
		BirthDay:      date,
		Now:           time.Now(),
		One:           1,
		Two:           2,
		Three:         3,
		MultiDimArray: [][]int{{1, 2, 3}, {1, 2, 3}},
		Sum: func(list []int) int {
			var ret int
			for _, el := range list {
				ret += el
			}
			return ret
		},
		Inc:    func(a int) int { return a + 1 },
		Nil:    nil,
		Tweets: []tweet{{"Oh My God!", date}, {"How you doin?", date}, {"Could I be wearing any more clothes?", date}},
	}

	tests := []struct {
		code string
		want interface{}
	}{
		{
			`ticket.string()`,
			`$100`,
		},
		{
			`ticket.priceDiv(25)`,
			4,
		},
		{
			`ticket.price`,
			100,
		},
		{
			`1`,
			int(1),
		},
		{
			`-.5`,
			float64(-.5),
		},
		{
			`true && false || false`,
			false,
		},
		{
			`int == 0 && int32 == 0 && int64 == 0 && float64 == 0 && bool && string == "string"`,
			true,
		},
		{
			`-int64 == 0`,
			true,
		},
		{
			`"a" != "b"`,
			true,
		},
		{
			`"a" != "b" || 1 == 2`,
			true,
		},
		{
			`int + 0`,
			0,
		},
		{
			`uint64 + 0`,
			int(0),
		},
		{
			`uint64 + int64`,
			int64(0),
		},
		{
			`int32 + int64`,
			int64(0),
		},
		{
			`float64 + 0`,
			float64(0),
		},
		{
			`0 + float64`,
			float64(0),
		},
		{
			`0 <= float64`,
			true,
		},
		{
			`float64 < 1`,
			true,
		},
		{
			`int < 1`,
			true,
		},
		{
			`2 + 2 == 4`,
			true,
		},
		{
			`8 % 3`,
			2,
		},
		{
			`2 ** 8`,
			float64(256),
		},
		{
			`-(2-5)**3-2/(+4-3)+-2`,
			float64(23),
		},
		{
			`"hello" + " " + "world"`,
			"hello world",
		},
		{
			`0 in -1..1 and 1 in 1..1`,
			true,
		},
		{
			`int in 0..1`,
			true,
		},
		{
			`int32 in 0..1`,
			true,
		},
		{
			`int64 in 0..1`,
			true,
		},
		{
			`1 in [1, 2, 3] && "foo" in {foo: 0, bar: 1} && "Price" in ticket`,
			true,
		},
		// {
		// 	`1 in [1.5] || 1 not in [1]`,
		// 	false,
		// },
		// {
		// 	`One in 0..1 && Two not in 0..1`,
		// 	true,
		// },
		{
			`int32 in [10, 20]`,
			false,
		},
		{
			`string matches "s.+"`,
			true,
		},
		{
			`string matches ("^" + string + "$")`,
			true,
		},
		{
			`"foobar" contains "bar"`,
			true,
		},
		{
			`"foobar" startsWith "foo"`,
			true,
		},
		{
			`"foobar" endsWith "bar"`,
			true,
		},
		{
			`(0..10)[5]`,
			5,
		},
		{
			`ticket.price`,
			100,
		},
		{
			`add(10, 5) + getInt()`,
			15,
		},
		{
			`len([1, 2, 3])`,
			3,
		},
		{
			`len([1, two, 3])`,
			3,
		},
		{
			`len(["hello", "world"])`,
			2,
		},
		{
			`len("hello, world")`,
			12,
		},
		{
			`len(array)`,
			5,
		},
		{
			`len({a: 1, b: 2, c: 2})`,
			3,
		},
		{
			`{foo: 0, bar: 1}`,
			map[string]interface{}{"foo": 0, "bar": 1},
		},
		{
			`{foo: 0, bar: 1}`,
			map[string]interface{}{"foo": 0, "bar": 1},
		},
		{
			`(true ? 0+1 : 2+3) + (false ? -1 : -2)`,
			-1,
		},
		{
			`filter(1..9, {# > 7})`,
			[]interface{}{8, 9},
		},
		{
			`map(1..3, {# * #})`,
			[]interface{}{1, 4, 9},
		},
		{
			`all(1..3, {# > 0})`,
			true,
		},
		{
			`none(1..3, {# == 0})`,
			true,
		},
		{
			`any([1,1,0,1], {# == 0})`,
			true,
		},
		{
			`one([1,1,0,1], {# == 0}) and not one([1,0,0,1], {# == 0})`,
			true,
		},
		{
			`count(1..30, {# % 3 == 0})`,
			10,
		},
		{
			`now.after(birthDay)`,
			true,
		},
		{
			`"a" < "b"`,
			true,
		},
		{
			`now.sub(now).string() == duration("0s").string()`,
			true,
		},
		{
			`8.5 * passengers.adults * len(segments)`,
			float64(17),
		},
		{
			`1 + 1`,
			2,
		},
		{
			`(one * two) * three == one * (two * three)`,
			true,
		},
		{
			`array[0]`,
			1,
		},
		{
			`sum(array)`,
			15,
		},
		{
			`array[0] < array[1]`,
			true,
		},
		{
			`sum(multiDimArray[0])`,
			6,
		},
		{
			`sum(multiDimArray[0]) + sum(multiDimArray[1])`,
			12,
		},
		{
			`inc(array[0] + array[1])`,
			4,
		},
		{
			`array[0] + array[1]`,
			3,
		},
		{
			`array[1:2]`,
			[]int{2},
		},
		{
			`array[0:5] == array`,
			true,
		},
		{
			`array[0:] == array`,
			true,
		},
		{
			`array[:5] == array`,
			true,
		},
		{
			`array[:] == array`,
			true,
		},
		{
			`1 + 2 + three`,
			6,
		},
		{
			`mapArg({foo: "bar"})`,
			"bar",
		},
		{
			`nilStruct`,
			(*time.Time)(nil),
		},
		{
			`0 == nil || "str" == nil || true == nil`,
			false,
		},
		{
			`variadic("head", 1, 2, 3)`,
			[]int{1, 2, 3},
		},
		{
			`variadic("empty")`,
			[]int{},
		},
		{
			`string[:]`,
			"string",
		},
		{
			`string[:3]`,
			"str",
		},
		{
			`string[:9]`,
			"string",
		},
		{
			`string[3:9]`,
			"ing",
		},
		{
			`string[7:9]`,
			"",
		},
		{
			`float(0)`,
			float64(0),
		},
		{
			`map(filter(tweets, {len(.text) > 10}), {format(.date)})`,
			[]interface{}{"23 Oct 17 18:30 UTC", "23 Oct 17 18:30 UTC"},
		},
		{
			`concat("a", 1, [])`,
			`a1[]`,
		},
	}

	for _, tt := range tests {
		program, err := jsexpr.Compile(tt.code, jsexpr.Env(&mockEnv{}))
		require.NoError(t, err, "compile error")

		got, err := jsexpr.Run(program, env)
		require.NoError(t, err, "execution error")

		assert.Equal(t, tt.want, got, tt.code)
	}

	for _, tt := range tests {
		program, err := jsexpr.Compile(tt.code, jsexpr.Optimize(false))
		require.NoError(t, err, "compile error")

		got, err := jsexpr.Run(program, env)
		require.NoError(t, err, "execution error")

		assert.Equal(t, tt.want, got, "unoptimized: "+tt.code)
	}

	for _, tt := range tests {
		got, err := jsexpr.Eval(tt.code, env)
		require.NoError(t, err, "eval error")

		assert.Equal(t, tt.want, got, "eval: "+tt.code)
	}
}

func TestExpr_eval_with_env(t *testing.T) {
	_, err := jsexpr.Eval("true", jsexpr.Env(map[string]interface{}{}))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "misused")
}

func TestExpr_fetch_from_func(t *testing.T) {
	_, err := jsexpr.Eval("foo.Value", map[string]interface{}{
		"foo": func() {},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot fetch Value from func()")
}

func TestExpr_map_default_values(t *testing.T) {
	env := map[string]interface{}{
		"foo": map[string]string{},
		"bar": map[string]*string{},
	}

	input := `foo['missing'] == '' && bar['missing'] == nil`

	program, err := jsexpr.Compile(input, jsexpr.Env(env))
	require.NoError(t, err)

	output, err := jsexpr.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, true, output)
}

func TestExpr_map_default_values_compile_check(t *testing.T) {
	tests := []struct {
		env   interface{}
		input string
	}{
		{
			mockMapStringStringEnv{"foo": "bar"},
			`Split(foo, sep)`,
		},
		{
			mockMapStringIntEnv{"foo": 1},
			`foo / bar`,
		},
	}
	for _, tt := range tests {
		_, err := jsexpr.Compile(tt.input, jsexpr.Env(tt.env), jsexpr.AllowUndefinedVariables())
		require.NoError(t, err)
	}
}

func TestExpr_calls_with_nil(t *testing.T) {
	env := map[string]interface{}{
		"equals": func(a, b interface{}) interface{} {
			assert.Nil(t, a, "a is not nil")
			assert.Nil(t, b, "b is not nil")
			return a == b
		},
		"is": is{},
	}

	p, err := jsexpr.Compile(
		"a == nil && equals(b, nil) && is.nil(c)",
		jsexpr.Env(env),
		jsexpr.Operator("==", "equals"),
		jsexpr.AllowUndefinedVariables(),
	)
	require.NoError(t, err)

	out, err := jsexpr.Run(p, env)
	require.NoError(t, err)
	require.Equal(t, true, out)
}

func TestExpr_call_floatarg_func_with_int(t *testing.T) {
	env := map[string]interface{}{
		"cnv": func(f float64) interface{} {
			return f
		},
	}
	for _, each := range []struct {
		input    string
		expected float64
	}{
		{"-1", -1.0},
		{"1+1", 2.0},
		{"+1", 1.0},
		{"1-1", 0.0},
		{"1/1", 1.0},
		{"1*1", 1.0},
	} {
		p, err := jsexpr.Compile(
			fmt.Sprintf("cnv(%s)", each.input),
			jsexpr.Env(env))
		require.NoError(t, err)

		out, err := jsexpr.Run(p, env)
		require.NoError(t, err)
		require.Equal(t, each.expected, out)
	}
}

func TestConstExpr_error(t *testing.T) {
	env := map[string]interface{}{
		"divide": func(a, b int) int { return a / b },
	}

	_, err := jsexpr.Compile(
		`1 + divide(1, 0)`,
		jsexpr.Env(env),
		jsexpr.ConstExpr("divide"),
	)
	require.Error(t, err)
	require.Equal(t, "compile error: integer divide by zero (1:5)\n | 1 + divide(1, 0)\n | ....^", err.Error())
}

func TestConstExpr_error_wrong_type(t *testing.T) {
	env := map[string]interface{}{
		"divide": 0,
	}

	_, err := jsexpr.Compile(
		`1 + divide(1, 0)`,
		jsexpr.Env(env),
		jsexpr.ConstExpr("divide"),
	)
	require.Error(t, err)
	require.Equal(t, "const expression \"divide\" must be a function", err.Error())
}

func TestConstExpr_error_no_env(t *testing.T) {
	_, err := jsexpr.Compile(
		`1 + divide(1, 0)`,
		jsexpr.ConstExpr("divide"),
	)
	require.Error(t, err)
	require.Equal(t, "no environment for const expression: divide", err.Error())
}

func TestPatch(t *testing.T) {
	program, err := jsexpr.Compile(
		`ticket == "$100" and "$90" != ticket + "0"`,
		jsexpr.Env(mockEnv{}),
		jsexpr.Patch(&stringerPatcher{}),
	)
	require.NoError(t, err)

	env := mockEnv{
		Ticket: &ticket{Price: 100},
	}
	output, err := jsexpr.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, true, output)
}

func TestPatch_length(t *testing.T) {
	program, err := jsexpr.Compile(
		`string.length == 5`,
		jsexpr.Env(mockEnv{}),
		jsexpr.Patch(&lengthPatcher{}),
	)
	require.NoError(t, err)

	env := mockEnv{String: "hello"}
	output, err := jsexpr.Run(program, env)
	require.NoError(t, err)
	require.Equal(t, true, output)
}

func TestCompile_exposed_error(t *testing.T) {
	_, err := jsexpr.Compile(`1 == true`)
	require.Error(t, err)

	fileError, ok := err.(*file.Error)
	require.True(t, ok, "error should be of type *file.Error")
	require.Equal(t, "invalid operation: == (mismatched types int and bool) (1:3)\n | 1 == true\n | ..^", fileError.Error())
	require.Equal(t, 2, fileError.Column)
	require.Equal(t, 1, fileError.Line)

	b, err := json.Marshal(err)
	require.NoError(t, err)
	require.Equal(t, `{"Line":1,"Column":2,"Message":"invalid operation: == (mismatched types int and bool)","Snippet":"\n | 1 == true\n | ..^"}`, string(b))
}

func TestAsBool_exposed_error_(t *testing.T) {
	_, err := jsexpr.Compile(`42`, jsexpr.AsBool())
	require.Error(t, err)

	_, ok := err.(*file.Error)
	require.False(t, ok, "error must not be of type *file.Error")
	require.Equal(t, "expected bool, but got int", err.Error())
}

func TestEval_exposed_error(t *testing.T) {
	_, err := jsexpr.Eval(`1/0`, nil)
	require.Error(t, err)

	fileError, ok := err.(*file.Error)
	require.True(t, ok, "error should be of type *file.Error")
	require.Equal(t, "runtime error: integer divide by zero (1:2)\n | 1/0\n | .^", fileError.Error())
	require.Equal(t, 1, fileError.Column)
	require.Equal(t, 1, fileError.Line)
}

func TestIssue105(t *testing.T) {
	type A struct {
		Field string
	}
	type B struct {
		Field int
	}
	type C struct {
		A
		B
	}
	type Env struct {
		C
	}

	code := `
		a.field == '' &&
		c.a.field == '' &&
		b.field == 0 &&
		c.b.field == 0
	`

	_, err := jsexpr.Compile(code, jsexpr.Env(Env{}))
	require.NoError(t, err)

	_, err = jsexpr.Compile(`field == ''`, jsexpr.Env(Env{}))
	require.Error(t, err)
	require.Contains(t, err.Error(), "ambiguous identifier field")
}

func TestIssue_nested_closures(t *testing.T) {
	code := `all(1..3, { all(1..3, { # > 0 }) and # > 0 })`

	program, err := jsexpr.Compile(code)
	require.NoError(t, err)

	output, err := jsexpr.Run(program, nil)
	require.NoError(t, err)
	require.True(t, output.(bool))
}

func TestIssue138(t *testing.T) {
	env := map[string]interface{}{}

	_, err := jsexpr.Compile(`1 / (1 - 1)`, jsexpr.Env(env))
	require.Error(t, err)
	require.Equal(t, "integer divide by zero (1:3)\n | 1 / (1 - 1)\n | ..^", err.Error())

	_, err = jsexpr.Compile(`1 % 0`, jsexpr.Env(env))
	require.Error(t, err)
}

//
// Mock types
//
type mockEnv struct {
	Any                  interface{}
	Int, One, Two, Three int
	Int32                int32
	Int64                int64
	Uint64               uint64
	Float64              float64
	Bool                 bool
	String               string
	Array                []int
	MultiDimArray        [][]int
	Sum                  func(list []int) int
	Inc                  func(int) int
	Ticket               *ticket
	Passengers           *passengers
	Segments             []*segment
	BirthDay             time.Time
	Now                  time.Time
	Nil                  interface{}
	NilStruct            *time.Time
	NilInt               *int
	NilSlice             []ticket
	Tweets               []tweet
}

func (e *mockEnv) GetInt() int {
	return e.Int
}

func (*mockEnv) Add(a, b int) int {
	return int(a + b)
}

func (*mockEnv) Duration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}

func (*mockEnv) MapArg(m map[string]interface{}) string {
	return m["foo"].(string)
}

func (*mockEnv) DateEqual(date time.Time, s string) bool {
	return date.Format("2006-01-02") == s
}

func (*mockEnv) StringerStringEqual(f fmt.Stringer, s string) bool {
	return f.String() == s
}

func (*mockEnv) StringStringerEqual(s string, f fmt.Stringer) bool {
	return s == f.String()
}

func (*mockEnv) StringerStringerEqual(f fmt.Stringer, g fmt.Stringer) bool {
	return f.String() == g.String()
}

func (*mockEnv) NotStringerStringEqual(f fmt.Stringer, s string) bool {
	return f.String() != s
}

func (*mockEnv) NotStringStringerEqual(s string, f fmt.Stringer) bool {
	return s != f.String()
}

func (*mockEnv) NotStringerStringerEqual(f fmt.Stringer, g fmt.Stringer) bool {
	return f.String() != g.String()
}

func (*mockEnv) Variadic(x string, xs ...int) []int {
	return xs
}

func (*mockEnv) Concat(list ...interface{}) string {
	out := ""
	for _, e := range list {
		out += fmt.Sprintf("%v", e)
	}
	return out
}

func (*mockEnv) Float(i interface{}) float64 {
	switch t := i.(type) {
	case int:
		return float64(t)
	case float64:
		return t
	default:
		panic("unexpected type")
	}
}

func (*mockEnv) Format(t time.Time) string { return t.Format(time.RFC822) }

type ticket struct {
	Price int
}

func (t *ticket) PriceDiv(p int) int {
	return t.Price / p
}

func (t *ticket) String() string {
	return fmt.Sprintf("$%v", t.Price)
}

type passengers struct {
	Adults   uint32
	Children uint32
	Infants  uint32
}

type segment struct {
	Origin      string
	Destination string
	Date        time.Time
}

type tweet struct {
	Text string
	Date time.Time
}

type mockMapStringStringEnv map[string]string

func (m mockMapStringStringEnv) Split(s, sep string) []string {
	return strings.Split(s, sep)
}

type mockMapStringIntEnv map[string]int

type is struct{}

func (is) Nil(a interface{}) bool {
	return a == nil
}

type patcher struct{}

func (p *patcher) Enter(_ *ast.Node) {}
func (p *patcher) Exit(node *ast.Node) {
	switch n := (*node).(type) {
	case *ast.PropertyNode:
		ast.Patch(node, &ast.FunctionNode{
			Name:      "get",
			Arguments: []ast.Node{n.Node, &ast.StringNode{Value: n.Property}},
		})
	}
}

var stringer = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

type stringerPatcher struct{}

func (p *stringerPatcher) Enter(_ *ast.Node) {}
func (p *stringerPatcher) Exit(node *ast.Node) {
	t := (*node).Type()
	if t == nil {
		return
	}
	if t.Implements(stringer) {
		ast.Patch(node, &ast.MethodNode{
			Node:   *node,
			Method: "string",
		})
	}

}

type lengthPatcher struct{}

func (p *lengthPatcher) Enter(_ *ast.Node) {}
func (p *lengthPatcher) Exit(node *ast.Node) {
	switch n := (*node).(type) {
	case *ast.PropertyNode:
		if n.Property == "length" {
			ast.Patch(node, &ast.BuiltinNode{
				Name:      "len",
				Arguments: []ast.Node{n.Node},
			})
		}
	}
}

// custom tests for debugging
type patcher1 struct{}

func (p *patcher1) Enter(_ *ast.Node) {}
func (p *patcher1) Exit(node *ast.Node) {
	n, ok := (*node).(*ast.IndexNode)
	if !ok {
		return
	}
	unary, ok := n.Index.(*ast.UnaryNode)
	if !ok {
		return
	}
	if unary.Operator == "-" {
		ast.Patch(&n.Index, &ast.BinaryNode{
			Operator: "-",
			Left:     &ast.BuiltinNode{Name: "len", Arguments: []ast.Node{n.Node}},
			Right:    unary.Node,
		})
	}
}

func Test_patcher1Index(t *testing.T) {
	env := map[string]interface{}{
		"list": []string{"1", "2", "3"},
		"a":    1,
	}

	code := `list[-a]` // will output 3

	program, err := jsexpr.Compile(code, jsexpr.Env(env), jsexpr.Patch(&patcher1{}))
	if err != nil {
		panic(err)
	}

	output, err := jsexpr.Run(program, env)
	if err != nil {
		panic(err)
	}
	fmt.Print(output)
	assert.Equal(t, "3", output)
}

// bytepower new feature test suite added below

func TestBytepowerExpr(t *testing.T) {
	type test struct {
		input    string
		expected interface{}
		env      interface{}
	}
	testPanda := &panda{
		Age: 10,
	}

	testKoala := &koala{
		Origin: "earth",
	}

	tests := []test{
		{
			`Date.now() == "test"`,
			true,
			bpMockEnv2{
				Date: dummy3{
					Now: func() string { return "test" },
				},
			},
		},
		{
			`Date.now() > 0`,
			true,
			nil,
		},
		{
			`Panda.age < 10`,
			true,
			bpMockEnv2{
				Panda: panda{
					Age: 8,
				},
			},
		},
		{
			`koala.origin == "earth"`,
			true,
			bpMockEnv{
				Koala: *testKoala,
			},
		},
		{
			`koala.HOWL()`,
			"fuck australia!",
			bpMockEnv{
				Koala: koala{
					Howl: func() string {
						return "fuck australia!"
					},
				},
			},
		},
		{
			`koala.Age < 10`,
			true,
			&bpMockEnv{
				Koala: koala{
					Age: 9,
				},
			},
		},
		{
			`panda.howl()`,
			"i'm from China",
			bpMockEnv{
				Panda: *testPanda,
			},
		},
		{
			`panda.age > 10`,
			false,
			&bpMockEnv{
				Panda: *testPanda,
			},
		},
		{
			`panda.age > 10`,
			false,
			&bpMockEnv{
				Panda: panda{
					Age: 8,
				},
			},
		},
		{
			`panda.age > 10`,
			true,
			bpMockEnv{
				Panda: panda{
					Age: 11,
				},
			},
		},
		{
			`panda.age`,
			10,
			bpMockEnv{
				Panda: panda{
					Age: 10,
				},
			},
		},
		{
			`parseInt("10", 16)`,
			16,
			nil,
		},
		{
			`parseInt("10")`,
			10,
			nil,
		},
		{
			`parseInt("10",16,1,1,1,1,"2","3")`,
			16,
			nil,
		},
		{
			`parseInt("10") < 10`,
			false,
			nil,
		},
		{
			`parseFloat(".5") < 1`,
			true,
			nil,
		},
		{
			`parseFloat(" 12.12.12 hey", 1, 3, 5)`,
			12.12,
			nil,
		},
		{
			`parseFloat(12.1, "ignored")`,
			12.1,
			nil,
		},
		{
			`pigat_get("player.level") + .5 < 555`,
			true,
			map[string]interface{}{
				"pigat_get": func(s string) int { return 1 },
			},
		},
		{
			`pigat_get("player.level") + .5 < 1`,
			false,
			map[string]interface{}{
				"pigat_get": func(s string) int { return 1 },
			},
		},
		{
			`1 < 2`,
			true,
			nil,
		},
		{
			`tracking.item.apple < 10`,
			false,
			map[string]interface{}{
				"tracking": mockTracking{},
			},
		},
		{ // this is traditional Object.Property evaluation, neither PropertyProvider not ValueProvider are implemented
			// to player Struct, thus, the `Level` property has to be capitalized as public-accessible property.
			`player.level < 10`,
			true,
			bpMockEnv{
				player{
					Level: 1,
				},
				panda{},
				koala{},
			},
		},
		{ // this is BP implemented PropertyProvider and ValueProvider, as player's object -- `mockPlayer` matched PropertyProvider
			// and level's object -- `mockLevel` matched ValueProvider, so that this expression could be evaluated
			`player.level < 10`,
			true,
			map[string]interface{}{"player": mockPlayer{}},
		},
		{
			`player.level.value < 10`,
			true,
			map[string]interface{}{"player": mockPlayer{}},
		},
	}

	for _, test := range tests {
		tree, err := parser.Parse(test.input)
		assert.Nil(t, err)

		program, err := compiler.Compile(tree, nil)
		assert.Nil(t, err)

		out, err := jsexpr.Run(program, test.env)
		assert.Nil(t, err)
		assert.Equal(t, test.expected, out)
	}
}

func TestBytepowerExprCompileError(t *testing.T) {
	type test struct {
		input string
		env   interface{}
	}
	tests := []test{
		{
			`pigat.player.level > 10`,
			map[string]interface{}{
				"pigat": mockPigat{},
			},
			// reason: from upon expression, player is not an accessible property of mockPigat{}, thus this is an error,
			// correct case: pass in a nil as env, instead of map
			// TODO: leverage the design in goja, this kind of abc.def.ghi could use json struct field to call references.
		},
	}
	for _, test := range tests {
		_, err := jsexpr.Compile(test.input, jsexpr.Env(test.env))
		assert.Error(t, err)
	}
}

type mockPigat struct{}

func (mP mockPigat) FetchProperty(property string) interface{} {
	return mockPlayer{}
}

func TestBytepowerExprRunError(t *testing.T) {
	type test struct {
		input string
		env   interface{}
	}
	tests := []test{
		{
			`noPropertyProvider.level < 10`,
			map[string]interface{}{
				"noPropertyProvider": dummy{},
			},
			// reason: dummy{} has no accessible property `level`
		},
		{
			`propertyProvider.level < 10`,
			map[string]interface{}{
				"propertyProvider": dummyPropertyProvider{},
			},
			// reason: the evaluated result of `propertyProvider.level` -- dummy{} -- neither has an evaluable operation
			// with 10, nor has implemented ValueProvider interface to provide a value
		},
		{
			`nestedPropertyProvider.propertyProvider < 10`,
			map[string]interface{}{
				"nestedPropertyProvider": nestedPropertyProvider{},
			},
			// reason: the evaluated result of `nestedPropertyProvider.propertyProvider` -- dummyPropertyProvider{} -- neither
			// has an evaluable operation with 10, nor has implemented ValueProvider interface to provide a value
		},
	}
	for _, test := range tests {
		prg, err := jsexpr.Compile(test.input, jsexpr.Env(nil))
		assert.Nil(t, err)

		_, err = jsexpr.Run(prg, test.env)
		assert.Error(t, err)
	}
}

type dummy struct{}

type dummyPropertyProvider struct{}

func (dPF dummyPropertyProvider) FetchProperty(property string) interface{} {
	return dummy{}
}

type nestedPropertyProvider struct{}

func (nPF nestedPropertyProvider) FetchProperty(property string) interface{} {
	return dummyPropertyProvider{}
}

type mockTracking struct{}

func (mT mockTracking) FetchProperty(property string) interface{} {
	return mockItem{}
}

type mockItem struct{}

func (mI mockItem) FetchProperty(property string) interface{} {
	return mockApple{}
}

func (mI mockItem) GetValue() interface{} {
	return "will be bypassed since I'm not leaf node of identifier"
}

type mockApple struct{}

func (mA mockApple) GetValue() interface{} {
	return 11
}

type mockPlayer struct{}

func (mP mockPlayer) FetchProperty(property string) interface{} {
	return mockLevel{}
}

type mockLevel struct {
	Value int
}

func (mL mockLevel) GetValue() interface{} {
	return 1
}

type player struct {
	Level int
}

type bpMockEnv struct {
	Player player
	Panda  panda
	Koala  koala
}

type panda struct {
	Age int
}

func (this panda) Howl() string {
	return "i'm from China"
}

type koala struct {
	Age    int           `jsexpr:"Age"`
	Howl   func() string `jsexpr:"HOWL"`
	Origin string
}

type bpMockEnv2 struct {
	Panda panda  `jsexpr:"Panda"`
	Date  dummy3 `jsexpr:"Date"`
}

type dummy3 struct {
	Now func() string
}
