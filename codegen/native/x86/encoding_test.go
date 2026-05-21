package x86_test

import (
	"bytes"
	"testing"

	"github.com/axiom-lang/axiom/codegen/native/x86"
)

// --------------------------------------------------------------------------
// Register Tests
// --------------------------------------------------------------------------

func TestPhysReg_IsGPR(t *testing.T) {
	if !x86.RAX.IsGPR() { t.Error("RAX should be GPR") }
	if !x86.R15.IsGPR() { t.Error("R15 should be GPR") }
	if x86.XMM0.IsGPR() { t.Error("XMM0 should not be GPR") }
}

func TestPhysReg_IsXMM(t *testing.T) {
	if !x86.XMM0.IsXMM() { t.Error("XMM0 should be XMM") }
	if !x86.XMM15.IsXMM() { t.Error("XMM15 should be XMM") }
	if x86.RAX.IsXMM() { t.Error("RAX should not be XMM") }
}

func TestPhysReg_HWReg(t *testing.T) {
	if x86.RAX.HWReg() != 0 { t.Errorf("RAX HWReg = %d", x86.RAX.HWReg()) }
	if x86.R8.HWReg() != 8 { t.Errorf("R8 HWReg = %d", x86.R8.HWReg()) }
	if x86.R15.HWReg() != 15 { t.Errorf("R15 HWReg = %d", x86.R15.HWReg()) }
	if x86.XMM0.HWReg() != 0 { t.Errorf("XMM0 HWReg = %d", x86.XMM0.HWReg()) }
}

func TestPhysReg_NeedsREX(t *testing.T) {
	if x86.RAX.NeedsREX() { t.Error("RAX should not need REX") }
	if x86.RDI.NeedsREX() { t.Error("RDI should not need REX") }
	if !x86.R8.NeedsREX() { t.Error("R8 should need REX") }
	if !x86.R15.NeedsREX() { t.Error("R15 should need REX") }
}

func TestPhysReg_String(t *testing.T) {
	if x86.RAX.String() != "rax" { t.Errorf("RAX.String() = %q", x86.RAX.String()) }
	if x86.R15.String() != "r15" { t.Errorf("R15.String() = %q", x86.R15.String()) }
	if x86.XMM0.String() != "xmm0" { t.Errorf("XMM0.String() = %q", x86.XMM0.String()) }
}

func TestAllocatableGPRs(t *testing.T) {
	regs := x86.AllocatableGPRs()
	if len(regs) != 14 { t.Errorf("expected 14 allocatable GPRs, got %d", len(regs)) }
	for _, r := range regs {
		if r == x86.RSP || r == x86.RBP {
			t.Errorf("RSP/RBP should not be allocatable, found %s", r)
		}
	}
}

// --------------------------------------------------------------------------
// ModRM Tests
// --------------------------------------------------------------------------

func TestModRM_RR(t *testing.T) {
	// MOV RAX, RCX → ModRM = 11 001 000 = 0xC8
	modrm, _, _ := x86.EncodeModRM_RR(x86.RCX, x86.RAX)
	if modrm != 0xC8 {
		t.Errorf("ModRM_RR(RCX, RAX) = 0x%02X, expected 0xC8", modrm)
	}
}

func TestModRM_RR_Extended(t *testing.T) {
	// R8, R9 — both need REX
	_, rex, needREX := x86.EncodeModRM_RR(x86.R8, x86.R9)
	if !needREX {
		t.Error("R8/R9 should need REX")
	}
	// REX.R for R8 (reg), REX.B for R9 (rm)
	if rex&x86.REXR == 0 { t.Error("expected REX.R for R8") }
	if rex&x86.REXB == 0 { t.Error("expected REX.B for R9") }
}

func TestEncodeREX(t *testing.T) {
	// REX.W only
	rex := x86.EncodeREX(true, false, false, false)
	if rex != 0x48 {
		t.Errorf("REX.W = 0x%02X, expected 0x48", rex)
	}
	// REX.WRB
	rex = x86.EncodeREX(true, true, false, true)
	if rex != 0x4D {
		t.Errorf("REX.WRB = 0x%02X, expected 0x4D", rex)
	}
}

