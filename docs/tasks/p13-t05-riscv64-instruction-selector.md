# p13-t05: RISC-V 64-bit Instruction Selector

## Purpose
Implement instruction selection for the RISC-V 64-bit architecture (RV64GC) in `codegen/native/riscv64/`. This translates AIR opcodes into RISC-V machine instructions, targeting the `riscv64-linux-gnu` ELF platform.

## Context
RISC-V is an open ISA with a modular extension system. The baseline for AXIOM is RV64GC:
- RV64I: 64-bit base integer instruction set
- M: Integer multiplication and division
- A: Atomic memory operations
- F: Single-precision floating-point
- D: Double-precision floating-point
- C: Compressed (16-bit) instructions (optional; emitter may or may not use them)

RISC-V is gaining adoption in embedded systems, server silicon (Sophon, Ventana), and RISC-V Linux servers. Supporting RV64GC positions AXIOM well for the growing RISC-V ecosystem. The ISA is simpler and more regular than x86-64 or ARM64, making the selector relatively straightforward.

RISC-V has 32 registers (x0-x31) with the ABI assigning them conventional names. Unlike ARM64, RISC-V has no flags register — comparisons return 0 or 1 in a general register, and conditional branches directly compare two registers.

## Inputs
- Optimized AIR (`OptimizedAIR`) from p11-t01
- Register allocation result (from p13-t06, which this task is paired with)
- Target triple: `riscv64-linux-gnu`
- AIR opcode definitions from `ir/air/opcodes.go`

## Outputs
- `codegen/native/riscv64/instr_selector.go` — instruction selector
- `codegen/native/riscv64/instructions.go` — RISC-V instruction definitions
- `codegen/native/riscv64/regs.go` — register definitions
- `codegen/native/riscv64/riscv64.go` — package entry and backend interface
- Test file: `codegen/native/riscv64/instr_selector_test.go`
- Golden tests: `tests/codegen/riscv64/golden/*.s`

## Dependencies
- p11-t01: AIR definition (input to selector)
- p09-t01: Native backend interface
- p13-t06: RISC-V ABI and register allocator (paired task, register assignment)

## Subsystems Affected
- `codegen/native/riscv64/` — new package
- `codegen/driver.go` — register RISC-V backend for `riscv64-linux-gnu` target
- `linker/elf/` — ELF64 RISC-V relocation support (may need additions)

## Detailed Requirements

### Register Definitions
```
x0  (zero): hardwired 0
x1  (ra):   return address
x2  (sp):   stack pointer
x3  (gp):   global pointer (not used by AXIOM)
x4  (tp):   thread pointer (used by runtime for actor local data)
x5  (t0):   temporary
x6  (t1):   temporary
x7  (t2):   temporary
x8  (s0/fp): saved / frame pointer
x9  (s1):   saved
x10 (a0):   arg/return
x11 (a1):   arg/return
x12 (a2):   arg
x13 (a3):   arg
x14 (a4):   arg
x15 (a5):   arg
x16 (a6):   arg
x17 (a7):   arg
x18 (s2):   saved
x19 (s3):   saved
...
x27 (s11):  saved
x28 (t3):   temporary
x29 (t4):   temporary
x30 (t5):   temporary
x31 (t6):   temporary

f0  (ft0) - f7  (ft7):  FP temporaries (caller-saved)
f8  (fs0) - f9  (fs1):  FP saved (callee-saved)
f10 (fa0) - f17 (fa7):  FP arg/return
f18 (fs2) - f27 (fs11): FP saved (callee-saved)
f28 (ft8) - f31 (ft11): FP temporaries (caller-saved)
```

### AIR → RISC-V Opcode Mapping

