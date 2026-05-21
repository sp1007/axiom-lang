package arm64_test

import (
	"encoding/binary"
	"testing"

	"github.com/axiom-lang/axiom/codegen/native/arm64"
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// Register Tests
// --------------------------------------------------------------------------

func TestPhysReg_IsGPR(t *testing.T) {
	if !arm64.X0.IsGPR() {
		t.Error("X0 should be GPR")
	}
	if !arm64.X30.IsGPR() {
		t.Error("X30 should be GPR")
	}
	if arm64.V0.IsGPR() {
		t.Error("V0 should not be GPR")
	}
}

func TestPhysReg_IsVec(t *testing.T) {
	if !arm64.V0.IsVec() {
		t.Error("V0 should be Vec")
	}
	if !arm64.V31.IsVec() {
		t.Error("V31 should be Vec")
	}
	if arm64.X0.IsVec() {
		t.Error("X0 should not be Vec")
	}
}

func TestPhysReg_String(t *testing.T) {
	if arm64.X0.String() != "x0" {
		t.Errorf("expected x0, got %s", arm64.X0.String())
	}
	if arm64.X29.String() != "x29" {
		t.Errorf("expected x29, got %s", arm64.X29.String())
	}
	if arm64.V0.String() != "v0" {
		t.Errorf("expected v0, got %s", arm64.V0.String())
	}
}

func TestAllocatableGPRs(t *testing.T) {
	gprs := arm64.AllocatableGPRs()
	if len(gprs) != 26 {
		t.Errorf("expected 26 allocatable GPRs, got %d", len(gprs))
	}
	// Should not include X18, X29, X30
	for _, r := range gprs {
		if r == arm64.X18 || r == arm64.X29 || r == arm64.X30 {
			t.Errorf("register %s should not be allocatable", r.String())
		}
	}
}

// --------------------------------------------------------------------------
// Encoding Tests
// --------------------------------------------------------------------------

func TestEncodeRet(t *testing.T) {
	code := arm64.EncodeRet()
	if len(code) != 4 {
		t.Fatalf("RET should be 4 bytes, got %d", len(code))
	}
	inst := binary.LittleEndian.Uint32(code)
	if inst != 0xD65F03C0 {
		t.Errorf("RET encoding = 0x%08X, expected 0xD65F03C0", inst)
	}
}

func TestEncodeNop(t *testing.T) {
	code := arm64.EncodeNop()
	inst := binary.LittleEndian.Uint32(code)
	if inst != 0xD503201F {
		t.Errorf("NOP encoding = 0x%08X, expected 0xD503201F", inst)
	}
}

func TestEncodeAddReg(t *testing.T) {
	code := arm64.EncodeAddReg(arm64.X0, arm64.X1, arm64.X2)
	inst := binary.LittleEndian.Uint32(code)
	// ADD X0, X1, X2: sf=1, op=0, S=0, shift=00, Rm=2, imm6=0, Rn=1, Rd=0
	// 1 0 0 01011 00 0 00010 000000 00001 00000
	if inst&0xFF000000 != 0x8B000000 {
		t.Errorf("ADD encoding top byte = 0x%08X", inst)
	}
}

func TestEncodeSubReg(t *testing.T) {
	code := arm64.EncodeSubReg(arm64.X3, arm64.X4, arm64.X5)
	inst := binary.LittleEndian.Uint32(code)
	if inst&0xFF000000 != 0xCB000000 {
		t.Errorf("SUB encoding top byte = 0x%08X", inst)
	}
}

func TestEncodeMovz(t *testing.T) {
	code := arm64.EncodeMovz(arm64.X0, 42, 0)
	if len(code) != 4 {
		t.Fatalf("MOVZ should be 4 bytes")
	}
	inst := binary.LittleEndian.Uint32(code)
	if inst&0xFF800000 != 0xD2800000 {
		t.Errorf("MOVZ encoding = 0x%08X", inst)
	}
}

func TestEncodeBl(t *testing.T) {
	code := arm64.EncodeBl(0x100)
	inst := binary.LittleEndian.Uint32(code)
	if inst&0xFC000000 != 0x94000000 {
		t.Errorf("BL encoding = 0x%08X", inst)
	}
}

func TestEncodeLdrStr(t *testing.T) {
	ldr := arm64.EncodeLdrImm(arm64.X1, arm64.X29, 16)
	inst := binary.LittleEndian.Uint32(ldr)
	if inst&0xFFC00000 != 0xF9400000 {
		t.Errorf("LDR encoding = 0x%08X", inst)
	}

	str := arm64.EncodeStrImm(arm64.X1, arm64.X29, 16)
	inst = binary.LittleEndian.Uint32(str)
	if inst&0xFFC00000 != 0xF9000000 {
		t.Errorf("STR encoding = 0x%08X", inst)
	}
}

func TestEncodeBCond(t *testing.T) {
	code := arm64.EncodeBCond(arm64.CondEQ, 0x20)
	inst := binary.LittleEndian.Uint32(code)
	if inst&0xFF000010 != 0x54000000 {
		t.Errorf("B.EQ encoding = 0x%08X", inst)
	}
}

func TestCondCode_String(t *testing.T) {
	if arm64.CondEQ.String() != "eq" {
		t.Errorf("expected 'eq', got %s", arm64.CondEQ.String())
	}
	if arm64.CondGT.String() != "gt" {
		t.Errorf("expected 'gt', got %s", arm64.CondGT.String())
	}
}

// --------------------------------------------------------------------------
// Instruction Selector Tests
// --------------------------------------------------------------------------

func TestSelect_SimpleReturn(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 42})
	fb.Emit(air.AirInst{Opcode: air.OpReturn, Src1: 1})
	fn := fb.Build()

	insts := arm64.Select(fn)
	if len(insts) == 0 {
		t.Fatal("expected non-empty MachInst list")
	}

	foundRet := false
	for _, inst := range insts {
		if inst.Op == arm64.MachRet {
			foundRet = true
		}
	}
	if !foundRet {
		t.Error("missing MachRet")
	}
}

