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

func TestOpcodeCount(t *testing.T) {
	if OpcodeCount > 65535 {
		t.Fatalf("OpcodeCount = %d exceeds uint16 max", OpcodeCount)
	}
}

func TestOpcodeNopIsZero(t *testing.T) {
	if OpcodeNop != 0 {
		t.Fatalf("OpcodeNop = %d, want 0 (sentinel value)", OpcodeNop)
	}
}
