# p15-t03: Actor System Initialization

## Purpose
Implement runtime initialization that wires together the allocator, scheduler, signal handlers, and actor infrastructure into a coherent system started automatically before `main()`.

## Context
AXIOM programs need the runtime initialized before any user code runs. This includes starting worker threads, registering signal handlers, initializing the global actor table, and setting up the root supervisor. The initialization must be transparent to the user — `axc` emits a hidden `__ax_runtime_init()` call at program entry.

## Inputs
- Scheduler from p15-t02
- AxAlloc from p14
- Panic handler from p07-t03
- Crash cleanup from p14-t07

## Outputs
- `runtime/runtime_init.c` — `ax_runtime_init()` and `ax_runtime_shutdown()`
- Emitted by codegen as hidden call before `main()`

## Dependencies
- p15-t02: scheduler — `ax_scheduler_init()`
- p14-t07: axalloc-crash-cleanup — `ax_register_crash_cleanup()`
- p07-t03: panic-handler — `ax_panic_handler_init()`

## Subsystems Affected
- All runtime subsystems: initialized in correct dependency order
- Codegen (p11-t15): emits `call __ax_runtime_init` in entry function prologue

## Detailed Requirements

```c
typedef struct AxRuntimeConfig {
    int   worker_threads;   // 0 = use CPU count
    int   max_actors;       // default 65536
    bool  debug_mode;       // enable extra checks
    bool  numa_aware;       // enable NUMA-aware allocation
} AxRuntimeConfig;

void ax_runtime_init(AxRuntimeConfig* config);
void ax_runtime_shutdown(void);

// Called by generated main() wrapper
void __ax_runtime_init(void);   // default config
void __ax_runtime_shutdown(void);
```

Initialization order:
1. `ax_panic_handler_init()` — earliest, needed for all subsequent failures.
2. `ax_register_crash_cleanup()` — register signal handlers.
3. `ax_global_alloc_init()` — initialize global fallback allocator.
4. `ax_actor_table_init(config.max_actors)` — allocate actor table.
5. `ax_scheduler_init(worker_count)` — create worker threads.
6. `ax_root_supervisor_spawn()` — create root supervisor actor.
7. `ax_scheduler_run()` — start scheduler (non-blocking; workers running).

Shutdown order (reverse):
1. `ax_scheduler_shutdown()` — wait for all actors to drain.
2. `ax_actor_table_destroy()`.
3. `ax_global_alloc_destroy()`.

Generated wrapper in `axc` codegen:
```c
int main(int argc, char** argv) {
    __ax_runtime_init();
    int result = ax_user_main(argc, argv);  // compiled user main()
    __ax_runtime_shutdown();
    return result;
}
```

## Implementation Steps

1. Create `runtime/runtime_init.c`.
2. Implement `ax_runtime_init()` with ordered subsystem startup.
3. Implement `ax_runtime_shutdown()` with reverse-order teardown.
4. Create `__ax_runtime_init()` thin wrapper with default config.
5. Wire into codegen: emit `call __ax_runtime_init` in generated `main`.
6. Test: AXIOM hello-world goes through full init/shutdown.

## Test Plan
- `TestRuntimeInit`: init with default config → no crash
- `TestRuntimeShutdown`: init → run 10 actors → shutdown → no leak
- `TestRuntimeInitOrder`: panic handler available immediately after init starts
- `TestRuntimeConfigThreads`: config.worker_threads=2 → only 2 workers created

## Validation Checklist
- [ ] Panic handler registered before any other init
- [ ] Scheduler not started before allocator ready
- [ ] Shutdown waits for all actors before freeing allocator
- [ ] Double-init protected (flag check)

## Acceptance Criteria
- Every AXIOM test program (from p16) initializes without assertion failures

## Definition of Done
- [ ] `runtime/runtime_init.c` implemented
- [ ] Generated `main` wrapper emitted by codegen

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Init order dependency cycle | Document explicit order; enforce with sequential calls |
| Shutdown race: actor sends after scheduler stopped | Drain all actor mailboxes before stopping scheduler |

## Future Follow-up Tasks
- Runtime configuration from environment variables (`AX_WORKERS=4`)
- Hot-reload runtime config without restart (phase post-18)