| AIR Opcode | RISC-V Instruction(s) | Notes |
|---|---|---|
| `OpIAdd` | `ADD rd, rs1, rs2` | RV64I |
| `OpIAddImm` | `ADDI rd, rs1, imm12` | signed 12-bit immediate |
| `OpISub` | `SUB rd, rs1, rs2` | |
| `OpIMul` | `MUL rd, rs1, rs2` | RV64M |
| `OpIDiv` | `DIV rd, rs1, rs2` | signed; RV64M |
| `OpIDivU` | `DIVU rd, rs1, rs2` | unsigned; RV64M |
| `OpIRem` | `REM rd, rs1, rs2` | signed remainder; RV64M |
| `OpIRemU` | `REMU rd, rs1, rs2` | unsigned remainder |
| `OpFAdd` | `FADD.D rd, rs1, rs2` | f64; use FADD.S for f32 |
| `OpFSub` | `FSUB.D rd, rs1, rs2` | |
| `OpFMul` | `FMUL.D rd, rs1, rs2` | |
| `OpFDiv` | `FDIV.D rd, rs1, rs2` | |
| `OpLoad` | `LD rd, offset(rs1)` | 64-bit load, sign-extends |
| `OpLoad32` | `LW rd, offset(rs1)` | 32-bit load, sign-extends |
| `OpLoad32U` | `LWU rd, offset(rs1)` | 32-bit load, zero-extends |
| `OpLoad16` | `LH rd, offset(rs1)` | 16-bit load |
| `OpLoad8` | `LB rd, offset(rs1)` | 8-bit load |
| `OpStore` | `SD rs2, offset(rs1)` | 64-bit store |
| `OpStore32` | `SW rs2, offset(rs1)` | 32-bit store |
| `OpStore8` | `SB rs2, offset(rs1)` | 8-bit store |
| `OpCall` | `JAL ra, label` or `AUIPC ra, hi; JALR ra, lo(ra)` | direct call; use AUIPC+JALR for long range |
| `OpCallIndirect` | `JALR ra, 0(rs1)` | indirect call through register |
| `OpReturn` | `RET` (alias `JALR x0, 0(x1)`) | |
| `OpBranch` | `JAL x0, label` or `J label` | unconditional |
| `OpBranchEQ` | `BEQ rs1, rs2, label` | branch if equal |
| `OpBranchNE` | `BNE rs1, rs2, label` | branch if not equal |
| `OpBranchLT` | `BLT rs1, rs2, label` | branch if less (signed) |
| `OpBranchGE` | `BGE rs1, rs2, label` | branch if ≥ (signed) |
| `OpBranchLTU` | `BLTU rs1, rs2, label` | unsigned less |
| `OpBranchGEU` | `BGEU rs1, rs2, label` | unsigned ≥ |
| `OpICmpSLT` | `SLT rd, rs1, rs2` | set rd=1 if rs1 < rs2 (signed) |
| `OpICmpSLTU` | `SLTU rd, rs1, rs2` | unsigned |
| `OpAnd` | `AND rd, rs1, rs2` | |
| `OpOr` | `OR rd, rs1, rs2` | |
| `OpXor` | `XOR rd, rs1, rs2` | |
| `OpShl` | `SLL rd, rs1, rs2` | logical shift left |
| `OpShr` | `SRL rd, rs1, rs2` | logical shift right |
| `OpAShr` | `SRA rd, rs1, rs2` | arithmetic shift right |
| `OpNeg` | `NEG rd, rs` (alias `SUB rd, x0, rs`) | |
| `OpNot` | `NOT rd, rs` (alias `XORI rd, rs, -1`) | |
| `OpMov` | `MV rd, rs` (alias `ADDI rd, rs, 0`) | register copy |
| `OpMovImm` | `LI rd, imm` (pseudo; expands to LUI+ADDI) | |
| `OpSExt32` | `ADDIW rd, rs, 0` | sign-extend 32→64 |
| `OpZExt32` | `SLLI rd, rs, 32; SRLI rd, rd, 32` | zero-extend 32→64 |

### RISC-V Comparison Model
Unlike x86-64 and ARM64, RISC-V has no flags register. Comparisons are done differently:
- For conditional branches: use `BEQ`, `BNE`, `BLT`, `BGE`, `BLTU`, `BGEU` — compare two registers directly
- For a boolean result (e.g., AIR `OpICmp` that stores bool): use `SLT`/`SLTU` (sets dest to 0 or 1)
- For equal/not-equal to zero: `BEQ rs, x0, label` / `BNE rs, x0, label`

