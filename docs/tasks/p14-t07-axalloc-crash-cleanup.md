# p14-t07: Allocator Crash Cleanup

## Purpose
Implement signal handlers and OS-level cleanup hooks that release all allocator resources (segments, pinned memory, file descriptors) on process crash, preventing resource leaks after panics or segfaults.

## Context
When an AXIOM program panics or receives SIGSEGV/SIGABRT, the default OS behavior frees anonymous mmap segments automatically. However, GPU-pinned memory, named shared memory segments, and file-backed allocations may leak. This task ensures all non-anonymous resources are released on crash.

## Inputs
- `AxAlloc`, `ActorHeap`, `AxGPUAlloc` instances from p14-t01 through p14-t06
- Signal handlers: SIGSEGV, SIGABRT, SIGTERM, SIGBUS
- Platform: Linux/macOS POSIX signals; Windows SEH

## Outputs
- `runtime/axalloc_crash.c` — crash cleanup signal handlers
- `ax_register_crash_cleanup()` — registers all signal handlers
- `ax_crash_cleanup()` — manually callable for graceful shutdown

## Dependencies
- p14-t01: axalloc-mvp — base allocator list
- p14-t06: axalloc-gpu-pinned — GPU memory needs explicit release
- p07-t03: panic-handler — panic calls crash cleanup before abort

## Subsystems Affected
- Runtime initialization: `ax_runtime_init()` calls `ax_register_crash_cleanup()`
- Panic handler: calls `ax_crash_cleanup()` before terminating

## Detailed Requirements

```c
// Global registry of all allocators (weak references)
#define AX_MAX_ALLOCATORS 64
static AxAllocBase* g_allocators[AX_MAX_ALLOCATORS];
static int g_allocator_count = 0;

void ax_register_allocator(AxAllocBase* alloc);
void ax_unregister_allocator(AxAllocBase* alloc);

// Signal handler installation
void ax_register_crash_cleanup(void);

// Crash cleanup (signal-safe: only async-signal-safe ops)
void ax_crash_cleanup(void);

// Individual cleanup hooks
void ax_cleanup_gpu_allocs(void);   // release pinned memory
void ax_cleanup_shared_mem(void);   // release named shared memory
void ax_cleanup_file_backed(void);  // close file-backed mmaps
```

Signal handler constraints (async-signal-safe only):
- Allowed: `write()`, `_exit()`, atomic loads
- Forbidden: `malloc()`, `printf()`, `pthread_mutex_lock()`
- Log crash to pre-allocated buffer using `write(STDERR_FILENO, ...)`

Windows SEH:
```c
LONG WINAPI ax_veh_handler(EXCEPTION_POINTERS* ep) {
    ax_crash_cleanup();
    return EXCEPTION_CONTINUE_SEARCH;
}
AddVectoredExceptionHandler(0, ax_veh_handler);
```

## Implementation Steps

1. Create `runtime/axalloc_crash.c`.
2. Implement global allocator registry (array, atomic counter).
3. Implement signal handler: save old handler, call ax_crash_cleanup(), restore.
4. Implement `ax_crash_cleanup()` using only async-signal-safe operations.
5. Wire GPU cleanup into crash handler.
6. Wire into `ax_runtime_init()`.
7. Test: force SIGSEGV → verify GPU memory released.

## Test Plan
- `TestCrashCleanupGPU`: force crash → verify cudaFreeHost called
- `TestCrashCleanupSharedMem`: shared mem → released after crash
- `TestCrashHandlerInstalled`: SIGSEGV triggers cleanup handler
- `TestCrashHandlerSignalSafe`: no malloc/printf in crash path

## Validation Checklist
- [ ] Signal handler registered for SIGSEGV, SIGABRT, SIGBUS, SIGTERM
- [ ] Crash cleanup uses only async-signal-safe syscalls
- [ ] GPU memory released before process exit
- [ ] No recursive signal handler invocation

## Acceptance Criteria
- Valgrind shows no GPU resource leaks after forced SIGSEGV

## Definition of Done
- [ ] `runtime/axalloc_crash.c` implemented
- [ ] Crash cleanup tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Crash handler calls non-async-signal-safe functions → undefined behavior | Audit all calls in handler; use write() not printf() |
| Double-free if crash during cleanup | Set allocator to NULL in registry before freeing |

## Future Follow-up Tasks
- Core dump annotation: write allocator state to crash report
- Integration with AXIOM supervisor/watchdog for actor crash recovery
