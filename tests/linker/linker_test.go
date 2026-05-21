package linker_test

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/axiom-lang/axiom/codegen"
	"github.com/axiom-lang/axiom/codegen/native/x86"
)

// TestLinkerBasicCall verifies that the linker can resolve a basic call
// between two object files, merging their text and patching the symbol.
func TestLinkerBasicCall(t *testing.T) {
	tmpDir := t.TempDir()

	// Object 1: contains main, calls target_func
	w1 := x86.NewELF64Writer()
	text1 := []byte{0x90, 0x90, 0x90, 0x90, 0xE8, 0x00, 0x00, 0x00, 0x00}
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
		Name:    "target_func",
		Value:   0,
		Size:    0,
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 0, // Undefined
	})

	w1.AddRelocation(x86.Relocation{
		Offset:  5,
		Kind:    x86.RelocPC32,
		SymName: 2,
		Addend:  -4,
	})

	obj1Path := filepath.Join(tmpDir, "obj1.o")
	if err := os.WriteFile(obj1Path, w1.Serialize(), 0644); err != nil {
		t.Fatal(err)
	}

	// Object 2: contains target_func definition
	w2 := x86.NewELF64Writer()
	text2 := []byte{0xC3}
	w2.SetText(text2)

	w2.AddSymbol(x86.ELF64Sym{
		Name:    "target_func",
		Value:   0,
		Size:    uint64(len(text2)),
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 1,
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

	if err := linker.Link(); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	outBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(outBytes) != len(text1)+len(text2) {
		t.Errorf("expected size %d, got %d", len(text1)+len(text2), len(outBytes))
	}
}

// TestLinkerUndefined verifies that undefined symbols trigger appropriate linker errors.
func TestLinkerUndefined(t *testing.T) {
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
		t.Fatal("expected linker to fail with undefined symbol error")
	}

	expectedErr := "undefined symbol: missing_function"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

// TestLinkerDuplicate verifies that duplicate symbol definitions trigger linker errors.
func TestLinkerDuplicate(t *testing.T) {
	tmpDir := t.TempDir()

	w1 := x86.NewELF64Writer()
	w1.SetText([]byte{0x90, 0xC3})
	w1.AddSymbol(x86.ELF64Sym{
		Name:    "main",
		Section: 1,
		Binding: x86.STB_GLOBAL,
	})
	w1.AddSymbol(x86.ELF64Sym{
		Name:    "duplicate_func",
		Section: 1,
		Binding: x86.STB_GLOBAL,
	})
	obj1Path := filepath.Join(tmpDir, "obj1.o")
	os.WriteFile(obj1Path, w1.Serialize(), 0644)

	w2 := x86.NewELF64Writer()
	w2.SetText([]byte{0xC3})
	w2.AddSymbol(x86.ELF64Sym{
		Name:    "duplicate_func",
		Section: 1,
		Binding: x86.STB_GLOBAL,
	})
	obj2Path := filepath.Join(tmpDir, "obj2.o")
	os.WriteFile(obj2Path, w2.Serialize(), 0644)

	outputPath := filepath.Join(tmpDir, "exec")
	linker := &codegen.AxiomLinker{
		InputFiles: []string{obj1Path, obj2Path},
		OutputPath: outputPath,
	}

	err := linker.Link()
	if err == nil {
		t.Fatal("expected duplicate symbol definition error")
	}

	expectedErr := "duplicate symbol definition: duplicate_func"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

// TestLinkerPCRelReloc verifies that the PC-relative relocation calculation matches
// the standard formula: targetAddress - (pc_addr + 4) + addend.
func TestLinkerPCRelReloc(t *testing.T) {
	tmpDir := t.TempDir()

	// Object 1: contains main, calls target_func
	w1 := x86.NewELF64Writer()
	// call target_func (offset 5)
	text1 := []byte{0x90, 0x90, 0x90, 0x90, 0xE8, 0x00, 0x00, 0x00, 0x00}
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
		Name:    "target_func",
		Value:   0,
		Size:    0,
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 0,
	})

	w1.AddRelocation(x86.Relocation{
		Offset:  5,
		Kind:    x86.RelocPC32,
		SymName: 2,
		Addend:  -4,
	})

	obj1Path := filepath.Join(tmpDir, "obj1.o")
	os.WriteFile(obj1Path, w1.Serialize(), 0644)

	// Object 2: contains target_func definition
	w2 := x86.NewELF64Writer()
	text2 := []byte{0xB8, 0x2A, 0x00, 0x00, 0x00, 0xC3}
	w2.SetText(text2)

	w2.AddSymbol(x86.ELF64Sym{
		Name:    "target_func",
		Value:   0,
		Size:    uint64(len(text2)),
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 1,
	})

	obj2Path := filepath.Join(tmpDir, "obj2.o")
	os.WriteFile(obj2Path, w2.Serialize(), 0644)

	outputPath := filepath.Join(tmpDir, "exec")
	linker := &codegen.AxiomLinker{
		InputFiles: []string{obj1Path, obj2Path},
		OutputPath: outputPath,
	}

	if err := linker.Link(); err != nil {
		t.Fatal(err)
	}

	outBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	// displacement = target (0x400009) - (pc (0x400005) + 4) + addend (-4) = -4
	val := binary.LittleEndian.Uint32(outBytes[5:9])
	if val != 0xFFFFFFFC {
		t.Errorf("expected displacement -4 (0xFFFFFFFC), got 0x%08X", val)
	}
}

