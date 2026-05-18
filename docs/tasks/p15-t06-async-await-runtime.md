# p15-t06: Async/Await Runtime

## Purpose
Implement the runtime state machine execution model for `async`/`await` in AXIOM, enabling non-blocking async functions to suspend and resume within the actor scheduler without blocking OS threads.

## Context
AXIOM's `async fn` compiles to a state machine (via p09-t09). The runtime must be able to: suspend a state machine when it hits `await`, return the OS thread to the scheduler, and resume the state machine when the awaited future completes. This enables high-concurrency I/O without thread-per-connection.

## Inputs
- `StateMachineInfo` from p09-t09 (AIR async lowering)
- Actor scheduler from p15-t02
- Future[T] type from p05-t05

## Outputs
- `runtime/async.c` — async state machine executor
- `AxFuture`, `AxTask` C structs
- `ax_future_await()`, `ax_future_resolve()` API

## Dependencies
- p09-t09: air-builder-async — state machine structure
- p15-t02: scheduler — task re-enqueue on resume
- p15-t01: actor-struct — async tasks run in actor context

## Subsystems Affected
- Scheduler: async tasks are schedulable units (like actors)
- I/O subsystem (future): I/O completion wakes async tasks

## Detailed Requirements

```c
typedef enum AxFutureState {
    AX_FUTURE_PENDING   = 0,
    AX_FUTURE_RESOLVED  = 1,
    AX_FUTURE_REJECTED  = 2,
} AxFutureState;

typedef struct AxFuture {
    _Atomic AxFutureState state;
    void*   result;          // resolved value (typed by type_id)
    uint32_t result_type_id;
    void*   error;           // if rejected
    struct AxTask* waiter;   // task waiting on this future
} AxFuture;

typedef struct AxTask {
    void*    state_machine;  // compiled async state machine instance
    AxStepFn step_fn;        // compiled step function (takes state, returns next_state or DONE)
    int      current_state;  // current state machine state
    AxFuture* awaiting;      // NULL if not suspended
    uint64_t actor_id;       // actor this task belongs to
} AxTask;

typedef int (*AxStepFn)(void* sm, int state, void** out_result);

// Poll a task: run until it suspends or completes
AxFutureState ax_task_poll(AxTask* task);

// Suspend task waiting on future
void ax_task_await(AxTask* task, AxFuture* future);

// Resolve future: wake up all waiters
void ax_future_resolve(AxFuture* future, void* result);
void ax_future_reject(AxFuture* future, void* error);

// Create a new future (in actor's heap)
AxFuture* ax_future_new(ActorHeap* heap);
```

Execution model:
1. `ax_task_poll(task)`: call `step_fn(state_machine, current_state)`.
2. If step returns `AWAITING`: task suspends; `ax_task_await(task, future)` records waiter.
3. When `ax_future_resolve()` called: re-enqueue task in scheduler.
4. Task resumes from `current_state`; step_fn transitions to next state.

## Implementation Steps

1. Create `runtime/async.c`.
2. Define `AxFuture`, `AxTask` structs.
3. Implement `ax_task_poll()` — run step_fn in loop until suspend or done.
4. Implement `ax_task_await()` — set awaiting, return to scheduler.
5. Implement `ax_future_resolve()` — CAS state, wake waiter (re-enqueue).
6. Implement `ax_future_new()` — allocate in actor heap.
7. Wire into scheduler: tasks are first-class schedulable units.

## Test Plan
- `TestAsyncSimple`: async fn returns value → future resolved correctly
- `TestAsyncAwait`: await another future → suspends, resumes after resolve
- `TestAsyncChain`: async fn awaiting another async fn → chain resolves in order
- `TestAsyncConcurrent`: 1000 concurrent tasks, all resolve correctly

## Validation Checklist
- [ ] Task suspends correctly (does not busy-wait)
- [ ] Future resolved exactly once (atomic CAS)
- [ ] Waiter re-enqueued after resolve
- [ ] No deadlock between tasks awaiting each other

## Acceptance Criteria
- 100K concurrent async tasks complete in < 1 second

## Definition of Done
- [ ] `runtime/async.c` implemented
- [ ] Async chain test passes

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Deadlock from circular await | Detect at compile time via effect system (p04-t09); panic at runtime |
| Task starved (never re-enqueued) | Verify re-enqueue path in all resolve branches |

## Future Follow-up Tasks
- I/O event loop integration (epoll/IOCP) for async I/O
- Timeout support: `await with_timeout(future, 5s)`
