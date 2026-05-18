# p11-t01: Target Triple Abstraction

## Purpose
Implement the target triple abstraction that describes the compilation target (architecture, OS, ABI), enabling cross-compilation and target-specific code paths throughout the backend.

## Context
The target triple (e.g., `x86_64-linux-gnu`, `aarch64-apple-macos13`, `riscv64-linux-gnu`) determines: pointer size, register count, calling convention, system call interface, and binary format. All backend code must consult the Target struct rather than hardcoding assumptions.

## Inputs
- CLI: `--target=TRIPLE` flag in `axc build`
- Host detection via `runtime.GOOS`, `runtime.GOARCH`

## Outputs
- `codegen/native/target.go` — Target struct and detection

## Dependencies
- p01-t01: repository-bootstrap — codegen/native/ directory

## Subsystems Affected
- All native backend phases: consult Target for architecture decisions
- Linker: determines output binary format (ELF/PE/Mach-O)
- ABI layer: different calling conventions per target

## Detailed Requirements

```go
type ArchKind uint8
const (ArchX86_64 ArchKind = iota; ArchARM64; ArchRISCV64)

type OSKind uint8
const (OSLinux OSKind = iota; OSWindows; OSmacOS)

type ABIKind uint8
const (ABISysV ABIKind = iota; ABIWin64; ABIAAPCS64; ABIRISCVpsABI)

type Target struct {
    Arch ArchKind
    OS   OSKind
    ABI  ABIKind
}

func (t Target) PointerSize() int  { if t.Arch == ArchRISCV64 { return 8 }; return 8 }
func (t Target) IntRegCount() int  { switch t.Arch { case ArchX86_64: return 16; case ArchARM64: return 31; case ArchRISCV64: return 32 } }
func (t Target) FloatRegCount() int { ... }
func (t Target) BinaryFormat() BinaryFmt { if t.OS==OSLinux||t.OS==OSLinux {return ELF}; ... }
func HostTarget() Target { ... }  // auto-detect from GOOS/GOARCH
func ParseTarget(triple string) (Target, error) { ... }
```

## Implementation Steps

1. Create `codegen/native/target.go`.
2. Implement all types and methods above.
3. Implement `ParseTarget("x86_64-linux-gnu")` string parser.
4. Implement `HostTarget()` using `runtime.GOOS`/`runtime.GOARCH`.
5. Add `--target` flag to `axc build` CLI.
6. Write tests: `TestParseTarget`, `TestHostTarget`, `TestTargetProperties`.

## Test Plan
- `TestParseTarget`: verify parsing of all 9 combinations (3 arch × 3 OS)
- `TestHostTarget`: verify auto-detection works on CI platforms
- `TestTargetProperties`: pointer size, reg count correct per arch

## Validation Checklist
- [ ] All 9 target combinations parseable
- [ ] HostTarget correct on Linux, Windows, macOS CI
- [ ] BinaryFormat returns correct format per OS
- [ ] IntRegCount correct per architecture

## Acceptance Criteria
- `--target=x86_64-linux-gnu` parses correctly and is used by all backend code

## Definition of Done
- [ ] `codegen/native/target.go` implemented
- [ ] `--target` flag working in CLI
- [ ] Unit tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| New targets added later break existing code | Use switch exhaustiveness checking |

## Future Follow-up Tasks
- p11-t02: x86 instruction set uses Target for feature detection
- p13-t01: arm64-instruction-selector uses Target.Arch
