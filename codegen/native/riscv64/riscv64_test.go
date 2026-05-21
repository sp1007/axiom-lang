package riscv64_test

import (
	"encoding/binary"
	"testing"

	"github.com/axiom-lang/axiom/codegen/native/riscv64"
)

// --------------------------------------------------------------------------
// Register Tests
// --------------------------------------------------------------------------

func TestPhysReg_IsGPR(t *testing.T) {
	if !riscv64.A0.IsGPR() {
		t.Error("A0 should be GPR")
	}
	if !riscv64.Zero.IsGPR() {
		t.Error("Zero should be GPR")
	}
	if riscv64.FA0.IsGPR() {
		t.Error("FA0 should not be GPR")
	}
}

func TestPhysReg_IsFP(t *testing.T) {
	if !riscv64.FA0.IsFP() {
		t.Error("FA0 should be FP")
	}
	if riscv64.A0.IsFP() {
		t.Error("A0 should not be FP")
	}
}

func TestPhysReg_String(t *testing.T) {
	tests := []struct {
		reg  riscv64.PhysReg
		name string
	}{
		{riscv64.Zero, "zero"},
		{riscv64.RA, "ra"},
		{riscv64.SP, "sp"},
		{riscv64.A0, "a0"},
		{riscv64.S0, "s0"},
		{riscv64.T0, "t0"},
	}
	for _, tt := range tests {
		if tt.reg.String() != tt.name {
			t.Errorf("expected %s, got %s", tt.name, tt.reg.String())
		}
	}
}

func TestAllocatableGPRs(t *testing.T) {
	gprs := riscv64.AllocatableGPRs()
	if len(gprs) < 25 {
		t.Errorf("expected at least 25 allocatable GPRs, got %d", len(gprs))
	}
	// Should not include zero, ra, sp, gp, tp
	for _, r := range gprs {
		if r == riscv64.Zero || r == riscv64.RA || r == riscv64.SP ||
			r == riscv64.GP || r == riscv64.TP {
			t.Errorf("register %s should not be allocatable", r.String())
		}
	}
}

// --------------------------------------------------------------------------
// Encoding Tests
// --------------------------------------------------------------------------

func TestEncodeNop(t *testing.T) {
	code := riscv64.EncodeNop()
	if len(code) != 4 {
		t.Fatalf("NOP should be 4 bytes, got %d", len(code))
	}
	inst := binary.LittleEndian.Uint32(code)
	// NOP = ADDI x0, x0, 0
	if inst != 0x00000013 {
		t.Errorf("NOP encoding = 0x%08X, expected 0x00000013", inst)
	}
}

func TestEncodeRet(t *testing.T) {
	code := riscv64.EncodeRet()
	inst := binary.LittleEndian.Uint32(code)
	// RET = JALR x0, x1, 0
	if inst != 0x00008067 {
		t.Errorf("RET encoding = 0x%08X, expected 0x00008067", inst)
	}
}

func TestEncodeAdd(t *testing.T) {
	code := riscv64.EncodeAdd(riscv64.A0, riscv64.A1, riscv64.A2)
	inst := binary.LittleEndian.Uint32(code)
	// Verify R-type opcode = 0x33
	if inst&0x7F != 0x33 {
		t.Errorf("ADD opcode = 0x%02X, expected 0x33", inst&0x7F)
	}
}

func TestEncodeSub(t *testing.T) {
	code := riscv64.EncodeSub(riscv64.A0, riscv64.A1, riscv64.A2)
	inst := binary.LittleEndian.Uint32(code)
	// funct7 should be 0x20 for SUB
	if (inst>>25)&0x7F != 0x20 {
		t.Errorf("SUB funct7 = 0x%02X, expected 0x20", (inst>>25)&0x7F)
	}
}

