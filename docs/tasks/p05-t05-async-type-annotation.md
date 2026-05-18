# p05-t05: Async Type Annotation

## Purpose
Add async function type annotation and `await` expression type checking, establishing the compile-time contract for asynchronous code. In MVP, execution is synchronous; the type system tracks async boundaries to enable future state machine generation in the AIR phase.

## Context
`async fn foo() -> i32` declares that foo is an async function. Its return type in the type system is `Future[i32]`, but callers using `await foo()` see `i32` as the result type. The type checker enforces that: (1) `await` only appears inside `async fn`, and (2) `await` applied to a `Future[T]` produces `T`. The actual async executor is implemented in Phase 15; in MVP, awaiting a future just calls the function synchronously.

## Inputs
- TypeTable (p04-t02) — `Future[T]` registered as generic struct
- Function TypeInfo — `IsAsync bool` field
- Effects system (p04-t09) — async is an effect

## Outputs
- `TypeID` for `Future[T]` generic struct in TypeTable
- `Symbol.IsAsync` flag populated
- Type checking for `async fn` and `await expr`
- `compiler/sema/async_types.go`

## Dependencies
- p04-t09: effects-system — async is an effect that propagates
- p05-t01: generic-type-representation — `Future[T]` is a generic type

## Subsystems Affected
- Type system: Future[T] generic type
- Effects: async effect propagation
- AIR builder (Phase 09): generates state machine code for async fns
- Runtime (Phase 15): executes async state machines

## Detailed Requirements

1. Register `Future[T]` as a generic struct in TypeTable during initialization.
2. `async fn foo() -> i32`: function TypeInfo gets `IsAsync=true`, declared return type is `i32`, but the function symbol's type is `fn() -> Future[i32]`.
3. `await expr`: if expr type is `Future[T]`, result type is `T`. Error if not Future[T].
4. `await` only valid inside `async fn` — check via context flag `inAsyncFn bool` during type checking.
5. Calling `async fn` without `await`: returns `Future[T]` value (can be stored, passed to `spawn`, etc.).
6. `spawn async_fn(args)` → spawns a new actor running the async function; returns `ActorRef`.
7. MVP behavior: `Future[T]` is implemented as a plain value holder (no actual async); `await` is an identity operation on the payload.
8. Type error: `await` in non-async function → "await can only be used inside async functions".

## Implementation Steps

1. Create `compiler/sema/async_types.go`.
2. In `NewTypeTable()`: register `Future` as a generic struct template.
3. Add `IsAsync bool` to function `TypeInfo` and parse from `async fn` syntax.
4. In type checker: when entering `async fn`, set `ctx.inAsyncFn = true`.
5. In `checkExpr` for `AwaitExpr`: verify `ctx.inAsyncFn`, verify operand type is `Future[T]`, return T.
6. In `checkExpr` for `SpawnExpr`: verify operand is a function call, result type is `ActorRef`.
7. Write tests: `TestAsyncFnType`, `TestAwaitType`, `TestAwaitOutsideAsync`, `TestSpawnType`.

## Test Plan

- `TestAsyncFnType`: `async fn foo() -> i32` → symbol type is `fn() -> Future[i32]`
- `TestAwaitType`: `let x = await foo()` where foo is `async fn -> i32` → x:i32
- `TestAwaitOutsideAsync`: `await foo()` in non-async fn → error
- `TestAwaitNonFuture`: `await 42` → "await requires Future[T], found i32"
- `TestSpawnType`: `let ref = spawn foo()` → ActorRef type
- `TestCompliance061`: compliance tests 061-070 (async/actor group) parse and type-check

## Validation Checklist

- [ ] `Future[T]` registered as generic type in TypeTable
- [ ] `async fn` correctly marks symbol as async
- [ ] `await` validates it's inside async function
- [ ] `await` result type is the unwrapped T
- [ ] Spawn returns ActorRef type
- [ ] Async effect propagates to callers

## Acceptance Criteria

- `async fn`, `await`, `spawn` type-check correctly
- Error messages are specific about async context violations

## Definition of Done

- [ ] `compiler/sema/async_types.go` implemented
- [ ] `Future[T]` in TypeTable
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Async type complexity grows before AIR state machine is ready | Keep MVP synchronous; only add state machine lowering in Phase 09 |
| Await in nested closures inside async fn | Track async context through closure boundaries |

## Future Follow-up Tasks

- p09-t09: AIR-builder-async generates state machine from async fn
- p15-t09: async-state-machine-executor runs the state machines
