# p15-t01: Actor Struct and Lifecycle

## Purpose
Define the core `AxActor` data structure and implement actor spawn, run, and termination lifecycle, forming the foundation of AXIOM's actor-based concurrency model.

## Context
AXIOM actors are lightweight concurrent entities with isolated heaps, message queues, and supervisor relationships. Each actor runs a message-processing loop until it receives a `Stop` signal or crashes. The runtime represents actors as `AxActor` structs managed by the scheduler.

## Inputs
- AXIOM actor syntax: `actor MyActor { fn handle(msg: Msg) { ... } }`
- `ActorHeap` from p14-t04 — per-actor memory
- Scheduler from p15-t02 — assigns actors to OS threads

## Outputs
- `runtime/actor.h` — `AxActor` struct definition
- `runtime/actor.c` — actor lifecycle implementation
- `ax_actor_spawn()`, `ax_actor_send()`, `ax_actor_stop()` C API

## Dependencies
- p14-t04: axalloc-actor-heap — per-actor allocator
- p07-t03: panic-handler — actor crash handling

## Subsystems Affected
- Scheduler (p15-t02): manages actor queue
- Message queue (p15-t05): embedded in AxActor
- Supervisor (p15-t07): tracks child actors

## Detailed Requirements

```c
typedef struct AxActor {
    uint64_t      id;              // unique actor ID (monotonic)
    ActorHeap     heap;            // isolated memory
    AxMsgQueue    mailbox;         // inbound message queue
    AxActorState  state;           // SPAWNING, RUNNING, STOPPING, DEAD
    AxHandlerFn   handler;         // compiled message handler fn ptr
    void*         state_data;      // actor user state (allocated in heap)
    uint64_t      supervisor_id;   // 0 = no supervisor
    uint32_t      restart_count;
    uint32_t      max_restarts;    // 0 = no limit
    AxRestartStrategy restart_strat; // ONE_FOR_ONE, ALL_FOR_ONE
} AxActor;

typedef enum AxActorState {
    AX_ACTOR_SPAWNING = 0,
    AX_ACTOR_RUNNING  = 1,
    AX_ACTOR_STOPPING = 2,
    AX_ACTOR_DEAD     = 3,
} AxActorState;

typedef void (*AxHandlerFn)(AxActor* self, void* msg, uint32_t msg_type);

// Spawn a new actor
uint64_t ax_actor_spawn(AxHandlerFn handler, void* init_state, size_t state_size,
                        uint64_t supervisor_id);

// Send message to actor (from any thread)
int ax_actor_send(uint64_t actor_id, void* msg, uint32_t msg_type, size_t msg_size);

// Request graceful stop
void ax_actor_stop(uint64_t actor_id);

// Internal: run one iteration of actor message loop
void ax_actor_step(AxActor* actor);
```

Actor ID: globally unique `uint64_t`, assigned by atomic counter.

Actor table: global `AxActor* g_actors[MAX_ACTORS]` with RW lock for registration.

## Implementation Steps

1. Create `runtime/actor.h` and `runtime/actor.c`.
2. Define `AxActor`, `AxActorState`, `AxHandlerFn`.
3. Implement `ax_actor_spawn()` — allocate actor, init heap, register in global table.
4. Implement `ax_actor_step()` — dequeue one message, call handler.
5. Implement `ax_actor_stop()` — set state to STOPPING, enqueue Stop message.
6. Implement `ax_actor_send()` — enqueue message in target actor's mailbox.
7. Write lifecycle tests.

## Test Plan
- `TestActorSpawn`: spawn actor → state = RUNNING, ID assigned
- `TestActorSend`: send message → handler called with correct args
- `TestActorStop`: stop signal → state transitions to DEAD
- `TestActorIsolation`: actor A cannot access actor B's heap

## Validation Checklist
- [ ] Actor IDs globally unique (atomic counter)
- [ ] Handler called for each dequeued message
- [ ] State machine: SPAWNING → RUNNING → STOPPING → DEAD
- [ ] Heap destroyed on DEAD

## Acceptance Criteria
- 10,000 actors spawned simultaneously without race conditions

## Definition of Done
- [ ] `runtime/actor.c` implemented
- [ ] Lifecycle tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Global actor table lock contention | Shard by actor_id % SHARD_COUNT |

## Future Follow-up Tasks
- p15-t02: scheduler assigns actors to OS threads
- p15-t07: supervisor tree links actors
