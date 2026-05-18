# p13-t06: RISC-V psABI Calling Convention

## Purpose
Implement the RISC-V psABI (Processor-Specific Application Binary Interface) calling convention and register allocator for the RV64GC backend, targeting `riscv64-linux-gnu`. This enables correct C interoperability, function calls between AXIOM functions, and stack frame management.

## Context
The RISC-V psABI is the standard ABI for RISC-V Linux platforms, defined by the RISC-V International psABI specification. It specifies:
- Which registers hold function arguments (a0-a7 for integers, fa0-fa7 for floats)
- Which registers are caller-saved (temporary) vs callee-saved
- Stack frame layout and alignment requirements
- How structs are passed (scalar-flattening rules)
- Return value conventions

The RISC-V ABI is simpler than AAPCS64 in some respects (no HFA concept, simpler struct passing) but has its own nuances (the "floating-point calling convention" variant for structs with float members).

## Inputs
- RISC-V register definitions from p13-t05 (`regs.go`)
- AIR function signatures from `ir/air/`
- AXIOM type system types
- RISC-V psABI specification (public, from riscv.org)

## Outputs
- `codegen/native/riscv64/abi.go` — argument classification and assignment
- `codegen/native/riscv64/regalloc.go` — linear-scan register allocator for RISC-V
- `codegen/native/riscv64/frame.go` — stack frame layout and prologue/epilogue
- `codegen/native/riscv64/call_lowering.go` — call/return instruction lowering
- Test file: `codegen/native/riscv64/abi_test.go`

## Dependencies
- p13-t05: RISC-V instruction selector (uses these outputs)

## Subsystems Affected
- `codegen/native/riscv64/` — new files in existing package

## Detailed Requirements

### Register Classification

#### Integer (GPR) Registers
```
Caller-saved (temporaries):   t0-t6 (x5-x7, x28-x31), ra (x1)
Callee-saved:                 s0-s11 (x8, x9, x18-x27)
Argument/return registers:    a0-a7 (x10-x17)
Special:                      zero (x0), sp (x2), gp (x3), tp (x4)
```

#### Floating-Point Registers
```
Caller-saved (FP temporaries): ft0-ft7 (f0-f7), ft8-ft11 (f28-f31)
Callee-saved (FP saved):       fs0-fs11 (f8-f9, f18-f27)
FP Argument/return:            fa0-fa7 (f10-f17)
```

### Argument Passing Rules (Integer ABI)

RISC-V uses a "flattening" approach:
1. Scalars (integer, pointer, f32, f64): passed directly in GPR or FPR
2. Structs: broken into XLEN-sized pieces and passed as if multiple integer args

#### Integer Argument Assignment
```
Available integer arg regs: a0-a7 (8 registers, XLEN=64 bits each)
Available FP arg regs: fa0-fa7 (8 registers)

For each argument (in left-to-right order):
  - Align the arg to its natural alignment (but at least 4 bytes)
  - If the argument is a floating-point scalar (f32 or f64):
      if NSRN < 8: place in fa[NSRN++]
      else: place on stack
  - If the argument is an integer or pointer:
      if NGRN < 8: place in a[NGRN++]
      else: place on stack
  - If the argument is a struct:
      fields are recursively flattened into GPR/FPR slots
      (RISC-V "softfloat calling convention" flattens everything to GPR)
```

#### Floating-Point Struct Passing (FP ABI variant)
Under the RISC-V LP64D ABI (which AXIOM targets):
- A struct with exactly 1 float field: pass field in an FPR
- A struct with exactly 2 fields where both are floats: pass in two FPRs
- A struct with 1 float + 1 integer field (≤ 16 bytes total): pass float in FPR, int in GPR
- Otherwise: flatten to GPR(s)

```go
type RVStructPassMethod int
const (
    PassInGPR  RVStructPassMethod = iota
    PassInFPR
    PassInGPRFPR  // float in FPR, int in GPR
    PassInGPRPair // two GPRs for a 16-byte struct
    PassOnStack
)
```

### Stack Argument Layout
- Arguments that don't fit in registers go on the stack
- Each argument is aligned to its natural alignment (min 8 bytes for XLEN=64)
- Stack grows downward; arguments pushed right-to-left conceptually
- SP must be 16-byte aligned at call site

### Return Value Convention
| Type | Register |
|---|---|
| i8-i64, u8-u64, pointer, bool | a0 (sign/zero-extended to 64 bits) |
| f32 | fa0 |
| f64 | fa0 |
| 128-bit integer | a0 (low), a1 (high) |
| struct ≤ 16 bytes | a0 (and a1 if > 8 bytes) |
| struct with float+int ≤ 16 bytes | fa0 (float), a0 (int) |
| larger struct | caller passes hidden pointer in a0; callee writes result |

