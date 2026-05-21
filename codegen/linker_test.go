package codegen_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/axiom-lang/axiom/codegen"
	"github.com/axiom-lang/axiom/codegen/native/x86"
)

// --------------------------------------------------------------------------
// p12-t06: Linker Tests
// --------------------------------------------------------------------------

func TestCOFF_ValidHeader(t *testing.T) {
	w := x86.NewCOFFWriter()
	textIdx := w.AddSection(".text", x86.IMAGE_SCN_CNT_CODE|x86.IMAGE_SCN_MEM_EXECUTE|x86.IMAGE_SCN_MEM_READ, []byte{0xC3})
	w.AddSymbol("main", textIdx, 0, true)

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

// TestLinkELF_Relocations verifies that AxiomLinker successfully parses multiple
// ELF64 object files and correctly resolves PC-relative relocations.
func TestLinkELF_Relocations(t *testing.T) {
	tmpDir := t.TempDir()

	// Create ELF Object 1: contains 'main', which makes a call using a PC-relative relocation
	w1 := x86.NewELF64Writer()
	// text size: 9 bytes, e.g., 0xE8 is the relative call opcode
	text1 := []byte{0x90, 0x90, 0x90, 0x90, 0xE8, 0x00, 0x00, 0x00, 0x00}
	w1.SetText(text1)

	// Add global defined symbol 'main' at offset 0 (sym idx 1 in ELF)
	w1.AddSymbol(x86.ELF64Sym{
		Name:    "main",
		Value:   0,
		Size:    uint64(len(text1)),
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 1, // .text
	})

	// Add undefined external symbol 'target_func' (sym idx 2 in ELF)
	w1.AddSymbol(x86.ELF64Sym{
		Name:    "target_func",
		Value:   0,
		Size:    0,
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 0, // Undefined
	})

	// Add PC32 relocation at offset 5 pointing to 'target_func' (sym idx 2)
	w1.AddRelocation(x86.Relocation{
		Offset:  5,
		Kind:    x86.RelocPC32,
		SymName: 2,
		Addend:  -4,
	})

	obj1Bytes := w1.Serialize()
	obj1Path := filepath.Join(tmpDir, "obj1.o")
	if err := os.WriteFile(obj1Path, obj1Bytes, 0644); err != nil {
		t.Fatal(err)
	}

	// Create ELF Object 2: contains 'target_func' definition
	w2 := x86.NewELF64Writer()
	text2 := []byte{0xB8, 0x2A, 0x00, 0x00, 0x00, 0xC3} // mov eax, 42; ret
	w2.SetText(text2)

	// Add global defined symbol 'target_func' at offset 0 (sym idx 1 in ELF)
	w2.AddSymbol(x86.ELF64Sym{
		Name:    "target_func",
		Value:   0,
		Size:    uint64(len(text2)),
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 1, // .text
	})

	obj2Bytes := w2.Serialize()
	obj2Path := filepath.Join(tmpDir, "obj2.o")
	if err := os.WriteFile(obj2Path, obj2Bytes, 0644); err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(tmpDir, "exec")
	linker := &codegen.AxiomLinker{
		InputFiles: []string{obj1Path, obj2Path},
		OutputPath: outputPath,
	}

	if err := linker.Link(); err != nil {
		t.Fatalf("linking failed: %v", err)
	}

	outBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	expectedSize := len(text1) + len(text2)
	if len(outBytes) != expectedSize {
		t.Fatalf("expected output size %d, got %d", expectedSize, len(outBytes))
	}

	// Verify the patched PC-relative displacement
	// targetAddress = 0x400000 + 9 (offset of target_func in merged text) = 0x400009
	// pc = 0x400000 + 5 = 0x400005
	// displacement = targetAddress - (pc + 4) + addend = 0x400009 - 0x400009 - 4 = -4
	// -4 as uint32 is 0xFFFFFFFC
	relVal := binary.LittleEndian.Uint32(outBytes[5:9])
	if relVal != 0xFFFFFFFC {
		t.Errorf("expected patched call target 0xFFFFFFFC (-4), got 0x%08X", relVal)
	}
}

// TestLinkCOFF_Relocations verifies that AxiomLinker successfully parses multiple
// PE/COFF object files and correctly resolves PC-relative relocations.
func TestLinkCOFF_Relocations(t *testing.T) {
	tmpDir := t.TempDir()

	// Create COFF Object 1: contains 'main', which makes a call using a PC-relative relocation
	w1 := x86.NewCOFFWriter()
	text1 := []byte{0x90, 0x90, 0x90, 0x90, 0xE8, 0x00, 0x00, 0x00, 0x00}
	textIdx1 := w1.AddSection(".text", x86.IMAGE_SCN_CNT_CODE|x86.IMAGE_SCN_MEM_EXECUTE|x86.IMAGE_SCN_MEM_READ, text1)

	// Add global defined symbol 'main' (sym idx 0 in COFF)
	mainIdx := w1.AddSymbol("main", textIdx1, 0, true)
	// Add undefined external symbol 'target_func' (sym idx 1 in COFF)
	targetIdx := w1.AddSymbol("target_func", 0, 0, true)

	// Add RelocPC32 (type 4) relocation at offset 5 pointing to targetIdx (1)
	w1.AddReloc(textIdx1, 5, targetIdx, x86.IMAGE_REL_AMD64_REL32)

	obj1Bytes := w1.Serialize()
	obj1Path := filepath.Join(tmpDir, "obj1.obj")
	if err := os.WriteFile(obj1Path, obj1Bytes, 0644); err != nil {
		t.Fatal(err)
	}

	// Create COFF Object 2: contains 'target_func' definition
	w2 := x86.NewCOFFWriter()
	text2 := []byte{0xB8, 0x2A, 0x00, 0x00, 0x00, 0xC3}
	textIdx2 := w2.AddSection(".text", x86.IMAGE_SCN_CNT_CODE|x86.IMAGE_SCN_MEM_EXECUTE|x86.IMAGE_SCN_MEM_READ, text2)

	// Add global defined symbol 'target_func' (sym idx 0 in COFF)
	w2.AddSymbol("target_func", textIdx2, 0, true)

	obj2Bytes := w2.Serialize()
	obj2Path := filepath.Join(tmpDir, "obj2.obj")
	if err := os.WriteFile(obj2Path, obj2Bytes, 0644); err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(tmpDir, "exec.exe")
	linker := &codegen.AxiomLinker{
		InputFiles: []string{obj1Path, obj2Path},
		OutputPath: outputPath,
	}

	if err := linker.Link(); err != nil {
		t.Fatalf("linking failed: %v", err)
	}

	outBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	expectedSize := len(text1) + len(text2)
	if len(outBytes) != expectedSize {
		t.Fatalf("expected output size %d, got %d", expectedSize, len(outBytes))
	}

	// In AxiomLinker, COFF RelocPC32 defaults to addend = -4
	// targetAddress = 0x400000 + 9 = 0x400009
	// pc = 0x400000 + 5 = 0x400005
	// displacement = targetAddress - (pc + 4) + addend = -4 = 0xFFFFFFFC
	relVal := binary.LittleEndian.Uint32(outBytes[5:9])
	if relVal != 0xFFFFFFFC {
		t.Errorf("expected patched call target 0xFFFFFFFC (-4), got 0x%08X", relVal)
	}

	_ = mainIdx // suppress unused var warning
}

// TestLinkUndefined verifies that linking fails if there is an undefined symbol
// that is not whitelisted.
func TestLinkUndefined(t *testing.T) {
	tmpDir := t.TempDir()

	w1 := x86.NewELF64Writer()
	text1 := []byte{0xE8, 0x00, 0x00, 0x00, 0x00}
	w1.SetText(text1)

	w1.AddSymbol(x86.ELF64Sym{
		Name:    "main",
		Value:   0,
		Size:    uint64(len(text1)),
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 1,
	})

	w1.AddSymbol(x86.ELF64Sym{
		Name:    "missing_function", // Undefined
		Value:   0,
		Size:    0,
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 0,
	})

	w1.AddRelocation(x86.Relocation{
		Offset:  1,
		Kind:    x86.RelocPC32,
		SymName: 2,
		Addend:  -4,
	})

	objPath := filepath.Join(tmpDir, "obj.o")
	if err := os.WriteFile(objPath, w1.Serialize(), 0644); err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(tmpDir, "exec")
	linker := &codegen.AxiomLinker{
		InputFiles: []string{objPath},
		OutputPath: outputPath,
	}

	err := linker.Link()
	if err == nil {
		t.Fatal("expected linker to fail with undefined symbol error, but it succeeded")
	}

	expectedErr := "undefined symbol: missing_function"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

// TestLinkDuplicate verifies that linking fails when the same symbol is defined twice.
func TestLinkDuplicate(t *testing.T) {
	tmpDir := t.TempDir()

	// Object 1: defines 'main' and 'foo'
	w1 := x86.NewELF64Writer()
	w1.SetText([]byte{0x90, 0xC3})
	w1.AddSymbol(x86.ELF64Sym{
		Name:    "main",
		Section: 1,
		Binding: x86.STB_GLOBAL,
	})
	w1.AddSymbol(x86.ELF64Sym{
		Name:    "foo",
		Section: 1,
		Binding: x86.STB_GLOBAL,
	})
	obj1Path := filepath.Join(tmpDir, "obj1.o")
	if err := os.WriteFile(obj1Path, w1.Serialize(), 0644); err != nil {
		t.Fatal(err)
	}

	// Object 2: also defines 'foo'
	w2 := x86.NewELF64Writer()
	w2.SetText([]byte{0xC3})
	w2.AddSymbol(x86.ELF64Sym{
		Name:    "foo",
		Section: 1,
		Binding: x86.STB_GLOBAL,
	})
	obj2Path := filepath.Join(tmpDir, "obj2.o")
	if err := os.WriteFile(obj2Path, w2.Serialize(), 0644); err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(tmpDir, "exec")
	linker := &codegen.AxiomLinker{
		InputFiles: []string{obj1Path, obj2Path},
		OutputPath: outputPath,
	}

	err := linker.Link()
	if err == nil {
		t.Fatal("expected linker to fail with duplicate symbol error, but it succeeded")
	}

	expectedErr := "duplicate symbol definition: foo"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

// TestIncrementalLinkerBenchmark measures linking speed and validates
// incremental relinking behaviors under file modification.
func TestIncrementalLinkerBenchmark(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial object file
	w := x86.NewELF64Writer()
	w.SetText([]byte{0x90, 0xC3})
	w.AddSymbol(x86.ELF64Sym{
		Name:    "main",
		Section: 1,
		Binding: x86.STB_GLOBAL,
	})
	objPath := filepath.Join(tmpDir, "obj.o")
	if err := os.WriteFile(objPath, w.Serialize(), 0644); err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(tmpDir, "exec")
	state := &codegen.IncrementalState{
		ObjectFiles: make(map[string]uint64),
		OutputPath:  outputPath,
	}

	linker := &codegen.AxiomLinker{
		InputFiles:  []string{objPath},
		OutputPath:  outputPath,
		Incremental: state,
	}

	// First link: performs compilation and writes the output file
	start := time.Now()
	if err := linker.Link(); err != nil {
		t.Fatalf("first link failed: %v", err)
	}
	firstDuration := time.Since(start)

	// Second link: no changes, should skip linking entirely
	start = time.Now()
	if err := linker.Link(); err != nil {
		t.Fatalf("second link failed: %v", err)
	}
	secondDuration := time.Since(start)

	t.Logf("First link took %v, second (incremental skipped) link took %v", firstDuration, secondDuration)

	// Modify object file
	w2 := x86.NewELF64Writer()
	w2.SetText([]byte{0x90, 0x90, 0xC3})
	w2.AddSymbol(x86.ELF64Sym{
		Name:    "main",
		Section: 1,
		Binding: x86.STB_GLOBAL,
	})
	if err := os.WriteFile(objPath, w2.Serialize(), 0644); err != nil {
		t.Fatal(err)
	}

	// Third link: modified file, should perform full relink
	start = time.Now()
	if err := linker.Link(); err != nil {
		t.Fatalf("third link failed: %v", err)
	}
	thirdDuration := time.Since(start)

	t.Logf("Third link (with modifications) took %v", thirdDuration)

	// Validate output file size is updated (from 2 bytes to 3 bytes)
	outBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(outBytes) != 3 {
		t.Errorf("expected modified output size 3, got %d", len(outBytes))
	}
}

// TestWindowsLinkerCompatibility validates generated COFF structures and fields
// against standard PE/COFF specifications to ensure toolchain compatibility.
func TestWindowsLinkerCompatibility(t *testing.T) {
	w := x86.NewCOFFWriter()
	textIdx := w.AddSection(".text", x86.IMAGE_SCN_CNT_CODE|x86.IMAGE_SCN_MEM_EXECUTE|x86.IMAGE_SCN_MEM_READ, []byte{0xB8, 0x2A, 0x00, 0x00, 0x00, 0xC3})
	w.AddSymbol("main", textIdx, 0, true)

	obj := w.Serialize()
	if len(obj) < 20 {
		t.Fatalf("COFF too small: %d bytes", len(obj))
	}

	machine := binary.LittleEndian.Uint16(obj[0:2])
	if machine != x86.IMAGE_FILE_MACHINE_AMD64 {
		t.Errorf("expected AMD64 machine (0x8664), got 0x%04X", machine)
	}

	numSections := binary.LittleEndian.Uint16(obj[2:4])
	if numSections != 1 {
		t.Errorf("expected 1 section, got %d", numSections)
	}

	symtabOff := binary.LittleEndian.Uint32(obj[8:12])
	symCount := binary.LittleEndian.Uint32(obj[12:16])
	if symCount != 1 {
		t.Errorf("expected 1 symbol, got %d", symCount)
	}

	// Verify section header fields
	secHeaderOff := 20
	secName := string(bytes.TrimRight(obj[secHeaderOff:secHeaderOff+8], "\x00"))
	if secName != ".text" {
		t.Errorf("expected section name '.text', got %q", secName)
	}

	rawSize := binary.LittleEndian.Uint32(obj[secHeaderOff+16 : secHeaderOff+20])
	if rawSize != 6 {
		t.Errorf("expected raw size 6, got %d", rawSize)
	}

	// Check symbol name in symbol table
	symOff := symtabOff
	symName := string(bytes.TrimRight(obj[symOff:symOff+8], "\x00"))
	if symName != "main" {
		t.Errorf("expected symbol name 'main', got %q", symName)
	}
}

// TestExternalLinkerCompatibility invokes standard system linkers (lld-link or link.exe)
// on the generated COFF objects to verify binary compatibility.
func TestExternalLinkerCompatibility(t *testing.T) {
	linkPath, err := exec.LookPath("lld-link")
	if err != nil {
		linkPath, err = exec.LookPath("link")
		if err != nil {
			t.Skip("No lld-link or link.exe found, skipping external linker compatibility test")
			return
		}
	}

	tmpDir := t.TempDir()
	w := x86.NewCOFFWriter()
	// main returning 42
	text := []byte{0xB8, 0x2A, 0x00, 0x00, 0x00, 0xC3}
	textIdx := w.AddSection(".text", x86.IMAGE_SCN_CNT_CODE|x86.IMAGE_SCN_MEM_EXECUTE|x86.IMAGE_SCN_MEM_READ, text)
	w.AddSymbol("main", textIdx, 0, true)

	objBytes := w.Serialize()
	objPath := filepath.Join(tmpDir, "main.obj")
	if err := os.WriteFile(objPath, objBytes, 0644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(tmpDir, "main.exe")
	// Run linker: link.exe /subsystem:console /entry:main /nodefaultlib /out:main.exe main.obj
	cmd := exec.Command(linkPath, "/subsystem:console", "/entry:main", "/nodefaultlib", "/out:"+outPath, objPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("External linker failed: %v\nOutput:\n%s", err, string(output))
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("Executable not generated by external linker: %v", err)
	}
	t.Logf("Successfully linked with external tool: %s", linkPath)
}
