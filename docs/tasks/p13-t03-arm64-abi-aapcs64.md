# p13-t03: ARM64 ABI — AAPCS64 Calling Convention

## Purpose
Implement the AAPCS64 (Procedure Call Standard for the Arm 64-bit Architecture) calling convention for the ARM64 backend. This governs how function arguments are passed, how return values are communicated, and how the stack is managed at call boundaries.

## Context
AAPCS64 is the standard ABI for 64-bit ARM platforms and is used on Linux (aarch64-linux-gnu), macOS (aarch64-apple-macos), and Windows (aarch64-windows). Correct ABI implementation is critical for interoperability with C libraries (libc, platform APIs) and for calling conventions between AXIOM functions.

Key ABI concerns:
- Argument passing in registers (first 8 integer args in X0-X7, first 8 float args in V0-V7)
- Stack argument passing for additional args (16-byte aligned on stack)
- Struct passing rules (small structs in registers, large structs via pointer)
- HFA (Homogeneous Float Aggregate) — up to 4 float/double members in V0-V3
- Return value conventions (X0 or X0+X1 for large integers; V0 for floats)
- Stack alignment (16-byte at function entry)
- Variadic function calling conventions

## Inputs
- ARM64 register definitions from p13-t02 (`regs.go`)
- AIR function signature (`FuncType` with parameter types and return type)
- AXIOM type system types (i8, i16, i32, i64, u8-u64, f32, f64, struct, array, pointer)
- AAPCS64 specification (public document from ARM Ltd)

## Outputs
- `codegen/native/arm64/abi.go` — AAPCS64 argument/return classification and assignment
- `codegen/native/arm64/call_lowering.go` — lower AIR `OpCall`/`OpReturn` using ABI rules
- Test file: `codegen/native/arm64/abi_test.go`

## Dependencies
- p13-t02: ARM64 register definitions and frame layout

## Subsystems Affected
- `codegen/native/arm64/` — new files
- `codegen/native/arm64/instr_selector.go` — call lowering integrates here

## Detailed Requirements

### Argument Classification

AAPCS64 classifies each argument into one of:
- `INTEGER` — integer, pointer, boolean, enum
- `FLOAT` — f32, f64
- `HFA` — struct with 2-4 identical float/double members
- `COMPOSITE` — struct or array not matching HFA rules
- `EMPTY` — zero-size type

#### Integer Argument Assignment (NGRN = Next General-Purpose Register Number)
```
NGRN starts at 0.
For each argument of class INTEGER:
  if NGRN < 8:
    assign X[NGRN], NGRN++
  else:
    SP + stack_offset, stack_offset += align(size, 8)
```

#### Float Argument Assignment (NSRN = Next SIMD/FP Register Number)
```
NSRN starts at 0.
For each argument of class FLOAT:
  if NSRN < 8:
    assign V[NSRN], NSRN++
  else:
    SP + stack_offset, stack_offset += 8
```

#### HFA (Homogeneous Float Aggregate)
A struct is an HFA if:
1. All members are the same floating-point type (all f32 or all f64)
2. 2 to 4 members
3. No padding between members

HFA passing: each member gets its own SIMD register. An HFA with 3 f32 members uses V[NSRN], V[NSRN+1], V[NSRN+2].

#### Composite (General Struct) Passing
```
size = sizeof(struct), rounded up to 8-byte boundary
if size <= 16:
  if NGRN + ceil(size/8) <= 8:
    use X[NGRN]..X[NGRN+n-1] (up to 2 registers)
    NGRN += n
  else:
    NGRN = 8  // no more register args after this
    pass on stack
else:
  caller allocates space on stack, passes pointer in X[NGRN]
  NGRN++
```

### Return Value Convention
| Return Type | Convention |
|---|---|
| void | No return value |
| i8, i16, i32, i64, u8-u64, pointer | X0 (zero/sign-extended to 64 bits) |
| f32 | S0 (bottom 32 bits of V0) |
| f64 | D0 (bottom 64 bits of V0) |
| struct ≤ 16 bytes (non-HFA) | X0 (and X1 if > 8 bytes) |
| HFA | V0, V1, V2, V3 (one per member) |
| struct > 16 bytes | Caller passes hidden pointer in X8; callee writes to *X8 |

### Stack Argument Layout
Arguments passed on stack:
- Each argument is aligned to its natural alignment (min 8 bytes)
- Stack grows downward; caller pushes args in right-to-left order
- Stack pointer at call site must be 16-byte aligned

### Variadic Functions
- Fixed args: passed per normal rules
- Variadic args: all variadic args go on stack (no register promotion for varargs)
- Callee must save VF registers to build `va_list` if it accesses them

