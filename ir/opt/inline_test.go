package opt_test

import (
	"testing"

	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/opt"
)

func buildTwoFuncModule(calleeInsts []air.AirInst, callerInsts []air.AirInst, calleeNameID, callerNameID uint32) *air.AirModule {
	// Build callee
	cbf := air.NewAirFuncBuilder(calleeNameID, 0)
	for _, inst := range calleeInsts {
		cbf.Emit(inst)
	}
	callee := cbf.Build()

	// Build caller
	clf := air.NewAirFuncBuilder(callerNameID, 0)
	for _, inst := range callerInsts {
		clf.Emit(inst)
	}
	caller := clf.Build()

	return &air.AirModule{
		Funcs: []air.AirFunc{*callee, *caller},
	}
}

func TestInline_SmallFunc(t *testing.T) {
	// Callee: identity(x) = return x  (small: 1 instruction)
	calleeInsts := []air.AirInst{
		{Opcode: air.OpCopy, TypeID: 3, Dest: 1, Src1: 1},
		{Opcode: air.OpReturn, Src1: 1},
	}

	// Caller: calls callee
	callerInsts := []air.AirInst{
		{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42},
		{Opcode: air.OpCall, TypeID: 3, Dest: 2, Src1: 100}, // call callee NameID=100
		{Opcode: air.OpReturn, Src1: 2},
	}

	mod := buildTwoFuncModule(calleeInsts, callerInsts, 100, 200)

	pass := &opt.InliningPass{Threshold: 30}
	changed := pass.Run(mod)

	if !changed {
		t.Fatal("expected inlining to occur")
	}

	// After inlining, caller should not have OpCall
	caller := &mod.Funcs[1]
	for _, inst := range caller.Insts {
		if inst.Opcode == air.OpCall {
			t.Error("expected OpCall to be replaced after inlining")
		}
	}
}

func TestInline_RecursiveNotInlined(t *testing.T) {
	// Callee calls itself (recursive)
	calleeInsts := []air.AirInst{
		{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 1},
		{Opcode: air.OpCall, TypeID: 3, Dest: 2, Src1: 100}, // calls itself
		{Opcode: air.OpReturn, Src1: 2},
	}

	callerInsts := []air.AirInst{
		{Opcode: air.OpCall, TypeID: 3, Dest: 1, Src1: 100},
		{Opcode: air.OpReturn, Src1: 1},
	}

	mod := buildTwoFuncModule(calleeInsts, callerInsts, 100, 200)

	pass := &opt.InliningPass{Threshold: 30}
	changed := pass.Run(mod)

	if changed {
		t.Error("recursive function should not be inlined")
	}
}

func TestInline_LargeFuncNotInlined(t *testing.T) {
	// Create a callee with > 30 instructions
	var calleeInsts []air.AirInst
	for i := 0; i < 35; i++ {
		calleeInsts = append(calleeInsts, air.AirInst{
			Opcode: air.OpIConst, TypeID: 3, Dest: uint32(i + 1), Src1: uint32(i),
		})
	}
	calleeInsts = append(calleeInsts, air.AirInst{Opcode: air.OpReturn})

	callerInsts := []air.AirInst{
		{Opcode: air.OpCall, TypeID: 3, Dest: 1, Src1: 100},
		{Opcode: air.OpReturn, Src1: 1},
	}

	mod := buildTwoFuncModule(calleeInsts, callerInsts, 100, 200)

	pass := &opt.InliningPass{Threshold: 30}
	changed := pass.Run(mod)

	if changed {
		t.Error("large function should not be inlined")
	}
}

func TestInline_SelfCallNotInlined(t *testing.T) {
	// Caller calls itself
	callerInsts := []air.AirInst{
		{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 1},
		{Opcode: air.OpCall, TypeID: 3, Dest: 2, Src1: 200}, // calls itself
		{Opcode: air.OpReturn, Src1: 2},
	}

	mod := &air.AirModule{
		Funcs: []air.AirFunc{
			func() air.AirFunc {
				fb := air.NewAirFuncBuilder(200, 0)
				for _, inst := range callerInsts {
					fb.Emit(inst)
				}
				return *fb.Build()
			}(),
		},
	}

	pass := &opt.InliningPass{Threshold: 30}
	changed := pass.Run(mod)

	if changed {
		t.Error("self-call should not be inlined")
	}
}

func TestInline_ExternNotInlined(t *testing.T) {
	mod := &air.AirModule{
		Funcs: []air.AirFunc{
			{Name: 100, IsExtern: true, Insts: []air.AirInst{{Opcode: air.OpReturn}}},
			func() air.AirFunc {
				fb := air.NewAirFuncBuilder(200, 0)
				fb.Emit(air.AirInst{Opcode: air.OpCall, Dest: 1, Src1: 100})
				fb.Emit(air.AirInst{Opcode: air.OpReturn})
				return *fb.Build()
			}(),
		},
	}

	pass := &opt.InliningPass{Threshold: 30}
	changed := pass.Run(mod)

	if changed {
		t.Error("extern function should not be inlined")
	}
}

func TestInline_RegisterRemapping(t *testing.T) {
	// Callee uses regs 1, 2
	calleeInsts := []air.AirInst{
		{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 10},
		{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 20},
		{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2},
		{Opcode: air.OpReturn, Src1: 3},
	}

	// Caller uses regs 1, 2 — need to verify no collision after inlining
	callerInsts := []air.AirInst{
		{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 99},
		{Opcode: air.OpCall, TypeID: 3, Dest: 2, Src1: 100},
		{Opcode: air.OpReturn, Src1: 2},
	}

	mod := buildTwoFuncModule(calleeInsts, callerInsts, 100, 200)

	pass := &opt.InliningPass{Threshold: 30}
	changed := pass.Run(mod)

	if !changed {
		t.Fatal("expected inlining to occur")
	}

	// After inlining, verify no duplicate Dest values (simple SSA check)
	caller := &mod.Funcs[1]
	dests := make(map[uint32]int)
	for _, inst := range caller.Insts {
		if inst.Opcode != air.OpNop && inst.Dest != 0 {
			dests[inst.Dest]++
		}
	}

	for dest, count := range dests {
		if count > 1 {
			t.Errorf("register %%d%d defined %d times (SSA violation)", dest, count)
		}
	}
}

func TestInline_NoCallsNoChange(t *testing.T) {
	// Function with no calls
	mod := buildSimpleFunc(
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42},
	)

	pass := &opt.InliningPass{Threshold: 30}
	changed := pass.Run(mod)

	if changed {
		t.Error("function with no calls should not trigger inlining")
	}
}

func TestInline_AsyncNotInlined(t *testing.T) {
	mod := &air.AirModule{
		Funcs: []air.AirFunc{
			{Name: 100, IsAsync: true, Insts: []air.AirInst{
				{Opcode: air.OpIConst, Dest: 1, Src1: 42},
				{Opcode: air.OpReturn, Src1: 1},
			}},
			func() air.AirFunc {
				fb := air.NewAirFuncBuilder(200, 0)
				fb.Emit(air.AirInst{Opcode: air.OpCall, Dest: 1, Src1: 100})
				fb.Emit(air.AirInst{Opcode: air.OpReturn})
				return *fb.Build()
			}(),
		},
	}

	pass := &opt.InliningPass{Threshold: 30}
	changed := pass.Run(mod)

	if changed {
		t.Error("async function should not be inlined")
	}
}
