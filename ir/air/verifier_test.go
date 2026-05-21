package air

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildValidFunc creates a well-formed single-block function:
//
//	%1: t1 = iconst 42
//	%2: t1 = iconst 58
//	%3: t1 = iadd %1, %2
//	ret %3
func buildValidFunc() *AirFunc {
	b := NewAirFuncBuilder(1, 1)

	r1 := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r1, Src1: 42})

	r2 := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r2, Src1: 58})

	r3 := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIAdd, TypeID: 1, Dest: r3, Src1: r1, Src2: r2})

	b.Emit(AirInst{Opcode: OpReturn, Src1: r3})

	return b.Build()
}

// hasError returns true if any VerifyError message contains substr.
func hasError(errs []VerifyError, substr string) bool {
	for _, e := range errs {
		if strings.Contains(e.Message, substr) {
			return true
		}
	}
	return false
}

func dumpErrors(t *testing.T, errs []VerifyError) {
	t.Helper()
	for _, e := range errs {
		t.Logf("  %s", e.Error())
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestVerify_ValidFunc(t *testing.T) {
	fn := buildValidFunc()
	errs := Verify(fn)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid func, got %d:", len(errs))
		dumpErrors(t, errs)
	}
}

func TestVerify_SSA_DuplicateDest(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	r := b.FreshReg() // %1
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r, Src1: 42})
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r, Src1: 58}) // dup!
	b.Emit(AirInst{Opcode: OpReturn, Src1: r})
	fn := b.Build()

	errs := Verify(fn)
	if !hasError(errs, "SSA violation") {
		t.Error("expected SSA violation for duplicate Dest")
		dumpErrors(t, errs)
	}
}

func TestVerify_SSA_DestZero_NotTracked(t *testing.T) {
	// Dest==0 is NoValue and should not be tracked for SSA.
	b := NewAirFuncBuilder(1, 1)
	b.Emit(AirInst{Opcode: OpStore, Dest: 0, Src1: 1})
	b.Emit(AirInst{Opcode: OpStore, Dest: 0, Src1: 2})
	b.Emit(AirInst{Opcode: OpReturn})
	fn := b.Build()

	errs := Verify(fn)
	if hasError(errs, "SSA violation") {
		t.Error("Dest==0 should not trigger SSA violation")
		dumpErrors(t, errs)
	}
}

func TestVerify_Terminator_Missing(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	r := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r, Src1: 42})
	// No terminator!
	fn := b.Build()

	errs := Verify(fn)
	if !hasError(errs, "not a terminator") {
		t.Error("expected error for missing terminator")
		dumpErrors(t, errs)
	}
}

func TestVerify_Terminator_NotLast(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	r := b.FreshReg()
	b.Emit(AirInst{Opcode: OpReturn, Src1: r})       // terminator
	b.Emit(AirInst{Opcode: OpIConst, Dest: 99, Src1: 1}) // after terminator!
	fn := b.Build()

	errs := Verify(fn)
	if !hasError(errs, "not the last instruction") {
		t.Error("expected error for terminator not at end")
		dumpErrors(t, errs)
	}
}

func TestVerify_EntryBlock_NotMarked(t *testing.T) {
	fn := &AirFunc{
		Blocks: []BasicBlock{
			{ID: 0, IsEntry: false, Instrs: []uint32{0}},
		},
		Insts: []AirInst{
			{Opcode: OpReturn},
		},
	}
	errs := Verify(fn)
	if !hasError(errs, "IsEntry=true") {
		t.Error("expected error for block 0 not marked as entry")
		dumpErrors(t, errs)
	}
}

func TestVerify_EntryBlock_HasPredecessors(t *testing.T) {
	fn := &AirFunc{
		Blocks: []BasicBlock{
			{ID: 0, IsEntry: true, Preds: []uint32{1}, Instrs: []uint32{0}},
			{ID: 1, Instrs: []uint32{1}},
		},
		Insts: []AirInst{
			{Opcode: OpReturn},
			{Opcode: OpJump, Src1: 0},
		},
	}
	errs := Verify(fn)
	if !hasError(errs, "no predecessors") {
		t.Error("expected error for entry block with predecessors")
		dumpErrors(t, errs)
	}
}

func TestVerify_PhiPlacement_AfterNonPhi(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	r1 := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r1, Src1: 1})
	r2 := b.FreshReg()
	b.Emit(AirInst{Opcode: OpPhi, TypeID: 1, Dest: r2}) // phi after non-phi!
	b.Emit(AirInst{Opcode: OpReturn, Src1: r1})
	fn := b.Build()

	errs := Verify(fn)
	if !hasError(errs, "phi instruction after non-phi") {
		t.Error("expected error for phi after non-phi")
		dumpErrors(t, errs)
	}
}

func TestVerify_PhiPlacement_AtStart_OK(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	r1 := b.FreshReg()
	b.Emit(AirInst{Opcode: OpReturn})

	blk1 := b.NewBlock()
	b.SwitchTo(blk1)
	b.Emit(AirInst{Opcode: OpPhi, TypeID: 1, Dest: r1})
	r2 := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r2, Src1: 1})
	b.Emit(AirInst{Opcode: OpReturn, Src1: r2})
	fn := b.Build()

	errs := Verify(fn)
	if hasError(errs, "phi instruction after non-phi") {
		t.Error("phi at start of block should not produce an error")
		dumpErrors(t, errs)
	}
}