// TestLinkerBSSLayout verifies empty object or bss-like sections merge successfully.
func TestLinkerBSSLayout(t *testing.T) {
	tmpDir := t.TempDir()

	w := x86.NewELF64Writer()
	w.SetText([]byte{}) // empty text section

	w.AddSymbol(x86.ELF64Sym{
		Name:    "empty_main",
		Section: 1,
		Binding: x86.STB_GLOBAL,
	})

	objPath := filepath.Join(tmpDir, "obj.o")
	os.WriteFile(objPath, w.Serialize(), 0644)

	outputPath := filepath.Join(tmpDir, "exec")
	linker := &codegen.AxiomLinker{
		InputFiles: []string{objPath},
		OutputPath: outputPath,
	}

	if err := linker.Link(); err != nil {
		t.Fatalf("linking empty section failed: %v", err)
	}

	outBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(outBytes) != 0 {
		t.Errorf("expected empty output size 0, got %d", len(outBytes))
	}
}

// TestLinkerIncrementalReuse verifies that unchanged object files correctly skip relinking.
func TestLinkerIncrementalReuse(t *testing.T) {
	tmpDir := t.TempDir()

	w := x86.NewELF64Writer()
	w.SetText([]byte{0x90, 0xC3})
	w.AddSymbol(x86.ELF64Sym{
		Name:    "main",
		Section: 1,
		Binding: x86.STB_GLOBAL,
	})

	objPath := filepath.Join(tmpDir, "obj.o")
	os.WriteFile(objPath, w.Serialize(), 0644)

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

	// Initial link
	if err := linker.Link(); err != nil {
		t.Fatalf("first link failed: %v", err)
	}

	// Keep track of modification time of outputPath
	stat1, err := os.Stat(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	// Sleep slightly to guarantee mod time difference if written again
	time.Sleep(10 * time.Millisecond)

	// Second link (should skip writing because unchanged)
	if err := linker.Link(); err != nil {
		t.Fatalf("second link failed: %v", err)
	}

	stat2, err := os.Stat(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	if stat1.ModTime() != stat2.ModTime() {
		t.Error("expected incremental link to skip writing the output file")
	}

	// Update object file to trigger relink
	w2 := x86.NewELF64Writer()
	w2.SetText([]byte{0x90, 0x90, 0xC3})
	w2.AddSymbol(x86.ELF64Sym{
		Name:    "main",
		Section: 1,
		Binding: x86.STB_GLOBAL,
	})
	os.WriteFile(objPath, w2.Serialize(), 0644)

	// Third link (should write output file)
	if err := linker.Link(); err != nil {
		t.Fatalf("third link failed: %v", err)
	}

	stat3, err := os.Stat(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	if stat2.ModTime() == stat3.ModTime() {
		t.Error("expected incremental link to rewrite modified output file")
	}
}

// TestLinkerHelloWorld verifies linking with FFI calls (such as printf)
// which are excluded from undefined symbol checks to permit standard runtime bindings.
func TestLinkerHelloWorld(t *testing.T) {
	tmpDir := t.TempDir()

	w := x86.NewELF64Writer()
	text := []byte{0x90, 0xE8, 0x00, 0x00, 0x00, 0x00, 0xC3}
	w.SetText(text)

	w.AddSymbol(x86.ELF64Sym{
		Name:    "main",
		Value:   0,
		Size:    uint64(len(text)),
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 1,
	})

	w.AddSymbol(x86.ELF64Sym{
		Name:    "printf", // whitelisted external FFI symbol
		Value:   0,
		Size:    0,
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 0,
	})

	w.AddRelocation(x86.Relocation{
		Offset:  2,
		Kind:    x86.RelocPC32,
		SymName: 2,
		Addend:  -4,
	})

	objPath := filepath.Join(tmpDir, "hello.o")
	os.WriteFile(objPath, w.Serialize(), 0644)

	outputPath := filepath.Join(tmpDir, "hello_exec")
	linker := &codegen.AxiomLinker{
		InputFiles: []string{objPath},
		OutputPath: outputPath,
	}

	if err := linker.Link(); err != nil {
		t.Fatalf("linking with FFI printf failed: %v", err)
	}
}
