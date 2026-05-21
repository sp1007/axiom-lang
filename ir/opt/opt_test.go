package opt_test

import (
	"testing"

	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/opt"
)

// --------------------------------------------------------------------------
// Pipeline Manager Tests
// --------------------------------------------------------------------------

func TestPipeline_O0_NoOp(t *testing.T) {
	p := opt.NewPipeline(opt.O0, false)
	mod := &air.AirModule{}
	stats := p.Run(mod)
	if stats.Iterations != 0 {
		t.Errorf("O0 should do 0 iterations, got %d", stats.Iterations)
	}
}

func TestPipeline_EmptyPasses(t *testing.T) {
	p := opt.NewPipeline(opt.O1, false)
	mod := &air.AirModule{}
	stats := p.Run(mod)
	if stats.Iterations != 0 {
		t.Errorf("empty pipeline should do 0 iterations, got %d", stats.Iterations)
	}
}

type noopPass struct{ name string }

func (p *noopPass) Name() string                  { return p.name }
func (p *noopPass) Run(_ *air.AirModule) bool { return false }

func TestPipeline_SinglePassNoChange(t *testing.T) {
	p := opt.NewPipeline(opt.O1, false)
	p.AddPass(&noopPass{"noop"})
	mod := &air.AirModule{}
	stats := p.Run(mod)
	if stats.Iterations != 1 {
		t.Errorf("expected 1 iteration, got %d", stats.Iterations)
	}
	if len(stats.Passes) != 1 {
		t.Errorf("expected 1 pass stat, got %d", len(stats.Passes))
	}
	if stats.Passes[0].PassName != "noop" {
		t.Errorf("expected pass name 'noop', got %q", stats.Passes[0].PassName)
	}
}

type countPass struct {
	count    int
	maxRuns  int
}

func (p *countPass) Name() string { return "counter" }
func (p *countPass) Run(_ *air.AirModule) bool {
	p.count++
	return p.count < p.maxRuns
}

func TestPipeline_Fixpoint(t *testing.T) {
	p := opt.NewPipeline(opt.O1, false)
	cp := &countPass{maxRuns: 3}
	p.AddPass(cp)
	mod := &air.AirModule{}
	stats := p.Run(mod)
	// Should run 3 iterations: pass returns true for first 2, false on 3rd
	if stats.Iterations != 3 {
		t.Errorf("expected 3 iterations (fixpoint at 3), got %d", stats.Iterations)
	}
	if cp.count != 3 {
		t.Errorf("expected pass called 3 times, got %d", cp.count)
	}
}

func TestPipeline_MaxIterations(t *testing.T) {
	p := opt.NewPipeline(opt.O1, false)
	p.SetMaxIterations(2)
	p.AddPass(&countPass{maxRuns: 100}) // always changes
	mod := &air.AirModule{}
	stats := p.Run(mod)
	if stats.Iterations != 2 {
		t.Errorf("expected max 2 iterations, got %d", stats.Iterations)
	}
}

func TestPipeline_WithVerify(t *testing.T) {
	p := opt.NewPipeline(opt.O1, true)
	p.AddPass(&noopPass{"verify-test"})
	mod := &air.AirModule{}
	stats := p.Run(mod)
	// Should complete without panic
	if stats.Level != opt.O1 {
		t.Errorf("expected level O1, got %s", stats.Level)
	}
}

func TestOptLevel_String(t *testing.T) {
	tests := []struct {
		level opt.OptLevel
		want  string
	}{
		{opt.O0, "O0"},
		{opt.O1, "O1"},
		{opt.O2, "O2"},
		{opt.O3, "O3"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("OptLevel(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestDefaultPipeline_O1(t *testing.T) {
	p := opt.DefaultPipeline(opt.O1, false)
	if p.Level() != opt.O1 {
		t.Errorf("expected O1, got %s", p.Level())
	}
}

// --------------------------------------------------------------------------
// Constant Folding Tests
// --------------------------------------------------------------------------

func buildSimpleFunc(insts ...air.AirInst) *air.AirModule {
	fb := air.NewAirFuncBuilder(1, 0)
	for _, inst := range insts {
		fb.Emit(inst)
	}
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()
	return &air.AirModule{Funcs: []air.AirFunc{*fn}}
}

func TestConstFold_Add(t *testing.T) {
	mod := buildSimpleFunc(
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 10},
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 20},
		air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2},
	)

	pass := &opt.ConstantFoldingPass{}
	changed := pass.Run(mod)

	if !changed {
		t.Fatal("expected constant folding to make changes")
	}

	// The IAdd should be replaced with IConst 30
	found := false
	for _, inst := range mod.Funcs[0].Insts {
		if inst.Opcode == air.OpIConst && inst.Dest == 3 && inst.Src1 == 30 {
			found = true
		}
	}
	if !found {
		t.Error("expected folded IConst 30 for 10 + 20")
	}
}