// --------------------------------------------------------------------------
// Instruction Encoding Tests
// --------------------------------------------------------------------------

func TestEncodeRet(t *testing.T) {
	got := x86.EncodeRet()
	if !bytes.Equal(got, []byte{0xC3}) {
		t.Errorf("RET = %X, expected C3", got)
	}
}

func TestEncodeNop(t *testing.T) {
	got := x86.EncodeNop()
	if !bytes.Equal(got, []byte{0x90}) {
		t.Errorf("NOP = %X, expected 90", got)
	}
}

func TestEncodePush(t *testing.T) {
	// PUSH RAX → 50
	got := x86.EncodePush(x86.RAX)
	if !bytes.Equal(got, []byte{0x50}) {
		t.Errorf("PUSH RAX = %X, expected 50", got)
	}
	// PUSH R8 → 41 50
	got = x86.EncodePush(x86.R8)
	if !bytes.Equal(got, []byte{0x41, 0x50}) {
		t.Errorf("PUSH R8 = %X, expected 41 50", got)
	}
}

func TestEncodePop(t *testing.T) {
	got := x86.EncodePop(x86.RBX)
	if !bytes.Equal(got, []byte{0x5B}) {
		t.Errorf("POP RBX = %X, expected 5B", got)
	}
}

func TestEncodeMovRR(t *testing.T) {
	// MOV RAX, RCX → 48 89 C8
	got := x86.EncodeMovRR(x86.RAX, x86.RCX)
	if len(got) != 3 || got[1] != 0x89 {
		t.Errorf("MOV RAX, RCX = %X", got)
	}
}

func TestEncodeMovRI(t *testing.T) {
	// MOV RAX, 42 → 48 C7 C0 2A 00 00 00
	got := x86.EncodeMovRI(x86.RAX, 42)
	if len(got) != 7 {
		t.Errorf("MOV RAX, 42 = %X (len=%d)", got, len(got))
	}
	if got[1] != 0xC7 {
		t.Errorf("opcode should be 0xC7, got 0x%02X", got[1])
	}
	// Check immediate
	imm := int32(got[3]) | int32(got[4])<<8 | int32(got[5])<<16 | int32(got[6])<<24
	if imm != 42 {
		t.Errorf("immediate = %d, expected 42", imm)
	}
}

func TestEncodeAddRR(t *testing.T) {
	got := x86.EncodeAddRR(x86.RAX, x86.RCX)
	if len(got) != 3 || got[1] != 0x01 {
		t.Errorf("ADD RAX, RCX = %X", got)
	}
}

func TestEncodeSubRR(t *testing.T) {
	got := x86.EncodeSubRR(x86.RAX, x86.RCX)
	if len(got) != 3 || got[1] != 0x29 {
		t.Errorf("SUB RAX, RCX = %X", got)
	}
}

func TestEncodeImulRR(t *testing.T) {
	got := x86.EncodeImulRR(x86.RAX, x86.RCX)
	if len(got) != 4 || got[1] != 0x0F || got[2] != 0xAF {
		t.Errorf("IMUL RAX, RCX = %X", got)
	}
}

func TestEncodeCqo(t *testing.T) {
	got := x86.EncodeCqo()
	if len(got) != 2 || got[1] != 0x99 {
		t.Errorf("CQO = %X", got)
	}
}

func TestEncodeIdivR(t *testing.T) {
	got := x86.EncodeIdivR(x86.RCX)
	if len(got) != 3 || got[1] != 0xF7 {
		t.Errorf("IDIV RCX = %X", got)
	}
}

func TestEncodeNegR(t *testing.T) {
	got := x86.EncodeNegR(x86.RAX)
	if len(got) != 3 || got[1] != 0xF7 {
		t.Errorf("NEG RAX = %X", got)
	}
}

func TestEncodeCmpRR(t *testing.T) {
	got := x86.EncodeCmpRR(x86.RAX, x86.RCX)
	if len(got) != 3 || got[1] != 0x39 {
		t.Errorf("CMP RAX, RCX = %X", got)
	}
}

