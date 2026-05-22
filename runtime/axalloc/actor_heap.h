/*
 * p14-t04: AxAlloc Actor Heap
 *
 * Per-actor isolated heap. Each actor has its own set of segments,
 * free lists, and bump allocators. No cross-actor sharing = no locks
 * needed for intra-actor allocation.
 */

#ifndef AXIOM_AXALLOC_ACTOR_HEAP_H
#define AXIOM_AXALLOC_ACTOR_HEAP_H

#include "size_classes.h"
#include "segment_manager.h"

#ifdef __cplusplus
extern "C" {
#endif

#define ACTOR_HEAP_MAGIC 0xAC704EA0

typedef struct {
    uint64_t     actor_id;                      /* owning actor ID */
    uint32_t     magic;                         /* validation magic */
    SegmentList  segments[NUM_SIZE_CLASSES];     /* per-size-class segments */
    FreeList     free_lists[NUM_SIZE_CLASSES];   /* per-size-class free lists */
    uint64_t     total_allocated;                /* bytes allocated */
    uint64_t     total_freed;                    /* bytes freed */
    uint64_t     alloc_count;                    /* number of alloc calls */
    uint64_t     free_count;                     /* number of free calls */
} ActorHeap;

/** Create a new actor heap for the given actor ID. */
ActorHeap* ax_actor_heap_create(uint64_t actor_id);

/** Destroy an actor heap, releasing all memory. */
void ax_actor_heap_destroy(ActorHeap* heap);

/** Allocate from an actor's heap. Returns user pointer or NULL. */
void* ax_actor_alloc(ActorHeap* heap, size_t user_size);

/** Free memory back to the actor's heap. */
void ax_actor_free(ActorHeap* heap, void* user_ptr);

/** Get heap statistics. */
typedef struct {
    uint64_t total_allocated;
    uint64_t total_freed;
    uint64_t alloc_count;
    uint64_t free_count;
    uint64_t live_bytes;      /* total_allocated - total_freed */
    int      segment_count;   /* total active + retired segments */
} ActorHeapStats;

void ax_actor_heap_stats(const ActorHeap* heap, ActorHeapStats* stats);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_AXALLOC_ACTOR_HEAP_H */
