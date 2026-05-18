# p13-t01: ARM64 Instruction Selector

## Purpose
Implement the instruction selection pass for the ARM64/AArch64 target architecture in `codegen/native/arm64/`. This pass translates AIR (AXIOM Intermediate Representation) opcodes into concrete ARM64 machine instructions, forming the foundation of the ARM64 native backend.

## Context
The ARM64 backend is needed to support Apple Silicon (M1/M2/M3) and ARM64 Linux targets. ARM64 uses a fixed-width 32-bit instruction encoding with a load-store architecture and 31 general-purpose 64-bit registers. The instruction selector operates on the post-register-allocation AIR and produces a list of `MachineInstr` structs that are later encoded into binary by the emitter.

ARM64 is the primary target for macOS (Apple Silicon) and is increasingly common on Linux servers (AWS Graviton, Ampere). Getting this right enables a large class of deployment targets.

## Inputs
- Optimized AIR (`OptimizedAIR`) from the AIR optimization pipeline (p11-t01 output)
- Register allocation result from p13-t02 (virtual register → physical register mapping)
- Target triple: `aarch64-apple-macos13` or `aarch64-linux-gnu`
- AIR opcode definitions from `ir/air/opcodes.go`

## Outputs
- `codegen/native/arm64/instr_selector.go` — instruction selector implementation
- `codegen/native/arm64/instructions.go` — ARM64 instruction type definitions
- `codegen/native/arm64/arm64.go` — package entry point and ISel interface
- Unit test file: `codegen/native/arm64/instr_selector_test.go`
- Golden test fixtures: `tests/codegen/arm64/golden/*.s` — expected assembly output

## Dependencies
- p11-t01: AIR definition and SSA form (provides input AIR)
- p09-t01: Native backend interface (`Backend`, `ISel` interfaces)
- p13-t02: ARM64 register allocator (for physical register assignment)

## Subsystems Affected
- `codegen/native/arm64/` — new package
- `codegen/driver.go` — register ARM64 backend
- `compiler/driver/` — `--target=aarch64-*` flag handling

## Detailed Requirements

### AIR → ARM64 Opcode Mapping

| AIR Opcode | ARM64 Instruction | Notes |
|---|---|---|
| `OpIAdd` | `ADD Xd, Xn, Xm` | 64-bit integer add |
| `OpIAddImm` | `ADD Xd, Xn, #imm12` | Immediate add (imm ≤ 4095) |
| `OpISub` | `SUB Xd, Xn, Xm` | 64-bit integer subtract |
| `OpIMul` | `MUL Xd, Xn, Xm` | Alias for `MADD Xd, Xn, Xm, XZR` |
| `OpIDiv` | `SDIV Xd, Xn, Xm` | Signed divide (no trap on div/0 in hardware) |
| `OpIRem` | `SDIV Xtmp, Xn, Xm; MSUB Xd, Xtmp, Xm, Xn` | Two-instruction sequence |
| `OpFAdd` | `FADD Dd, Dn, Dm` | f64 add; use Sd for f32 |
| `OpFMul` | `FMUL Dd, Dn, Dm` | f64 multiply |
| `OpFDiv` | `FDIV Dd, Dn, Dm` | f64 divide |
| `OpLoad` | `LDR Xd, [Xn, #offset]` | 64-bit load; offset must be 8-byte aligned for scaled form |
| `OpLoad32` | `LDR Wd, [Xn, #offset]` | 32-bit load, zero-extends |
| `OpLoad8` | `LDRB Wd, [Xn, #offset]` | 8-bit load, zero-extends |
| `OpStore` | `STR Xs, [Xn, #offset]` | 64-bit store |
| `OpStore32` | `STR Ws, [Xn, #offset]` | 32-bit store |
| `OpStore8` | `STRB Ws, [Xn, #offset]` | 8-bit store |
| `OpCall` | `BL label` | Branch with link (saves PC+4 in X30) |
| `OpCallIndirect` | `BLR Xn` | Indirect call through register |
| `OpReturn` | `RET` | Branches to address in X30 |
| `OpBranch` | `B label` | Unconditional branch |
| `OpBranchCond` | `B.cond label` | Conditional branch (cond from flags) |
| `OpICmp` | `CMP Xn, Xm` (alias `SUBS XZR, Xn, Xm`) | Sets NZCV flags |
| `OpAnd` | `AND Xd, Xn, Xm` | Bitwise AND |
| `OpOr` | `ORR Xd, Xn, Xm` | Bitwise OR |
| `OpXor` | `EOR Xd, Xn, Xm` | Bitwise XOR |
| `OpShl` | `LSL Xd, Xn, Xm` | Logical shift left |
| `OpShr` | `LSR Xd, Xn, Xm` | Logical shift right (unsigned) |
| `OpAShr` | `ASR Xd, Xn, Xm` | Arithmetic shift right (signed) |
| `OpNeg` | `NEG Xd, Xn` (alias `SUB Xd, XZR, Xn`) | Negate |
| `OpNot` | `MVN Xd, Xn` | Bitwise NOT |
| `OpMov` | `MOV Xd, Xn` (alias `ORR Xd, XZR, Xn`) | Register copy |
| `OpMovImm` | `MOVZ Xd, #imm16` or `MOV Xd, #imm` | Load immediate |
| `OpSIMDAdd` | `FADD V0.4S, V1.4S, V2.4S` | f32x4 SIMD add via NEON |
| `OpSIMDMul` | `FMUL V0.4S, V1.4S, V2.4S` | f32x4 SIMD multiply |

