package codegen_test

import (
	"encoding/binary"
	"testing"

	"github.com/axiom-lang/axiom/codegen"
	"github.com/axiom-lang/axiom/codegen/native/x86"
)

// --------------------------------------------------------------------------
// p12-t06: Linker Tests
// --------------------------------------------------------------------------

func TestCOFF_ValidHeader(t *testing.T) {
	w := x86.NewCOFFWriter()
	w.SetText([]byte{0xC3})
	w.AddSymbol(x86.ELF64Sym{
		Name:    "main",
		Value:   0,
		Size:    1,
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 1,
	})

	obj := w.Serialize()
	if len(obj) < 20 {
		t.Fatalf("COFF too small: %d bytes", len(obj))
	}

	machine := binary.LittleEndian.Uint16(obj[0:2])
	if machine != x86.IMAGE_FILE_MACHINE_AMD64 {
		t.Errorf("machine = 0x%04X, expected 0x8664", machine)
	}

	numSections := binary.LittleEndian.Uint16(obj[2:4])
	if numSections != 1 {
		t.Errorf("numSections = %d, expected 1", numSections)
	}
}

func TestMachO_ValidHeader(t *testing.T) {
	w := x86.NewMachOWriter()
	w.SetText([]byte{0xC3})
	w.AddSymbol(x86.ELF64Sym{
		Name:    "main",
		Value:   0,
		Size:    1,
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 1,
	})

	obj := w.Serialize()
	if len(obj) < 32 {
		t.Fatalf("Mach-O too small: %d bytes", len(obj))
	}

	magic := binary.LittleEndian.Uint32(obj[0:4])
	if magic != x86.MH_MAGIC_64 {
		t.Errorf("magic = 0x%08X, expected 0xFEEDFACF", magic)
	}

	cpuType := binary.LittleEndian.Uint32(obj[4:8])
	if cpuType != x86.CPU_TYPE_X86_64 {
		t.Errorf("cpuType = 0x%08X, expected x86_64", cpuType)
	}

	fileType := binary.LittleEndian.Uint32(obj[12:16])
	if fileType != x86.MH_OBJECT {
		t.Errorf("fileType = %d, expected MH_OBJECT (1)", fileType)
	}
}

func TestDemangleDisplay(t *testing.T) {
	tests := []struct {
		mangled  string
		expected string
	}{
		{"_AX_math_add_ii_i", "math::add(i32, i32) -> i32"},
		{"_AX_main_main_v_v", "main::main() -> void"},
		{"_AX_io_write_tl_o", "io::write(str, i64) -> bool"},
	}

	for _, tt := range tests {
		got := codegen.DemangleDisplay(tt.mangled)
		if got != tt.expected {
			t.Errorf("DemangleDisplay(%q) = %q, expected %q", tt.mangled, got, tt.expected)
		}
	}
}

func TestDemangleDisplay_NonMangled(t *testing.T) {
	// Non-mangled names should be returned as-is
	got := codegen.DemangleDisplay("printf")
	if got != "printf" {
		t.Errorf("expected 'printf', got %q", got)
	}
}

func TestIncrementalState(t *testing.T) {
	state := &codegen.IncrementalState{
		ObjectFiles: map[string]uint64{
			"foo.o": 12345,
		},
	}

	if !state.NeedsRelink("foo.o", 99999) {
		t.Error("changed hash should need relink")
	}
	if state.NeedsRelink("foo.o", 12345) {
		t.Error("same hash should not need relink")
	}
	if !state.NeedsRelink("bar.o", 0) {
		t.Error("new file should need relink")
	}
}
