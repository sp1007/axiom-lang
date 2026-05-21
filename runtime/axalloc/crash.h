/*
 * p14-t07: AxAlloc Crash Cleanup
 *
 * Signal/SEH handlers for graceful cleanup on crash.
 * Releases non-anonymous resources (GPU pinned, shared mem).
 * All crash handler code must be async-signal-safe.
 */

#ifndef AXIOM_AXALLOC_CRASH_H
#define AXIOM_AXALLOC_CRASH_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

#define AX_MAX_REGISTERED_ALLOCATORS 64

typedef enum {
    AX_ALLOC_TYPE_HEAP = 0,
    AX_ALLOC_TYPE_GPU_PINNED,
    AX_ALLOC_TYPE_SHARED_MEM,
    AX_ALLOC_TYPE_FILE_BACKED,
} AxAllocType;

typedef struct {
    void*       base;
    size_t      size;
    AxAllocType type;
    int         active;      /* 1=registered, 0=freed */
} AxAllocRegistry;

/**
 * Register a crash cleanup handler.
 * Must be called once during runtime init.
 */
void ax_register_crash_cleanup(void);

/**
 * Register an allocator for crash cleanup.
 * Returns a handle for unregistration.
 */
int ax_register_allocator(void* base, size_t size, AxAllocType type);

/**
 * Unregister an allocator.
 */
void ax_unregister_allocator(int handle);

/**
 * Emergency cleanup — async-signal-safe.
 * Called from signal handler / SEH.
 */
void ax_crash_cleanup(void);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_AXALLOC_CRASH_H */