### Stack Frame Layout
```
High addresses
+------------------------+  ← caller's SP (16-byte aligned)
| RA save slot           |  8 bytes (x1 / ra)
| FP save slot           |  8 bytes (x8 / s0)
+------------------------+
| Callee-saved GPRs      |  8 bytes each (s1, s2, ... as needed)
| Callee-saved FP regs   |  8 bytes each (fs0, fs1, ... as needed)
+------------------------+
| Spill slots            |  8 bytes each
+------------------------+
| Local variables        |  aligned as needed
+------------------------+
| Outgoing arg area      |  (stack args for calls this function makes)
+------------------------+  ← SP during function body (16-byte aligned)
Low addresses
```

### Prologue/Epilogue
```asm
; Prologue
ADDI sp, sp, -frame_size    ; allocate frame
SD   ra, offset_ra(sp)      ; save return address
SD   s0, offset_s0(sp)      ; save frame pointer
ADDI s0, sp, frame_size     ; set frame pointer (s0 = caller SP)
SD   s1, offset_s1(sp)      ; save s1 if used
; ... save other callee-saved regs

; Epilogue
LD   s1, offset_s1(sp)      ; restore s1
; ... restore other callee-saved regs
LD   s0, offset_s0(sp)      ; restore frame pointer
LD   ra, offset_ra(sp)      ; restore return address
ADDI sp, sp, +frame_size    ; deallocate frame
RET                          ; = JALR x0, 0(ra)
```

### Register Allocator
Use linear-scan allocation, same algorithm as p13-t02 but adapted for RISC-V register conventions:
- Allocatable caller-saved: t0-t6 (7 registers) + a0-a7 (8 registers) = 15
- Allocatable callee-saved: s1-s11 (11 registers)
- Total allocatable GPR: 26 (not including s0=FP, sp, gp, tp, zero, ra)
- Float allocatable: 24 FP regs (ft0-ft7, fa0-fa7, ft8-ft11 = 20 caller-saved + fs0-fs11 = 12 callee-saved)

Prefer caller-saved registers first to avoid unnecessary callee-save overhead in leaf functions.

## Implementation Steps

### Step 1: Register Classification
Create `codegen/native/riscv64/regs.go`:
```go
package riscv64

type Reg uint8
const (
    ZERO Reg = iota; RA; SP; GP; TP
    T0; T1; T2
    S0; S1  // S0 = FP
    A0; A1; A2; A3; A4; A5; A6; A7
    S2; S3; S4; S5; S6; S7; S8; S9; S10; S11
    T3; T4; T5; T6
    // FP registers
    FT0; FT1; FT2; FT3; FT4; FT5; FT6; FT7
    FS0; FS1
    FA0; FA1; FA2; FA3; FA4; FA5; FA6; FA7
    FS2; FS3; FS4; FS5; FS6; FS7; FS8; FS9; FS10; FS11
    FT8; FT9; FT10; FT11
)

var IntArgRegs  = []Reg{A0,A1,A2,A3,A4,A5,A6,A7}
var FPArgRegs   = []Reg{FA0,FA1,FA2,FA3,FA4,FA5,FA6,FA7}
var CalleeSaved = []Reg{S1,S2,S3,S4,S5,S6,S7,S8,S9,S10,S11}
var CallerSaved = []Reg{T0,T1,T2,T3,T4,T5,T6}
```

### Step 2: Argument Classification and Assignment
Create `codegen/native/riscv64/abi.go`:
```go
func AssignArgs(sig *types.FuncType) *ABIAssignment {
    ngrn, nsrn := 0, 0
    stackOffset := 0
    locs := make([]ArgLoc, len(sig.Params))

    for i, param := range sig.Params {
        switch t := param.(type) {
        case *types.F32Type:
            if nsrn < 8 {
                locs[i] = ArgLoc{InReg: true, Reg: FPArgRegs[nsrn], IsFP: true}
                nsrn++
            } else {
                locs[i] = ArgLoc{InReg: false, Offset: stackOffset}
                stackOffset += 8
            }
        case *types.F64Type:
            if nsrn < 8 {
                locs[i] = ArgLoc{InReg: true, Reg: FPArgRegs[nsrn], IsFP: true}
                nsrn++
            } else {
                locs[i] = ArgLoc{InReg: false, Offset: stackOffset}
                stackOffset += 8
            }
        case *types.StructType:
            locs[i] = assignStructArg(t, &ngrn, &nsrn, &stackOffset)
        default:
            if ngrn < 8 {
                locs[i] = ArgLoc{InReg: true, Reg: IntArgRegs[ngrn]}
                ngrn++
            } else {
                sz := align(typeSize(param), 8)
                locs[i] = ArgLoc{InReg: false, Offset: stackOffset}
                stackOffset += sz
            }
        }
    }
    return &ABIAssignment{Args: locs, StackSize: align(stackOffset, 16)}
}
```

