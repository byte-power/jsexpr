package vm_test

import (
	"strings"
	"testing"

	"github.com/byte-power/jsexpr/vm"
)

func TestProgram_Disassemble(t *testing.T) {
	for op := vm.OpPush; op < vm.OpEnd; op++ {
		program := vm.Program{
			Constants: []interface{}{true},
			Bytecode:  []byte{op},
		}
		d := program.Disassemble()
		if strings.Contains(d, "\t0x") {
			t.Errorf("cannot disassemble all opcodes")
		}
	}
}
