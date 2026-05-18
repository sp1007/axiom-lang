# p11-t09: Win64 ABI Implementation

## Purpose
Implement the Windows x64 calling convention for parameter passing, return values, shadow space, and callee-saved XMM registers — required for AXIOM programs targeting Windows.

## Context
The Win64 ABI differs significantly from System V: only 4 integer registers (RCX, RDX, R8, R9), a mandatory 32-byte shadow space, and XMM6-XMM15 are callee-saved. AXIOM on Windows must follow this ABI to call system APIs and C runtime functions.

## Inputs
- Function call sites in MachInst stream
- TypeInfo for parameter and return types
- Physical register assignments from p11-t05

## Outputs
- `codegen/native/x86/abi_win64.go` — Win64 ABI parameter/return marshaling

## Dependencies
- p11-t07: x86-stack-frame — shadow space allocation in frame
- p11-t05: linear-scan-regalloc — physical register assignments
- p11-t08: x86-abi-sysv — reference implementation for comparison

## Subsystems Affected
- Function calls: Win64 argument setup and return value handling
- Stack frame: 32-byte shadow space mandatory before every CALL

## Detailed Requirements

Win64 ABI rules:
- Integer arguments: RCX, RDX, R8, R9 (first 4 only)
- Float arguments: XMM0, XMM1, XMM2, XMM3 (first 4, positional — slot shared with int)
- Stack arguments: 5th+ arg on stack, 8-byte slots, right-to-left
- Shadow space: 32 bytes ALWAYS allocated before CALL (even with 0 args)
- Return value: RAX (integer ≤ 8 bytes), XMM0 (float/double)
- Caller-saved: RAX, RCX, RDX, R8, R9, R10, R11, XMM0-XMM5
- Callee-saved: RBX, RBP, RDI, RSI, R12, R13, R14, R15, XMM6-XMM15

```go
var Win64IntArgs  = []X86Reg{RCX, RDX, R8, R9}
var Win64FloatArgs = []X86Reg{XMM0, XMM1, XMM2, XMM3}
var Win64CalleeSaved = []X86Reg{RBX, RBP, RDI, RSI, R12, R13, R14, R15}
var Win64CalleeSavedXMM = []X86Reg{XMM6, XMM7, XMM8, XMM9, XMM10, XMM11, XMM12, XMM13, XMM14, XMM15}
const Win64ShadowSpace = 32

type Win64ABI struct{}

func (a *Win64ABI) EmitCallSetup(args []TypedReg, stackFrame *StackFrame) []MachInst
func (a *Win64ABI) EmitReturnFetch(retType uint32) []MachInst
func (a *Win64ABI) CalleeSavedRegs() []X86Reg
func (a *Win64ABI) CalleeSavedXMMRegs() []X86Reg
func (a *Win64ABI) ShadowSpaceBytes() int
```

Argument passing (positional — slot index matters, not type count):
- Slot 0: if integer → RCX; if float → XMM0
- Slot 1: if integer → RDX; if float → XMM1
- Slot 2: if integer → R8;  if float → XMM2
- Slot 3: if integer → R9;  if float → XMM3
- Slot 4+: push to stack at [RSP + 32 + (slot-4)*8]

Shadow space: SUB RSP, 32 before every CALL (part of frame or explicit allocation).

XMM callee-saved: if XMM6-XMM15 used, emit MOVAPS [rbp-offset], XMMn in prologue.

## Implementation Steps

1. Create `codegen/native/x86/abi_win64.go`.
2. Implement `EmitCallSetup()` — positional slot assignment (RCX/XMM0 for slot 0, etc.).
3. Implement shadow space: SUB RSP, 32 before CALL; ADD RSP, 32 after.
4. Implement stack arg placement for 5th+ args.
5. Implement `EmitReturnFetch()` — MOV from RAX or MOVSD from XMM0.
6. Implement `CalleeSavedXMMRegs()` — list XMM6-XMM15 for prologue save.
7. Implement `EmitFuncEntry()` — receive args from Win64 regs into VRegs.
8. Test with Windows system calls and C runtime calls.

## Test Plan
- `TestWin64IntArgs`: 4 int args → RCX, RDX, R8, R9
- `TestWin64FloatArg`: float in slot 1 → XMM1 (not XMM0)
- `TestWin64MixedArgs`: (int, float, int, float) → (RCX, XMM1, R8, XMM3)
- `TestWin64ShadowSpace`: CALL preceded by SUB RSP, 32
- `TestWin64StackArg`: 5th arg → [RSP+32]
- `TestWin64CalleeSavedXMM`: function using XMM7 → saved/restored

## Validation Checklist
- [ ] Positional slot assignment (not type-counted)
- [ ] Shadow space always 32 bytes before CALL
- [ ] 5th+ args at [RSP + 32 + (n-4)*8]
- [ ] XMM6-XMM15 saved in prologue if used
- [ ] Return in RAX or XMM0

## Acceptance Criteria
- AXIOM function calling `WriteFile` Windows API passes correct arguments

## Definition of Done
- [ ] `codegen/native/x86/abi_win64.go` implemented
- [ ] Unit tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Shadow space forgotten → stack corruption | Assert SUB RSP, 32 present before every CALL in verifier |
| XMM callee-saved not restored → corruption across calls | Save/restore all used XMM6-XMM15 in prologue/epilogue |

## Future Follow-up Tasks
- p11-t10: machine code emitter uses ABI to emit correct call sequences
