package opt_test

import (
	"testing"

	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/opt"
)

// --------------------------------------------------------------------------
// p10-t11: Differential Tests
//
// Verify that optimizations preserve program semantics by running the
// comptime interpreter on both unoptimized (O0) and optimized (O1/O2)
// AIR and comparing results.
// --------------------------------------------------------------------------

func buildDiffTestModule(insts []air.AirInst) (*air.AirModule, *air.AirModule) {
	// Build two identical modules
	build := func() *air.AirModule {
		fb := air.NewAirFuncBuilder(1, 3)
		for _, inst := range insts {
			fb.Emit(inst)
		}
		fn := fb.Build()
		return &air.AirModule{Funcs: []air.AirFunc{*fn}}
	}
	return build(), build()
}

func runInterp(t *testing.T, mod *air.AirModule) opt.Value {
	t.Helper()
	interp := opt.NewCompTimeInterpreter(mod)
	result, err := interp.Interpret(&mod.Funcs[0], nil)
	if err != nil {
		t.Fatalf("interpreter error: %v", err)
	}
	return result
}

// TestDiff_ConstFold_Preserves ensures constant folding produces
// the same result as unoptimized execution.
func TestDiff_ConstFold_Preserves(t *testing.T) {
	insts := []air.AirInst{
		{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 10},
		{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 20},
		{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2},
		{Opcode: air.OpReturn, Src1: 3},
	}

	o0, o1 := buildDiffTestModule(insts)

	// Run O0 (unoptimized)
	resultO0 := runInterp(t, o0)

	// Run O1 (optimized)
	pipeline := opt.DefaultPipeline(opt.O1, false)
	pipeline.Run(o1)
	resultO1 := runInterp(t, o1)

	if resultO0.IVal != resultO1.IVal {
		t.Errorf("O0 result (%d) != O1 result (%d)", resultO0.IVal, resultO1.IVal)
	}
	if resultO0.IVal != 30 {
		t.Errorf("expected 30, got %d", resultO0.IVal)
	}
}

// TestDiff_ChainedOps ensures chained constant folding is semantically correct.
func TestDiff_ChainedOps(t *testing.T) {
	insts := []air.AirInst{
		{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 2},
		{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 3},
		{Opcode: air.OpIMul, TypeID: 3, Dest: 3, Src1: 1, Src2: 2},
		{Opcode: air.OpIConst, TypeID: 3, Dest: 4, Src1: 7},
		{Opcode: air.OpIAdd, TypeID: 3, Dest: 5, Src1: 3, Src2: 4},
		{Opcode: air.OpReturn, Src1: 5},
	}

	o0, o1 := buildDiffTestModule(insts)
	resultO0 := runInterp(t, o0)

	pipeline := opt.DefaultPipeline(opt.O1, true)
	pipeline.Run(o1)
	resultO1 := runInterp(t, o1)

	if resultO0.IVal != resultO1.IVal {
		t.Errorf("O0 (%d) != O1 (%d)", resultO0.IVal, resultO1.IVal)
	}
	if resultO0.IVal != 13 {
		t.Errorf("expected 13 (2*3+7), got %d", resultO0.IVal)
	}
}

// TestDiff_Comparison ensures comparison ops are preserved.
func TestDiff_Comparison(t *testing.T) {
	insts := []air.AirInst{
		{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 5},
		{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 10},
		{Opcode: air.OpLt, TypeID: 11, Dest: 3, Src1: 1, Src2: 2},
		{Opcode: air.OpReturn, Src1: 3},
	}

	o0, o1 := buildDiffTestModule(insts)
	resultO0 := runInterp(t, o0)

	pipeline := opt.DefaultPipeline(opt.O1, false)
	pipeline.Run(o1)
	resultO1 := runInterp(t, o1)

	if resultO0.IVal != resultO1.IVal {
		t.Errorf("O0 (%d) != O1 (%d)", resultO0.IVal, resultO1.IVal)
	}
	if resultO0.IVal != 1 {
		t.Errorf("expected 1 (5 < 10 → true), got %d", resultO0.IVal)
	}
}

// TestDiff_Negation ensures unary ops are preserved.
func TestDiff_Negation(t *testing.T) {
	insts := []air.AirInst{
		{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42},
		{Opcode: air.OpNeg, TypeID: 3, Dest: 2, Src1: 1},
		{Opcode: air.OpReturn, Src1: 2},
	}

	o0, o1 := buildDiffTestModule(insts)
	resultO0 := runInterp(t, o0)

	pipeline := opt.DefaultPipeline(opt.O1, false)
	pipeline.Run(o1)
	resultO1 := runInterp(t, o1)

	if resultO0.IVal != resultO1.IVal {
		t.Errorf("O0 (%d) != O1 (%d)", resultO0.IVal, resultO1.IVal)
	}
	if resultO0.IVal != -42 {
		t.Errorf("expected -42, got %d", resultO0.IVal)
	}
}

// TestDiff_BitwiseOps ensures bitwise operations are preserved.
func TestDiff_BitwiseOps(t *testing.T) {
	insts := []air.AirInst{
		{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 0xFF00},
		{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 0x0FF0},
		{Opcode: air.OpAnd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2},
		{Opcode: air.OpReturn, Src1: 3},
	}

	o0, o1 := buildDiffTestModule(insts)
	resultO0 := runInterp(t, o0)

	pipeline := opt.DefaultPipeline(opt.O1, false)
	pipeline.Run(o1)
	resultO1 := runInterp(t, o1)

	if resultO0.IVal != resultO1.IVal {
		t.Errorf("O0 (%d) != O1 (%d)", resultO0.IVal, resultO1.IVal)
	}
	expected := int64(0xFF00 & 0x0FF0) // 0x0F00
	if resultO0.IVal != expected {
		t.Errorf("expected %d, got %d", expected, resultO0.IVal)
	}
}

// TestDiff_VoidReturn ensures void functions are handled.
func TestDiff_VoidReturn(t *testing.T) {
	insts := []air.AirInst{
		{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 99}, // unused
		{Opcode: air.OpReturn},
	}

	o0, o1 := buildDiffTestModule(insts)
	resultO0 := runInterp(t, o0)

	pipeline := opt.DefaultPipeline(opt.O1, false)
	pipeline.Run(o1)
	resultO1 := runInterp(t, o1)

	// Void return → IVal should be 0 for both
	if resultO0.IVal != resultO1.IVal {
		t.Errorf("O0 (%d) != O1 (%d)", resultO0.IVal, resultO1.IVal)
	}
}
