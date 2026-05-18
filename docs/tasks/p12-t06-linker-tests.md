# p12-t06: Linker Integration Tests

## Purpose
Validate the complete link pipeline (symbol resolution, relocation, section layout, executable generation) with a comprehensive test suite spanning all supported object file formats and target platforms.

## Context
The linker is a critical system-level component where bugs manifest as silent corruption, SIGSEGV, or wrong output rather than compile errors. Integration tests must exercise the full pipeline end-to-end, including cross-format linking (ELF, PE-COFF, Mach-O) and edge cases like circular dependencies, weak symbols, and large binaries.

## Inputs
- Hand-crafted .o files with known content (for precise relocation testing)
- AXIOM-compiled .o files from Phase 11 native backend
- Platform: Linux (ELF), Windows (PE-COFF), macOS (Mach-O)

## Outputs
- `tests/linker/linker_test.go` — linker integration test suite
- `tests/linker/fixtures/` — pre-built .o fixture files for each test

## Dependencies
- p12-t05: incremental-linker — linker under test
- p11-t16: native-differential-tests — provides AXIOM .o files
- p12-t01 through p12-t04: all linker-adjacent components

## Subsystems Affected
- CI pipeline: linker tests run in platform-specific CI jobs

## Detailed Requirements

Test categories:

**Symbol Resolution:**
- Two objects, one defines `add`, other calls `add` → linked executable runs correctly
- Undefined symbol → error listing all undefined names
- Duplicate definition → error with both defining object names
- Weak symbol (lower priority) → strong definition wins

**Relocation:**
- PC-relative CALL → correct 4-byte offset in output
- Absolute 64-bit reference → correct VA in output
- PLT stub relocation → correct stub generated

**Section Layout:**
- All .text sections merged in input order
- .rodata 8-byte aligned
- .bss section not materialized in file (NOBITS)
- Total executable size < sum of .o sizes (BSS not materialized)

**Executable Structure:**
- ELF ET_EXEC: entry point symbol at expected VA
- `./output` runs without SIGSEGV
- `readelf -l output` shows PT_LOAD segments

**Incremental Linking:**
- Changing one .o → only that .o re-read (measure with file access counters)
- Output binary identical to full relink

**Edge Cases:**
- Empty .o file (no sections) → links without error
- .o with only .bss → BSS tracked, no bytes in file
- Very large .o (>64KB .text) → links correctly

```go
func TestLinkerBasicCall(t *testing.T)
func TestLinkerUndefined(t *testing.T)
func TestLinkerDuplicate(t *testing.T)
func TestLinkerPCRelReloc(t *testing.T)
func TestLinkerBSSLayout(t *testing.T)
func TestLinkerIncrementalReuse(t *testing.T)
func TestLinkerHelloWorld(t *testing.T)
```

## Implementation Steps

1. Create `tests/linker/linker_test.go`.
2. Create `tests/linker/fixtures/` with hand-crafted .o files in binary (or generated programmatically).
3. Implement fixture generator using ELF64Writer from p11-t12.
4. Write all test cases with clear failure messages.
5. Add platform guards (ELF tests on Linux only, COFF on Windows only).
6. Add CI job matrix: run linker tests on Linux, Windows, macOS.

## Test Plan
See "Detailed Requirements" — all test functions listed there constitute the test plan.

## Validation Checklist
- [ ] All symbol resolution error cases covered
- [ ] Relocation accuracy verified with hex comparison
- [ ] Executable actually runs (not just links)
- [ ] Incremental mode tested with timing assertions

## Acceptance Criteria
- All linker tests pass on Linux, Windows, macOS in CI

## Definition of Done
- [ ] `tests/linker/linker_test.go` implemented
- [ ] All tests pass on at least Linux
- [ ] CI job added for linker tests

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Platform-specific tests hard to run in CI | Use Docker containers for Linux tests on all platforms |
| Fixture .o files become stale | Regenerate fixtures as part of test setup using ELF64Writer |

## Future Follow-up Tasks
- Fuzz the linker with randomly generated .o files
- Benchmark: AXIOM linker vs lld on 1000-file link