### Condition Code Mapping
ARM64 condition codes for `B.cond`:
- EQ (equal), NE (not equal)
- LT/LE/GT/GE (signed comparisons)
- LO/LS/HI/HS (unsigned comparisons — use these for pointer comparisons)
- MI (minus/negative), PL (plus/non-negative)

### Large Immediate Handling
Immediates that don't fit in 12-bit fields require materialization:
```
; For a 32-bit immediate 0xDEADBEEF:
MOVZ X0, #0xBEEF         ; load lower 16 bits
MOVK X0, #0xDEAD, LSL#16 ; merge upper 16 bits
```
For 64-bit immediates, up to 4 MOVZ/MOVK instructions.

### NEON SIMD Instructions
When AIR contains vectorized operations (from the vectorizer in p11-t12):
- `f32x4` maps to `V.4S` (four 32-bit floats)
- `f64x2` maps to `V.2D` (two 64-bit doubles)
- `i32x4` maps to `V.4S` (four 32-bit integers)
- Load vector: `LD1 {V0.4S}, [Xn]`
- Store vector: `ST1 {V0.4S}, [Xn]`

### Instruction Encoding (for the emitter)
ARM64 instructions are fixed 32-bit. The selector emits `MachineInstr` structs; the emitter in p13-t04 does binary encoding:
```go
type MachineInstr struct {
    Opcode   ARM64Opcode
    Rd       Reg   // destination register
    Rn       Reg   // first source register
    Rm       Reg   // second source register (or XZR)
    Imm      int64 // immediate value
    Label    string // for branch targets
    Cond     CondCode
    VecShape VecShape // for NEON
}
```

## Implementation Steps

### Step 1: Define ARM64 Instruction Types
Create `codegen/native/arm64/instructions.go`:
```go
package arm64

type ARM64Opcode int

const (
    ADD ARM64Opcode = iota
    SUB
    MUL
    SDIV
    MSUB
    LDR
    STR
    LDRB
    STRB
    BL
    BLR
    RET
    B
    BCond
    CMP
    AND
    ORR
    EOR
    LSL
    LSR
    ASR
    NEG
    MVN
    MOVZ
    MOVK
    MOV
    FADD
    FMUL
    FDIV
    FSUB
    NOP
)

type CondCode int
const (
    EQ CondCode = iota; NE; LT; LE; GT; GE; LO; LS; HI; HS; MI; PL
)
```

