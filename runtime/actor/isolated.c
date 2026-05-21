/*
 * p15-t04: Isolated[T] Runtime — Implementation
 */

#include "isolated.h"
#include <stdlib.h>
#include <string.h>

AxIsolated ax_isolated_wrap(void* data, size_t size, uint64_t source_actor,
                            uint32_t type_tag) {
    AxIsolated iso;
    iso.data = data;
    iso.size = size;
    iso.source_actor = source_actor;
    iso.type_tag = type_tag;
    iso.consumed = 0;
    return iso;
}

void* ax_isolated_unwrap(AxIsolated* iso, uint64_t target_actor) {
    if (!iso || iso->consumed) return NULL;
    (void)target_actor;

    /* Mark as consumed — ownership transferred */
    iso->consumed = 1;

    /*
     * In a full implementation, this would:
     * 1. Detach the segment from source actor's heap
     * 2. Attach the segment to target actor's heap
     * 3. Update all internal pointers if needed
     *
     * For now, we simply transfer the raw pointer since
     * both actors share the same address space.
     */
    return iso->data;
}

int ax_isolated_is_consumed(const AxIsolated* iso) {
    if (!iso) return 1;
    return iso->consumed;
}

void* ax_isolated_deep_copy(const void* data, size_t size) {
    if (!data || size == 0) return NULL;
    void* copy = malloc(size);
    if (copy) {
        memcpy(copy, data, size);
    }
    return copy;
}
