# p06-t04: Escape Analysis

## Purpose
Determine statically whether each allocation escapes its declaring scope (and must be heap-allocated) or stays within its scope (and can be stack-allocated). Stack allocation eliminates heap allocation overhead and makes CTGC automatic without runtime tracking.

## Context
Escape analysis uses the Connection Graph built by p06-t02. A value "escapes" if it has an `EscapesTo` edge pointing to a node with a longer lifetime. Values that don't escape can be placed on the stack with zero runtime cost. This is the same analysis performed by Go's escape analysis and Java's JIT, but done statically at compile time in AXIOM.

## Inputs
- Populated ConnectionGraph from p06-t02 (EscapesTo edges already set)
- SymbolTable — scope depth (lifetime) of each symbol
- TypeTable — type sizes for stack allocation decisions

## Outputs
- Escape annotation on each VarDecl node: `Flags |= EscapesToHeap` or stays clear (stack)
- `EscapeReport{allocatesOnHeap []uint32, allocatesOnStack []uint32}` for profiling

## Dependencies
- p06-t02: ownership-rules — ConnectionGraph with EscapesTo edges populated
- p06-t01: connection-graph — Escapes() API

## Subsystems Affected
- Code generation: stack allocs use local variables, heap allocs use ax_alloc()
- CTGC (p06-t05): only heap allocations need =destroy injection
- Performance: stack allocation is the common case for well-written code

## Detailed Requirements

1. `EscapeAnalysis` struct: `cg *ConnectionGraph, tt *TypeTable, tree *AstTree`
2. `Analyze(funcNodeIdx uint32)` — runs escape analysis for one function:
   - For each VarDecl in the function: call `cg.Escapes(nodeID)`
   - If escapes → set `node.Flags |= EscapesToHeap`
   - If not → `node.Flags` clear (stack allocation)
3. A value escapes if ANY of:
   - It is returned from the function (`EscapesTo` RETURN_SLOT)
   - It is stored in a field of a heap-allocated value (`EscapesTo` heap node)
   - It is passed to `spawn` (actor communication)
   - It is captured in a closure that outlives the declaring scope
   - Its address is taken and the pointer escapes (unsafe block)
4. A value does NOT escape if:
   - Used only locally (reads, field access, arithmetic)
   - Passed as `lent T` (borrow, does not extend lifetime)
   - Passed as `!T` (sink) to a non-escaping function
5. Very large values (> 1KB): always heap-allocate regardless of escape status (configurable threshold).
6. `--escape-report` flag: print which allocations are on stack vs heap (useful for profiling).

## Implementation Steps

1. Create `compiler/sema/escape.go`.
2. Implement `Analyze()` walking all VarDecl nodes in a function.
3. For each VarDecl: call `cg.Escapes(cg.NodeOfSym(symID))`.
4. Set `AstNode.Flags |= FlagEscapesToHeap` when escaping.
5. Implement size threshold: `if typeInfo.Size > 1024 { setEscapesToHeap() }`.
6. Implement `--escape-report` output in `cmd/axc/`.
7. Write tests: `TestEscapeReturn`, `TestNoEscapeLocal`, `TestEscapeClosure`, `TestEscapeHeapStore`.

## Test Plan

- `TestNoEscapeSimple`: `let x = 5; return x + 1` — x is i32, used locally → stack
- `TestEscapeReturn`: `let x = Foo{}; return x` — x returned → heap
- `TestEscapeHeapStore`: `let x = Foo{}; heap_vec.push(x)` — stored in heap → x escapes
- `TestEscapeSpawn`: `let x = Foo{}; spawn worker(x)` — must be Isolated, escapes to actor
- `TestBorrowNoEscape`: `fn use(f: lent Foo)` called with x — x not moved, no escape
- `TestClosureCapture`: closure capturing x — x escapes if closure outlives scope

## Validation Checklist

- [ ] Returned values marked as EscapesToHeap
- [ ] Local-only values NOT marked as EscapesToHeap
- [ ] Size threshold applied (> 1KB → heap even if no escape)
- [ ] Spawn site values correctly marked
- [ ] Borrow (lent) does not cause escape

## Acceptance Criteria

- Simple functions (add two ints) have zero heap allocations
- Fibonacci returns a value that correctly escapes to heap if returned from main
- `--escape-report` flag shows allocation sites

## Definition of Done

- [ ] `compiler/sema/escape.go` implemented
- [ ] Escape flags set correctly on all VarDecl nodes
- [ ] Unit tests pass
- [ ] Integrated into compiler pipeline before codegen

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Conservative over-approximation (too many heap allocations) | Acceptable for correctness; optimize with flow sensitivity later |
| Escape analysis misses closure captures | Explicit closure capture analysis in future RFC |

## Future Follow-up Tasks

- p06-t05: CTGC-destroy-injection only injects for heap-allocated values
- p08-t06: cgen-generational-checks only emits gen_id checks for heap values
- p10-t05: opt-ctgc-air further optimizes based on escape results