When lowering `if a < b then ... else ...`:
```asm
BLT a0, a1, then_label
j else_label
then_label:
  ...
else_label:
  ...
```

### Large Immediate Handling
RISC-V immediates are 12-bit signed. For larger values:
```asm
; Load 32-bit immediate 0xDEAD_BEEF:
LUI  t0, 0xDEADB      ; upper 20 bits (sign-adjusted)
ADDI t0, t0, 0xEEF    ; lower 12 bits (sign-extended, so adjust upper if lower is negative)

; For 64-bit: use multiple instructions or a constant pool with AUIPC+LD
```

The LUI+ADDI pair requires care when the 12-bit part has the sign bit set — subtract 1 from the upper 20 bits to compensate.

### Function Call Distance
RISC-V JAL has a ±1MB range (20-bit offset). For larger programs, use:
```asm
AUIPC ra, %hi(target)
JALR  ra, %lo(target)(ra)
```
This is the standard "call" pseudo-instruction expanded form.

## Implementation Steps

### Step 1: Define RISC-V Instruction Types
Create `codegen/native/riscv64/instructions.go`:
```go
package riscv64

type RV64Opcode int
const (
    ADD RV64Opcode = iota
    ADDI
    SUB
    MUL; DIV; REM; DIVU; REMU
    LD; LW; LWU; LH; LHU; LB; LBU
    SD; SW; SH; SB
    JAL; JALR
    BEQ; BNE; BLT; BGE; BLTU; BGEU
    SLT; SLTU
    AND; OR; XOR; ANDI; ORI; XORI
    SLL; SRL; SRA; SLLI; SRLI; SRAI
    LUI; AUIPC
    ADDIW; ADDW; SUBW; MULW; DIVW
    FADD_D; FSUB_D; FMUL_D; FDIV_D
    FADD_S; FSUB_S; FMUL_S; FDIV_S
    FLD; FSD; FLW; FSW
    FCVT_D_W; FCVT_W_D  // int↔float conversions
)

type MachineInstr struct {
    Opcode  RV64Opcode
    Rd      Reg
    Rs1     Reg
    Rs2     Reg
    Imm     int64
    Label   string
}
```

### Step 2: Implement the Selector
Create `codegen/native/riscv64/instr_selector.go`:
```go
func (s *InstrSelector) selectInstr(instr air.Instr) error {
    switch instr.Op {
    case air.OpIAdd:
        s.emit(MachineInstr{Opcode: ADD, Rd: s.reg(instr.Dst), Rs1: s.reg(instr.Src1), Rs2: s.reg(instr.Src2)})
    case air.OpIAddImm:
        s.emit(MachineInstr{Opcode: ADDI, Rd: s.reg(instr.Dst), Rs1: s.reg(instr.Src1), Imm: instr.Imm})
    case air.OpLoad:
        s.emit(MachineInstr{Opcode: LD, Rd: s.reg(instr.Dst), Rs1: s.reg(instr.Src1), Imm: instr.Offset})
    case air.OpBranchCond:
        rv64cond := airCondToRV64(instr.Cond)
        s.emit(MachineInstr{Opcode: rv64cond, Rs1: s.reg(instr.Src1), Rs2: s.reg(instr.Src2), Label: instr.Target})
    case air.OpCall:
        // Short range: JAL ra, label
        // Long range: AUIPC + JALR (let linker decide via relocation)
        s.emit(MachineInstr{Opcode: JAL, Rd: RA, Label: instr.Target})
    case air.OpReturn:
        s.emit(MachineInstr{Opcode: JALR, Rd: ZERO, Rs1: RA, Imm: 0})
    default:
        return fmt.Errorf("riscv64: unimplemented opcode %s", instr.Op)
    }
    return nil
}
```

### Step 3: Immediate Materialization
```go
func (s *InstrSelector) materializeImm(rd Reg, imm int64) {
    if imm >= -2048 && imm <= 2047 {
        s.emit(MachineInstr{Opcode: ADDI, Rd: rd, Rs1: ZERO, Imm: imm})
        return
    }
    // 32-bit: LUI + ADDI
    upper := (imm + 0x800) >> 12  // adjust for sign extension
    lower := imm - (upper << 12)
    s.emit(MachineInstr{Opcode: LUI, Rd: rd, Imm: upper})
    if lower != 0 {
        s.emit(MachineInstr{Opcode: ADDI, Rd: rd, Rs1: rd, Imm: lower})
    }
}
```

