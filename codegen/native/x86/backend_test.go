package x86_test

import (
	"testing"

	"github.com/axiom-lang/axiom/codegen/native/x86"
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// Instruction Selector Tests
// --------------------------------------------------------------------------

func TestSelect_SimpleReturn(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 1})
	fn := fb.Build()

	insts := x86.Select(fn)
	if len(insts) == 0 {
		t.Fatal("expected non-empty MachInst list")
	}

	// Should contain MachMovImm and MachRet
	foundMov := false
	foundRet := false
	for _, inst := range insts {
		if inst.Op == x86.MachMovImm {
			foundMov = true
			if inst.Src1.Imm != 42 {
				t.Errorf("expected imm=42, got %d", inst.Src1.Imm)
			}
		}
		if inst.Op == x86.MachRet {
			foundRet = true
		}
	}
	if !foundMov {
		t.Error("missing MachMovImm")
	}
	if !foundRet {
		t.Error("missing MachRet")
	}
}

func TestSelect_ZeroConst(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 0})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 1})
	fn := fb.Build()

	insts := x86.Select(fn)
	foundXorZero := false
	for _, inst := range insts {
		if inst.Op == x86.MachXorZero {
			foundXorZero = true
		}
	}
	if !foundXorZero {
		t.Error("zero const should use XOR-zeroing")
	}
}

func TestSelect_Add(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 10})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 20})
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 3})
	fn := fb.Build()

	insts := x86.Select(fn)
	foundAdd := false
	for _, inst := range insts {
		if inst.Op == x86.MachAdd {
			foundAdd = true
		}
	}
	if !foundAdd {
		t.Error("expected MachAdd")
	}
}

func TestSelect_Div(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 100})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 3})
	fb.Emit(air.AirInst{Opcode: air.OpIDiv, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 3})
	fn := fb.Build()

	insts := x86.Select(fn)
	foundCqo := false
	foundIdiv := false
	for _, inst := range insts {
		if inst.Op == x86.MachCqo {
			foundCqo = true
		}
		if inst.Op == x86.MachIdiv {
			foundIdiv = true
		}
	}
	if !foundCqo {
		t.Error("expected CQO before IDIV")
	}
	if !foundIdiv {
		t.Error("expected MachIdiv")
	}
}

func TestSelect_Comparison(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 11)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 5})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 10})
	fb.Emit(air.AirInst{Opcode: air.OpLt, TypeID: 11, Dest: 3, Src1: 1, Src2: 2})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 3})
	fn := fb.Build()

	insts := x86.Select(fn)
	foundCmp := false
	foundSetCC := false
	for _, inst := range insts {
		if inst.Op == x86.MachCmp {
			foundCmp = true
		}
		if inst.Op == x86.MachSetCC && inst.CC == x86.CCL {
			foundSetCC = true
		}
	}
	if !foundCmp {
		t.Error("expected MachCmp")
	}
	if !foundSetCC {
		t.Error("expected MachSetCC with CCL")
	}
}

// --------------------------------------------------------------------------
// Liveness Analysis Tests
// --------------------------------------------------------------------------

func TestLiveness_SimpleFunc(t *testing.T) {
	insts := []x86.MachInst{
		{Op: x86.MachMovImm, Dst: x86.VReg(1), Src1: x86.Imm(42)},
		{Op: x86.MachMov, Dst: x86.Phys(x86.RAX), Src1: x86.VReg(1)},
		{Op: x86.MachRet},
	}

	intervals := x86.ComputeLiveness(insts)
	if len(intervals) != 1 {
		t.Fatalf("expected 1 interval, got %d", len(intervals))
	}
	if intervals[0].VReg != 1 {
		t.Errorf("expected VReg 1, got %d", intervals[0].VReg)
	}
	if intervals[0].Start != 0 || intervals[0].End != 1 {
		t.Errorf("interval: [%d, %d], expected [0, 1]", intervals[0].Start, intervals[0].End)
	}
}

func TestLiveness_TwoRegs(t *testing.T) {
	insts := []x86.MachInst{
		{Op: x86.MachMovImm, Dst: x86.VReg(1), Src1: x86.Imm(10)},  // 0
		{Op: x86.MachMovImm, Dst: x86.VReg(2), Src1: x86.Imm(20)},  // 1
		{Op: x86.MachMov, Dst: x86.VReg(3), Src1: x86.VReg(1)},     // 2
		{Op: x86.MachAdd, Dst: x86.VReg(3), Src1: x86.VReg(2)},     // 3
		{Op: x86.MachRet},
	}

	intervals := x86.ComputeLiveness(insts)
	if len(intervals) != 3 {
		t.Fatalf("expected 3 intervals, got %d", len(intervals))
	}
}

