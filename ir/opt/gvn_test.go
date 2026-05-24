package opt_test

import (
	"testing"

	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/opt"
)

func TestGVN_Basic(t *testing.T) {
	// Let's build a function:
	//   %1 = iconst 10
	//   %2 = iconst 20
	//   %3 = iadd %1, %2
	//   %4 = iadd %1, %2   <-- redundant, should be mapped to %3
	//   ret %4
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 10})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 20})
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 4, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 4})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	gvn := &opt.GVNPass{}
	changed := gvn.Run(mod)

	if !changed {
		t.Fatal("expected GVN to optimize redundant iadd")
	}

	// The second iadd (dest 4) should be replaced with OpCopy from %3
	foundCopy := false
	for _, inst := range mod.Funcs[0].Insts {
		if inst.Dest == 4 {
			if inst.Opcode == air.OpCopy && inst.Src1 == 3 {
				foundCopy = true
			}
		}
	}

	if !foundCopy {
		t.Error("expected redundant instruction to be rewritten to a copy of %3")
	}
}

func TestGVN_MutableOperandSkipped(t *testing.T) {
	// Let's build a function with a mutated register:
	//   %1 = iconst 10
	//   %1 = copy %2       <-- %1 defined twice (defCounts > 1)
	//   %3 = iadd %1, %9
	//   %4 = iadd %1, %9   <-- GVN should skip because %1 is mutated
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 10})
	fb.Emit(air.AirInst{Opcode: air.OpCopy, TypeID: 3, Dest: 1, Src1: 2})
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 9})
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 4, Src1: 1, Src2: 9})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 4})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	gvn := &opt.GVNPass{}
	changed := gvn.Run(mod)

	if changed {
		t.Fatal("expected GVN to skip optimization when operand is mutated")
	}
}

func TestGVN_IntegrationWithCopyPropAndDCE(t *testing.T) {
	// Let's run GVN + CopyProp + DCE together:
	//   %1 = iconst 10
	//   %2 = iconst 20
	//   %3 = iadd %1, %2
	//   %4 = iadd %1, %2   <-- GVN replaces with %4 = copy %3
	//   ret %4             <-- CopyProp replaces with ret %3
	//                      <-- DCE deletes %4 = copy %3
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 10})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 20})
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 4, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 4})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	// Run passes
	(&opt.GVNPass{}).Run(mod)
	(&opt.CopyPropagationPass{}).Run(mod)
	(&opt.DCEPass{}).Run(mod)

	// In the final instruction array:
	// - No instruction should have Dest == 4 (fully removed by DCE!)
	// - OpReturn should return %3
	found4 := false
	foundReturn3 := false

	for _, inst := range mod.Funcs[0].Insts {
		if inst.Opcode == air.OpNop {
			continue
		}
		if inst.Dest == 4 {
			found4 = true
		}
		if inst.Opcode == air.OpReturn && inst.Src1 == 3 {
			foundReturn3 = true
		}
	}

	if found4 {
		t.Error("expected redundant instruction %4 to be completely deleted by GVN + CopyProp + DCE")
	}
	if !foundReturn3 {
		t.Error("expected return instruction to use %3 instead of %4")
	}
}