### Step 2: Implement the ISel Pass
Create `codegen/native/arm64/instr_selector.go`:
```go
package arm64

type InstrSelector struct {
    air     *air.Function
    regmap  map[air.VReg]Reg
    instrs  []MachineInstr
    consts  map[int64]Reg // materialized constants cache
}

func (s *InstrSelector) Select(fn *air.Function) ([]MachineInstr, error) {
    for _, block := range fn.Blocks {
        s.emitLabel(block.Label)
        for _, instr := range block.Instrs {
            if err := s.selectInstr(instr); err != nil {
                return nil, err
            }
        }
    }
    return s.instrs, nil
}

func (s *InstrSelector) selectInstr(instr air.Instr) error {
    switch instr.Op {
    case air.OpIAdd:
        rd := s.regmap[instr.Dst]
        rn := s.regmap[instr.Src1]
        rm := s.regmap[instr.Src2]
        s.emit(MachineInstr{Opcode: ADD, Rd: rd, Rn: rn, Rm: rm})
    case air.OpLoad:
        rd := s.regmap[instr.Dst]
        rn := s.regmap[instr.Src1]
        s.emit(MachineInstr{Opcode: LDR, Rd: rd, Rn: rn, Imm: instr.Offset})
    case air.OpIMul:
        rd := s.regmap[instr.Dst]
        rn := s.regmap[instr.Src1]
        rm := s.regmap[instr.Src2]
        // MUL is an alias for MADD Xd, Xn, Xm, XZR
        s.emit(MachineInstr{Opcode: MUL, Rd: rd, Rn: rn, Rm: rm})
    case air.OpIDiv:
        rd := s.regmap[instr.Dst]
        rn := s.regmap[instr.Src1]
        rm := s.regmap[instr.Src2]
        s.emit(MachineInstr{Opcode: SDIV, Rd: rd, Rn: rn, Rm: rm})
    case air.OpIRem:
        // tmp = Xn / Xm; Xd = Xn - tmp * Xm
        tmp := s.allocTempReg()
        s.emit(MachineInstr{Opcode: SDIV, Rd: tmp, Rn: s.regmap[instr.Src1], Rm: s.regmap[instr.Src2]})
        s.emit(MachineInstr{Opcode: MSUB, Rd: s.regmap[instr.Dst], Rn: tmp, Rm: s.regmap[instr.Src2], Ra: s.regmap[instr.Src1]})
    case air.OpCall:
        s.emit(MachineInstr{Opcode: BL, Label: instr.Target})
    case air.OpReturn:
        s.emit(MachineInstr{Opcode: RET})
    // ... all other opcodes
    default:
        return fmt.Errorf("arm64: unimplemented opcode %s", instr.Op)
    }
    return nil
}
```

### Step 3: Implement Immediate Materialization
```go
func (s *InstrSelector) materializeImm(rd Reg, imm int64) {
    if imm >= 0 && imm <= 0xFFFF {
        s.emit(MachineInstr{Opcode: MOVZ, Rd: rd, Imm: imm})
        return
    }
    // Multi-chunk for larger immediates
    chunks := splitImm64(imm)
    first := true
    for shift, chunk := range chunks {
        if chunk == 0 { continue }
        if first {
            s.emit(MachineInstr{Opcode: MOVZ, Rd: rd, Imm: chunk, Shift: shift * 16})
            first = false
        } else {
            s.emit(MachineInstr{Opcode: MOVK, Rd: rd, Imm: chunk, Shift: shift * 16})
        }
    }
}
```

### Step 4: Implement NEON Vector Selection
```go
func (s *InstrSelector) selectSIMD(instr air.Instr) {
    switch instr.Op {
    case air.OpSIMDFAdd:
        vd := s.vecRegmap[instr.Dst]
        vn := s.vecRegmap[instr.Src1]
        vm := s.vecRegmap[instr.Src2]
        s.emit(MachineInstr{
            Opcode: FADD, Rd: vd, Rn: vn, Rm: vm,
            VecShape: Vec4S,  // .4S = four f32
        })
    }
}
```

