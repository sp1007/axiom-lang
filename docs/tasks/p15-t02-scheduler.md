# p15-t02: M:N Work-Stealing Scheduler

## Purpose
Implement the AXIOM runtime scheduler that maps N actors onto M OS threads using work-stealing, enabling efficient CPU utilization without OS-level context switching per actor.

## Context
AXIOM's actor model uses an M:N scheduler: M OS threads (workers) run N actors. Each worker has a local run queue; when idle it steals actors from other workers' queues. This provides scalable throughput without the overhead of one-thread-per-actor. MVP uses 1:1 (one worker per CPU), with full work-stealing in Phase 15.

## Inputs
- `AxActor` structs from p15-t01
- CPU count from `runtime.NumCPU()`
- Actor run queues (work-stealing deques)

## Outputs
- `runtime/scheduler.c` — work-stealing scheduler
- `ax_scheduler_init()`, `ax_scheduler_run()`, `ax_scheduler_shutdown()` API

## Dependencies
- p15-t01: actor-struct — AxActor, ax_actor_step()
- p14-t04: axalloc-actor-heap — actor heap initialization

## Subsystems Affected
- Actor runtime: all actors run via scheduler
- OS threads: scheduler creates and manages worker threads

## Detailed Requirements

```c
#define AX_MAX_WORKERS 256
#define AX_RUNQ_SIZE   4096  // power of 2

typedef struct AxRunQueue {
    AxActor* ring[AX_RUNQ_SIZE];
    _Atomic uint32_t head;
    _Atomic uint32_t tail;
    // Chase-Lev deque: local push/pop at tail, steal from head
} AxRunQueue;

typedef struct AxWorker {
    pthread_t     thread;
    int           id;
    AxRunQueue    runq;
    uint64_t      steal_attempts;
    uint64_t      actors_run;
} AxWorker;

typedef struct AxScheduler {
    AxWorker  workers[AX_MAX_WORKERS];
    int       worker_count;
    _Atomic int running;
} AxScheduler;

void ax_scheduler_init(int worker_count);
void ax_scheduler_submit(AxActor* actor);  // enqueue to least-loaded worker
void ax_scheduler_run(void);               // start all worker threads
void ax_scheduler_shutdown(void);          // stop all workers gracefully
```

Work-stealing algorithm (Chase-Lev deque):
- Local push/pop: atomic ops on tail (no contention from owner thread)
- Steal: CAS on head (contention only between thieves)
- Worker loop: run local queue until empty → steal from random worker → sleep briefly → repeat

Actor scheduling:
- Ready actors submitted to scheduler's least-loaded queue.
- After `ax_actor_step()`, if mailbox non-empty, resubmit actor.
- Sleeping actors (empty mailbox) removed from run queue.

## Implementation Steps

1. Create `runtime/scheduler.c`.
2. Implement `AxRunQueue` as Chase-Lev work-stealing deque.
3. Implement worker thread function: run local queue → steal → yield.
4. Implement `ax_scheduler_init()` — create worker threads.
5. Implement `ax_scheduler_submit()` — load-balanced enqueue.
6. Implement work-stealing: random victim selection, CAS-based steal.
7. Implement graceful shutdown: drain queues, join threads.
8. Write benchmarks: actor throughput vs thread count.

## Test Plan
- `TestSchedulerBasic`: submit 100 actors → all run to completion
- `TestSchedulerWorkStealing`: imbalanced load → work stolen between workers
- `TestSchedulerShutdown`: all actors complete before shutdown returns
- `TestSchedulerThroughput`: 1M simple actors in < 1 second on 8 cores

## Validation Checklist
- [ ] Chase-Lev deque: no lost actors under concurrent push/pop/steal
- [ ] All submitted actors eventually run (liveness)
- [ ] Shutdown waits for all actors to complete
- [ ] No busy-wait loops without yield/sleep

## Acceptance Criteria
- 1M ping-pong messages between 2 actors completes in < 100ms on 4 cores

## Definition of Done
- [ ] `runtime/scheduler.c` implemented
- [ ] Work-stealing demonstrated under load imbalance

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Chase-Lev deque ABA problem | Use epoch-based hazard pointers or tagged pointers |
| Thundering herd on steal | Exponential backoff + random victim selection |

## Future Follow-up Tasks
- NUMA-aware scheduling: prefer workers on same NUMA node
- p15-t03: actor system init wires scheduler to runtime startup
