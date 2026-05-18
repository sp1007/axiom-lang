# p11-t12: ELF64 Object File Emitter

## Purpose
Produce valid ELF64 relocatable object files (`.o`) from assembled machine code, symbol tables, and relocation entries — enabling linking with system linkers (`ld`, `lld`) on Linux.

## Context
ELF (Executable and Linkable Format) is the standard object file format on Linux. The AXIOM native backend must produce `.o` files that `ld` or `lld` can link into executables. This requires correct ELF header, section headers, symbol table, string table, and relocation sections.

## Inputs
- `[]byte` — `.text` bytes from p11-t10
- `[]ELFRela` — relocations from p11-t11
- Function symbol names and offsets from emitter
- Global data (`.rodata`, `.data`) from constant pool

## Outputs
- `codegen/native/x86/elf64.go` — ELF64 object file writer
- `[]byte` — complete `.o` file content

## Dependencies
- p11-t10: x86-machine-code-emitter — .text bytes
- p11-t11: relocation-backpatcher — .rela.text entries

## Subsystems Affected
- Linker integration: produced .o files linked by system linker
- Debug info: DWARF sections added in p11-t13

## Detailed Requirements

ELF64 structure:
```
ELF Header (64 bytes)
Section: .text       (SHT_PROGBITS, SHF_ALLOC|SHF_EXECINSTR)
Section: .rela.text  (SHT_RELA, link=.symtab, info=.text)
Section: .rodata     (SHT_PROGBITS, SHF_ALLOC)
Section: .data       (SHT_PROGBITS, SHF_ALLOC|SHF_WRITE)
Section: .bss        (SHT_NOBITS, SHF_ALLOC|SHF_WRITE)
Section: .symtab     (SHT_SYMTAB)
Section: .strtab     (SHT_STRTAB)
Section: .shstrtab   (SHT_STRTAB)
Section Header Table
```

```go
type ELF64Writer struct {
    Sections  []*ELFSection
    Symbols   []ELFSym
    Strtab    *StringTable
    Shstrtab  *StringTable
}

type ELFSection struct {
    Name    string
    Type    uint32  // SHT_PROGBITS, SHT_RELA, SHT_SYMTAB, SHT_STRTAB, SHT_NOBITS
    Flags   uint64
    Data    []byte
    Link    uint32  // section index
    Info    uint32  // section index or symbol count
    Align   uint64
    EntSize uint64
}

func (w *ELF64Writer) AddTextSection(code []byte) int
func (w *ELF64Writer) AddRelaSection(textIdx int, relas []ELFRela) int
func (w *ELF64Writer) AddSymbol(name string, sectionIdx int, offset, size uint64, binding, stype uint8)
func (w *ELF64Writer) Serialize() []byte
```

ELF header constants:
- `e_ident`: `\x7fELF` + 64-bit + LE + v1 + Linux ABI
- `e_type = ET_REL (1)` — relocatable
- `e_machine = EM_X86_64 (62)`

Symbol table: local symbols first (STB_LOCAL), then global (STB_GLOBAL). `sh_info` = index of first global symbol.

Relocation entry (Elf64_Rela):
```go
type ELFRela struct {
    Offset uint64
    Info   uint64  // (symIdx << 32) | relocType
    Addend int64
}
```

## Implementation Steps

1. Create `codegen/native/x86/elf64.go`.
2. Define ELF64 header, section header, symbol, and relocation structs with correct field sizes.
3. Implement `StringTable` — growing byte slice with null-terminated strings, returns offsets.
4. Implement section assembly: collect all sections, compute offsets.
5. Implement symbol table serialization: locals first, then globals.
6. Implement `Serialize()` — write ELF header → sections → section header table (little-endian).
7. Test with `readelf -a` to verify structural correctness.
8. Test with `objdump -d` to verify .text disassembly.

## Test Plan
- `TestELFHeader`: magic bytes, machine EM_X86_64, type ET_REL correct
- `TestELFTextSection`: .text data bytes match input
- `TestELFSymbolTable`: exported functions appear as STB_GLOBAL symbols
- `TestELFRelocation`: CALL to printf → R_X86_64_PLT32 in .rela.text
- `TestELFLinkable`: produced .o links with `gcc -nostdlib` without errors

## Validation Checklist
- [ ] ELF magic bytes correct: `\x7fELF\x02\x01\x01\x00`
- [ ] Machine = EM_X86_64 (62)
- [ ] Section header table at end of file
- [ ] .shstrtab section referenced by `e_shstrndx`
- [ ] Symbol table: locals before globals

## Acceptance Criteria
- `readelf -h output.o` shows valid ELF64 relocatable object
- `gcc -x c /dev/null output.o -o prog` links without errors

## Definition of Done
- [ ] `codegen/native/x86/elf64.go` implemented
- [ ] Unit tests pass; produced .o files verified with `readelf`

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Section offset miscalculation → corrupted ELF | Compute all offsets in single pass, verify with readelf |
| String table null termination missing | StringTable always appends \x00 after each string |

## Future Follow-up Tasks
- p11-t13: DWARF line info adds .debug_line section to ELF
- p12-t02: PE-COFF emitter for Windows targets
- p12-t03: Mach-O emitter for macOS targets
