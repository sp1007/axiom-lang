# p15-t07: Supervisor Tree

## Purpose
Implement the supervisor tree runtime infrastructure: parent actors that monitor children and apply restart strategies (ONE_FOR_ONE, ALL_FOR_ONE, REST_FOR_ONE) when child actors crash.

## Context
AXIOM's fault tolerance model follows Erlang/OTP supervision. When an actor crashes (unhandled panic), its supervisor receives a `ChildDown` message and decides whether to restart the child, restart all children, or escalate to its own supervisor. This enables self-healing systems.

## Inputs
- `AxActor` with `supervisor_id` field from p15-t01
- Panic handler crash signal from p07-t03
- Restart strategy definitions

## Outputs
- `runtime/supervisor.c` — supervisor implementation
- `ax_supervisor_spawn()`, `ax_supervisor_handle_child_down()` API

## Dependencies
- p15-t01: actor-struct — supervisor_id linkage
- p07-t03: panic-handler — notifies supervisor on actor crash
- p15-t05: actor-message-queue — ChildDown message delivery

## Subsystems Affected
- Actor lifecycle: all non-root actors should have a supervisor
- Panic handler: on actor crash, enqueue ChildDown to supervisor

## Detailed Requirements

```c
typedef enum AxRestartStrategy {
    AX_RESTART_ONE_FOR_ONE  = 0,  // restart only crashed child
    AX_RESTART_ALL_FOR_ONE  = 1,  // restart all children
    AX_RESTART_REST_FOR_ONE = 2,  // restart crashed + all started after it
} AxRestartStrategy;

typedef struct AxSupervisorSpec {
    AxRestartStrategy strategy;
    uint32_t max_restarts;    // max in time window
    uint32_t time_window_sec; // restart window
} AxSupervisorSpec;

typedef struct AxChildSpec {
    AxHandlerFn handler;
    void*       init_state;
    size_t      state_size;
    char        name[64];
} AxChildSpec;

typedef struct AxSupervisor {
    AxSupervisorSpec spec;
    uint64_t  children[AX_MAX_CHILDREN];  // actor IDs
    uint32_t  child_count;
    uint32_t  restart_counts[AX_MAX_CHILDREN];
    uint64_t  last_restart_time;
} AxSupervisor;

// Spawn a supervisor actor
uint64_t ax_supervisor_spawn(AxSupervisorSpec spec, AxChildSpec* children, int n);

// Called by runtime when child crashes
void ax_supervisor_child_down(uint64_t supervisor_id, uint64_t child_id, const char* reason);

// Internal: supervisor message handler
void ax_supervisor_handle(AxActor* self, void* msg, uint32_t msg_type);
```

Restart flow:
1. Child actor panics → `ax_crash_cleanup()` → `ax_supervisor_child_down(supervisor_id, child_id, reason)`.
2. Supervisor receives `MSG_CHILD_DOWN`.
3. Check restart count in time window → if exceeded, crash supervisor (escalate up).
4. Apply strategy: ONE_FOR_ONE → respawn only child; ALL_FOR_ONE → stop + respawn all.
5. Respawn: call `ax_actor_spawn()` with original ChildSpec.

## Implementation Steps

1. Create `runtime/supervisor.c`.
2. Define `AxRestartStrategy`, `AxSupervisorSpec`, `AxChildSpec`, `AxSupervisor`.
3. Implement `ax_supervisor_spawn()` — spawn supervisor actor + children.
4. Implement `ax_supervisor_handle()` — process ChildDown messages.
5. Implement restart strategy dispatch.
6. Implement restart frequency check (max_restarts in time_window_sec).
7. Wire panic handler to call `ax_supervisor_child_down()`.
8. Write fault injection tests.

## Test Plan
- `TestSupervisorOneForOne`: child crashes → only that child restarted
- `TestSupervisorAllForOne`: child crashes → all children restarted
- `TestSupervisorMaxRestarts`: exceed max_restarts → supervisor itself crashes (escalates)
- `TestSupervisorEscalation`: root supervisor crash → process terminates cleanly

## Validation Checklist
- [ ] Crashed child always reported to supervisor (no silent crash)
- [ ] Restart count resets after time_window_sec
- [ ] Escalation occurs when max_restarts exceeded
- [ ] New child actor has fresh heap (not reusing crashed actor's memory)

## Acceptance Criteria
- Actor crashing 3 times within 1 second triggers supervisor escalation

## Definition of Done
- [ ] `runtime/supervisor.c` implemented
- [ ] Fault injection tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Restart storm: fast-crashing child consumes all resources | Enforce max_restarts with exponential backoff |
| Supervisor itself crashes during restart → orphaned children | Root supervisor never crashes; always escalates to process exit |

## Future Follow-up Tasks
- Dynamic supervision (add/remove children at runtime)
- Supervisor telemetry: crash rate metrics
