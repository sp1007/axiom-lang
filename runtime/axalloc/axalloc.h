/**
 * axalloc.h — AXIOM Runtime Memory Allocator (MVP)
 *
 * Every heap allocation in AXIOM carries a 16-byte header containing
 * a generation counter (gen_id) and the user-requested allocation size.
 * This enables the generational reference safety system used throughout
 * the AXIOM runtime for use-after-free detection.
 *
 * When a pointer is freed, the generation counter is incremented.
 * Any held reference (AxRef) still carrying the old generation value
 * will fail the validity check in ax_deref, catching use-after-free
 * at runtime without a garbage collector.
 */
#pragma once

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/**
 * AxHeader: 16 bytes prepended to every heap allocation.
 * Layout: [gen_id:8B][size:8B][user data...]
 * gen_id starts at 1 for live allocations.
 * gen_id is incremented on free, invalidating live AxRef values.
 */
typedef struct {
    uint64_t gen_id;  /* generation counter; 1=live, incremented on free */
    uint64_t size;    /* user allocation size (not including header) */
} AxHeader;

/* Compile-time assertion: AxHeader must be exactly 16 bytes and 8-byte aligned */
_Static_assert(sizeof(AxHeader) == 16, "AxHeader must be 16 bytes");
_Static_assert(sizeof(AxHeader) % 8 == 0, "AxHeader must be 8-byte aligned");

/**
 * AxRef: fat pointer with captured generation ID.
 * Used for safe reference validation via ax_deref.
 */
typedef struct {
    void*    ptr;     /* points to the byte immediately AFTER the AxHeader */
    uint64_t gen_id;  /* generation at the time this reference was created */
} AxRef;

/**
 * ax_alloc — Allocate `size` bytes of heap memory.
 * Returns pointer to user data (after the internal header).
 * Never returns NULL; calls ax_panic on OOM.
 */
void* ax_alloc(size_t size);

/**
 * ax_free — Free a pointer previously returned by ax_alloc.
 * Increments the header's gen_id before freeing, invalidating
 * all outstanding AxRef values.
 * ax_free(NULL) is a no-op.
 */
void ax_free(void* ptr);

/**
 * ax_realloc — Resize an existing allocation.
 * Preserves the existing gen_id. Returns new pointer.
 * ax_realloc(NULL, size) is equivalent to ax_alloc(size).
 * Never returns NULL; calls ax_panic on OOM.
 */
void* ax_realloc(void* ptr, size_t new_size);

/**
 * ax_alloc_size — Return the user-visible size of the allocation.
 * Does NOT include the header size.
 * ax_alloc_size(NULL) returns 0.
 */
size_t ax_alloc_size(void* ptr);

/**
 * ax_get_header — Get the AxHeader for a user pointer.
 * The header is located immediately before the user data.
 */
static inline AxHeader* ax_get_header(void* ptr) {
    return ((AxHeader*)ptr) - 1;
}

/**
 * ax_make_ref — Create an AxRef capturing the current generation.
 * Panics if ptr is NULL.
 */
static inline AxRef ax_make_ref(void* ptr) {
    AxRef ref;
    if (!ptr) {
        ref.ptr = NULL;
        ref.gen_id = 0;
        return ref;
    }
    ref.ptr = ptr;
    ref.gen_id = ax_get_header(ptr)->gen_id;
    return ref;
}

#ifdef __cplusplus
}
#endif

/* Include generational reference checking (ax_deref, ax_invalidate, etc.) */
#include "genref.h"
