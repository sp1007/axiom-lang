# p13-t02: ARM64 Register Allocation and Stack Frame

## Purpose
Implement register allocation and stack frame layout for the ARM64 backend. This pass assigns physical ARM64 registers to AIR virtual registers and constructs the function prologue/epilogue, ensuring correct callee-save/caller-save discipline per the AAPCS64 ABI.

## Context
ARM64 has 31 general-purpose 64-bit registers (X0-X30) plus XZR (zero register) and SP (stack pointer). The register file is divided into caller-saved and callee-saved sets. Register allocation must respect this division to avoid corrupting live values across calls. Stack frames must be 16-byte aligned at all times.

This task builds a linear-scan register allocator adapted for ARM64's register conventions. Linear scan is chosen over graph coloring for compilation speed — it is O(n) in the number of live intervals versus O(n^3) for graph coloring, which matters for fast incremental builds.

## Inputs
- AIR function with virtual registers (VRegs) from p11-t01
- Live interval analysis output (computed from AIR liveness analysis)
- ARM64 physical register definitions (this task defines them)
- AAPCS64 specification (defined in p13-t03, but register classes needed here)

## Outputs
- `codegen/native/arm64/regalloc.go` — linear-scan allocator for ARM64
- `codegen/native/arm64/frame.go` — stack frame layout and prologue/epilogue emission
- `codegen/native/arm64/regs.go` — physical register definitions and classifications
- Test file: `codegen/native/arm64/regalloc_test.go`

## Dependencies
- p13-t01: ARM64 instruction selector (uses regalloc output)
- p11-t05: Register allocator interface and live interval analysis (provides base infrastructure)

## Subsystems Affected
- `codegen/native/arm64/` — new files in existing package
- `codegen/native/arm64/instr_selector.go` — consumes regmap from regalloc

## Detailed Requirements

### Register Definitions

#### General-Purpose Registers (64-bit Xn, 32-bit Wn)
```
X0-X7:   Argument/result registers (caller-saved)
X8:      Indirect result location register (caller-saved)
X9-X15:  Temporary registers (caller-saved)
X16-X17: Intra-procedure scratch (IP0, IP1) — linker stubs use these
X18:     Platform register (avoid on macOS; usable on Linux)
X19-X28: Callee-saved registers
X29:     Frame pointer (FP) — must be preserved
X30:     Link register (LR) — return address, must be preserved
XZR:     Zero register (reads as 0, writes discarded)
SP:      Stack pointer (not in GPR file)
```

#### Floating-Point / NEON Registers (128-bit Vn, 64-bit Dn, 32-bit Sn)
```
V0-V7:   FP argument/result registers (caller-saved)
V8-V15:  Callee-saved (only bottom 64 bits need saving per AAPCS64)
V16-V31: Temporary FP registers (caller-saved)
```

#### Allocatable Integer Register Sets
```go
var CallerSavedIntRegs = []Reg{X0, X1, X2, X3, X4, X5, X6, X7, X8, X9, X10, X11, X12, X13, X14, X15}
var CalleeSavedIntRegs = []Reg{X19, X20, X21, X22, X23, X24, X25, X26, X27, X28}
// X16, X17 reserved for linker; X18 avoided; X29=FP, X30=LR handled separately
```

### Stack Frame Layout
```
High addresses
+---------------------------+
| Caller's frame            |
+---------------------------+  ← incoming SP (16-byte aligned)
| Return address (X30)      |  \
| Frame pointer (X29)       |   > saved by STP X29, X30, [SP, #-frame_size]!
+---------------------------+
| Callee-saved registers    |  (X19-X28 that are used, 8 bytes each)
| Callee-saved FP regs      |  (V8-V15 lower 64 bits, 8 bytes each)
+---------------------------+
| Local variable spill slots|  (8 bytes each, aligned)
+---------------------------+
| Outgoing argument area    |  (for calls with > 8 args)
+---------------------------+  ← SP during function body (16-byte aligned)
Low addresses
```

Frame size computation:
```
frame_size = align16(
    8 * 2 +                          // X29 + X30
    8 * len(used_callee_saved_gpr) + // X19-X28 used
    8 * len(used_callee_saved_fpr) + // V8-V15 used
    8 * num_spill_slots +             // spilled VRegs
    8 * max_outgoing_args             // args beyond 8 on stack
)
```

### Prologue Emission
```asm
; Save FP and LR atomically (required by AAPCS64 for frame record)
STP X29, X30, [SP, #-frame_size]!   ; pre-index: SP = SP - frame_size, then store
MOV X29, SP                          ; update frame pointer

; Save callee-saved GPRs (pairs for efficiency)
STP X19, X20, [SP, #offset_19_20]
STP X21, X22, [SP, #offset_21_22]
; ... for each used pair

; Save callee-saved FP regs (only lower 64 bits per AAPCS64)
STR D8, [SP, #offset_d8]
STR D9, [SP, #offset_d9]
; ...
```

