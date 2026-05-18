# p12-t05: Incremental Linker

## Purpose
Implement a minimal custom linker that combines AXIOM object files and runtime library into a single executable, supporting incremental relinking for fast development iteration.

## Context
While AXIOM can delegate to system linkers (ld, lld) for production builds, a custom linker enables faster development builds by relinking only changed object files. The AXIOM linker handles ELF64 on Linux, with PE-COFF and Mach-O as follow-up targets. Incremental linking is critical for the self-hosting compiler (phase 18) where recompile-relink cycles must be fast.

## Inputs
- `[]string` — list of `.o` file paths to link
- `[]string` — shared library dependencies (`-l` flags)
- Entry point symbol name (default: `main`)
- Output file path and format (ELF executable)

## Outputs
- `linker/linker.go` — `AxiomLinker` type
- ELF64 executable (type `ET_EXEC`) or shared library (`ET_DYN`)

## Dependencies
- p11-t11: relocation-backpatcher — relocation types and resolution
- p11-t12: elf64-emitter — input .o file parsing
- p12-t01: symbol-mangling — symbol resolution by mangled name

## Subsystems Affected
- Compiler driver: `axc link` command invokes custom linker
- CI pipeline: builds AXIOM test executables without external linker dependency

## Detailed Requirements

```go
type AxiomLinker struct {
    InputFiles  []string
    LibPaths    []string
    EntryPoint  string
    OutputPath  string
    OutputType  OutputType  // ET_EXEC, ET_DYN
}

type LinkerSymbol struct {
    Name    string
    Section int
    Offset  uint64
    Size    uint64
    Defined bool
}

func (l *AxiomLinker) Link() error
func (l *AxiomLinker) loadObject(path string) (*ObjectFile, error)
func (l *AxiomLinker) resolveSymbols() error
func (l *AxiomLinker) applyRelocations() error
func (l *AxiomLinker) layoutSections() error
func (l *AxiomLinker) writeExecutable() error
```

Linking algorithm:
1. Load all .o files, parse ELF, extract sections and symbol tables.
2. Build global symbol table: for each symbol, find its defining object.
3. Detect undefined symbols → error with list.
4. Lay out sections: merge all .text sections, then .rodata, .data, .bss.
5. Assign virtual addresses (base: 0x400000 for ET_EXEC).
6. Apply relocations: for each reloc entry, compute target VA, patch bytes.
7. Write ELF executable with PT_LOAD segments and program headers.

Incremental mode:
- Cache object file content hashes.
- On relink: reload only changed objects; reuse previous section layout for unchanged sections.
- Patch only changed relocations.

## Implementation Steps

1. Create `linker/linker.go`.
2. Implement ELF .o parser (read sections, symbols, relocations).
3. Implement symbol resolution with duplicate/undefined error reporting.
4. Implement section layout with address assignment.
5. Implement relocation application (PC-relative, absolute).
6. Implement ELF ET_EXEC output with PT_LOAD program headers.
7. Implement incremental hash cache.
8. Write integration tests.

## Test Plan
- `TestLinkTwoObjects`: link a.o + b.o where a calls b → correct executable
- `TestLinkUndefined`: undefined symbol → error message with symbol name
- `TestLinkDuplicate`: two definitions of same symbol → error
- `TestLinkIncremental`: change a.o, relink → only a.o reprocessed (timed)
- `TestLinkHelloWorld`: link hello.o + runtime.o → executable that prints "hello"

## Validation Checklist
- [ ] All undefined symbols reported before linking fails
- [ ] PC-relative relocations correct (test with known offset)
- [ ] ELF executable has PT_LOAD segments
- [ ] Incremental relink faster than full relink

## Acceptance Criteria
- AXIOM hello-world links in < 10ms without external linker

## Definition of Done
- [ ] `linker/linker.go` implemented
- [ ] Integration tests pass
- [ ] Incremental linking demonstrated

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Complex relocation types (GOT, PLT) | Start with R_X86_64_PC32 and R_X86_64_64 only; add PLT later |
| ELF program header layout errors → SIGSEGV | Unit test PT_LOAD with known VA ranges; use /proc/self/maps for verification |

## Future Follow-up Tasks
- PE-COFF linker for Windows
- LTO (link-time optimization) support
- p18: self-hosting compiler uses custom linker exclusively
