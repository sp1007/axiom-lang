# p04-t09: Effects System

## Purpose
Implement the effects propagation system that tracks which side effects (raises, async, pure) each function has, and enforces that callers either handle or propagate those effects. This enables compile-time verification of error handling completeness and purity guarantees.

## Context
AXIOM's effects system is annotation-based: `{.raises: [IOError].}` declares that a function can raise IOError. If a function calls another with `raises:[IOError]` and does not handle it (via match/try), it must also declare `raises:[IOError]` in its signature. The `pure` effect means a function has no side effects. `async` means a function can suspend. Effects are checked statically without runtime overhead.

## Inputs
- Typed AST with resolved symbols
- Function TypeInfo (from TypeTable) — includes declared effects
- SymbolTable

## Outputs
- `EffectSet{Raises []TypeID, IsPure bool, IsAsync bool}` per function symbol
- `[]Diagnostic` — unhandled effect errors

## Dependencies
- p04-t07: type-checker-expressions — call sites identified
- p04-t02: type-table-primitives — error types registered as TypeIDs

## Subsystems Affected
- Type checker: effects are validated alongside types
- Code generation: async effect drives state machine generation (Phase 09)
- Standard library: `std.fs`, `std.net` declare `raises` effects

## Detailed Requirements

1. `EffectSet` struct:
   ```go
   type EffectSet struct {
       Raises  []uint32 // TypeIDs of exception types
       IsPure  bool
       IsAsync bool
   }
   ```
2. Parse effect annotations from function return type syntax: `-> i32 {.raises: [IOError].}` → `EffectSet{Raises:[TypeIDOfIOError]}`.
3. `@pure` attribute on function → `IsPure=true`.
4. `async fn` → `IsAsync=true`.
5. Effect propagation algorithm:
   - For each call in function body: get callee's EffectSet
   - For each `raise` in callee: check if handled by enclosing match/try block; if not, add to caller's required EffectSet
   - If caller's declared EffectSet doesn't include a required effect → error
6. `effectsOf(fn_name)` intrinsic: returns EffectSet of a function (for use in generic constraints).
7. `pure` function calling non-pure → error: "pure function cannot call impure function".
8. MVP: `try expr` syntax → unwrap Result type, propagate error on Err variant. `?` operator deferred to RFC.

## Implementation Steps

1. Create `compiler/sema/effects.go`.
2. Parse `{.raises: [...].}` syntax in type checker — extract TypeIDs from the raises list.
3. Build `FuncEffects map[uint32]EffectSet` (symID → EffectSet) for all functions.
4. Propagation pass: for each function, walk call sites, collect required raises, compare with declared raises.
5. Emit diagnostic: `"function 'main' calls 'fs.read' which raises IOError, but does not handle or declare it"`.
6. Handle `match result { Ok(v): ... | Err(e): ... }` as error handling that satisfies the IOError raise.
7. Handle `async fn` call from non-async context → error: "cannot await in non-async function".
8. Write tests: propagation, missing handler, correct declaration.

## Test Plan

- `TestEffectPropagation`: fn A calls fn B{raises:[IOError]} without handling → A must declare raises or error
- `TestEffectHandled`: fn A calls fn B{raises:[IOError]} with try/match → no error
- `TestEffectPure`: pure fn calling impure fn → error
- `TestEffectAsync`: `await` in non-async fn → error
- `TestEffectsOf`: `effectsOf(fs.read)` returns `{raises:[IOError]}`

## Validation Checklist

- [ ] All `raises` declarations parsed correctly
- [ ] Unhandled effects propagate to caller
- [ ] Handled effects (match/try) do not propagate
- [ ] Pure functions cannot call impure functions
- [ ] Async effects propagate correctly
- [ ] `effectsOf()` intrinsic returns correct EffectSet

## Acceptance Criteria

- Compliance tests 081-090 (stdlib) pass with correct effect checking
- A function using `std.fs.open` without declaring `raises:[IOError]` fails compilation

## Definition of Done

- [ ] `compiler/sema/effects.go` implemented
- [ ] `go test ./compiler/sema/ -run TestEffect` passes
- [ ] Effects integrated into compiler pipeline

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Effect propagation cycles (mutual recursion) | Use iterative fixed-point algorithm, not recursive |
| Over-strict pure checking rejecting valid stdlib use | Start with relaxed pure checking; tighten with explicit tests |

## Future Follow-up Tasks

- p09-t09: AIR builder generates state machines for async functions
- p16-t08: std.fs declares its raises effects using this system