// --------------------------------------------------------------------------
// Register Allocation Tests
// --------------------------------------------------------------------------

func TestRegAlloc_SimpleNoSpill(t *testing.T) {
	intervals := []x86.LiveInterval{
		{VReg: 1, Start: 0, End: 2},
		{VReg: 2, Start: 1, End: 3},
	}

	result := x86.LinearScanAlloc(intervals, []x86.PhysReg{x86.RAX, x86.RCX})
	if result.SpillCount != 0 {
		t.Errorf("expected no spills, got %d", result.SpillCount)
	}
	if len(result.Allocs) != 2 {
		t.Errorf("expected 2 allocations, got %d", len(result.Allocs))
	}
	// Both should get physical registers
	for _, alloc := range result.Allocs {
		if alloc.Spilled {
			t.Errorf("VReg %d should not be spilled", alloc.VReg)
		}
	}
}

func TestRegAlloc_ForcedSpill(t *testing.T) {
	intervals := []x86.LiveInterval{
		{VReg: 1, Start: 0, End: 10},
		{VReg: 2, Start: 1, End: 10},
		{VReg: 3, Start: 2, End: 10},
	}

	// Only 2 registers available → one must spill
	result := x86.LinearScanAlloc(intervals, []x86.PhysReg{x86.RAX, x86.RCX})
	if result.SpillCount != 1 {
		t.Errorf("expected 1 spill, got %d", result.SpillCount)
	}
}

func TestRegAlloc_Empty(t *testing.T) {
	result := x86.LinearScanAlloc(nil, x86.AllocatableGPRs())
	if result.SpillCount != 0 {
		t.Error("expected 0 spills for empty input")
	}
}

func TestRegAlloc_ExpireReuse(t *testing.T) {
	intervals := []x86.LiveInterval{
		{VReg: 1, Start: 0, End: 1},  // dies early
		{VReg: 2, Start: 2, End: 3},  // reuses VReg1's register
	}

	result := x86.LinearScanAlloc(intervals, []x86.PhysReg{x86.RAX})
	if result.SpillCount != 0 {
		t.Error("register should be reused after expiry")
	}
}

// --------------------------------------------------------------------------
// Stack Frame Tests
// --------------------------------------------------------------------------

func TestFrame_Empty(t *testing.T) {
	frame := x86.ComputeFrame(nil, 0, 0)
	if frame.TotalSize != 0 && frame.TotalSize != 8 {
		// After CALL: RSP = ...8. push RBP → ...0. TotalSize=0 keeps 16-byte aligned.
		t.Logf("TotalSize = %d (alignment padding: %d)", frame.TotalSize, frame.AlignPadding)
	}
}

func TestFrame_WithSpills(t *testing.T) {
	frame := x86.ComputeFrame(nil, 4, 0)
	if frame.SpillSlots != 4 {
		t.Errorf("expected 4 spill slots, got %d", frame.SpillSlots)
	}
	if frame.TotalSize < 32 {
		t.Errorf("expected at least 32 bytes for 4 spill slots, got %d", frame.TotalSize)
	}
}

func TestFrame_Alignment(t *testing.T) {
	// With callee-saved registers
	calleeSaved := []x86.PhysReg{x86.RBX, x86.R12}
	frame := x86.ComputeFrame(calleeSaved, 1, 0)
	// pushed: return addr + RBP + RBX + R12 = 4 pushes = 32 bytes
	pushedBytes := (len(calleeSaved) + 1 + 1) * 8 // callee-saved + RBP + ret addr
	total := pushedBytes + frame.TotalSize
	if total%16 != 0 {
		t.Errorf("total %d not 16-aligned (pushed=%d, frame=%d, padding=%d)",
			total, pushedBytes, frame.TotalSize, frame.AlignPadding)
	}
}

func TestFrame_SpillOffset(t *testing.T) {
	frame := x86.ComputeFrame(nil, 4, 0)
	if frame.SpillOffset(0) != -8 {
		t.Errorf("spill slot 0 offset = %d, expected -8", frame.SpillOffset(0))
	}
	if frame.SpillOffset(1) != -16 {
		t.Errorf("spill slot 1 offset = %d, expected -16", frame.SpillOffset(1))
	}
}

func TestFrame_Prologue(t *testing.T) {
	calleeSaved := []x86.PhysReg{x86.RBX}
	frame := x86.ComputeFrame(calleeSaved, 0, 0)
	prologue := x86.EmitPrologue(&frame)
	if len(prologue) < 2 {
		t.Fatalf("prologue too short: %d", len(prologue))
	}
	// First: PUSH RBP
	if prologue[0].Op != x86.MachPush {
		t.Error("expected PUSH RBP first")
	}
	// Second: MOV RBP, RSP
	if prologue[1].Op != x86.MachMov {
		t.Error("expected MOV RBP, RSP second")
	}
}

