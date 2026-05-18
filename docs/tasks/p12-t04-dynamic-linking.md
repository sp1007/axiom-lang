# p12-t04: Dynamic Linking Support

## Purpose
Enable AXIOM programs to link against shared libraries (`.so`, `.dll`, `.dylib`) at runtime, and to produce AXIOM shared libraries consumable by other languages.

## Context
Dynamic linking allows AXIOM programs to call C libraries (libc, libm, OpenSSL) without statically embedding them. It also allows AXIOM modules to be packaged as shared libraries. This requires PLT/GOT stubs on ELF, import libraries on Windows, and two-level namespace on Mach-O.

## Inputs
- `extern` function declarations in AXIOM source (resolved by semantic layer)
- `#[export]` annotations on AXIOM functions
- Target platform (ELF/PE/Mach-O) from p11-t01

## Outputs
- `codegen/native/dynlink.go` — dynamic linking support
- `.plt` and `.got` sections for ELF shared library calls
- Import library (`.lib`) for Windows DLL exports

## Dependencies
- p11-t11: relocation-backpatcher — PLT32/GOTPCREL relocation types
- p11-t12: elf64-emitter — adds .plt, .got, .dynamic sections
- p12-t01: symbol-mangling — export names

## Subsystems Affected
- ELF emitter: .plt, .got.plt, .dynamic sections
- PE-COFF emitter: import library, export table
- Mach-O emitter: LC_LOAD_DYLIB commands

## Detailed Requirements

**ELF dynamic linking:**
```
.dynamic section: DT_NEEDED entries for each shared library
.dynsym: exported symbols with STB_GLOBAL
.dynstr: dynamic symbol string table
.plt: procedure linkage table stubs
.got.plt: GOT entries for PLT
.rela.plt: R_X86_64_JUMP_SLOT relocations
```

PLT stub (16 bytes):
```asm
__printf_plt:
    JMP  [RIP + printf@GOTPCREL]  ; 6 bytes: FF 25 <32-bit offset>
    PUSH imm32                    ; reloc index
    JMP  PLT0                     ; back to resolver
```

**PE-COFF dynamic linking:**
- Import table: `IMAGE_IMPORT_DESCRIPTOR` per DLL
- Import Address Table (IAT): one slot per imported function
- Call via IAT: `CALL [RIP + func@IAT]`

**Mach-O dynamic linking:**
- `LC_LOAD_DYLIB` per shared library
- Stubs section: indirect call via `__got`
- `CALL _printf` → resolved to `__got.__printf` at load time

```go
type DynLinkSpec struct {
    Library  string   // "libc.so.6", "kernel32.dll", "libSystem.B.dylib"
    Symbols  []string // imported symbol names
}

func (e *ELF64Writer) AddDynamicLinking(specs []DynLinkSpec)
func (e *ELF64Writer) AddExportedSymbols(syms []ExportSym)
```

## Implementation Steps

1. Create `codegen/native/dynlink.go` with platform-agnostic `DynLinkSpec`.
2. For ELF: implement PLT stub generation, GOT.PLT entries, .dynamic section.
3. For PE: implement import table and IAT construction.
4. For Mach-O: implement LC_LOAD_DYLIB load commands.
5. Wire into respective emitters via `AddDynamicLinking()`.
6. Implement `#[export]` annotation handling: add to .dynsym/.export table.
7. Test: AXIOM program calling `printf` via dynamic linking.

## Test Plan
- `TestELFPLTStub`: PLT stub bytes match expected encoding
- `TestELFDynamic`: produced .so has .dynamic section with DT_NEEDED
- `TestDynCallPrintf`: AXIOM hello-world using dynamic printf runs
- `TestExportedSymbol`: `#[export] fn add(...)` visible in shared library

## Validation Checklist
- [ ] PLT stubs 16 bytes each, correct JMP encoding
- [ ] .dynsym has STB_GLOBAL for exported symbols
- [ ] DT_NEEDED entries for all required shared libs
- [ ] `ldd output` shows correct library dependencies

## Acceptance Criteria
- `axc compile --shared math.ax -o libmath.so` → `nm -D libmath.so` shows exported symbols

## Definition of Done
- [ ] `codegen/native/dynlink.go` implemented
- [ ] ELF PLT/GOT generation working
- [ ] hello-world with dynamic printf passes

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| PLT resolver bootstrap (PLT0) complex | Use system dynamic linker (ld-linux.so); emit minimal PLT |
| ASLR/PIE complications | Use -no-pie for initial MVP; add PIE support post-MVP |

## Future Follow-up Tasks
- Position-independent executable (PIE) support
- Weak symbol handling for optional library functions
