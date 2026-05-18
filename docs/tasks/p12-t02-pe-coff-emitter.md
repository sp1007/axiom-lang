# p12-t02: PE-COFF Object File Emitter

## Purpose
Produce valid PE-COFF relocatable object files (`.obj`) for Windows targets, enabling AXIOM native code to be linked with MSVC toolchain, MinGW, or LLVM's `lld-link`.

## Context
PE-COFF (Portable Executable / Common Object File Format) is the Windows standard for object files. It differs from ELF in section layout, relocation types, symbol table format, and string table placement. The Win64 ABI (p11-t09) must be used for all Windows function calls.

## Inputs
- `.text` bytes from x86-64 emitter (p11-t10)
- Win64 relocation entries from p11-t11
- Mangled symbol names from p12-t01
- DWARF/CodeView debug info (optional)

## Outputs
- `codegen/native/x86/coff.go` — PE-COFF object writer
- `[]byte` — complete `.obj` file content

## Dependencies
- p11-t10: x86-machine-code-emitter — .text bytes
- p11-t11: relocation-backpatcher — relocation entries (COFF format)
- p12-t01: symbol-mangling — mangled symbol names
- p11-t09: x86-abi-win64 — Win64 calling convention

## Subsystems Affected
- Windows build pipeline: produces .obj files linkable with lld-link or link.exe

## Detailed Requirements

COFF file structure:
```
COFF File Header (20 bytes)
Optional Header (0 bytes for .obj)
Section Headers (40 bytes each)
Section Data (.text, .rdata, .data, .bss)
Relocation Tables
Symbol Table
String Table
```

```go
type COFFWriter struct {
    Sections []COFFSection
    Symbols  []COFFSymbol
    Strings  []byte
}

type COFFSection struct {
    Name        [8]byte
    VirtSize    uint32
    VirtAddr    uint32  // 0 for .obj
    RawSize     uint32
    RawDataPtr  uint32
    RelocsPtr   uint32
    NumRelocs   uint16
    Characteristics uint32
}

type COFFSymbol struct {
    Name        [8]byte  // or 0+offset into string table
    Value       uint32
    SectionNum  int16
    Type        uint16
    StorageClass uint8
    NumAux      uint8
}

type COFFReloc struct {
    VirtAddr    uint32
    SymTableIdx uint32
    Type        uint16  // IMAGE_REL_AMD64_REL32=4, IMAGE_REL_AMD64_ADDR64=1
}

func (w *COFFWriter) AddSection(name string, flags uint32, data []byte) int
func (w *COFFWriter) AddSymbol(name string, sectionIdx int, offset uint32, external bool) int
func (w *COFFWriter) AddReloc(sectionIdx, offset, symIdx int, relocType uint16)
func (w *COFFWriter) Serialize() []byte
```

Machine type: `IMAGE_FILE_MACHINE_AMD64 (0x8664)`.

Section characteristics:
- `.text`: `0x60500020` (code + execute + read + align16)
- `.rdata`: `0x40400040` (initialized data + read + align8)
- `.data`: `0xC0400040` (initialized data + read + write + align8)

## Implementation Steps

1. Create `codegen/native/x86/coff.go`.
2. Implement COFF header struct with correct field sizes/offsets.
3. Implement section header with characteristics constants.
4. Implement symbol table: short names inline, long names → string table offset.
5. Implement string table (4-byte length prefix + null-separated strings).
6. Implement `Serialize()` — write all components in COFF layout order.
7. Test with `dumpbin /SYMBOLS output.obj` and `lld-link`.

## Test Plan
- `TestCOFFHeader`: machine = AMD64, correct section count
- `TestCOFFTextSection`: .text data bytes present
- `TestCOFFSymbolTable`: exported function symbols visible
- `TestCOFFRelocation`: external call → IMAGE_REL_AMD64_REL32 entry
- `TestCOFFLinkable`: produced .obj links with lld-link into .exe

## Validation Checklist
- [ ] Machine type 0x8664 (AMD64)
- [ ] String table 4-byte length prefix
- [ ] Relocations point to correct symbol indices
- [ ] Section alignments match characteristics flags

## Acceptance Criteria
- `lld-link /entry:main output.obj /out:prog.exe` produces executable

## Definition of Done
- [ ] `codegen/native/x86/coff.go` implemented
- [ ] Unit tests pass; `dumpbin` verification

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Symbol name > 8 bytes requires string table reference | Always use string table for names > 8 chars (offset format) |
| Relocation type mismatch (REL32 vs ADDR64) | Unit test each reloc type with known linker inputs |

## Future Follow-up Tasks
- CodeView debug info for Windows debugger (PDB format)
- p12-t04: dynamic linking .dll exports via PE export table
