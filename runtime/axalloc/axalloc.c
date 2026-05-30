/**
 * axalloc.c — AXIOM Runtime Canary-Based Debug Allocator
 */
#include "axalloc.h"

#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <stdint.h>

#ifdef _WIN32
#include <malloc.h>
#else
#include <malloc.h>
#endif

extern void ax_panic(const char* msg);

#define MAGIC_HEADER 0xDEADBEEFCAFEBABEULL
#define MAGIC_FOOTER 0xBABEC0FEDEADF00DULL

struct debug_block {
    struct debug_block* next;
    struct debug_block* prev;
    size_t size;
    uint64_t magic_header;
};

struct debug_footer {
    uint64_t magic_footer;
    uint64_t padding;
};

static struct debug_block* debug_head = NULL;

void verify_all_allocations(const char* context) {
#ifdef AX_ALLOC_DEBUG
    struct debug_block* curr = debug_head;
    while (curr != NULL) {
        if (curr->magic_header != MAGIC_HEADER) {
            fprintf(stderr, "[CRITICAL] [%s] Heap corruption (Header Magic Mismatch) detected at block %p! Size: %zu\n",
                    context, (void*)(curr + 1), curr->size);
            fflush(stderr);
            ax_panic("Heap corruption: Header Magic Mismatch");
        }
        struct debug_footer* footer = (struct debug_footer*)((char*)(curr + 1) + curr->size);
        if (footer->magic_footer != MAGIC_FOOTER) {
            fprintf(stderr, "[CRITICAL] [%s] Heap corruption (Footer Canary/Overflow) detected at block %p! Size: %zu, footer address: %p\n",
                    context, (void*)(curr + 1), curr->size, (void*)footer);
            fflush(stderr);
            ax_panic("Heap corruption: Footer Canary/Overflow");
        }
        curr = curr->next;
    }
#endif
}

void* ax_alloc(size_t size) {
    // Verify all existing allocations first to catch any lazy write corruption immediately!
    verify_all_allocations("ax_alloc entry");

    size_t total_size = sizeof(struct debug_block) + size + sizeof(struct debug_footer);
    struct debug_block* block = (struct debug_block*)malloc(total_size);
    if (!block) {
        ax_panic("ax_alloc: out of memory");
        return NULL;
    }

    block->size = size;
    block->magic_header = MAGIC_HEADER;

    struct debug_footer* footer = (struct debug_footer*)((char*)(block + 1) + size);
    footer->magic_footer = MAGIC_FOOTER;
    footer->padding = 0;

    // Link block
    block->next = debug_head;
    block->prev = NULL;
    if (debug_head) {
        debug_head->prev = block;
    }
    debug_head = block;

    return (void*)(block + 1);
}

void ax_free(void* ptr) {
    if (!ptr) return;

    verify_all_allocations("ax_free entry");

    struct debug_block* block = (struct debug_block*)ptr - 1;
    if (block->magic_header != MAGIC_HEADER) {
        fprintf(stderr, "[CRITICAL] ax_free: Invalid block header magic at %p!\n", ptr);
        fflush(stderr);
        ax_panic("ax_free: Invalid block header magic");
    }

    struct debug_footer* footer = (struct debug_footer*)((char*)ptr + block->size);
    if (footer->magic_footer != MAGIC_FOOTER) {
        fprintf(stderr, "[CRITICAL] ax_free: Heap corruption (Footer Canary) detected at block %p! Size: %zu\n",
                ptr, block->size);
        fflush(stderr);
        ax_panic("ax_free: Footer Canary Mismatch");
    }

    // Unlink block
    if (block->next) {
        block->next->prev = block->prev;
    }
    if (block->prev) {
        block->prev->next = block->next;
    } else {
        debug_head = block->next;
    }

    free(block);
}

void* ax_realloc(void* ptr, size_t new_size) {
    if (!ptr) {
        return ax_alloc(new_size);
    }

    verify_all_allocations("ax_realloc entry");

    struct debug_block* block = (struct debug_block*)ptr - 1;
    if (block->magic_header != MAGIC_HEADER) {
        fprintf(stderr, "[CRITICAL] ax_realloc: Invalid block header magic at %p!\n", ptr);
        fflush(stderr);
        ax_panic("ax_realloc: Invalid block header magic");
    }

    // Allocate new block
    void* new_ptr = ax_alloc(new_size);
    if (!new_ptr) {
        return NULL;
    }

    // Copy data
    size_t copy_size = block->size < new_size ? block->size : new_size;
    memcpy(new_ptr, ptr, copy_size);

    // Free old block
    ax_free(ptr);

    return new_ptr;
}

size_t ax_alloc_size(void* ptr) {
    if (!ptr) return 0;
    struct debug_block* block = (struct debug_block*)ptr - 1;
    if (block->magic_header != MAGIC_HEADER) {
        return 0;
    }
    return block->size;
}

/* --------------------------------------------------------------------------
 * AXIOM-native Allocator Segment Manager Global State Shim
 * -------------------------------------------------------------------------- */
#define AXIOM_MAX_SEGMENTS 4096

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
