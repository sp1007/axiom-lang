# p11-t08: System V AMD64 ABI Implementation

## Purpose
Implement the System V AMD64 calling convention (used on Linux and macOS x86-64) for parameter passing, return values, and register usage.

## Context
The ABI defines which registers hold function arguments, which register holds the return value, and which registers must be preserved across calls. AXIOM functions on Linux/macOS must follow this ABI to interoperate with C code and the operating system.

## Inputs
- Function call sites in MachInst stream
- TypeInfo for parameter and return types
- Physical register assignments from p11-t05

## Outputs
- `codegen/native/x86/abi_sysv.go` — ABI parameter/return marshaling
- MachInst insertions for argument setup and return value retrieval

## Dependencies
- p11-t07: x86-stack-frame — stack argument layout
- p11-t05: linear-scan-regalloc — physical register assignments

## Subsystems Affected
- Function calls: argument setup and return value handling
- Prologue/epilogue: which callee-saved regs to save

## Detailed Requirements

System V AMD64 ABI:
- Integer arguments: RDI, RSI, RDX, RCX, R8, R9 (first 6)
- Float arguments: XMM0-XMM7 (first 8)
- Stack arguments: 7th+ integer arg pushed right-to-left, 8-byte aligned
- Return value: RAX (integer ≤ 8 bytes), RAX:RDX (16-byte struct), XMM0 (float)
- Caller-saved: RAX, RCX, RDX, RSI, RDI, R8, R9, R10, R11, XMM0-XMM15
- Callee-saved: RBX, R12, R13, R14, R15, RBP, XMM6-XMM15 (wait — on SysV: XMM8-XMM15 are callee-saved? Actually on SysV, all XMM are caller-saved. Callee-saved: RBX, R12-R15, RBP)

```go
var SysVIntArgs = []X86Reg{RDI, RSI, RDX, RCX, R8, R9}
var SysVFloatArgs = []X86Reg{XMM0, XMM1, XMM2, XMM3, XMM4, XMM5, XMM6, XMM7}
var SysVCalleeSaved = []X86Reg{RBX, R12, R13, R14, R15}

func (a *SysVABI) EmitCallSetup(args []TypedReg, stackFrame *StackFrame) []MachInst
func (a *SysVABI) EmitReturnFetch(retType uint32) []MachInst
func (a *SysVABI) CalleeSavedRegs() []X86Reg
```

For AXIOM calls (all args known at compile time):
1. For each argument: if integer and argIdx < 6 → MOV SysVIntArgs[argIdx], src_reg; if float → MOVSS/MOVSD XMM(argIdx), src_reg; else push to stack.
2. CALL instruction.
3. After CALL: fetch return from RAX or XMM0 into dest_reg.

## Implementation Steps

1. Create `codegen/native/x86/abi_sysv.go`.
2. Implement `EmitCallSetup()` — emit MOVs for each arg in correct register.
3. Implement `EmitReturnFetch()` — MOV result from RAX/XMM0.
4. Implement `EmitFuncEntry()` — for functions receiving args, move from ABI registers to local VRegs.
5. Implement `CalleeSavedRegs()` — used by stack frame computation.
6. Test with a C-to-AXIOM call and AXIOM-to-C call.

## Test Plan
- `TestSysVArgPassing`: function with 6 int args → first in RDI, second in RSI, ...
- `TestSysVFloatArg`: float arg → XMM0
- `TestSysVStackArg`: 7th int arg → pushed to stack
- `TestSysVReturnInt`: return i32 → value in RAX
- `TestSysVReturnFloat`: return f64 → value in XMM0

## Validation Checklist
- [ ] First 6 int args in RDI/RSI/RDX/RCX/R8/R9
- [ ] Float args in XMM0-XMM7
- [ ] Stack args 8-byte aligned, pushed right-to-left
- [ ] Return in RAX (int) or XMM0 (float)
- [ ] Callee-saved list correct (RBX, R12-R15)

## Acceptance Criteria
- AXIOM function calling `printf("hello\n")` works correctly via SysV ABI

## Definition of Done
- [ ] `codegen/native/x86/abi_sysv.go` implemented
- [ ] Unit tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Variadic functions (printf) need special handling (AL = # XMM args) | Handle in FFI layer; set AL before call to variadic |

## Future Follow-up Tasks
- p11-t09: Win64 ABI for Windows targets
