/*
 * p14-t05: AxAlloc NUMA-Aware Allocation (Stub)
 *
 * Infrastructure for NUMA-aware memory allocation.
 * Allocates segments from the NUMA node closest to the running thread.
 *
 * p14-t06: AxAlloc GPU-Pinned Memory (Stub)
 *
 * Infrastructure for GPU-pinned memory allocation for zero-copy
 * data transfer between CPU and GPU actors.
 */

#ifndef AXIOM_AXALLOC_PLATFORM_EXT_H
#define AXIOM_AXALLOC_PLATFORM_EXT_H

#include "size_classes.h"
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* --------------------------------------------------------------------------
 * p14-t05: NUMA-Aware Allocation
 * -------------------------------------------------------------------------- */

typedef struct {
    int      node_id;        /* NUMA node index */
    uint64_t total_memory;   /* total memory on node (bytes) */
    uint64_t free_memory;    /* free memory on node (bytes) */
} NumaNodeInfo;

/** Get the NUMA node for the current thread. */
int ax_numa_current_node(void);

/** Get the number of NUMA nodes in the system. */
int ax_numa_node_count(void);

/** Allocate memory on a specific NUMA node. */
void* ax_numa_alloc(size_t size, int node_id);

/** Free NUMA-allocated memory. */
void ax_numa_free(void* ptr, size_t size);

/** Get NUMA node information. */
int ax_numa_node_info(int node_id, NumaNodeInfo* info);

/* --------------------------------------------------------------------------
 * p14-t06: GPU-Pinned Memory
 * -------------------------------------------------------------------------- */

/** GPU memory allocation flags. */
typedef enum {
    AX_GPU_FLAG_NONE      = 0,
    AX_GPU_FLAG_WRITE_COMBINED = 1,  /* write-combined for CPU→GPU transfers */
    AX_GPU_FLAG_HOST_MAPPED    = 2,  /* accessible from both CPU and GPU */
    AX_GPU_FLAG_PORTABLE       = 4,  /* usable across multiple GPU contexts */
} AxGpuFlags;

/** Allocate GPU-pinned (page-locked) host memory. */
void* ax_gpu_pinned_alloc(size_t size, AxGpuFlags flags);

/** Free GPU-pinned memory. */
void ax_gpu_pinned_free(void* ptr);

/** Query if a pointer is GPU-pinned. */
int ax_is_gpu_pinned(const void* ptr);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_AXALLOC_PLATFORM_EXT_H */
