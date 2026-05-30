/*
 * p15-t03: Actor System Initialization — Implementation
 */

#include "runtime_init.h"
#include "actor.h"
#include "scheduler.h"
#include <string.h>

/* --------------------------------------------------------------------------
 * Global State
 * -------------------------------------------------------------------------- */

static int g_runtime_initialized = 0;
static AxScheduler g_scheduler;

/* --------------------------------------------------------------------------
 * Default Config
 * -------------------------------------------------------------------------- */

AxRuntimeConfig ax_runtime_default_config(void) {
    AxRuntimeConfig config;
    config.worker_threads = 4;      /* sensible default */
    config.max_actors = 65536;
    config.debug_mode = 0;
    config.numa_aware = 0;
    return config;
}

/* --------------------------------------------------------------------------
 * Init
 * -------------------------------------------------------------------------- */

int ax_runtime_init(const AxRuntimeConfig* config) {
    if (g_runtime_initialized) return 0; /* idempotent */

    AxRuntimeConfig cfg;
    if (config) {
        cfg = *config;
    } else {
        cfg = ax_runtime_default_config();
    }

    /* Step 1: Panic handler (already registered by Go runtime) */

    /* Step 2: Crash cleanup */
    /* ax_register_crash_cleanup(); — linked from axalloc */

    /* Step 3: Global allocator */
    extern void ax_segment_manager_init(void);
    ax_segment_manager_init();

    /* Step 4: Actor table */
    ax_actor_table_init();

    /* Step 5: Scheduler */
    uint32_t workers = cfg.worker_threads;
    if (workers == 0) workers = 4; /* auto-detect placeholder */
    if (workers > AX_MAX_WORKERS) workers = AX_MAX_WORKERS;

    if (ax_scheduler_init(&g_scheduler, workers) != 0) {
        return -1;
    }

    g_runtime_initialized = 1;
    return 0;
}

/* --------------------------------------------------------------------------
 * Shutdown
 * -------------------------------------------------------------------------- */

void ax_runtime_shutdown(void) {
    if (!g_runtime_initialized) return;

    /* Single-threaded cooperative scheduler: run all actors to completion before shutdown */
    ax_scheduler_run(&g_scheduler);

    /* Step 1: Scheduler shutdown */
    ax_scheduler_shutdown(&g_scheduler);

    /* Step 2: Actor table destroy */
    ax_actor_table_destroy();

    /* Step 3: Global allocator */
    /* ax_segment_manager_shutdown(); — linked from axalloc */

    g_runtime_initialized = 0;
}

/* --------------------------------------------------------------------------
 * Query
 * -------------------------------------------------------------------------- */

int ax_runtime_is_init(void) {
    return g_runtime_initialized;
}

struct AxScheduler* ax_get_global_scheduler(void) {
    if (!g_runtime_initialized) return NULL;
    return &g_scheduler;
}

/* --------------------------------------------------------------------------
 * Codegen Entry Points
 * -------------------------------------------------------------------------- */

void __ax_runtime_init(void) {
    ax_runtime_init(NULL);
}

void __ax_runtime_shutdown(void) {
    ax_runtime_shutdown();
}

void ax___ax_runtime_init(void) {
    ax_runtime_init(NULL);
}

void ax___ax_runtime_shutdown(void) {
    ax_runtime_shutdown();
}

