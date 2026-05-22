/*
 * p14-t02: AxAlloc Segment Manager
 *
 * Manages 64KB memory segments used as bump regions for size-class allocation.
 * Segments are acquired from the OS and recycled via a pool.
 */

#ifndef AXIOM_AXALLOC_SEGMENT_MANAGER_H
#define AXIOM_AXALLOC_SEGMENT_MANAGER_H

#include "size_classes.h"

#ifdef __cplusplus
extern "C" {
#endif

/* --------------------------------------------------------------------------
 * Segment Definitions
 * -------------------------------------------------------------------------- */

#define SEGMENT_SIZE (64 * 1024)   /* 64KB per segment */
#define MAX_SEGMENTS 4096          /* max total segments in pool */
#define SEGMENT_MAGIC 0xAF5E6000   /* magic value for validation */

typedef struct Segment {
    char*           base;       /* start of usable memory */
    char*           bump;       /* current bump pointer */
    char*           limit;      /* end of segment */
    SizeClass       sclass;     /* size class this segment serves */
    struct Segment* next;       /* next in linked list */
    uint32_t        magic;      /* validation magic */
} Segment;

typedef struct {
    Segment* active;    /* currently active segment for this size class */
    Segment* retired;   /* list of retired (full) segments */
    size_t   count;     /* number of segments in this list */
} SegmentList;

/* --------------------------------------------------------------------------
 * Segment Manager API
 * -------------------------------------------------------------------------- */

/**
 * Acquire a new segment for the given size class.
 * Returns NULL if OS memory allocation fails or pool is exhausted.
 */
Segment* ax_segment_acquire(SizeClass sc);

/**
 * Release a segment back to the pool.
 */
void ax_segment_release(Segment* seg);

/**
 * Get the active segment for a size class, acquiring a new one if needed.
 */
Segment* ax_segment_get_active(SegmentList* list, SizeClass sc);

/**
 * Release all segments in a list.
 */
void ax_segment_list_release_all(SegmentList* list);

/**
 * Allocate from a segment's bump region.
 * Returns NULL if segment is exhausted.
 */
void* ax_segment_bump_alloc(Segment* seg, SizeClass sc);

/**
 * Returns the utilization of a segment (0.0 to 1.0).
 */
double ax_segment_utilization(const Segment* seg);

/**
 * Initialize the segment manager. Must be called once at startup.
 */
void ax_segment_manager_init(void);

/**
 * Shutdown the segment manager. Releases all pooled segments.
 */
void ax_segment_manager_shutdown(void);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_AXALLOC_SEGMENT_MANAGER_H */
