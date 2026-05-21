/*
 * p14-t05/t06: NUMA + GPU Platform Extensions — Stub Implementation
 *
 * Fallback implementations for systems without NUMA/GPU support.
 * These stubs allow compilation on all platforms.
 */

#include "platform_ext.h"
#include <string.h>

/* --------------------------------------------------------------------------
 * p14-t05: NUMA Stubs (Non-NUMA fallback)
 * -------------------------------------------------------------------------- */

int ax_numa_current_node(void) {
    return 0;  /* single-node fallback */
}

int ax_numa_node_count(void) {
    return 1;  /* single-node fallback */
}

void* ax_numa_alloc(size_t size, int node_id) {
    (void)node_id;
    /* Fall back to regular OS allocation */
#ifdef _WIN32
    #include <windows.h>
    return VirtualAlloc(NULL, size, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
#else
    #include <sys/mman.h>
    void* p = mmap(NULL, size, PROT_READ | PROT_WRITE,
                    MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
    return (p == (void*)-1) ? NULL : p;
#endif
}

void ax_numa_free(void* ptr, size_t size) {
    if (!ptr) return;
#ifdef _WIN32
    (void)size;
    VirtualFree(ptr, 0, MEM_RELEASE);
#else
    munmap(ptr, size);
#endif
}

int ax_numa_node_info(int node_id, NumaNodeInfo* info) {
    if (!info) return -1;
    info->node_id = node_id;
    info->total_memory = 0;  /* unknown in stub */
    info->free_memory = 0;
    return 0;
}

/* --------------------------------------------------------------------------
 * p14-t06: GPU-Pinned Memory Stubs
 * -------------------------------------------------------------------------- */

void* ax_gpu_pinned_alloc(size_t size, AxGpuFlags flags) {
    (void)flags;
    /* Fall back to regular allocation — not page-locked */
#ifdef _WIN32
    return VirtualAlloc(NULL, size, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
#else
    void* p = mmap(NULL, size, PROT_READ | PROT_WRITE,
                    MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
    return (p == (void*)-1) ? NULL : p;
#endif
}

void ax_gpu_pinned_free(void* ptr) {
    if (!ptr) return;
    /* In a real implementation, this would unpin and free */
    /* Stub: we can't munmap because we don't know the size */
    (void)ptr;
}

int ax_is_gpu_pinned(const void* ptr) {
    (void)ptr;
    return 0;  /* nothing is GPU-pinned in stub mode */
}