### Epilogue Emission
```asm
; Restore callee-saved FP regs
LDR D8, [SP, #offset_d8]
LDR D9, [SP, #offset_d9]

; Restore callee-saved GPRs
LDP X19, X20, [SP, #offset_19_20]
LDP X21, X22, [SP, #offset_21_22]
; ...

; Restore FP and LR, deallocate frame
LDP X29, X30, [SP], #frame_size     ; post-index: load then SP = SP + frame_size

RET   ; branches to X30
```

### Linear-Scan Allocator Algorithm
```
1. Compute live intervals for all VRegs (start_instr, end_instr)
2. Sort intervals by start point
3. For each new interval I:
   a. Expire intervals that ended before I.start → free their registers
   b. If a free register is available in the preferred class:
      - Assign it to I
   c. Else:
      - Spill: pick the interval with latest end point
      - If spilled.end > I.end: spill 'spilled', assign its reg to I
      - Else: spill I (assign a stack slot to I)
4. Emit spill/reload instructions around uses of spilled VRegs
```

### Spill Slot Management
```go
type SpillSlot struct {
    Offset int    // byte offset from SP
    Size   int    // 4 or 8 bytes
}

type FrameLayout struct {
    TotalSize    int
    SpillSlots   map[VReg]SpillSlot
    SavedRegs    []SavedReg
    OutArgSize   int
}
```

Spill emit: before a use of spilled VReg:
```asm
LDR Xscratch, [SP, #spill_offset]
; use Xscratch
STR Xscratch, [SP, #spill_offset]  ; if modified
```

### Register Pairing for STP/LDP
When saving/restoring callee-saved registers, use `STP`/`LDP` to save two registers at once (more efficient than individual `STR`/`LDR`). Pair registers that are adjacent in the callee-saved list.

## Implementation Steps

### Step 1: Define Physical Register Types
Create `codegen/native/arm64/regs.go`:
```go
package arm64

type Reg uint8

const (
    X0 Reg = iota; X1; X2; X3; X4; X5; X6; X7
    X8; X9; X10; X11; X12; X13; X14; X15
    X16; X17; X18; X19; X20; X21; X22; X23
    X24; X25; X26; X27; X28
    X29  // FP
    X30  // LR
    XZR
    SP
    // Float
    V0; V1; V2; V3; V4; V5; V6; V7
    V8; V9; V10; V11; V12; V13; V14; V15
    V16; V17; V18; V19; V20; V21; V22; V23
    V24; V25; V26; V27; V28; V29; V30; V31
)

func (r Reg) IsCalleeSaved() bool {
    return (r >= X19 && r <= X28) || r == X29 || r == X30 ||
           (r >= V8 && r <= V15)
}
```

### Step 2: Implement Live Interval Computation
```go
func ComputeLiveIntervals(fn *air.Function) map[VReg]Interval {
    intervals := make(map[VReg]Interval)
    instrIdx := 0
    for _, block := range fn.Blocks {
        for _, instr := range block.Instrs {
            for _, use := range instr.Uses() {
                iv := intervals[use]
                iv.End = instrIdx
                intervals[use] = iv
            }
            if instr.Dst != VRegNone {
                iv := intervals[instr.Dst]
                if iv.Start == 0 {
                    iv.Start = instrIdx
                }
                iv.End = instrIdx
                intervals[instr.Dst] = iv
            }
            instrIdx++
        }
    }
    return intervals
}
```

### Step 3: Implement Linear-Scan Core
```go
type LinearScanAllocator struct {
    intervals    []LiveInterval  // sorted by start
    active       []LiveInterval  // sorted by end
    freeIntRegs  []Reg
    freeFPRegs   []Reg
    regmap       map[VReg]Reg
    spillSlots   map[VReg]int    // VReg → stack slot offset
    nextSpill    int
}

func (a *LinearScanAllocator) Allocate(fn *air.Function) (*AllocationResult, error) {
    a.freeIntRegs = append([]Reg{}, CallerSavedIntRegs...)
    a.freeIntRegs = append(a.freeIntRegs, CalleeSavedIntRegs...)
    // Process intervals in start order
    for _, iv := range a.intervals {
        a.expireOldIntervals(iv)
        if len(a.freeIntRegs) == 0 {
            a.spillAtInterval(iv)
        } else {
            reg := a.freeIntRegs[0]
            a.freeIntRegs = a.freeIntRegs[1:]
            a.regmap[iv.VReg] = reg
            a.active = insertByEnd(a.active, iv)
        }
    }
    return &AllocationResult{RegMap: a.regmap, SpillSlots: a.spillSlots}, nil
}
```