### C Interoperability
The `extern "C"` FFI in AXIOM must produce AAPCS64-compatible calls. When calling libc functions (e.g., `write`, `malloc`), the generated code must exactly follow these rules.

## Implementation Steps

### Step 1: Argument Classification
Create `codegen/native/arm64/abi.go`:
```go
package arm64

type ArgClass int
const (
    ClassInteger ArgClass = iota
    ClassFloat
    ClassHFA
    ClassComposite
    ClassEmpty
)

func ClassifyArg(t types.Type) ArgClass {
    switch t := t.(type) {
    case *types.IntType, *types.PtrType, *types.BoolType:
        return ClassInteger
    case *types.F32Type, *types.F64Type:
        return ClassFloat
    case *types.StructType:
        if isHFA(t) { return ClassHFA }
        return ClassComposite
    case *types.VoidType:
        return ClassEmpty
    default:
        return ClassInteger // safe default
    }
}

func isHFA(st *types.StructType) bool {
    if len(st.Fields) < 2 || len(st.Fields) > 4 { return false }
    baseType := st.Fields[0].Type
    if _, ok := baseType.(*types.F32Type); !ok {
        if _, ok := baseType.(*types.F64Type); !ok { return false }
    }
    for _, f := range st.Fields[1:] {
        if !typesEqual(f.Type, baseType) { return false }
    }
    return true
}
```

### Step 2: Argument Assignment
```go
type ArgLoc struct {
    InReg  bool
    Reg    Reg    // if InReg
    Offset int    // if !InReg, offset from SP at call site
    Size   int
}

type ABIAssignment struct {
    Args   []ArgLoc
    Return ArgLoc
    StackSize int  // total stack space for stack args
}

func AssignArgs(sig *types.FuncType) *ABIAssignment {
    ngrn, nsrn := 0, 0
    stackOffset := 0
    locs := make([]ArgLoc, len(sig.Params))

    for i, param := range sig.Params {
        class := ClassifyArg(param)
        switch class {
        case ClassInteger:
            if ngrn < 8 {
                locs[i] = ArgLoc{InReg: true, Reg: intArgRegs[ngrn], Size: 8}
                ngrn++
            } else {
                sz := align(typeSize(param), 8)
                locs[i] = ArgLoc{InReg: false, Offset: stackOffset, Size: sz}
                stackOffset += sz
            }
        case ClassFloat:
            if nsrn < 8 {
                locs[i] = ArgLoc{InReg: true, Reg: fpArgRegs[nsrn], Size: fpTypeSize(param)}
                nsrn++
            } else {
                locs[i] = ArgLoc{InReg: false, Offset: stackOffset, Size: 8}
                stackOffset += 8
            }
        case ClassComposite:
            sz := typeSize(param)
            if sz <= 16 && ngrn + (sz+7)/8 <= 8 {
                n := (sz + 7) / 8
                locs[i] = ArgLoc{InReg: true, Reg: intArgRegs[ngrn], Size: sz}
                ngrn += n
            } else {
                ngrn = 8 // no more reg args
                locs[i] = ArgLoc{InReg: false, Offset: stackOffset, Size: align(sz, 8)}
                stackOffset += align(sz, 8)
            }
        }
    }
    return &ABIAssignment{Args: locs, StackSize: align(stackOffset, 16)}
}
```

### Step 3: Call Lowering
Create `codegen/native/arm64/call_lowering.go`:
```go
func LowerCall(call *air.CallInstr, abi *ABIAssignment, sel *InstrSelector) []MachineInstr {
    instrs := []MachineInstr{}

    // Allocate stack space for stack args
    if abi.StackSize > 0 {
        instrs = append(instrs, MachineInstr{
            Opcode: SUB, Rd: SP, Rn: SP, Imm: int64(abi.StackSize),
        })
    }

    // Move args to their assigned locations
    for i, loc := range abi.Args {
        src := sel.regmap[call.Args[i]]
        if loc.InReg {
            instrs = append(instrs, MachineInstr{Opcode: MOV, Rd: loc.Reg, Rn: src})
        } else {
            instrs = append(instrs, MachineInstr{
                Opcode: STR, Rn: src, Base: SP, Offset: int64(loc.Offset),
            })
        }
    }

    // Emit call
    instrs = append(instrs, MachineInstr{Opcode: BL, Label: call.Target})

    // Recover return value
    if call.Dst != air.VRegNone {
        dst := sel.regmap[call.Dst]
        retReg := X0 // or V0 for float return
        if isFloatType(call.RetType) { retReg = V0 }
        instrs = append(instrs, MachineInstr{Opcode: MOV, Rd: dst, Rn: retReg})
    }

    // Deallocate stack arg space
    if abi.StackSize > 0 {
        instrs = append(instrs, MachineInstr{
            Opcode: ADD, Rd: SP, Rn: SP, Imm: int64(abi.StackSize),
        })
    }
    return instrs
}
```