func TestEncodeAndRR(t *testing.T) {
	got := x86.EncodeAndRR(x86.RAX, x86.RCX)
	if len(got) != 3 || got[1] != 0x21 {
		t.Errorf("AND RAX, RCX = %X", got)
	}
}

func TestEncodeOrRR(t *testing.T) {
	got := x86.EncodeOrRR(x86.RAX, x86.RCX)
	if len(got) != 3 || got[1] != 0x09 {
		t.Errorf("OR RAX, RCX = %X", got)
	}
}

func TestEncodeXorRR(t *testing.T) {
	got := x86.EncodeXorRR(x86.RAX, x86.RCX)
	if len(got) != 3 || got[1] != 0x31 {
		t.Errorf("XOR RAX, RCX = %X", got)
	}
}

func TestEncodeXorZero(t *testing.T) {
	// XOR EAX, EAX → 31 C0 (2 bytes, no REX)
	got := x86.EncodeXorZero(x86.RAX)
	if len(got) != 2 || got[0] != 0x31 {
		t.Errorf("XOR EAX,EAX = %X", got)
	}
}

func TestEncodeJmpRel32(t *testing.T) {
	got := x86.EncodeJmpRel32(0x100)
	if len(got) != 5 || got[0] != 0xE9 {
		t.Errorf("JMP rel32 = %X", got)
	}
}

func TestEncodeJccRel32(t *testing.T) {
	got := x86.EncodeJccRel32(x86.CCE, 0x42)
	if len(got) != 6 || got[0] != 0x0F || got[1] != 0x84 {
		t.Errorf("JE rel32 = %X", got)
	}
}

func TestEncodeCallRel32(t *testing.T) {
	got := x86.EncodeCallRel32(0)
	if len(got) != 5 || got[0] != 0xE8 {
		t.Errorf("CALL rel32 = %X", got)
	}
}

func TestEncodeSetCC(t *testing.T) {
	got := x86.EncodeSetCC(x86.CCE, x86.RAX)
	found0F := false
	for _, b := range got {
		if b == 0x0F { found0F = true }
	}
	if !found0F {
		t.Errorf("SETE RAX = %X, expected 0x0F prefix", got)
	}
}

func TestEncodeMovLoad(t *testing.T) {
	got := x86.EncodeMovLoad(x86.RAX, x86.RBP, -8)
	if len(got) < 3 {
		t.Errorf("MOV RAX, [RBP-8] too short: %X", got)
	}
	// Should contain 0x8B opcode
	if got[1] != 0x8B {
		t.Errorf("expected opcode 0x8B, got 0x%02X in %X", got[1], got)
	}
}

func TestEncodeMovStore(t *testing.T) {
	got := x86.EncodeMovStore(x86.RBP, -16, x86.RCX)
	if len(got) < 3 {
		t.Errorf("MOV [RBP-16], RCX too short: %X", got)
	}
	if got[1] != 0x89 {
		t.Errorf("expected opcode 0x89, got 0x%02X in %X", got[1], got)
	}
}

func TestCondCode_String(t *testing.T) {
	if x86.CCE.String() != "e" { t.Errorf("CCE = %q", x86.CCE.String()) }
	if x86.CCL.String() != "l" { t.Errorf("CCL = %q", x86.CCL.String()) }
	if x86.CCG.String() != "g" { t.Errorf("CCG = %q", x86.CCG.String()) }
}

// --------------------------------------------------------------------------
// Encoding with extended registers (R8-R15)
// --------------------------------------------------------------------------

func TestEncodeMovRR_Extended(t *testing.T) {
	got := x86.EncodeMovRR(x86.R8, x86.R9)
	// Should have REX prefix with both R and B bits
	if len(got) != 3 {
		t.Errorf("MOV R8, R9 = %X (len=%d)", got, len(got))
	}
	rex := got[0]
	if rex&x86.REXW == 0 { t.Error("expected REX.W") }
}

func TestEncodePushPop_Extended(t *testing.T) {
	push := x86.EncodePush(x86.R12)
	if len(push) != 2 {
		t.Errorf("PUSH R12 = %X", push)
	}
	pop := x86.EncodePop(x86.R12)
	if len(pop) != 2 {
		t.Errorf("POP R12 = %X", pop)
	}
}
