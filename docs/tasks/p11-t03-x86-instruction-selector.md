# p11-t03: x86-64 Instruction Selector

## Purpose
Map AIR opcodes to x86-64 machine instructions, selecting the best instruction for each AIR operation while handling addressing modes and peephole optimizations like macro-op fusion.

## Context
Instruction selection bridges AIR (target-independent) and machine instructions (target-specific). Each AIR instruction may map to one or more x86-64 instructions. The selector uses a pattern-matching approach: for common patterns like `cmp + branch`, it emits a fused compare-and-branch pair.

## Inputs
- `AirFunc` — instructions to select for
- x86 encoding tables from p11-t02

## Outputs
- `codegen/native/x86/selector.go`
- `[]MachInst` per basic block (machine instructions before register assignment)

## Dependencies
- p11-t02: x86-instruction-set — encoding tables
- p09-t01: air-instruction-set — AIR opcodes

## Subsystems Affected
- Register allocator (p11-t05): operates on MachInst with VRegs
- Machine code emitter (p11-t10): emits actual bytes from MachInst

## Detailed Requirements

```go
type MachOpKind uint16  // x86 opcode identifier
type OperandKind uint8
const (OpKindReg OperandKind = iota; OpKindImm; OpKindMem; OpKindLabel)

type MachOperand struct {
    Kind  OperandKind
    VReg  uint32   // virtual register (before alloc) or PhysReg (after)
    Imm   int64    // for OpKindImm
    Base  uint32   // for OpKindMem: base VReg
    Disp  int32    // for OpKindMem: displacement
    Scale uint8    // for SIB: 1, 2, 4, 8
}

type MachInst struct {
    Op      MachOpKind
    Dst     MachOperand
    Src1    MachOperand
    Src2    MachOperand
    Size    uint8  // operand size: 1, 2, 4, 8 bytes
}
```

AIR → x86 mapping (key patterns):
- `OpIConst → MOV reg, imm64`
- `OpIAdd → ADD r64, r64`
- `OpISub → SUB r64, r64`
- `OpIMul → IMUL r64, r64`
- `OpIDiv → IDIV r64` (with CDQ/CQO for sign extension into RDX)
- `OpIConst 0 → XOR reg, reg` (optimization: XOR is shorter and sets flags)
- `OpLoad → MOV r64, [base+disp]`
- `OpStore → MOV [base+disp], r64`
- `OpGEP → LEA r64, [base+disp*scale]`
- `OpCall → CALL rel32` (or `CALL r64` for indirect)
- `OpReturn → RET` (after moving return value to RAX/XMM0)
- `OpJump → JMP rel32`
- `OpBranch (cmp + jcc) → CMP r64, r64; Jcc rel32` (fused pair)
- `OpAlloc → CALL _AX_ax_alloc`
- `OpDeref → CALL _AX_ax_deref` (gen_id check)
- `OpSIMDAdd (width=8) → VADDPS ymm, ymm, ymm` (AVX2)

## Implementation Steps

1. Create `codegen/native/x86/selector.go`.
2. Implement `Select(fn *AirFunc, target Target) [][]MachInst` (per-block).
3. Handle each AIR opcode with a case in a large switch.
4. Implement branch fusion: scan for `OpICmp` followed by `OpBranch` → emit single fused CMP+Jcc.
5. Implement `XOR reg,reg` optimization for iconst 0.
6. Implement IDIV: emit `CQO; IDIV src`.
7. Write unit tests per opcode pattern.

## Test Plan
- `TestSelectAdd`: `OpIAdd %0, %1` → `ADD r0_vreg, r1_vreg`
- `TestSelectConst`: `OpIConst 0` → `XOR vreg, vreg`
- `TestSelectBranchFusion`: `OpICmpLt + OpBranch` → `CMP + JL`
- `TestSelectDiv`: `OpIDiv` → `CQO; IDIV`
- `TestSelectLoad`: `OpLoad %addr, 0` → `MOV vreg, [addr+0]`

## Validation Checklist
- [ ] All AIR opcodes have a selection rule
- [ ] CMP+Jcc fusion applied for all compare+branch patterns
- [ ] XOR-zeroing applied for iconst 0
- [ ] IDIV includes CQO sign extension
- [ ] MachInst uses VRegs (before register allocation)

## Acceptance Criteria
- All AIR instructions mapped to valid x86-64 MachInsts
- Fused CMP+Jcc produces single compare-and-branch instead of two instructions

## Definition of Done
- [ ] `codegen/native/x86/selector.go` implemented
- [ ] Unit tests pass
- [ ] AIR for hello world → valid MachInst sequence

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| IDIV semantics (quotient in RAX, remainder in RDX) confuses caller | Explicitly handle RAX/RDX clobbering in selector |

## Future Follow-up Tasks
- p11-t04: liveness-analysis operates on MachInst VRegs
- p11-t10: x86-machine-code-emitter converts MachInst to bytes
