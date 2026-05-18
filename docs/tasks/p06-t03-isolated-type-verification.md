# p06-t03: Isolated[T] Type Verification

## Purpose
Verify at compile time that values of type `Isolated[T]` have no external references — their entire object subgraph is self-contained. This guarantees safe zero-copy message passing between actors without data races or use-after-send errors.

## Context
`Isolated[T]` is AXIOM's mechanism for safe inter-actor communication. When an actor sends an `Isolated[T]` value to another actor, ownership transfers atomically — no copying needed. The compile-time verifier proves that no external code holds references to the isolated value, making the transfer safe. This uses the Connection Graph built by p06-t02.

## Inputs
- Populated ConnectionGraph from p06-t02
- TypeTable — to identify `Isolated[T]` typed variables
- SpawnExpr and SendExpr AST nodes — the points where Isolated[T] is consumed

## Outputs
- `[]Diagnostic` — "value has external references, cannot be Isolated[T]"
- Boolean annotation on variables: `IsProvenIsolated bool`

## Dependencies
- p06-t02: ownership-rules — ConnectionGraph must be populated first
- p06-t01: connection-graph — `InEdges()` API used

## Subsystems Affected
- Ownership safety: Isolated[T] is the zero-copy message passing primitive
- Actor runtime (Phase 15): spawn/send calls that pass Isolated[T]
- Type system: Isolated[T] is a marker type with compile-time semantics

## Detailed Requirements

1. `IsolatedVerifier` struct: `cg *ConnectionGraph, tt *TypeTable`
2. `VerifyIsolated(valueNodeID uint32) (bool, []uint32)` — returns (isIsolated, external_node_ids_that_violate):
   - Collect subgraph of valueNodeID: all nodes reachable via Owns/FlowsTo edges
   - For each node in subgraph: check all incoming edges
   - An edge is "external" if its source is NOT in the subgraph
   - External incoming edges = violation of isolation
3. At each `spawn foo(data)` and `actor.send(data)` call site: verify the `data` argument is `Isolated[T]` and passes `VerifyIsolated`.
4. At assignment `let x: Isolated[T] = expr`: verify that expr is proven isolated at that point.
5. Error message: `"value of type Isolated[Foo] has external references from: [bar, baz]; cannot pass to spawn"`.
6. `@[unsafe_isolated]` attribute: bypass check (for raw FFI code that the programmer guarantees is isolated).
7. Special case: freshly allocated values with no borrows taken are automatically isolated.

## Implementation Steps

1. Create `compiler/sema/isolated.go` with `IsolatedVerifier`.
2. Implement subgraph collection via DFS on Owns/FlowsTo edges.
3. Implement external edge detection by checking all InEdges of each subgraph node.
4. Hook into type checker: when an `Isolated[T]` variable is used in send/spawn context, call `VerifyIsolated`.
5. Hook into assignment: when assigning to `Isolated[T]` variable, call `VerifyIsolated` on RHS.
6. Implement `@[unsafe_isolated]` bypass.
7. Write unit tests.

## Test Plan

- `TestIsolatedFreshAlloc`: `let x = Isolated(Foo{})` — no external refs → OK
- `TestIsolatedWithBorrow`: borrow `y = &x.field; let z = Isolated(x)` → external ref from y → error
- `TestIsolatedTransitive`: struct containing another struct, no external refs → OK
- `TestIsolatedAfterMove`: move x into Isolated wrapper → original invalidated → no external refs
- `TestIsolatedUnsafeBypass`: `@[unsafe_isolated]` on function parameter → no check performed

## Validation Checklist

- [ ] Fresh allocation with no borrows → proven isolated
- [ ] Any outstanding borrow → isolation violation detected
- [ ] External struct field ref → detected
- [ ] Error names the external reference holders
- [ ] `@[unsafe_isolated]` bypass works

## Acceptance Criteria

- Compliance tests 061-070 (concurrency group) type-check correctly
- Actor spawn with non-isolated value rejected with clear error

## Definition of Done

- [ ] `compiler/sema/isolated.go` implemented
- [ ] Integrated into type checker at spawn/send sites
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| False positives rejecting valid isolated values | Err on the side of allowing; `@[unsafe_isolated]` escape hatch |
| Performance: subgraph DFS for every spawn | Cache isolation status per symbol; only re-verify on mutation |

## Future Follow-up Tasks

- p15-t05: actor-spawn-mailbox calls spawn with Isolated[T] values
- p06-t04: escape analysis uses similar subgraph analysis