### Step 3: Linear-Scan Register Allocator
```go
// Reuse the same linear-scan algorithm from p13-t02 arm64 allocator.
// Key difference: RISC-V has more allocatable GPRs (26 vs ARM64's 26 similarly).
// Prefer caller-saved regs first (t0-t6, a0-a7) to minimize callee-save overhead.

func (a *LinearScanAllocator) preferenceOrder() []Reg {
    return append(
        append([]Reg{}, CallerSaved...),  // t0-t6 first
        CalleeSaved...,                   // s1-s11 second
    )
}
```

### Step 4: Frame Layout and Prologue/Epilogue
```go
func EmitPrologue(frame FrameLayout) []MachineInstr {
    instrs := []MachineInstr{}
    instrs = append(instrs, MachineInstr{Opcode: ADDI, Rd: SP, Rs1: SP, Imm: -int64(frame.TotalSize)})
    instrs = append(instrs, MachineInstr{Opcode: SD, Rs1: SP, Rs2: RA, Imm: int64(frame.RAOffset)})
    instrs = append(instrs, MachineInstr{Opcode: SD, Rs1: SP, Rs2: S0, Imm: int64(frame.FPOffset)})
    instrs = append(instrs, MachineInstr{Opcode: ADDI, Rd: S0, Rs1: SP, Imm: int64(frame.TotalSize)})
    for _, sr := range frame.SavedRegs {
        instrs = append(instrs, MachineInstr{Opcode: SD, Rs1: SP, Rs2: sr.Reg, Imm: int64(sr.Offset)})
    }
    return instrs
}
```

## Test Plan

### Unit Tests
1. `TestIntArgRegsAssignment` — 8 i64 args → a0-a7
2. `TestNinthArgOnStack` — 9th i64 arg → stack offset 0
3. `TestFPArgRegs` — 8 f64 args → fa0-fa7
4. `TestMixedArgs` — `(f64, i32, f64, i32)` → fa0, a0, fa1, a1
5. `TestStructFlattening` — struct{i64, i64} → a0, a1
6. `TestStructFPPlusInt` — struct{f32, i32} → fa0, a0
7. `TestLargeStructPointer` — struct{i64, i64, i64} → pointer in a0
8. `TestReturnF64` — f64 return → fa0
9. `TestFrameSizeAligned` — frame size always multiple of 16
10. `TestPrologueSavesRA` — prologue emits `SD ra, offset(sp)`

### Integration Tests
1. Call C `printf` with i32 arg — verify correct register usage
2. Function returning struct{f64, i32} — verify fa0 + a0

## Validation Checklist
- [ ] Integer args a0-a7 assigned before stack
- [ ] FP args fa0-fa7 assigned independently of integer args
- [ ] Structs with float+int fields use FPR+GPR split
- [ ] RA and S0 always saved in prologue
- [ ] Frame size is 16-byte aligned
- [ ] Callee-saved regs restored in reverse order of save

## Acceptance Criteria
1. ABI matches RISC-V psABI LP64D specification
2. C FFI test: `write(1, "hello\n", 6)` via AXIOM FFI works on RISC-V Linux
3. All unit tests pass
4. Struct passing rules match psABI for single-float, two-float, float+int cases

## Definition of Done
- `abi.go`, `regalloc.go`, `frame.go`, `call_lowering.go` implemented
- Unit tests ≥ 90% coverage
- Integration with `axc build --target=riscv64-linux-gnu` functional
- No silent ABI violations

## Risks & Mitigations
| Risk | Mitigation |
|---|---|
| FP struct passing rules complex | Test each case against C compiler (gcc -march=rv64gc) output |
| ADDI frame allocation out of 12-bit range (large frames) | Use LI + ADD sequence for large frames |
| TP register (x4) used by both runtime and ABI | Reserve TP for actor-local storage; document this constraint |

## Future Follow-up Tasks
- p13-t07: QEMU-based integration tests for cross-compilation
- RISC-V V extension ABI (vector registers) — future task
