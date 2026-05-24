/*
 * p15-t03: Actor System Initialization
 *
 * Ordered init/shutdown of all runtime subsystems.
 * Called transparently before/after user main().
 */

#ifndef AXIOM_RUNTIME_INIT_H
#define AXIOM_RUNTIME_INIT_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* --------------------------------------------------------------------------
 * Runtime Configuration
 * -------------------------------------------------------------------------- */

typedef struct {
    uint32_t worker_threads;   /* 0 = auto-detect CPU count */
    uint32_t max_actors;       /* max concurrent actors */
    int      debug_mode;       /* enable debug logging */
    int      numa_aware;       /* enable NUMA-local allocation */
} AxRuntimeConfig;

/** Default runtime configuration. */
AxRuntimeConfig ax_runtime_default_config(void);

/* --------------------------------------------------------------------------
 * Runtime Init / Shutdown
 * -------------------------------------------------------------------------- */

/**
 * Initialize the AXIOM runtime with the given configuration.
 *
 * Init order:
 * 1. Panic handler
 * 2. Crash cleanup handlers
 * 3. Global allocator (segment manager)
 * 4. Actor table
 * 5. Scheduler (spawn worker threads)
 *
 * Returns 0 on success, -1 on failure.
 */
int ax_runtime_init(const AxRuntimeConfig* config);

/**
 * Shutdown the AXIOM runtime.
 *
 * Shutdown order (reverse of init):
 * 1. Scheduler shutdown (drain + join workers)
 * 2. Actor table destroy (drain mailboxes)
 * 3. Global allocator destroy
 */
void ax_runtime_shutdown(void);

/**
 * Check if the runtime is initialized.
 */
int ax_runtime_is_init(void);

struct AxScheduler;
struct AxScheduler* ax_get_global_scheduler(void);

/**
 * Default-config init/shutdown wrappers.
 * These are emitted by codegen into the generated main() wrapper.
 */
void __ax_runtime_init(void);
void __ax_runtime_shutdown(void);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_RUNTIME_INIT_H */
