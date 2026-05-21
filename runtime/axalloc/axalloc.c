/**
 * axalloc.c — AXIOM Runtime Memory Allocator Implementation (MVP)
 *
 * Wraps malloc/free with a 16-byte AxHeader prepended to every allocation.
 * The header carries a generation counter for use-after-free detection
 * and the user-requested allocation size.
 */
#include "axalloc.h"

#include <stdlib.h>
#include <string.h>
#include <stdio.h>

/* ax_panic is defined in the panic handler module (p07-t03).
 * For standalone compilation, a stub must be provided. */
extern void ax_panic(const char* msg);

void* ax_alloc(size_t size) {
    AxHeader* hdr = (AxHeader*)malloc(sizeof(AxHeader) + size);
    if (!hdr) {
        ax_panic("ax_alloc: out of memory");
        return NULL; /* unreachable if ax_panic aborts */
    }
    hdr->gen_id = 1;
    hdr->size   = (uint64_t)size;
    return (void*)(hdr + 1);
}

void ax_free(void* ptr) {
    if (!ptr) return;
    AxHeader* hdr = ax_get_header(ptr);
    hdr->gen_id++; /* invalidate all live AxRef values */
    free(hdr);
}

void* ax_realloc(void* ptr, size_t new_size) {
    if (!ptr) return ax_alloc(new_size);
    AxHeader* old_hdr = ax_get_header(ptr);
    uint64_t gen = old_hdr->gen_id;
    AxHeader* new_hdr = (AxHeader*)realloc(old_hdr, sizeof(AxHeader) + new_size);
    if (!new_hdr) {
        ax_panic("ax_realloc: out of memory");
        return NULL; /* unreachable if ax_panic aborts */
    }
    new_hdr->gen_id = gen;        /* preserve generation */
    new_hdr->size   = (uint64_t)new_size;
    return (void*)(new_hdr + 1);
}

size_t ax_alloc_size(void* ptr) {
    if (!ptr) return 0;
    return (size_t)ax_get_header(ptr)->size;
}
