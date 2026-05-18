# p11-t15: Native Backend Integration

## Purpose
Wire all Phase 11 components (instruction selector, register allocator, spill code, stack frame, ABI, emitter, ELF writer, DWARF, .axmeta) into a single coherent native backend pipeline callable from the compiler driver.

## Context
Phases 11-t01 through t14 implement individual components. This task integrates them into a pipeline: AIR Module → ELF64 object file. The compiler driver (`axc`) invokes the native backend for `--target=x86_64-linux` or similar targets.

## Inputs
- `AirModule` from AIR builder (p09)
- Optimized AIR from optimization pipeline (p10)
- `Target` struct from p11-t01
- Compiler flags: `--debug` (enable DWARF), `--opt-level` (O0-O3)

## Outputs
- `codegen/native/backend.go` — `NativeBackend` type orchestrating the pipeline
- `NativeBackend.Compile(mod AirModule, target Target, flags Flags) ([]byte, error)` — returns .o file bytes

## Dependencies
- p11-t01 through p11-t14: all native backend components
- p10-t01: opt-pipeline-manager — receives optimized AIR

## Subsystems Affected
- Compiler driver (`axc`): calls `NativeBackend.Compile()`
- Testing: integration tests compile real AXIOM programs

## Detailed Requirements

```go
type NativeBackend struct {
    Target  Target
    ABI     ABI       // SysVABI or Win64ABI based on target
}

type ABI interface {
    EmitCallSetup(args []TypedReg, frame *StackFrame) []MachInst
    EmitReturnFetch(retType uint32) []MachInst
    CalleeSavedRegs() []X86Reg
    ShadowSpaceBytes() int
}

func NewNativeBackend(target Target) *NativeBackend
func (nb *NativeBackend) Compile(mod *AirModule, flags CompileFlags) ([]byte, error)
```

Pipeline within `Compile()`:
```
AirModule
  → for each AirFunc:
      InstructionSelector.Select(func) → [][]MachInst
      LivenessAnalysis.Compute(blocks) → []LiveInterval
      LinearScanAllocator.Allocate(intervals) → VRegMap
      SpillCodeInserter.Insert(blocks, vrmap) → updated blocks
      StackFrame.Compute(vrmap, spillCount, localBytes) → StackFrame
      Emitter.EmitFunc(func, vrmap, frame) → ([]byte, []Relocation)
      BackPatcher.ResolveLocal() → remaining Relocs
  → ELF64Writer.AddTextSection(allBytes)
  → ELF64Writer.AddSymbols(...)
  → ELF64Writer.AddRelaSection(relocs)
  → if flags.Debug: DWARFWriter.Serialize() → ELF64Writer.AddSection(".debug_line", ...)
  → ELF64Writer.AddSection(".axmeta", SerializeAxMeta(...))
  → ELF64Writer.Serialize() → []byte
```

## Implementation Steps

1. Create `codegen/native/backend.go`.
2. Implement `NewNativeBackend(target)` — select SysVABI or Win64ABI based on target.OS.
3. Implement `Compile()` — run the full pipeline for each function, collect bytes.
4. Concatenate all function code into single `.text` section.
5. Collect all symbols and relocations across functions.
6. Assemble ELF object with all sections.
7. Add integration test: compile `fn add(a, b: i32) -> i32 { a + b }` → produce .o → link → run.

## Test Plan
- `TestNativeBackendAdd`: add(a,b) → ELF .o → link → run → correct result
- `TestNativeBackendCall`: function calling another AXIOM function → correct ABI
- `TestNativeBackendDebug`: with --debug flag → .debug_line section present
- `TestNativeBackendAxmeta`: .axmeta section present in output

## Validation Checklist
- [ ] All phases invoked in correct order
- [ ] ABI selected based on target OS (SysV for Linux/macOS, Win64 for Windows)
- [ ] Debug sections only emitted when --debug flag set
- [ ] Produced .o links successfully

## Acceptance Criteria
- `axc compile --target=x86_64-linux hello.ax -o hello.o && gcc hello.o -o hello && ./hello` works

## Definition of Done
- [ ] `codegen/native/backend.go` implemented
- [ ] Integration test compiles and runs correctly

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Phase ordering error → use-before-define | Enforce ordering in Compile() with explicit intermediate types |
| Cross-function symbol resolution | Collect all symbols before ELF serialization |

## Future Follow-up Tasks
- p11-t16: differential tests compare native vs C backend output
- p12: multi-format object file emission (PE-COFF, Mach-O)