### Step 4: Return Lowering
```go
func LowerReturn(ret *air.ReturnInstr, abi *ABIAssignment, sel *InstrSelector) []MachineInstr {
    if ret.Value == air.VRegNone {
        return []MachineInstr{{Opcode: RET}}
    }
    src := sel.regmap[ret.Value]
    retReg := X0
    if isFloatType(ret.Type) { retReg = V0 }
    return []MachineInstr{
        {Opcode: MOV, Rd: retReg, Rn: src},
        {Opcode: RET},
    }
}
```

### Step 5: C FFI Integration
For `extern "C"` declarations, the AXIOM type system must map to AAPCS64 types:
- `i8` → `signed char` → INTEGER class, 8-bit
- `*T` → pointer → INTEGER class, 64-bit
- `f64` → `double` → FLOAT class, 64-bit

Ensure that `extern "C" fn write(fd: i32, buf: *u8, count: u64) -> i64` is lowered correctly:
- `fd` → W0 (32-bit, zero-extended to X0)
- `buf` → X1
- `count` → X2
- return → X0

## Test Plan

### Unit Tests
1. `TestClassifyInteger` — i8, i16, i32, i64, u64, pointer all → ClassInteger
2. `TestClassifyFloat` — f32, f64 → ClassFloat
3. `TestClassifyHFA` — struct{f32, f32} → ClassHFA; struct{f32, f64} → ClassComposite
4. `TestAssignFirst8IntArgs` — 8 integer args all get X0-X7
5. `TestAssignNinthIntArg` — 9th integer arg goes to stack at offset 0
6. `TestAssignMixedArgs` — `(f64, i32, f32)` → V0=f64, X0=i32, V1=f32
7. `TestStructLe16InRegs` — struct{i64, i64} → X0+X1
8. `TestStructGt16OnStack` — struct{i64, i64, i64} → pointer in X0
9. `TestHFAAssignment` — struct{f32, f32, f32} → V0, V1, V2
10. `TestReturnFloat` — f64 return → V0
11. `TestReturnLargeStruct` — > 16 bytes → X8 hidden pointer

### Integration Tests
1. Call a C function (`printf`) with correct args, verify output
2. Return a struct from AXIOM function, verify caller receives it correctly
3. Variadic call (e.g., `printf(fmt, arg1, arg2)`) produces correct stack layout

## Validation Checklist
- [ ] Integer args X0-X7 assigned in order
- [ ] Float args V0-V7 assigned in order, independently of integer args
- [ ] 9th+ integer args go to stack (NGRN ≥ 8)
- [ ] Composite structs ≤ 16 bytes use register pair when available
- [ ] Composite structs > 16 bytes: caller passes pointer, callee writes through it
- [ ] HFA with 2-4 identical float/double members uses SIMD register sequence
- [ ] Stack arg area is 16-byte aligned
- [ ] Return values for f32/f64 use V0, not X0
- [ ] `extern "C"` FFI produces ABI-compatible call sequences

## Acceptance Criteria
1. ABI assignment matches AAPCS64 specification for all type categories
2. C interop test: AXIOM program calls `write(1, "hello\n", 6)` via FFI and succeeds on ARM64 Linux
3. Struct return test: function returning a 16-byte struct writes to X0+X1 correctly
4. HFA test: struct with 3 f32 fields returns in V0, V1, V2
5. All unit tests pass

## Definition of Done
- `abi.go` and `call_lowering.go` implemented and reviewed
- Unit tests ≥ 90% coverage
- Integration test with real C call passes on ARM64 Linux or macOS
- No silent ABI violations — all edge cases (9th arg, large struct, HFA) tested

## Risks & Mitigations
| Risk | Mitigation |
|---|---|
| HFA detection incorrect | Test against known C struct layouts using `offsetof`/`sizeof` |
| Stack alignment off by one | Use asserts; misalignment causes SIGBUS on ARM64 (not just performance degradation) |
| macOS vs Linux ABI differences | Target-aware: macOS uses X18 as platform register, must not allocate it |
| C varargs ABI complex | Defer varargs AXIOM syntax to Phase 9; for now, only call C varargs with fixed-type wrappers |

## Future Follow-up Tasks
- p13-t04: Mach-O integration (uses this ABI for Apple Silicon)
- p16-t07: `std.io` uses `extern "C"` calls that must follow AAPCS64
- Windows ARM64 ABI: Microsoft uses a slightly different calling convention (fewer HFA rules) — future task
