package opt_test

import (
	"testing"

	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/opt"
)

func TestCTGC_FreeAllocReuse(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)
	ptrReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpAlloc, TypeID: 3, Dest: ptrReg})
	fb.Emit(air.AirInst{Opcode: air.OpFree, Src1: ptrReg})
	allocReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpAlloc, TypeID: 3, Dest: allocReg})
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	pass := &opt.CTGCOptPass{}
	changed := pass.Run(mod)

	if !changed {
		t.Fatal("expected CTGC to replace free+alloc with alias_reuse")
	}

	hasReuse := false
	hasFree := false
	for _, inst := range mod.Funcs[0].Insts {
		if inst.Opcode == air.OpAliasReuse {
			hasReuse = true
			if inst.Src1 != ptrReg {
				t.Errorf("alias_reuse should reference freed reg %d, got %d", ptrReg, inst.Src1)
			}
		}
		if inst.Opcode == air.OpFree {
			hasFree = true
		}
	}
	if !hasReuse {
		t.Error("expected OpAliasReuse")
	}
	if hasFree {
		t.Error("OpFree should be NOPed")
	}
}

func TestCTGC_FreeAllocBlockedByUse(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)
	ptrReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpAlloc, TypeID: 3, Dest: ptrReg})
	fb.Emit(air.AirInst{Opcode: air.OpFree, Src1: ptrReg})
	// Use the freed register before the next alloc → blocks reuse
	derefReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpDeref, TypeID: 3, Dest: derefReg, Src1: ptrReg})
	allocReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpAlloc, TypeID: 3, Dest: allocReg})
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	pass := &opt.CTGCOptPass{}
	changed := pass.Run(mod)

	// The free+alloc should NOT be replaced because the freed reg is used in between
	for _, inst := range mod.Funcs[0].Insts {
		if inst.Opcode == air.OpAliasReuse {
			t.Error("should not create alias_reuse when freed reg is used in between")
		}
	}
	_ = changed
}

func TestCTGC_GenIdCheckElim(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)
	ptrReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpAlloc, TypeID: 3, Dest: ptrReg})
	refReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpMakeRef, TypeID: 3, Dest: refReg, Src1: ptrReg})
	derefReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpDeref, TypeID: 3, Dest: derefReg, Src1: refReg})
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	pass := &opt.CTGCOptPass{}
	changed := pass.Run(mod)

	if !changed {
		t.Fatal("expected CTGC to eliminate redundant gen_id check")
	}

	// OpDeref should be replaced with OpCopy
	hasDeref := false
	hasCopy := false
	for _, inst := range mod.Funcs[0].Insts {
		if inst.Opcode == air.OpDeref {
			hasDeref = true
		}
		if inst.Opcode == air.OpCopy && inst.Dest == derefReg && inst.Src1 == refReg {
			hasCopy = true
		}
	}
	if hasDeref {
		t.Error("OpDeref should be replaced with OpCopy")
	}
	if !hasCopy {
		t.Error("expected OpCopy replacing redundant OpDeref")
	}
}

func TestCTGC_GenIdCheckNotElimAfterEscape(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)
	ptrReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpAlloc, TypeID: 3, Dest: ptrReg})
	refReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpMakeRef, TypeID: 3, Dest: refReg, Src1: ptrReg})
	// Store might cause escape
	fb.Emit(air.AirInst{Opcode: air.OpStore, Src1: refReg, Src2: ptrReg})
	derefReg := fb.FreshReg()
	fb.Emit(air.AirInst{Opcode: air.OpDeref, TypeID: 3, Dest: derefReg, Src1: refReg})
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	pass := &opt.CTGCOptPass{}
	pass.Run(mod)

	// OpDeref should NOT be replaced because the ref may have escaped via store
	for _, inst := range mod.Funcs[0].Insts {
		if inst.Opcode == air.OpCopy && inst.Dest == derefReg {
			t.Error("OpDeref should NOT be eliminated after potential escape via store")
		}
	}
}

func TestCTGC_NoChangeOnCleanCode(t *testing.T) {
	mod := buildSimpleFunc(
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42},
	)

	pass := &opt.CTGCOptPass{}
	changed := pass.Run(mod)

	if changed {
		t.Error("no allocations means no CTGC changes")
	}
}
