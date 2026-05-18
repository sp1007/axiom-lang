# p11-t11: Relocation Back-Patcher

## Purpose
Resolve unresolved symbol references in emitted machine code by recording relocations and applying fixups once all symbols and their addresses are known.

## Context
When the emitter encounters a CALL to an external symbol or a reference to a global variable, it cannot know the final address at emission time. Instead it records a `Relocation` entry and patches the instruction bytes during the linking phase (or object file writing phase for ELF relocations). This module handles both intra-object fixups (already resolved) and inter-object relocations (written into the `.rela.text` section).

## Inputs
- `[]byte` — raw `.text` section bytes from p11-t10
- `[]Relocation` — unresolved references from p11-t10
- `SymbolTable` — maps symbol names to offsets within the current object

## Outputs
- `codegen/native/x86/reloc.go` — relocation types and back-patcher
- Updated `[]byte` with intra-object references resolved
- `[]ELFRelocation` or `[]COFFRelocation` for external symbols

## Dependencies
- p11-t10: x86-machine-code-emitter — produces Relocation list
- p11-t12: elf64-emitter — consumes ELFRelocation entries

## Subsystems Affected
- Object file emitter: writes relocation table into object file sections
- Linker (phase 12): applies remaining relocations at link time

## Detailed Requirements

```go
type RelocType uint8
const (
    R_X86_64_PC32   RelocType = 1  // 32-bit PC-relative (CALL, JMP)
    R_X86_64_PLT32  RelocType = 2  // PLT-relative (external functions)
    R_X86_64_32     RelocType = 3  // 32-bit absolute
    R_X86_64_64     RelocType = 4  // 64-bit absolute (data refs)
    R_X86_64_GOTPCREL RelocType = 5 // GOT-PC-relative
)

type Relocation struct {
    Offset   int
    Symbol   string
    Type     RelocType
    Addend   int32
}

type BackPatcher struct {
    Text    []byte
    Relocs  []Relocation
    Symbols map[string]int  // local symbol → offset in .text
}

func (bp *BackPatcher) ResolveLocal() []Relocation  // patch intra-object refs, return remaining
func (bp *BackPatcher) ToELFRelocs(symtab *ELFSymTab) []ELFRela
func (bp *BackPatcher) ToCOFFRelocs(symtab *COFFSymTab) []COFFReloc
```

Resolution algorithm:
1. For each Relocation:
   - If `Symbol` found in local `Symbols` map → patch bytes at `Offset` with PC-relative offset.
   - Else → forward to object file as external relocation entry.
2. PC-relative patch: `int32(target - (patchSite + 4))` written as little-endian at `Text[Offset]`.
3. For ELF: emit `Elf64_Rela{r_offset, r_info=(symIdx << 32 | type), r_addend}`.

## Implementation Steps

1. Create `codegen/native/x86/reloc.go`.
2. Define `RelocType` constants matching ELF and COFF reloc types.
3. Implement `ResolveLocal()` — patch intra-object call/jump targets.
4. Implement PC-relative offset calculation with addend.
5. Implement `ToELFRelocs()` — produce Elf64_Rela entries.
6. Implement `ToCOFFRelocs()` — produce IMAGE_RELOCATION entries.
7. Write unit tests: patch a CALL instruction with known target offset.

## Test Plan
- `TestResolveLocalCall`: CALL to function in same object → bytes patched correctly
- `TestExternalReloc`: CALL to `printf` → produces ELF R_X86_64_PLT32 entry
- `TestPCRelativeCalc`: verify `target - (site+4)` formula
- `TestUnresolvedSymbol`: unknown symbol → error, not silent zero

## Validation Checklist
- [ ] Local symbols resolved in-place before object file is written
- [ ] External symbols produce correct ELF/COFF relocation entries
- [ ] PC-relative offset formula correct (accounts for instruction length)
- [ ] Addend field set correctly per relocation type

## Acceptance Criteria
- Object file with external `printf` call passes `readelf -r` inspection showing correct PLT32 reloc

## Definition of Done
- [ ] `codegen/native/x86/reloc.go` implemented
- [ ] Unit tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Off-by-4 in PC-relative offset (forget +4 for instruction length) | Unit test with known distance, assert exact byte values |

## Future Follow-up Tasks
- p11-t12: ELF64 emitter writes relocation table into .rela.text section
- p12-t04: linker resolves remaining external relocations at link time