func TestConstFold_Sub(t *testing.T) {
	mod := buildSimpleFunc(
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 50},
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 8},
		air.AirInst{Opcode: air.OpISub, TypeID: 3, Dest: 3, Src1: 1, Src2: 2},
	)

	pass := &opt.ConstantFoldingPass{}
	pass.Run(mod)

	for _, inst := range mod.Funcs[0].Insts {
		if inst.Opcode == air.OpIConst && inst.Dest == 3 && inst.Src1 == 42 {
			return // success
		}
	}
	t.Error("expected folded IConst 42 for 50 - 8")
}

func TestConstFold_Mul(t *testing.T) {
	mod := buildSimpleFunc(
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 6},
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 7},
		air.AirInst{Opcode: air.OpIMul, TypeID: 3, Dest: 3, Src1: 1, Src2: 2},
	)

	pass := &opt.ConstantFoldingPass{}
	pass.Run(mod)

	for _, inst := range mod.Funcs[0].Insts {
		if inst.Opcode == air.OpIConst && inst.Dest == 3 && inst.Src1 == 42 {
			return
		}
	}
	t.Error("expected folded IConst 42 for 6 * 7")
}

func TestConstFold_DivByZero(t *testing.T) {
	mod := buildSimpleFunc(
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42},
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 0},
		air.AirInst{Opcode: air.OpIDiv, TypeID: 3, Dest: 3, Src1: 1, Src2: 2},
	)

	pass := &opt.ConstantFoldingPass{}
	pass.Run(mod)

	// Division by zero should NOT be folded
	for _, inst := range mod.Funcs[0].Insts {
		if inst.Dest == 3 && inst.Opcode == air.OpIDiv {
			return // good — not folded
		}
	}
	t.Error("division by zero should not be folded")
}

func TestConstFold_Comparison(t *testing.T) {
	mod := buildSimpleFunc(
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 10},
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 20},
		air.AirInst{Opcode: air.OpLt, TypeID: 11, Dest: 3, Src1: 1, Src2: 2},
	)

	pass := &opt.ConstantFoldingPass{}
	pass.Run(mod)

	for _, inst := range mod.Funcs[0].Insts {
		if inst.Opcode == air.OpIConst && inst.Dest == 3 && inst.Src1 == 1 {
			return // 10 < 20 → true (1)
		}
	}
	t.Error("expected folded comparison 10 < 20 → 1")
}

func TestConstFold_Neg(t *testing.T) {
	mod := buildSimpleFunc(
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42},
		air.AirInst{Opcode: air.OpNeg, TypeID: 3, Dest: 2, Src1: 1},
	)

	pass := &opt.ConstantFoldingPass{}
	pass.Run(mod)

	for _, inst := range mod.Funcs[0].Insts {
		if inst.Opcode == air.OpIConst && inst.Dest == 2 && int32(inst.Src1) == -42 {
			return
		}
	}
	t.Error("expected folded -42")
}

