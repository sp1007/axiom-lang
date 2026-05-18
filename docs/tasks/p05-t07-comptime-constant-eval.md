# p05-t07: Compile-Time Constant Evaluation (`#run` Stub)

## Purpose
Implement basic compile-time evaluation for `#run` expressions that are pure constant expressions. This is the MVP stub — it evaluates arithmetic, boolean logic, and string concatenation at compile time, substituting the result as a constant in the AST. A full bytecode interpreter is deferred to p10-t06.

## Context
Plan §9.1 lists `#run compile-time eval` as an early optimization (Phase 4-5) with the note: _"Pure functions only"_. Plan §12.7 confirms: _"`#run` (full VM): constant folding only"_ for the stub phase.

`#run` allows compile-time execution:
```axiom
const PI = 3.14159
const TAU = #run PI * 2.0    // evaluated at compile time → 6.28318
const MSG = #run "hello" + " " + "world"  // → "hello world"
```

## Inputs
- TypedAST with `#run` expression nodes
- TypeTable for constant type verification
- Pure function detection from effects system (p04-t09)

## Outputs
- `compiler/sema/comptime.go` — compile-time evaluator for constant expressions
- Updated TypedAST: `#run` nodes replaced with literal constant nodes

## Dependencies
- p04-t09: effects-system — pure function detection
- p04-t07: type-checker-expressions — expression types resolved
- p05-t01: generic-type-representation — generic constants possible

## Subsystems Affected
- Type checker: `#run` expressions evaluated during type checking
- AST: `#run` nodes replaced with constant literal nodes

## Detailed Requirements

### 1. Supported Operations (MVP)

| Category | Operations |
|----------|-----------|
| Integer arithmetic | `+`, `-`, `*`, `/`, `%` on `i8`–`i64`, `u8`–`u64` |
| Float arithmetic | `+`, `-`, `*`, `/` on `f32`, `f64` |
| Boolean logic | `and`, `or`, `not`, comparisons |
| String concat | `+` on `string` literals |
| Constant refs | Named `const` values |

### 2. NOT Supported (Deferred to p10-t06)

- Function calls (even pure ones)
- Loops or conditionals
- Variable declarations
- I/O operations
- Type construction (struct literals)

### 3. Evaluator API

```go
type ComptimeEvaluator struct {
    types  *TypeTable
    consts map[uint32]ComptimeValue  // symbolIdx → evaluated value
}

type ComptimeValue struct {
    Kind     TypeID
    IntVal   int64
    FloatVal float64
    StrVal   string
    BoolVal  bool
}

func (e *ComptimeEvaluator) Eval(nodeIdx uint32, ast *Ast) (ComptimeValue, error)
```

### 4. Error Handling

If `#run` contains non-constant expressions:
```
error[E1500]: cannot evaluate at compile time
  --> main.ax:5:15
   |
 5 | const X = #run read_file("data.txt")
   |               ^^^^^^^^^^^^^^^^^^^^^^ function calls not supported in #run (MVP)
   |
note: full #run support available in a future version
```

## Implementation Steps

1. Create `compiler/sema/comptime.go`.
2. Implement `Eval()` — recursive evaluation of constant expressions.
3. Handle integer ops with overflow detection.
4. Handle float ops.
5. Handle string concatenation.
6. Handle `const` references (look up in `consts` map).
7. Integrate into type checker: when encountering `#run` node, call `Eval()` and replace node.
8. Write tests for each operation type.

## Test Plan

- `TestComptimeIntArith`: `#run 2 + 3` → 5
- `TestComptimeFloatArith`: `#run 3.14 * 2.0` → 6.28
- `TestComptimeBoolLogic`: `#run true and false` → false
- `TestComptimeStringConcat`: `#run "a" + "b"` → "ab"
- `TestComptimeConstRef`: `const A = 5; const B = #run A * 2` → 10
- `TestComptimeOverflow`: `#run 9223372036854775807 + 1` → compile error
- `TestComptimeNonConstant`: `#run foo()` → error "cannot evaluate"

## Validation Checklist

- [ ] Integer arithmetic evaluated correctly at compile time
- [ ] Float arithmetic evaluated correctly
- [ ] String concatenation works
- [ ] Const references resolved
- [ ] Non-constant expressions produce clear error messages
- [ ] Overflow detected and reported

## Acceptance Criteria

- `const TAU = #run 3.14159 * 2.0` results in `TAU = 6.28318` in the TypedAST
- Non-evaluable `#run` expressions produce clear compile errors

## Definition of Done

- [ ] `compiler/sema/comptime.go` implemented
- [ ] Unit tests pass
- [ ] Integrated into type checker pipeline

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Infinite recursion in const references | Detect cycles via "Evaluating" flag on each const symbol |
| Floating point precision differences across platforms | Use IEEE 754 double precision; accept platform differences for now |

## Future Follow-up Tasks

- p10-t06: Full compile-time interpreter with function calls, loops, and conditionals
- p09-t06: AIR builder emits constants directly for `#run` results
