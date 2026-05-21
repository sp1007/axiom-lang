package air

import (
	"strings"
	"testing"
)

func TestPrintFunc_SimpleReturn(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	r1 := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r1, Src1: 42})
	r2 := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r2, Src1: 58})
	r3 := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIAdd, TypeID: 1, Dest: r3, Src1: r1, Src2: r2})
	b.Emit(AirInst{Opcode: OpReturn, Src1: r3})
	fn := b.Build()
	fn.Params = []uint32{1, 1}

	got := SprintFunc(fn)
	t.Logf("output:\n%s", got)

	expect := []string{
		"fn @1(t1, t1) -> t1:",
		"block_0:  ; entry exit",
		"%1: t1 = iconst",
		"%2: t1 = iconst",
		"%3: t1 = iadd %1, %2",
		"ret %3",
	}
	for _, s := range expect {
		if !strings.Contains(got, s) {
			t.Errorf("output missing %q", s)
		}
	}
}

func TestPrintFunc_VoidReturn(t *testing.T) {
	b := NewAirFuncBuilder(1, 0)
	b.Emit(AirInst{Opcode: OpReturn}) // void ret
	fn := b.Build()

	got := SprintFunc(fn)
	t.Logf("output:\n%s", got)

	if !strings.Contains(got, "ret\n") {
		t.Error("void return should print as 'ret'")
	}
	// Should NOT have "-> t0" in header since RetType=0
	if strings.Contains(got, "-> t0") {
		t.Error("RetType=0 should not print return type")
	}
}

func TestPrintFunc_NopSkipped(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	b.Emit(AirInst{Opcode: OpNop})
	b.Emit(AirInst{Opcode: OpReturn})
	fn := b.Build()

	got := SprintFunc(fn)
	t.Logf("output:\n%s", got)

	if strings.Contains(got, "nop") {
		t.Error("NOP should be skipped in output")
	}
}

func TestPrintFunc_JumpAndBranch(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	blk1 := b.NewBlock()
	blk2 := b.NewBlock()
	blk3 := b.NewBlock()

	r := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r, Src1: 1})
	b.Emit(AirInst{Opcode: OpBranch, Src1: r, Src2: blk1, Dest: blk2})
	b.AddEdge(0, blk1)
	b.AddEdge(0, blk2)

	b.SwitchTo(blk1)
	b.Emit(AirInst{Opcode: OpJump, Src1: blk3})
	b.AddEdge(blk1, blk3)

	b.SwitchTo(blk2)
	b.Emit(AirInst{Opcode: OpJump, Src1: blk3})
	b.AddEdge(blk2, blk3)

	b.SwitchTo(blk3)
	b.Emit(AirInst{Opcode: OpReturn})

	fn := b.Build()
	got := SprintFunc(fn)
	t.Logf("output:\n%s", got)

	expect := []string{
		"branch %1 block_1 block_2",
		"jump block_3",
		"block_3:  ; exit",
	}
	for _, s := range expect {
		if !strings.Contains(got, s) {
			t.Errorf("output missing %q", s)
		}
	}
}

func TestPrintFunc_UnaryOps(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	r1 := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r1, Src1: 42})
	r2 := b.FreshReg()
	b.Emit(AirInst{Opcode: OpNeg, TypeID: 1, Dest: r2, Src1: r1})
	r3 := b.FreshReg()
	b.Emit(AirInst{Opcode: OpNot, TypeID: 1, Dest: r3, Src1: r1})
	b.Emit(AirInst{Opcode: OpReturn, Src1: r2})
	fn := b.Build()

	got := SprintFunc(fn)
	t.Logf("output:\n%s", got)

	if !strings.Contains(got, "%2: t1 = neg %1") {
		t.Error("expected unary neg format")
	}
	if !strings.Contains(got, "%3: t1 = not %1") {
		t.Error("expected unary not format")
	}
}

func TestPrintFunc_NoTypeID(t *testing.T) {
	b := NewAirFuncBuilder(1, 0)
	r := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 0, Dest: r, Src1: 42})
	b.Emit(AirInst{Opcode: OpReturn, Src1: r})
	fn := b.Build()

	got := SprintFunc(fn)
	t.Logf("output:\n%s", got)

	// TypeID=0 should not show type annotation
	if strings.Contains(got, "t0") {
		t.Error("TypeID=0 should not be printed")
	}
	if !strings.Contains(got, "%1 = iconst") {
		t.Errorf("expected '%%1 = iconst', got:\n%s", got)
	}
}

func TestPrintFunc_VoidInstruction(t *testing.T) {
	b := NewAirFuncBuilder(1, 0)
	b.Emit(AirInst{Opcode: OpStore, Dest: 0, Src1: 1})
	b.Emit(AirInst{Opcode: OpReturn})
	fn := b.Build()

	got := SprintFunc(fn)
	t.Logf("output:\n%s", got)

	if !strings.Contains(got, "store %1") {
		t.Error("void store should print as 'store %1'")
	}
}

func TestPrintModule(t *testing.T) {
	b1 := NewAirFuncBuilder(1, 1)
	b1.Emit(AirInst{Opcode: OpReturn})
	b2 := NewAirFuncBuilder(2, 1)
	b2.Emit(AirInst{Opcode: OpReturn})

	mod := &AirModule{
		Funcs: []AirFunc{*b1.Build(), *b2.Build()},
	}

	var sb strings.Builder
	PrintModule(&sb, mod)
	got := sb.String()
	t.Logf("output:\n%s", got)

	if !strings.Contains(got, "fn @1") {
		t.Error("missing first function")
	}
	if !strings.Contains(got, "fn @2") {
		t.Error("missing second function")
	}
}

func TestPrintFunc_EntryAndExitAnnotations(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	blk1 := b.NewBlock()

	r := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r, Src1: 1})
	b.Emit(AirInst{Opcode: OpJump, Src1: blk1})
	b.AddEdge(0, blk1)

	b.SwitchTo(blk1)
	b.Emit(AirInst{Opcode: OpReturn, Src1: r})
	fn := b.Build()

	got := SprintFunc(fn)
	t.Logf("output:\n%s", got)

	if !strings.Contains(got, "; entry") {
		t.Error("block 0 should be annotated as entry")
	}
	if !strings.Contains(got, "; exit") {
		t.Error("exit block should be annotated as exit")
	}
}

func TestSprintFunc_ReturnType(t *testing.T) {
	b := NewAirFuncBuilder(1, 5)
	b.Emit(AirInst{Opcode: OpReturn})
	fn := b.Build()

	got := SprintFunc(fn)
	if !strings.Contains(got, "-> t5") {
		t.Errorf("expected '-> t5' in output, got:\n%s", got)
	}
}

func TestPrintFunc_DiamondCFG(t *testing.T) {
	fn := buildDiamondCFG()
	got := SprintFunc(fn)
	t.Logf("diamond CFG output:\n%s", got)

	// Just verify it doesn't panic and produces something reasonable.
	if !strings.Contains(got, "block_0:") {
		t.Error("missing block_0")
	}
	if !strings.Contains(got, "block_3:") {
		t.Error("missing block_3")
	}
}
