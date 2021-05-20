package compiler_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/byte-power/jsexpr"
	"github.com/byte-power/jsexpr/compiler"
	"github.com/byte-power/jsexpr/conf"
	"github.com/byte-power/jsexpr/parser"
	"github.com/byte-power/jsexpr/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

func TestCompile_debug(t *testing.T) {
	input := `false && true && true`

	tree, err := parser.Parse(input)
	require.NoError(t, err)

	_, err = compiler.Compile(tree, nil)
	require.NoError(t, err)
}

func TestCompile(t *testing.T) {
	type test struct {
		input   string
		program vm.Program
	}
	var tests = []test{
		{
			`pigat_get("player.level") + 1 < 10`,
			vm.Program{},
		},
		{
			`true && true || true`,
			vm.Program{
				Bytecode: []byte{
					vm.OpTrue,
					vm.OpJumpIfFalse, 2, 0,
					vm.OpPop,
					vm.OpTrue,
					vm.OpJumpIfTrue, 2, 0,
					vm.OpPop,
					vm.OpTrue,
				},
			},
		},
		{
			`65535`,
			vm.Program{
				Constants: []interface{}{
					int(math.MaxUint16),
				},
				Bytecode: []byte{
					vm.OpPush, 0, 0,
				},
			},
		},
		{
			`.5`,
			vm.Program{
				Constants: []interface{}{
					float64(.5),
				},
				Bytecode: []byte{
					vm.OpPush, 0, 0,
				},
			},
		},
		{
			`true`,
			vm.Program{
				Bytecode: []byte{
					vm.OpTrue,
				},
			},
		},
		{
			`Name`,
			vm.Program{
				Constants: []interface{}{
					"Name",
				},
				Bytecode: []byte{
					vm.OpFetch, 0, 0,
				},
			},
		},
		{
			`"string"`,
			vm.Program{
				Constants: []interface{}{
					"string",
				},
				Bytecode: []byte{
					vm.OpPush, 0, 0,
				},
			},
		},
		{
			`"string" == "string"`,
			vm.Program{
				Constants: []interface{}{
					"string",
				},
				Bytecode: []byte{
					vm.OpPush, 0, 0,
					vm.OpPush, 0, 0,
					vm.OpEqual,
				},
			},
		},
		{
			`1000000 == 1000000`,
			vm.Program{
				Constants: []interface{}{
					int64(1000000),
				},
				Bytecode: []byte{
					vm.OpPush, 0, 0,
					vm.OpPush, 0, 0,
					vm.OpEqual,
				},
			},
		},
		{
			`-1`,
			vm.Program{
				Constants: []interface{}{1},
				Bytecode: []byte{
					vm.OpPush, 0, 0,
					vm.OpNegate,
				},
			},
		},
	}

	for _, test := range tests {
		node, err := parser.Parse(test.input)
		require.NoError(t, err)

		program, err := compiler.Compile(node, nil)
		require.NoError(t, err, test.input)

		assert.Equal(t, test.program.Disassemble(), program.Disassemble(), test.input)
	}
}

func TestCompile_cast(t *testing.T) {
	input := `1`
	expected := &vm.Program{
		Constants: []interface{}{
			1,
		},
		Bytecode: []byte{
			vm.OpPush, 0, 0,
			vm.OpCast, 1, 0,
		},
	}

	tree, err := parser.Parse(input)
	require.NoError(t, err)

	program, err := compiler.Compile(tree, &conf.Config{Expect: reflect.Float64})
	require.NoError(t, err)

	assert.Equal(t, expected.Disassemble(), program.Disassemble())
}

func TestConstantsMarshallingAfterCompilation(t *testing.T) {
	type test struct {
		input    string
		expected interface{}
		env      interface{}
	}

	tests := []test{
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
		// {
		// 	`player.level.value + 1 < 10`,
		// 	false,
		// 	nil,
		// },
		{
			`tracking.item.apple < 10`,
			false,
			map[string]interface{}{
				"tracking": mockTracking{},
			},
		},
		{
			`Player.Level < 10`,
			true,
			mockEnv{
				player{
					Level: 1,
				},
			},
		},
		{
			`player.level < 10`,
			true,
			map[string]interface{}{"player": mockPlayer{}},
		},
	}

	for _, test := range tests {
		tree, err := parser.Parse(test.input)
		assert.Nil(t, err)

		program, err := compiler.Compile(tree, nil)
		assert.Nil(t, err)

		bytes, err := msgpack.Marshal(*program)
		assert.Nil(t, err)

		var newProgram vm.Program
		err = msgpack.Unmarshal(bytes, &newProgram)
		assert.Nil(t, err)
		//assert.Equal(t, program, &newProgram)

		out, err := jsexpr.Run(program, test.env)
		assert.Nil(t, err)
		assert.Equal(t, test.expected, out)
	}
}

type mockTracking struct{}

func (mT mockTracking) FetchProperty(property string) interface{} {
	return mockItem{}
}

type mockItem struct{}

func (mI mockItem) FetchProperty(property string) interface{} {
	return mockApple{}
}

type mockApple struct{}

func (mA mockApple) GetValue() interface{} {
	return 11
}

type mockPlayer struct{}

func (mP mockPlayer) FetchProperty(property string) interface{} {
	return mockLevel{}
}

type mockLevel struct{}

func (mL mockLevel) GetValue() interface{} {
	return 1
}

type player struct {
	Level int
}

type mockEnv struct {
	Player player
}