func TestVerify_BranchTarget_OutOfRange(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	r := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r, Src1: 1})
	b.Emit(AirInst{Opcode: OpJump, Src1: 999}) // invalid target
	fn := b.Build()

	errs := Verify(fn)
	if !hasError(errs, "out of range") {
		t.Error("expected error for jump to invalid block")
		dumpErrors(t, errs)
	}
}

func TestVerify_BranchTarget_BothInvalid(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	r := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r, Src1: 1})
	b.Emit(AirInst{Opcode: OpBranch, Src1: r, Src2: 100, Dest: 200})
	fn := b.Build()

	errs := Verify(fn)
	trueErr := hasError(errs, "true target block_100")
	falseErr := hasError(errs, "false target block_200")
	if !trueErr || !falseErr {
		t.Errorf("expected both branch targets invalid: true=%v false=%v", trueErr, falseErr)
		dumpErrors(t, errs)
	}
}

func TestVerify_SuccessorConsistency_Missing(t *testing.T) {
	// Build a function where jump targets block 1 but Succs is empty.
	b := NewAirFuncBuilder(1, 1)
	blk1 := b.NewBlock()
	b.SwitchTo(0)
	r := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r, Src1: 1})
	b.Emit(AirInst{Opcode: OpJump, Src1: blk1})
	// Intentionally NOT calling AddEdge
	b.SwitchTo(blk1)
	b.Emit(AirInst{Opcode: OpReturn})
	fn := b.Build()

	errs := Verify(fn)
	if !hasError(errs, "not in Succs") {
		t.Error("expected successor consistency error")
		dumpErrors(t, errs)
	}
}

func TestVerify_SuccessorConsistency_Extra(t *testing.T) {
	// Succs has a block that the terminator doesn't target.
	b := NewAirFuncBuilder(1, 1)
	blk1 := b.NewBlock()
	blk2 := b.NewBlock()
	b.SwitchTo(0)
	r := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r, Src1: 1})
	b.Emit(AirInst{Opcode: OpJump, Src1: blk1})
	b.AddEdge(0, blk1)
	b.AddEdge(0, blk2) // spurious edge
	b.SwitchTo(blk1)
	b.Emit(AirInst{Opcode: OpReturn})
	b.SwitchTo(blk2)
	b.Emit(AirInst{Opcode: OpReturn})
	fn := b.Build()

	errs := Verify(fn)
	if !hasError(errs, "terminator does not target") {
		t.Error("expected error for extra successor")
		dumpErrors(t, errs)
	}
}

func TestVerify_EmptyBlock_Warning(t *testing.T) {
	b := NewAirFuncBuilder(1, 1)
	b.Emit(AirInst{Opcode: OpReturn})
	b.NewBlock() // empty block
	fn := b.Build()

	errs := Verify(fn)
	if !hasError(errs, "warning") {
		t.Error("expected warning for empty block")
		dumpErrors(t, errs)
	}
}

func TestVerify_DiamondCFG_Valid(t *testing.T) {
	fn := buildDiamondCFG()
	errs := Verify(fn)
	if len(errs) != 0 {
		t.Errorf("diamond CFG should be valid, got %d errors:", len(errs))
		dumpErrors(t, errs)
	}
}

func TestVerify_NopsIgnored(t *testing.T) {
	// A block with only NOPs and a terminator should be valid.
	b := NewAirFuncBuilder(1, 1)
	b.Emit(AirInst{Opcode: OpNop})
	b.Emit(AirInst{Opcode: OpNop})
	b.Emit(AirInst{Opcode: OpReturn})
	fn := b.Build()

	errs := Verify(fn)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d:", len(errs))
		dumpErrors(t, errs)
	}
}

func TestVerify_BranchDest_NotSSA(t *testing.T) {
	// OpBranch uses Dest as false target, not as SSA def.
	// So the same "Dest" value used in a branch should not conflict with
	// a real SSA def at that register.
	b := NewAirFuncBuilder(1, 1)
	blk1 := b.NewBlock()
	blk2 := b.NewBlock()
	r := b.FreshReg()
	b.Emit(AirInst{Opcode: OpIConst, TypeID: 1, Dest: r, Src1: 1})
	// branch uses Dest=blk2 (which might collide with a register number)
	b.Emit(AirInst{Opcode: OpBranch, Src1: r, Src2: blk1, Dest: blk2})
	b.AddEdge(0, blk1)
	b.AddEdge(0, blk2)

	b.SwitchTo(blk1)
	b.Emit(AirInst{Opcode: OpReturn})
	b.SwitchTo(blk2)
	b.Emit(AirInst{Opcode: OpReturn})
	fn := b.Build()

	errs := Verify(fn)
	if hasError(errs, "SSA violation") {
		t.Error("OpBranch.Dest should not be treated as SSA def")
		dumpErrors(t, errs)
	}
}

func TestVerifyError_String(t *testing.T) {
	e := VerifyError{BlockID: 2, InstIdx: 5, Message: "bad thing"}
	want := "block_2[5]: bad thing"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}

	e2 := VerifyError{BlockID: 0, InstIdx: ^uint32(0), Message: "block-level"}
	got := e2.Error()
	if !strings.HasPrefix(got, "block_0:") {
		t.Errorf("block-level Error() = %q, expected prefix 'block_0:'", got)
	}
}