### Step 4: Condition Code Mapping
```go
func airCondToRV64(cond air.Cond) RV64Opcode {
    switch cond {
    case air.CondEQ:  return BEQ
    case air.CondNE:  return BNE
    case air.CondLT:  return BLT
    case air.CondGE:  return BGE
    case air.CondLTU: return BLTU
    case air.CondGEU: return BGEU
    default: panic(fmt.Sprintf("riscv64: unknown cond %v", cond))
    }
}
```

### Step 5: Register with Backend Driver
```go
case "riscv64-linux-gnu":
    return riscv64.NewBackend(triple), nil
```

## Test Plan

### Unit Tests
1. `TestSelectADD` — `OpIAdd` → `ADD rd, rs1, rs2`
2. `TestSelectLD` — `OpLoad` → `LD rd, offset(rs1)`
3. `TestSelectBranchEQ` — `OpBranchCond(EQ)` → `BEQ rs1, rs2, label`
4. `TestSelectCall` — `OpCall` → `JAL ra, label`
5. `TestSelectReturn` — `OpReturn` → `JALR x0, 0(x1)`
6. `TestMaterializeSmallImm` — imm=42 → single `ADDI`
7. `TestMaterializeLargeImm` — imm=0xDEADBEEF → `LUI + ADDI` pair
8. `TestMaterializeNegImm` — imm=-1 → `ADDI x0, x0, -1`
9. `TestSelectFloatAdd` — `OpFAdd(f64)` → `FADD.D`

### Golden Tests
For each test in `tests/codegen/riscv64/`:
1. `arith.ax` — integer arithmetic
2. `float_ops.ax` — floating-point
3. `branches.ax` — conditional branches
4. `calls.ax` — function calls and returns
5. `memory.ax` — loads and stores

Compare `axc dump-riscv64-asm <file>` output with `golden/<file>.s`.

## Validation Checklist
- [ ] All 39 integer instructions from RV64I/M selected correctly
- [ ] Float instructions use `.D` suffix for f64, `.S` for f32
- [ ] Conditional branches use direct register comparison (no flags)
- [ ] Immediate materialization handles all 64-bit immediate ranges
- [ ] `LD`/`SD` offsets are 12-bit signed (emit error if out of range)
- [ ] `JAL` label is PC-relative (linker fills in via R_RISCV_JAL relocation)
- [ ] No global mutable state in the selector

## Acceptance Criteria
1. All AIR opcodes in the mapping table produce correct RISC-V instruction sequences
2. Golden tests pass for all 5 test programs
3. Float operations use the correct `.D`/`.S` suffix based on AIR type
4. Immediate materialization correct for 12-bit, 32-bit, and 64-bit constants
5. Unit test coverage ≥ 90%

## Definition of Done
- `instructions.go`, `regs.go`, `instr_selector.go`, `riscv64.go` implemented and reviewed
- All unit tests pass
- All golden tests pass
- Backend registered in `codegen/driver.go` for `riscv64-linux-gnu`
- No panics or unimplemented opcodes in submitted code

## Risks & Mitigations
| Risk | Mitigation |
|---|---|
| LUI+ADDI sign adjustment off by one | Unit test every boundary: 0x7FF, 0x800, -0x800, -0x801 |
| JAL range exceeded for large programs | Emit AUIPC+JALR by default; let linker relax to JAL if in range |
| No flags register — comparison model different | Explicitly test all AIR conditional branch types |
| C extension (compressed) encoding complexity | Skip C extension in MVP; emit all instructions as 32-bit |

## Future Follow-up Tasks
- p13-t06: RISC-V ABI and register allocator (paired task)
- p13-t07: Cross-compile integration tests using QEMU user-mode
- RISC-V V extension (vector) for SIMD — future task post-Phase 13
- Compressed instruction emission (smaller code size) — optimization pass in Phase 17+
