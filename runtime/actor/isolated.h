/*
 * p15-t04: Isolated[T] Runtime Support
 *
 * Isolated types enable zero-copy transfer of heap segments between actors.
 * When a value is wrapped in Isolated[T], its owning segment is detached
 * from the sender's heap and attached to the receiver's heap.
 */

#ifndef AXIOM_RUNTIME_ISOLATED_H
#define AXIOM_RUNTIME_ISOLATED_H

#include "actor.h"
#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/* --------------------------------------------------------------------------
 * Isolated[T] Wrapper
 * -------------------------------------------------------------------------- */

typedef struct {
    void*       data;           /* pointer to the isolated data */
    size_t      size;           /* size of the data in bytes */
    uint64_t    source_actor;   /* actor that created this isolation */
    uint32_t    type_tag;       /* runtime type tag for safety */
    int         consumed;       /* 1 = already consumed (moved) */
} AxIsolated;

/**
 * Wrap a value in an Isolated container.
 * The value is detached from the source actor's heap.
 */
AxIsolated ax_isolated_wrap(void* data, size_t size, uint64_t source_actor,
                            uint32_t type_tag);

/**
 * Unwrap an Isolated value into the target actor's heap.
 * The isolated container is consumed (cannot be used again).
 * Returns the pointer to the data in the new heap, or NULL on error.
 */
void* ax_isolated_unwrap(AxIsolated* iso, uint64_t target_actor);

/**
 * Check if an isolated value has been consumed.
 */
int ax_isolated_is_consumed(const AxIsolated* iso);

/**
 * Deep-copy fallback for non-segment-transferable types.
 */
void* ax_isolated_deep_copy(const void* data, size_t size);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_RUNTIME_ISOLATED_H */
