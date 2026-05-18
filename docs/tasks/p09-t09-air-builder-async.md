# p09-t09: AIR Builder — Async State Machine

## Purpose
Lower async functions to state machine AIR. In MVP, async functions execute synchronously; this task establishes the state machine structure that Phase 15's async executor will use for real async execution.

## Context
An async function is split at each `await` point into numbered states. Each state is a basic block group that runs until the next await. The state machine tracks which state to resume from, enabling cooperative multitasking. In MVP (no real async scheduler yet), `await` is a synchronous call — this task adds the structural annotation without breaking the MVP C-backend.

## Inputs
- AirFunc with `IsAsync=true` flag
- `await` points identified in the CFG (OpSpawn, OpRecv instructions)
- AirFuncBuilder

## Outputs
- State machine structure in AirFunc: state number annotations in block metadata
- `ir/air/builder/async.go`

## Dependencies
- p09-t08: air-builder-control-flow — block structure built first
- p05-t05: async-type-annotation — IsAsync flag set in type checker
- p09-t03: air-metadata-table — state numbers stored in metadata

## Subsystems Affected
- AIR: async functions have state machine structure
- C-backend: in MVP, ignores state machine (emits sync code)
- Runtime (Phase 15): uses state machine for real async

## Detailed Requirements

1. After lowering an async function's body (p09-t08), run the async splitter:
   - Find all blocks containing OpRecv or explicit await points
   - Split function at each await into "before await" and "after await" groups
   - Number states starting from 0 (0 = initial entry state)
   - Add `state_number` to block metadata
2. `AirFunc.StateMachine *StateMachineInfo` (nil for sync functions):
   ```go
   type StateMachineInfo struct {
       NumStates   uint32
       StateBlocks [][]uint32  // state N → block IDs
       ResumeBlock uint32      // entry block for resuming
   }
   ```
3. In MVP C-backend: if IsAsync, emit function as normal (ignore state machine — synchronous execution).
4. In AIR printer: show state number in `; state=N` comment when state machine exists.
5. Future Phase 15: C-backend emits switch statement over state number for real async.

## Implementation Steps

1. Create `ir/air/builder/async.go`.
2. Implement `buildStateMachine(fn *AirFunc) *StateMachineInfo`.
3. Find await points: scan all instructions for OpRecv, and CallExpr to async functions.
4. Group blocks into states by await boundaries.
5. Store StateMachineInfo in AirFunc.
6. Update AIR printer to show state annotations.
7. MVP: ensure C-backend works correctly with async functions (sync execution).
8. Write tests: `TestAsyncSingleAwait`, `TestAsyncMultipleAwaits`, `TestSyncFuncNoStateMachine`.

## Test Plan

- `TestAsyncSingleAwait`: `async fn foo(): let x = await bar()` → 2 states
- `TestAsyncMultipleAwaits`: 3 awaits → 4 states
- `TestSyncFuncNoStateMachine`: regular function → StateMachine is nil
- `TestAsyncMVP`: async function compiles and runs correctly in MVP (synchronous)

## Validation Checklist

- [ ] All await points create state boundaries
- [ ] State numbers sequential starting from 0
- [ ] Non-async functions have nil StateMachine
- [ ] MVP C-backend produces working (synchronous) code for async functions
- [ ] AIR verifier passes on async functions

## Acceptance Criteria

- Compliance tests 061-070 (async group) pass in MVP synchronous mode
- State machine structure correctly identifies await points

## Definition of Done

- [ ] `ir/air/builder/async.go` implemented
- [ ] StateMachineInfo populated for async functions
- [ ] MVP synchronous execution works correctly
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| State machine interferes with MVP sync execution | MVP C-backend explicitly ignores state machine; sync path unchanged |
| Await inside loop creates complex state structure | Document limitation; defer proper async loop support to Phase 15 |

## Future Follow-up Tasks

- p15-t09: async-state-machine-executor implements real resumption
- p15-t08: async-executor-io wires I/O completion to actor mailbox