func TestConstFold_ChainedFold(t *testing.T) {
	// 3 + 4 = 7, then 7 * 6 = 42
	mod := buildSimpleFunc(
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 3},
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 4},
		air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2},
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 4, Src1: 6},
		air.AirInst{Opcode: air.OpIMul, TypeID: 3, Dest: 5, Src1: 3, Src2: 4},
	)

	pass := &opt.ConstantFoldingPass{}
	pass.Run(mod)

	// After first fold: %3 = iconst 7, after second fold: %5 = iconst 42
	found7 := false
	found42 := false
	for _, inst := range mod.Funcs[0].Insts {
		if inst.Opcode == air.OpIConst && inst.Dest == 3 && inst.Src1 == 7 {
			found7 = true
		}
		if inst.Opcode == air.OpIConst && inst.Dest == 5 && inst.Src1 == 42 {
			found42 = true
		}
	}
	if !found7 {
		t.Error("expected intermediate fold 3 + 4 = 7")
	}
	if !found42 {
		t.Error("expected chained fold 7 * 6 = 42")
	}
}

func TestConstFold_NoChange(t *testing.T) {
	// Non-constant operand — should not fold
	mod := buildSimpleFunc(
		air.AirInst{Opcode: air.OpCopy, TypeID: 3, Dest: 1, Src1: 0},
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 10},
		air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2},
	)

	pass := &opt.ConstantFoldingPass{}
	changed := pass.Run(mod)

	if changed {
		t.Error("should not fold when one operand is non-constant")
	}
}

// --------------------------------------------------------------------------
// DCE Tests
// --------------------------------------------------------------------------

func TestDCE_UnusedConst(t *testing.T) {
	mod := buildSimpleFunc(
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42},
		air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 100}, // unused
	)

	pass := &opt.DCEPass{}
	changed := pass.Run(mod)

	if !changed {
		t.Fatal("expected DCE to eliminate unused constant")
	}

	// %1 is used by return (Src1=0 means void return, but let's check if %2 is NOPed)
	for _, inst := range mod.Funcs[0].Insts {
		if inst.Dest == 2 && inst.Opcode != air.OpNop {
			t.Error("expected unused %2 to be NOPed")
		}
	}
}

func TestDCE_UsedConst(t *testing.T) {
	// %1 is used by return
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 1})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	pass := &opt.DCEPass{}
	changed := pass.Run(mod)

	if changed {
		t.Error("should not eliminate used constant")
	}
}

func TestDCE_SideEffectPreserved(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)
	fb.Emit(air.AirInst{Opcode: air.OpCall, Dest: 1, Src1: 99}) // call has side effects
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	pass := &opt.DCEPass{}
	changed := pass.Run(mod)

	// Call should NOT be eliminated even though result is unused
	if changed {
		t.Error("DCE should not eliminate side-effectful instructions")
	}
}

func TestDCE_StorePreserved(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 0)
	fb.Emit(air.AirInst{Opcode: air.OpStore, Dest: 1, Src1: 2, Src2: 3})
	fb.Emit(air.AirInst{Opcode: air.OpReturn})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	pass := &opt.DCEPass{}
	changed := pass.Run(mod)

	if changed {
		t.Error("DCE should not eliminate stores")
	}
}

// --------------------------------------------------------------------------
// Integration Tests: Pipeline + ConstFold + DCE
// --------------------------------------------------------------------------

func TestPipeline_ConstFoldThenDCE(t *testing.T) {
	// 10 + 20 = 30, then returned
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 10})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 20})
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 3})
	fn := fb.Build()
	mod := &air.AirModule{Funcs: []air.AirFunc{*fn}}

	p := opt.DefaultPipeline(opt.O1, true) // verify enabled
	stats := p.Run(mod)

	if stats.Iterations < 1 {
		t.Error("expected at least 1 iteration")
	}

	// After fold + DCE: %3 should be iconst 30, %1 and %2 should be NOPed
	hasConst30 := false
	for _, inst := range mod.Funcs[0].Insts {
		if inst.Opcode == air.OpIConst && inst.Dest == 3 && inst.Src1 == 30 {
			hasConst30 = true
		}
	}
	if !hasConst30 {
		t.Error("expected folded constant 30")
	}
}

func TestPipeline_EmptyModule(t *testing.T) {
	mod := &air.AirModule{}
	p := opt.DefaultPipeline(opt.O1, true)
	stats := p.Run(mod)
	// Should not crash
	if stats.Level != opt.O1 {
		t.Errorf("expected O1, got %s", stats.Level)
	}
}