func TestSelect_Add(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 10})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 20})
	fb.Emit(air.AirInst{Opcode: air.OpIAdd, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fn := fb.Build()

	insts := arm64.Select(fn)
	foundAdd := false
	for _, inst := range insts {
		if inst.Op == arm64.MachAdd {
			foundAdd = true
			// ARM64 ADD is three-operand
			if inst.Src2.Kind == arm64.OpndNone {
				t.Error("ARM64 ADD should have Src2")
			}
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
	fn := fb.Build()

	insts := arm64.Select(fn)
	foundSdiv := false
	for _, inst := range insts {
		if inst.Op == arm64.MachSdiv {
			foundSdiv = true
		}
	}
	if !foundSdiv {
		t.Error("expected MachSdiv")
	}
}

func TestSelect_Mod(t *testing.T) {
	fb := air.NewAirFuncBuilder(1, 3)
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 1, Src1: 10})
	fb.Emit(air.AirInst{Opcode: air.OpIConst, TypeID: 3, Dest: 2, Src1: 3})
	fb.Emit(air.AirInst{Opcode: air.OpIMod, TypeID: 3, Dest: 3, Src1: 1, Src2: 2})
	fn := fb.Build()

	insts := arm64.Select(fn)
	foundSdiv := false
	foundMsub := false
	for _, inst := range insts {
		if inst.Op == arm64.MachSdiv {
			foundSdiv = true
		}
		if inst.Op == arm64.MachMsub {
			foundMsub = true
		}
	}
	if !foundSdiv {
		t.Error("expected SDIV for modulo")
	}
	if !foundMsub {
		t.Error("expected MSUB for modulo")
	}
}

// --------------------------------------------------------------------------
// ABI Tests
// --------------------------------------------------------------------------

func TestABI_AAPCS64(t *testing.T) {
	abi := &arm64.AAPCS64{}
	if abi.Name() != "aapcs64" {
		t.Errorf("expected aapcs64, got %s", abi.Name())
	}
	if len(abi.IntArgRegs()) != 8 {
		t.Errorf("expected 8 int arg regs, got %d", len(abi.IntArgRegs()))
	}
	if len(abi.CalleeSavedRegs()) != 10 {
		t.Errorf("expected 10 callee-saved, got %d", len(abi.CalleeSavedRegs()))
	}
	if abi.ReturnReg() != arm64.X0 {
		t.Error("return reg should be X0")
	}
	if abi.StackAlignment() != 16 {
		t.Error("stack alignment should be 16")
	}
}

// --------------------------------------------------------------------------
// Frame Tests
// --------------------------------------------------------------------------

func TestFrame_Basic(t *testing.T) {
	frame := arm64.ComputeFrame(nil, 2, 0)
	// 16 (FP+LR) + 16 (2 spill slots) = 32 → already aligned
	if frame.TotalSize < 32 {
		t.Errorf("expected at least 32 bytes, got %d", frame.TotalSize)
	}
	if frame.TotalSize%16 != 0 {
		t.Errorf("frame size %d not 16-aligned", frame.TotalSize)
	}
}

func TestFrame_Prologue(t *testing.T) {
	frame := arm64.ComputeFrame([]arm64.PhysReg{arm64.X19}, 0, 0)
	prologue := arm64.EmitPrologue(&frame)
	if len(prologue) < 2 {
		t.Fatalf("prologue too short: %d", len(prologue))
	}
}

func TestFrame_Epilogue(t *testing.T) {
	frame := arm64.ComputeFrame(nil, 0, 0)
	epilogue := arm64.EmitEpilogue(&frame)
	last := epilogue[len(epilogue)-1]
	if last.Op != arm64.MachRet {
		t.Error("expected RET at end of epilogue")
	}
}