func TestEncodeAddi(t *testing.T) {
	code := riscv64.EncodeAddi(riscv64.A0, riscv64.Zero, 42)
	inst := binary.LittleEndian.Uint32(code)
	if inst&0x7F != 0x13 {
		t.Errorf("ADDI opcode = 0x%02X, expected 0x13", inst&0x7F)
	}
	// Check immediate
	imm := int32(inst) >> 20
	if imm != 42 {
		t.Errorf("ADDI immediate = %d, expected 42", imm)
	}
}

func TestEncodeLi_Small(t *testing.T) {
	code := riscv64.EncodeLi(riscv64.A0, 42)
	if len(code) != 4 {
		t.Errorf("LI(42) should be 4 bytes (single ADDI), got %d", len(code))
	}
}

func TestEncodeLi_Large(t *testing.T) {
	code := riscv64.EncodeLi(riscv64.A0, 0x12345)
	if len(code) != 8 {
		t.Errorf("LI(0x12345) should be 8 bytes (LUI+ADDI), got %d", len(code))
	}
}

func TestEncodeLdSd(t *testing.T) {
	ld := riscv64.EncodeLd(riscv64.A0, riscv64.SP, 16)
	inst := binary.LittleEndian.Uint32(ld)
	if inst&0x7F != 0x03 {
		t.Errorf("LD opcode = 0x%02X, expected 0x03", inst&0x7F)
	}

	sd := riscv64.EncodeSd(riscv64.A0, riscv64.SP, 16)
	inst = binary.LittleEndian.Uint32(sd)
	if inst&0x7F != 0x23 {
		t.Errorf("SD opcode = 0x%02X, expected 0x23", inst&0x7F)
	}
}

func TestEncodeBeq(t *testing.T) {
	code := riscv64.EncodeBeq(riscv64.A0, riscv64.A1, 0x10)
	inst := binary.LittleEndian.Uint32(code)
	if inst&0x7F != 0x63 {
		t.Errorf("BEQ opcode = 0x%02X, expected 0x63", inst&0x7F)
	}
}

func TestEncodeJal(t *testing.T) {
	code := riscv64.EncodeJal(riscv64.RA, 0x100)
	inst := binary.LittleEndian.Uint32(code)
	if inst&0x7F != 0x6F {
		t.Errorf("JAL opcode = 0x%02X, expected 0x6F", inst&0x7F)
	}
}

// --------------------------------------------------------------------------
// ABI Tests
// --------------------------------------------------------------------------

func TestABI_LP64D(t *testing.T) {
	abi := &riscv64.RV64ABI{}
	if abi.Name() != "lp64d" {
		t.Errorf("expected lp64d, got %s", abi.Name())
	}
	if len(abi.IntArgRegs()) != 8 {
		t.Errorf("expected 8 int arg regs, got %d", len(abi.IntArgRegs()))
	}
	if len(abi.CalleeSavedRegs()) != 12 {
		t.Errorf("expected 12 callee-saved, got %d", len(abi.CalleeSavedRegs()))
	}
	if abi.ReturnReg() != riscv64.A0 {
		t.Error("return reg should be A0")
	}
	if abi.StackAlignment() != 16 {
		t.Error("stack alignment should be 16")
	}
}

// --------------------------------------------------------------------------
// Frame Tests
// --------------------------------------------------------------------------

func TestFrame_Basic(t *testing.T) {
	frame := riscv64.ComputeFrame(nil, 2, 0)
	// 8 (RA) + 16 (2 spills) = 24 → pad to 32
	if frame.TotalSize%16 != 0 {
		t.Errorf("frame size %d not 16-aligned", frame.TotalSize)
	}
}

func TestFrame_SpillOffset(t *testing.T) {
	frame := riscv64.ComputeFrame([]riscv64.PhysReg{riscv64.S0, riscv64.S1}, 2, 0)
	off0 := frame.SpillOffset(0)
	off1 := frame.SpillOffset(1)
	if off1 != off0+8 {
		t.Errorf("spill offsets: [0]=%d [1]=%d, expected 8-byte spacing", off0, off1)
	}
}