### Step 4: Implement Frame Layout
```go
func (f *FrameBuilder) Build(fn *air.Function, alloc *AllocationResult) FrameLayout {
    usedCalleeGPR := collectUsedCalleeGPR(alloc.RegMap)
    usedCalleeFPR := collectUsedCalleeFPR(alloc.RegMap)
    numSpills := len(alloc.SpillSlots)
    
    size := 16 // X29 + X30
    size += 8 * len(usedCalleeGPR)
    size += 8 * len(usedCalleeFPR)
    size += 8 * numSpills
    size += f.outArgSize
    size = align(size, 16)  // 16-byte alignment
    
    return FrameLayout{TotalSize: size, ...}
}
```

### Step 5: Emit Prologue/Epilogue
```go
func EmitPrologue(frame FrameLayout) []MachineInstr {
    instrs := []MachineInstr{}
    instrs = append(instrs, MachineInstr{
        Opcode: STP, Rn: X29, Rm: X30,
        Base: SP, Offset: -frame.TotalSize, PreIndex: true,
    })
    instrs = append(instrs, MachineInstr{Opcode: MOV, Rd: X29, Rn: SP})
    // ... save callee-saved regs
    return instrs
}

func EmitEpilogue(frame FrameLayout) []MachineInstr {
    instrs := []MachineInstr{}
    // ... restore callee-saved regs
    instrs = append(instrs, MachineInstr{
        Opcode: LDP, Rd: X29, Rn: X30,
        Base: SP, Offset: frame.TotalSize, PostIndex: true,
    })
    instrs = append(instrs, MachineInstr{Opcode: RET})
    return instrs
}
```

## Test Plan

### Unit Tests
1. `TestCalleeSavedDetection` — verify `IsCalleeSaved()` returns correct results for all regs
2. `TestLiveIntervalComputation` — compute intervals for a simple 3-instruction function, verify ranges
3. `TestLinearScanNoSpill` — 4 VRegs with non-overlapping intervals → no spills, 4 distinct regs
4. `TestLinearScanWithSpill` — 17 simultaneously live VRegs (more than 16 allocatable) → at least 1 spill
5. `TestFrameSizeAlignment` — frame sizes are always multiples of 16
6. `TestPrologueEpilogue` — verify STP X29,X30 appears in prologue and LDP X29,X30 in epilogue
7. `TestCalleeSavedPreserved` — if X19 is used, it appears in both prologue save and epilogue restore

### Integration Tests
1. Compile a function that uses > 16 local variables → verify spill/reload in output
2. Compile a function that calls another function → verify LR saved/restored
3. Verify frame pointer chain is correct (X29 → caller's X29)

## Validation Checklist
- [ ] All 31 GPRs and 32 FP registers defined correctly
- [ ] Caller-saved and callee-saved sets are disjoint and complete
- [ ] Frame size is always 16-byte aligned
- [ ] Prologue always saves X29 and X30 with STP
- [ ] Epilogue always restores X29 and X30 with LDP
- [ ] Spill slots assigned at unique non-overlapping offsets
- [ ] LinearScan allocator is O(n log n) where n = number of VRegs
- [ ] No physical register assigned to two live VRegs simultaneously

## Acceptance Criteria
1. Linear-scan allocator produces valid allocation for all test functions (no register conflicts)
2. Frame layout is 16-byte aligned for all test cases
3. Functions with > 16 simultaneously live integer VRegs spill correctly and produce correct results
4. Prologue/epilogue generated correctly for leaf functions (no calls) and non-leaf functions
5. All unit tests pass

## Definition of Done
- `regs.go`, `regalloc.go`, `frame.go` implemented and reviewed
- Unit test coverage ≥ 90%
- Integration tests pass (requires p13-t01 selector to be functional)
- No register conflicts detected by a post-allocation verifier pass
- Spill slots do not overlap in the frame layout

## Risks & Mitigations
| Risk | Mitigation |
|---|---|
| Frame pointer required by macOS | Always save/restore X29 even in leaf functions on macOS target |
| 16-byte alignment violated | Add assert in frame builder; CI fails if misaligned |
| Spill in tight loop causes performance regression | Track spill count as a metric; flag in diagnostics if > threshold |
| X16/X17 conflict with linker stubs | Exclude X16, X17 from allocatable set unconditionally |

## Future Follow-up Tasks
- p13-t03: AAPCS64 ABI — arg register assignment at call sites depends on this allocator
- p11-t05: The live interval algorithm may be shared with the x86-64 allocator — consider abstracting
- Coalescing optimization: merge copy instructions by assigning same physical register to copy src/dst
