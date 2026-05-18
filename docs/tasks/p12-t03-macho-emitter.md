# p12-t03: Mach-O Object File Emitter

## Purpose
Produce valid Mach-O 64-bit relocatable object files (`.o`) for macOS targets, enabling AXIOM native code to be linked with Apple's `ld64` linker or LLVM's `lld`.

## Context
Mach-O (Mach Object) is Apple's binary format used on macOS and iOS. It differs from ELF in using load commands instead of section headers, a two-level namespace for dynamic linking, and different relocation types. macOS uses the System V AMD64 ABI (same as Linux) but Mach-O for the container format.

## Inputs
- `.text` bytes from x86-64 emitter (p11-t10)
- Relocation entries from p11-t11 (converted to Mach-O format)
- Mangled symbol names from p12-t01

## Outputs
- `codegen/native/x86/macho.go` — Mach-O object writer
- `[]byte` — complete Mach-O `.o` file content

## Dependencies
- p11-t10: x86-machine-code-emitter — .text bytes
- p11-t11: relocation-backpatcher — relocation entries
- p12-t01: symbol-mangling — mangled symbol names
- p11-t08: x86-abi-sysv — SysV ABI (macOS uses same ABI as Linux)

## Subsystems Affected
- macOS build pipeline: produces .o files linkable with Apple ld64

## Detailed Requirements

Mach-O 64-bit structure:
```
mach_header_64 (32 bytes)
Load Commands:
  LC_SEGMENT_64 (__TEXT, __text)
  LC_SEGMENT_64 (__DATA, __data)
  LC_SYMTAB
  LC_DYSYMTAB
Section data (__text, __data, __cstring, etc.)
Relocation entries
Symbol table
String table
```

```go
type MachOWriter struct {
    Sections  []MachOSection
    Symbols   []MachONList64
    Strings   []byte
    Relocs    [][]MachOReloc  // per section
}

type MachOSection struct {
    SectName  [16]byte
    SegName   [16]byte  // __TEXT or __DATA
    Addr      uint64
    Size      uint64
    Offset    uint32
    Align     uint32    // log2 alignment
    RelocOff  uint32
    NReloc    uint32
    Flags     uint32
    Data      []byte
}

type MachONList64 struct {
    StrOffset uint32
    Type      uint8   // N_SECT|N_EXT for exported
    Sect      uint8   // section index (1-based)
    Desc      uint16
    Value     uint64
}

type MachOReloc struct {
    Addr    int32   // section-relative offset
    SymNum  uint32  // (symNum:24, pcRel:1, len:2, extern:1, type:4)
}
```

Mach-O magic: `0xFEEDFACF` (MH_MAGIC_64, little-endian).

File type: `MH_OBJECT (0x1)`.

CPU type: `CPU_TYPE_X86_64 (0x01000007)`, subtype: `CPU_SUBTYPE_X86_64_ALL (3)`.

Relocation types: `X86_64_RELOC_BRANCH (2)` for CALL, `X86_64_RELOC_SIGNED (1)` for data.

## Implementation Steps

1. Create `codegen/native/x86/macho.go`.
2. Define Mach-O header, load command, section, nlist64, and reloc structs.
3. Implement `LC_SEGMENT_64` construction for `__TEXT` and `__DATA`.
4. Implement `LC_SYMTAB` pointing to symbol and string tables.
5. Implement symbol table: local symbols first (by convention), then external.
6. Implement relocation table per section.
7. Implement `Serialize()` with correct file layout.
8. Test with `otool -v output.o` and Apple `ld`.

## Test Plan
- `TestMachOHeader`: magic 0xFEEDFACF, cputype CPU_TYPE_X86_64
- `TestMachOTextSection`: __TEXT/__text data present
- `TestMachOSymbols`: exported functions in symbol table with N_EXT flag
- `TestMachOReloc`: CALL → X86_64_RELOC_BRANCH entry
- `TestMachOLinkable`: `ld -r output.o -o linked.o` works

## Validation Checklist
- [ ] Magic bytes 0xFEEDFACF (little-endian 64-bit)
- [ ] MH_OBJECT file type
- [ ] String table null-terminated, starts with \x00
- [ ] Symbol table sorted: local before external
- [ ] Relocation pcRel bit set for CALL

## Acceptance Criteria
- `otool -tv output.o` shows correct disassembly of __text section

## Definition of Done
- [ ] `codegen/native/x86/macho.go` implemented
- [ ] Unit tests pass; `otool` verification

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| macOS code signing requirements for executables | Signing is linker/post-link step; .o files don't need signing |
| Section alignment in Mach-O stricter than ELF | Always align __text to 16 bytes (log2=4) |

## Future Follow-up Tasks
- ARM64 (Apple Silicon) Mach-O target (p13)
- Code signing integration for macOS app bundle
