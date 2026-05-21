/*
 * p14-t04: AxAlloc Actor Heap — Implementation
 */

#include "actor_heap.h"
#include <string.h>
#include <stdlib.h>

/* --------------------------------------------------------------------------
 * Lifecycle
 * -------------------------------------------------------------------------- */

ActorHeap* ax_actor_heap_create(uint64_t actor_id) {
    ActorHeap* heap = (ActorHeap*)calloc(1, sizeof(ActorHeap));
    if (!heap) return NULL;

    heap->actor_id = actor_id;
    heap->magic = ACTOR_HEAP_MAGIC;
    return heap;
}

void ax_actor_heap_destroy(ActorHeap* heap) {
    if (!heap || heap->magic != ACTOR_HEAP_MAGIC) return;

    /* Release all segments for each size class */
    for (int sc = 0; sc < NUM_SIZE_CLASSES; sc++) {
        ax_segment_list_release_all(&heap->segments[sc]);
    }

    heap->magic = 0;
    free(heap);
}

/* --------------------------------------------------------------------------
 * Allocation
 * -------------------------------------------------------------------------- */

void* ax_actor_alloc(ActorHeap* heap, size_t user_size) {
    if (!heap || heap->magic != ACTOR_HEAP_MAGIC) return NULL;

    SizeClass sc = ax_size_class_for(user_size);

    if (sc == SIZE_CLASS_LARGE) {
        void* ptr = ax_large_alloc(user_size);
        if (ptr) {
            heap->total_allocated += user_size;
            heap->alloc_count++;
        }
        return ptr;
    }

    /* Try free list first */
    void* block = ax_free_list_pop(&heap->free_lists[sc]);

    if (!block) {
        /* Get active segment (or acquire new one) */
        Segment* seg = ax_segment_get_active(&heap->segments[sc], sc);
        if (!seg) return NULL;

        block = ax_segment_bump_alloc(seg, sc);
        if (!block) {
            /* Current segment exhausted, try a fresh one */
            seg = ax_segment_get_active(&heap->segments[sc], sc);
            if (!seg) return NULL;
            block = ax_segment_bump_alloc(seg, sc);
            if (!block) return NULL;
        }
    }

    /* Initialize header */
    AxHeader* hdr = (AxHeader*)block;
    hdr->gen_id = 1;
    hdr->flags = (uint32_t)sc;

    size_t block_size = SIZE_CLASS_SIZES[sc];
    heap->total_allocated += block_size;
    heap->alloc_count++;

    return ax_block_to_user(block);
}

/* --------------------------------------------------------------------------
 * Deallocation
 * -------------------------------------------------------------------------- */

void ax_actor_free(ActorHeap* heap, void* user_ptr) {
    if (!heap || !user_ptr || heap->magic != ACTOR_HEAP_MAGIC) return;

    void* block = ax_user_to_block(user_ptr);
    AxHeader* hdr = (AxHeader*)block;

    SizeClass sc = (SizeClass)(hdr->flags & 0xF);
    hdr->gen_id = 0; /* invalidate */

    if (sc == SIZE_CLASS_LARGE || sc >= NUM_SIZE_CLASSES) {
        /* Large allocs cannot be recycled into free list */
        return;
    }

    size_t block_size = SIZE_CLASS_SIZES[sc];
    heap->total_freed += block_size;
    heap->free_count++;

    ax_free_list_push(&heap->free_lists[sc], block);
}

/* --------------------------------------------------------------------------
 * Statistics
 * -------------------------------------------------------------------------- */

void ax_actor_heap_stats(const ActorHeap* heap, ActorHeapStats* stats) {
    if (!heap || !stats) return;

    stats->total_allocated = heap->total_allocated;
    stats->total_freed = heap->total_freed;
    stats->alloc_count = heap->alloc_count;
    stats->free_count = heap->free_count;
    stats->live_bytes = heap->total_allocated - heap->total_freed;

    int seg_count = 0;
    for (int sc = 0; sc < NUM_SIZE_CLASSES; sc++) {
        if (heap->segments[sc].active) seg_count++;
        /* Count retired segments */
        const Segment* s = heap->segments[sc].retired;
        while (s) {
            seg_count++;
            s = s->next;
        }
    }
    stats->segment_count = seg_count;
}
