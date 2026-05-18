# p09-t06: AIR Builder — Expressions

## Purpose
Implement the AIR builder for expressions — the component that lowers typed AST expression nodes to AIR instructions. Each expression returns a virtual register containing its value.

## Context
The AIR builder walks the typed, ownership-analyzed AST and emits AIR instructions. Expression lowering is the core of this process: every sub-expression recursively produces a virtual register. The builder uses the `AirFuncBuilder` from p09-t02 to emit instructions and allocate registers.

## Inputs
- Typed AST with resolved symbols and TypeIDs
- AirFuncBuilder (p09-t02) for instruction emission
- TypeTable (p04-t02) for type information

## Outputs
- `ir/air/builder/expr.go` — `lowerExpr(nodeIdx) uint32` function (returns register)

## Dependencies
- p09-t02: air-basic-blocks — AirFuncBuilder used
- p09-t01: air-instruction-set — opcodes used
- p04-t07: type-checker-expressions — TypeIDs set in AST nodes

## Subsystems Affected
- AIR builder: expressions are the leaves of the lowering process
- Optimization: arithmetic patterns recognized by constant folder

## Detailed Requirements

1. `lowerExpr(nodeIdx uint32) uint32` — returns virtual register containing expression value.
2. Expression lowering by NodeKind:
   - `Literal (IntLit)`: `%r = iconst value_lo, value_hi`
   - `Literal (FloatLit)`: `%r = fconst bits_lo, bits_hi` (IEEE 754 bit pattern)
   - `Literal (StringLit)`: `%r = load @string_literal_N` (string in rodata)
   - `Literal (BoolLit)`: `%r = iconst 1/0` (TypeBool)
   - `Literal (Nil)`: `%r = iconst 0` (null pointer)
   - `Ident`: if stack var → `%r = load %addr`; if param → `%r = param_N`
   - `BinaryExpr (+)`: lower left → %l, lower right → %r, emit `%result = iadd %l, %r`
   - `BinaryExpr (and)`: short-circuit: lower left, branch if false, lower right
   - `BinaryExpr (==)`: emit `%result = icmpeq %l, %r`
   - `UnaryExpr (-)`: `%result = neg %operand`
   - `UnaryExpr (not)`: `%result = bnot %operand` (boolean not = XOR with 1)
   - `CallExpr`: lower all args to registers, emit `%r = call @sym, %arg0, %arg1, ...`
   - `FieldExpr`: lower object → %ptr, emit `%r = load %field_ptr` where `%field_ptr = gep %ptr, field_offset`
   - `IndexExpr`: lower array/slice → %ptr, lower index → %idx, bounds-check emit, `%elem_ptr = gep %ptr, %idx`, `%r = load %elem_ptr`
   - `CastExpr (as i32)`: emit appropriate conversion opcode (IToF, FToI, ZExt, SExt, Trunc)
   - `DerefExpr (ptr.*)`: `%real_ptr = deref %ref_reg` (gen_id check), `%r = load %real_ptr`
   - `AwaitExpr`: in MVP, just lower the sub-expression (synchronous)
3. Short-circuit evaluation: `a and b` — evaluate b only if a is true.
4. String literals: stored in a string table, referenced by index in rodata.

## Implementation Steps

1. Create `ir/air/builder/expr.go`.
2. Implement `lowerExpr()` dispatch on NodeKind.
3. Implement each case above.
4. Handle short-circuit for `and`/`or`: create branch blocks.
5. Add `lowerArgs(callNodeIdx) []uint32` helper for call expressions.
6. Add `emitBoundsCheck(ptr, idx uint32)` helper for index expressions.
7. Write unit tests per expression kind.

## Test Plan

- `TestLowerIntLit`: `42` → single OpIConst instruction, result register returned
- `TestLowerBinaryAdd`: `1 + 2` → two OpIConst + one OpIAdd
- `TestLowerShortCircuit`: `a and b` → branch, not eager evaluation
- `TestLowerCallExpr`: `fn foo(x: i32)` call → args lowered, OpCall emitted
- `TestLowerFieldExpr`: `s.x` → OpGEP + OpLoad
- `TestLowerIndex`: `arr[i]` → bounds check + OpGEP + OpLoad
- `TestLowerCast`: `5 as f64` → OpIToF

## Validation Checklist

- [ ] All expression kinds produce a valid virtual register
- [ ] Short-circuit for and/or uses branch blocks
- [ ] Index access includes bounds check
- [ ] Dereference includes OpDeref (gen_id check)
- [ ] Each case produces verifier-valid AIR

## Acceptance Criteria

- Fibonacci function lowers to correct AIR (verified by printer + golden test)
- AIR verifier produces 0 errors on lowered expressions

## Definition of Done

- [ ] `ir/air/builder/expr.go` implemented
- [ ] Unit tests pass
- [ ] AIR verifier passes on all test cases

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Short-circuit logic creates complex CFG structure | Use helper `emitShortCircuit(op, left, right) uint32` |
| String literal indexing into rodata | Maintain string pool, emit OpLoad from rodata base + offset |

## Future Follow-up Tasks

- p09-t07: air-builder-statements uses lowerExpr for initializers
- p10-t02: constant folder optimizes OpIConst sequences