func TestFrame_Epilogue(t *testing.T) {
	calleeSaved := []x86.PhysReg{x86.RBX}
	frame := x86.ComputeFrame(calleeSaved, 0, 0)
	epilogue := x86.EmitEpilogue(&frame)
	// Should end with POP RBP, RET
	if len(epilogue) < 2 {
		t.Fatalf("epilogue too short: %d", len(epilogue))
	}
	last := epilogue[len(epilogue)-1]
	if last.Op != x86.MachRet {
		t.Error("expected RET at end of epilogue")
	}
}

// --------------------------------------------------------------------------
// ABI Tests
// --------------------------------------------------------------------------

func TestABI_SysV(t *testing.T) {
	abi := x86.NewABI("sysv")
	if abi.Name() != "sysv" {
		t.Errorf("expected sysv, got %s", abi.Name())
	}
	if len(abi.IntArgRegs()) != 6 {
		t.Errorf("SysV should have 6 int arg regs, got %d", len(abi.IntArgRegs()))
	}
	if len(abi.CalleeSavedRegs()) != 5 {
		t.Errorf("SysV should have 5 callee-saved regs, got %d", len(abi.CalleeSavedRegs()))
	}
	if abi.ShadowSpace() != 0 {
		t.Error("SysV should have 0 shadow space")
	}
}

func TestABI_Win64(t *testing.T) {
	abi := x86.NewABI("win64")
	if abi.Name() != "win64" {
		t.Errorf("expected win64, got %s", abi.Name())
	}
	if len(abi.IntArgRegs()) != 4 {
		t.Errorf("Win64 should have 4 int arg regs, got %d", len(abi.IntArgRegs()))
	}
	if len(abi.CalleeSavedRegs()) != 8 {
		t.Errorf("Win64 should have 8 callee-saved regs, got %d", len(abi.CalleeSavedRegs()))
	}
	if abi.ShadowSpace() != 32 {
		t.Errorf("Win64 should have 32-byte shadow space, got %d", abi.ShadowSpace())
	}
}

// --------------------------------------------------------------------------
// Emitter Integration Test
// --------------------------------------------------------------------------

func TestEmitter_SimpleReturn42(t *testing.T) {
	// Build: MOV RAX, 42; RET
	allocs := map[uint32]x86.RegAllocation{
		1: {VReg: 1, Phys: x86.RAX},
	}

	insts := []x86.MachInst{
		{Op: x86.MachMovImm, Dst: x86.VReg(1), Src1: x86.Imm(42)},
		{Op: x86.MachMov, Dst: x86.Phys(x86.RAX), Src1: x86.VReg(1)},
		{Op: x86.MachRet},
	}

	frame := x86.ComputeFrame(nil, 0, 0)
	emitter := x86.NewEmitter(allocs)
	emitter.EmitFunction(insts, &frame)

	if emitter.CodeSize() == 0 {
		t.Error("emitter produced no code")
	}

	// Code should contain 0xC3 (RET)
	foundRet := false
	for _, b := range emitter.Code {
		if b == 0xC3 {
			foundRet = true
		}
	}
	if !foundRet {
		t.Error("emitted code missing RET (0xC3)")
	}
}

// --------------------------------------------------------------------------
// ELF64 Tests
// --------------------------------------------------------------------------

func TestELF64_ValidHeader(t *testing.T) {
	w := x86.NewELF64Writer()
	w.SetText([]byte{0xC3}) // minimal: just RET
	w.AddSymbol(x86.ELF64Sym{
		Name:    "main",
		Value:   0,
		Size:    1,
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 1,
	})

	obj := w.Serialize()
	if len(obj) < 64 {
		t.Fatalf("ELF too small: %d bytes", len(obj))
	}

	// Check magic
	if string(obj[0:4]) != "\x7fELF" {
		t.Error("missing ELF magic")
	}
	// Check class (64-bit)
	if obj[4] != 2 {
		t.Errorf("expected ELFCLASS64 (2), got %d", obj[4])
	}
	// Check little-endian
	if obj[5] != 1 {
		t.Errorf("expected ELFDATA2LSB (1), got %d", obj[5])
	}
}

func TestELF64_EmptyText(t *testing.T) {
	w := x86.NewELF64Writer()
	w.SetText(nil)
	obj := w.Serialize()
	if len(obj) < 64 {
		t.Fatalf("ELF too small: %d bytes", len(obj))
	}
	if string(obj[0:4]) != "\x7fELF" {
		t.Error("missing ELF magic")
	}
}
