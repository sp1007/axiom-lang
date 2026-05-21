package air

import (
	"testing"
	"unsafe"
)

func TestAirInstSize(t *testing.T) {
	const want = 16
	got := unsafe.Sizeof(AirInst{})
	if got != want {
		t.Fatalf("AirInst size = %d bytes, want %d bytes. "+
			"AirInst layout is FROZEN. Do not add fields without an RFC.", got, want)
	}
}

func TestAirInstFieldOffsets(t *testing.T) {
	var inst AirInst
	base := uintptr(unsafe.Pointer(&inst))
	check := func(name string, got, want uintptr) {
		t.Helper()
		if got != want {
			t.Errorf("AirInst.%s offset = %d, want %d", name, got, want)
		}
	}
	check("Opcode", uintptr(unsafe.Pointer(&inst.Opcode))-base, 0)
	check("TypeID", uintptr(unsafe.Pointer(&inst.TypeID))-base, 2)
	check("Dest", uintptr(unsafe.Pointer(&inst.Dest))-base, 4)
	check("Src1", uintptr(unsafe.Pointer(&inst.Src1))-base, 8)
	check("Src2", uintptr(unsafe.Pointer(&inst.Src2))-base, 12)
}

func TestOpcodeNopIsZero(t *testing.T) {
	if OpNop != 0 {
		t.Fatalf("OpNop = 0x%04X, want 0x0000 (sentinel value)", OpNop)
	}
}

func TestOpcodeBackwardCompat(t *testing.T) {
	// Verify backward-compat aliases match their new names.
	cases := []struct {
		old, new Opcode
		name     string
	}{
		{OpcodeNop, OpNop, "Nop"},
		{OpcodeConst, OpIConst, "Const/IConst"},
		{OpcodeAdd, OpIAdd, "Add/IAdd"},
		{OpcodeSub, OpISub, "Sub/ISub"},
		{OpcodeLoad, OpLoad, "Load"},
		{OpcodeStore, OpStore, "Store"},
		{OpcodeAlloc, OpAlloc, "Alloc"},
		{OpcodeDealloc, OpFree, "Dealloc/Free"},
		{OpcodeCall, OpCall, "Call"},
		{OpcodeReturn, OpReturn, "Return"},
		{OpcodeJump, OpJump, "Jump"},
		{OpcodeBranch, OpBranch, "Branch"},
		{OpcodePhi, OpPhi, "Phi"},
		{OpcodeSpawn, OpSpawn, "Spawn"},
		{OpcodeAwait, OpAwait, "Await"},
		{OpcodeDestroyVal, OpDestroy, "DestroyVal/Destroy"},
	}
	for _, tc := range cases {
		if tc.old != tc.new {
			t.Errorf("backward-compat alias %s: old=0x%04X != new=0x%04X", tc.name, tc.old, tc.new)
		}
	}
}

func TestOpcodeClasses(t *testing.T) {
	if OpAlloc.Class() != 0x01 {
		t.Errorf("OpAlloc.Class() = 0x%02X, want 0x01", OpAlloc.Class())
	}
	if OpIAdd.Class() != 0x02 {
		t.Errorf("OpIAdd.Class() = 0x%02X, want 0x02", OpIAdd.Class())
	}
	if OpJump.Class() != 0x03 {
		t.Errorf("OpJump.Class() = 0x%02X, want 0x03", OpJump.Class())
	}
	if OpSIMDLoad.Class() != 0x04 {
		t.Errorf("OpSIMDLoad.Class() = 0x%02X, want 0x04", OpSIMDLoad.Class())
	}
	if OpComptime.Class() != 0x05 {
		t.Errorf("OpComptime.Class() = 0x%02X, want 0x05", OpComptime.Class())
	}
	if OpNop.Class() != 0x00 {
		t.Errorf("OpNop.Class() = 0x%02X, want 0x00", OpNop.Class())
	}
}

func TestIsMemory(t *testing.T) {
	memOps := []Opcode{OpAlloc, OpFree, OpLoad, OpStore, OpGEP, OpCopy, OpMove, OpMakeRef, OpDeref, OpArenaAlloc, OpDestroy, OpAliasReuse}
	for _, op := range memOps {
		if !op.IsMemory() {
			t.Errorf("%s (0x%04X) should be IsMemory", op.Mnemonic(), op)
		}
	}
	if OpIAdd.IsMemory() {
		t.Error("OpIAdd should not be IsMemory")
	}
}

func TestIsControl(t *testing.T) {
	ctrlOps := []Opcode{OpJump, OpBranch, OpCall, OpReturn, OpPhi, OpLoopBegin, OpLoopEnd, OpSpawn, OpSend, OpRecv, OpAwait}
	for _, op := range ctrlOps {
		if !op.IsControl() {
			t.Errorf("%s (0x%04X) should be IsControl", op.Mnemonic(), op)
		}
	}
	if OpIAdd.IsControl() {
		t.Error("OpIAdd should not be IsControl")
	}
}

func TestIsTerminator(t *testing.T) {
	terms := []Opcode{OpJump, OpBranch, OpReturn}
	for _, op := range terms {
		if !op.IsTerminator() {
			t.Errorf("%s should be terminator", op.Mnemonic())
		}
	}
	nonTerms := []Opcode{OpCall, OpPhi, OpIAdd, OpLoad, OpNop}
	for _, op := range nonTerms {
		if op.IsTerminator() {
			t.Errorf("%s should NOT be terminator", op.Mnemonic())
		}
	}
}

func TestIsBinaryALU(t *testing.T) {
	bins := []Opcode{OpIAdd, OpISub, OpIMul, OpIDiv, OpIMod, OpFAdd, OpFSub, OpFMul, OpFDiv, OpEq, OpNe, OpLt, OpLe, OpGt, OpGe, OpAnd, OpOr, OpXor, OpShl, OpShr}
	for _, op := range bins {
		if !op.IsBinaryALU() {
			t.Errorf("%s should be IsBinaryALU", op.Mnemonic())
		}
	}
	nonBins := []Opcode{OpNeg, OpNot, OpIToF, OpFToI, OpZExt, OpSExt, OpTrunc, OpNop, OpLoad}
	for _, op := range nonBins {
		if op.IsBinaryALU() {
			t.Errorf("%s should NOT be IsBinaryALU", op.Mnemonic())
		}
	}
}

func TestMnemonics(t *testing.T) {
	cases := []struct {
		op   Opcode
		want string
	}{
		{OpNop, "nop"},
		{OpIAdd, "iadd"},
		{OpLoad, "load"},
		{OpStore, "store"},
		{OpReturn, "ret"},
		{OpPhi, "phi"},
		{OpSIMDFMA, "vfma"},
		{OpComptime, "comptime"},
	}
	for _, tc := range cases {
		if got := tc.op.Mnemonic(); got != tc.want {
			t.Errorf("Opcode(0x%04X).Mnemonic() = %q, want %q", tc.op, got, tc.want)
		}
	}
	// Unknown opcode
	if got := Opcode(0x9999).Mnemonic(); got != "???" {
		t.Errorf("unknown opcode mnemonic = %q, want %q", got, "???")
	}
}