### Step 5: Add Condition Code Translation
```go
func airCondToARM64(cond air.Cond) CondCode {
    switch cond {
    case air.CondEQ: return EQ
    case air.CondNE: return NE
    case air.CondLT: return LT
    case air.CondLE: return LE
    case air.CondGT: return GT
    case air.CondGE: return GE
    default: panic(fmt.Sprintf("unknown cond: %v", cond))
    }
}
```

### Step 6: Register with Backend Driver
In `codegen/driver.go`:
```go
case "aarch64-apple-macos13", "aarch64-linux-gnu":
    return arm64.NewBackend(triple), nil
```

## Test Plan

### Unit Tests
1. Test each AIR opcode individually: emit one AIR instruction, verify correct `MachineInstr` produced
2. Test immediate materialization: 0, 1, 0xFFFF, 0x10000, 0xDEADBEEF, max-int64
3. Test IRem two-instruction sequence: verify SDIV + MSUB pair
4. Test condition code translation: all 8 AIR conditions map correctly
5. Test NEON opcodes: f32x4 add, mul, load, store

### Golden Tests
For each test program in `tests/codegen/arm64/`:
1. `tests/codegen/arm64/arith.ax` — arithmetic ops
2. `tests/codegen/arm64/load_store.ax` — memory operations
3. `tests/codegen/arm64/branches.ax` — conditional branches
4. `tests/codegen/arm64/calls.ax` — function calls
5. `tests/codegen/arm64/simd.ax` — NEON vector operations

Run: `axc dump-arm64-asm <file.ax>` → compare with `golden/<file>.s`

### Error Cases
- Unsupported opcode returns error (no panic)
- Offset out of range for scaled LDR/STR → fall back to unscaled form
- Verify no duplicate instruction emission

## Validation Checklist
- [ ] All AIR opcodes from the opcode table have a corresponding ARM64 lowering
- [ ] `MachineInstr` structs are produced (not binary — that is the emitter's job)
- [ ] NEON selection only triggers for vector-typed AIR values
- [ ] Large immediate materialization tested for all boundary values
- [ ] No global mutable state in the selector
- [ ] Selector is stateless between functions (new InstrSelector per function)
- [ ] Golden tests all pass
- [ ] `go test ./codegen/native/arm64/...` passes

## Acceptance Criteria
1. All AIR opcodes listed in the mapping table are lowered without panicking
2. Golden test output matches expected ARM64 assembly for all 5 test programs
3. NEON instructions generated correctly for vector AIR values
4. Unit test coverage ≥ 90% for `instr_selector.go`
5. The selector integrates with `axc build --target=aarch64-linux-gnu` without errors

## Definition of Done
- `codegen/native/arm64/instr_selector.go` implemented and reviewed
- All unit tests pass
- All golden tests pass
- `axc build --target=aarch64-linux-gnu tests/hello.ax` produces a `MachineInstr` list (encoding done in p13-t04)
- No `TODO` or `panic("unimplemented")` in submitted code
- Code reviewed for architecture boundary violations (no parser/semantic imports)

## Risks & Mitigations
| Risk | Mitigation |
|---|---|
| ARM64 opcode encoding complexity | Keep selector and emitter separate; selector only produces `MachineInstr` structs |
| NEON register bank differences | Use a separate `VReg` namespace for NEON registers |
| Offset range limits for LDR/STR | Implement offset range check; emit `ADD + LDR` sequence for large offsets |
| IRem requires temp register | Allocate a spill register in the selector; coordinate with register allocator |

## Future Follow-up Tasks
- p13-t02: Register allocator (required before instruction selection can use physical registers)
- p13-t03: AAPCS64 calling convention (required for call/return lowering)
- p13-t04: Mach-O binary emitter (encodes `MachineInstr` → bytes)
- Peephole optimizer for ARM64 (merge ADD+CMP into ADDS, etc.) — Phase 17+
