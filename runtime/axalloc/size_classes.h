/*
 * p14-t01: AxAlloc Size-Classed Allocation
 *
 * Size class definitions, AxHeader layout, free list operations,
 * and allocation/deallocation primitives for AXIOM's custom allocator.
 */

#ifndef AXIOM_AXALLOC_SIZE_CLASSES_H
#define AXIOM_AXALLOC_SIZE_CLASSES_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/* --------------------------------------------------------------------------
 * Size Class Definitions
 * -------------------------------------------------------------------------- */

#define NUM_SIZE_CLASSES 10

typedef enum {
    SIZE_CLASS_8    = 0,
    SIZE_CLASS_16   = 1,
    SIZE_CLASS_32   = 2,
    SIZE_CLASS_64   = 3,
    SIZE_CLASS_128  = 4,
    SIZE_CLASS_256  = 5,
    SIZE_CLASS_512  = 6,
    SIZE_CLASS_1024 = 7,
    SIZE_CLASS_2048 = 8,
    SIZE_CLASS_4096 = 9,
    SIZE_CLASS_LARGE = 10   /* > 4096: direct OS alloc */
} SizeClass;

static const size_t SIZE_CLASS_SIZES[NUM_SIZE_CLASSES] = {
    8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096
};

/* --------------------------------------------------------------------------
 * AxHeader — 8-byte header for every managed allocation
 * -------------------------------------------------------------------------- */

typedef struct {
    uint32_t gen_id;    /* generational ID (1=live, 0=freed) */
    uint32_t flags;     /* lower 4 bits: SizeClass; upper bits: GC/type flags */
} AxHeader;

#define AX_HEADER_SIZE sizeof(AxHeader)  /* 8 bytes */

/* --------------------------------------------------------------------------
 * Pointer Conversion
 * -------------------------------------------------------------------------- */

static inline void* ax_block_to_user(void* block) {
    return (char*)block + AX_HEADER_SIZE;
}

static inline void* ax_user_to_block(void* user_ptr) {
    return (char*)user_ptr - AX_HEADER_SIZE;
}

static inline AxHeader* ax_get_header(void* user_ptr) {
    return (AxHeader*)ax_user_to_block(user_ptr);
}

/* --------------------------------------------------------------------------
 * Size Class Selection
 * -------------------------------------------------------------------------- */

/**
 * Returns the size class for a given user allocation size.
 * The total block size includes the 8-byte AxHeader.
 */
SizeClass ax_size_class_for(size_t user_size);

/* --------------------------------------------------------------------------
 * Free List
 * -------------------------------------------------------------------------- */

typedef struct FreeSlot {
    struct FreeSlot* next;
} FreeSlot;

typedef struct {
    FreeSlot* head;
    size_t    count;
} FreeList;

/** Push a freed block onto the free list. */
static inline void ax_free_list_push(FreeList* list, void* block) {
    FreeSlot* slot = (FreeSlot*)((char*)block + AX_HEADER_SIZE);
    slot->next = list->head;
    list->head = slot;
    list->count++;
}

/** Pop a block from the free list. Returns NULL if empty. */
static inline void* ax_free_list_pop(FreeList* list) {
    if (!list->head) return NULL;
    FreeSlot* slot = list->head;
    list->head = slot->next;
    list->count--;
    return (char*)slot - AX_HEADER_SIZE;
}

/* --------------------------------------------------------------------------
 * Bump Allocator
 * -------------------------------------------------------------------------- */

/**
 * Allocate from a bump region. Returns NULL if region is exhausted.
 * bump: pointer to current bump position (updated on success).
 * limit: end of the bump region.
 * sc: size class to allocate.
 */
void* ax_bump_alloc(char** bump, char* limit, SizeClass sc);

/* --------------------------------------------------------------------------
 * Size-Class Allocation API
 * -------------------------------------------------------------------------- */

/**
 * Allocate user_size bytes from free list + bump region.
 * Returns user pointer (header is hidden before it).
 * Returns NULL if both free list and bump region are exhausted.
 */
void* ax_size_class_alloc(FreeList* free_list, char** bump, char* limit,
                          size_t user_size);

/**
 * Free a size-class allocation. Pushes block onto the appropriate free list.
 * free_lists: array of NUM_SIZE_CLASSES FreeList entries.
 */
void ax_size_class_free(FreeList* free_lists, void* user_ptr);

/* --------------------------------------------------------------------------
 * Large Allocation (> 4096 bytes)
 * -------------------------------------------------------------------------- */

/** Allocate a large block using OS memory mapping. */
void* ax_large_alloc(size_t user_size);

/** Free a large block. */
void ax_large_free(void* user_ptr, size_t user_size);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_AXALLOC_SIZE_CLASSES_H */
