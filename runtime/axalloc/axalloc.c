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

/* --------------------------------------------------------------------------
 * AXIOM-native Allocator Segment Manager Global State Shim
 * -------------------------------------------------------------------------- */
#include <stdint.h>

#define AXIOM_MAX_SEGMENTS 4096

// Must match the memory layout of AXIOM's Segment structure
typedef struct {
    char*           base;
    char*           bump;
    char*           limit;
    int32_t         sclass;
    void*           next;
    uint32_t        magic;
} AxiomSegment;

static AxiomSegment axiom_segment_slab[AXIOM_MAX_SEGMENTS];
static int64_t      axiom_segment_slab_used = 0;
static void*        axiom_free_segment_pool = NULL;

int64_t* std_mem_alloc_get_slab_used(void) {
    return &axiom_segment_slab_used;
}

void* std_mem_alloc_get_slab(void) {
    return axiom_segment_slab;
}

void** std_mem_alloc_get_free_pool(void) {
    return &axiom_free_segment_pool;
}

void ax_segment_manager_init(void) {
    // Stub for MVP allocator
}

