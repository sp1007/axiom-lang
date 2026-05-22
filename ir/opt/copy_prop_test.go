package opt_test

import (
	"testing"

	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/opt"
)

func TestCopyProp_Basic(t *testing.T) {
	// Let's build a function:
	//   %1 = iconst 42
	//   %2 = copy %1
	//   %3 = iadd %2, %2
	//   ret %3
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42})
	fb.Emit(air.AirInst{Opcode: air.OpCopy, TypeID: 3, Dest: 2, Src1: 1})
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 2, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 3})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	pass := &opt.CopyPropagationPass{}
	changed := pass.Run(mod)

	if !changed {
		t.Fatal("expected copy propagation to optimize copies")
	}

	// In IAdd, Src1 and Src2 should now be %1 instead of %2
	foundFolded := false
	for _, inst := range mod.Funcs[0].Insts {
		if inst.Opcode == air.OpIAdd {
			if inst.Src1 == 1 && inst.Src2 == 1 {
				foundFolded = true
			}
		}
	}

	if !foundFolded {
		t.Error("expected IAdd operands to be propagated to %1")
	}
}

func TestCopyProp_CallArgs(t *testing.T) {
	// Let's build a function:
	//   %1 = iconst 100
	//   %2 = copy %1
	//   %3 = call %99(%2)
	//   ret
	fb := air.NewAirFuncBuilder(1, 0)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 100})
	fb.Emit(air.AirInst{Opcode: air.OpCopy, TypeID: 3, Dest: 2, Src1: 1})

	// Setup call args in Extras
	argStart := fb.EmitExtra(1) // count
	fb.EmitExtra(2)            // arg1: %2

	fb.Emit(air.AirInst{Opcode: air.OpCall, Dest: 3, Src1: 99, Src2: argStart})
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	pass := &opt.CopyPropagationPass{}
	changed := pass.Run(mod)

	if !changed {
		t.Fatal("expected copy propagation to change call args")
	}

	// Extras arg should be resolved to 1 (%1)
	if mod.Funcs[0].Extras[argStart+1] != 1 {
		t.Errorf("expected call arg in Extras to be propagated to 1, got %d", mod.Funcs[0].Extras[argStart+1])
	}
}
