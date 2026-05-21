/*
 * p14-t01: AxAlloc Size-Classed Allocation — Implementation
 *
 * Core allocation primitives: size class selection, bump allocation,
 * free-list recycling, and large allocation via OS memory mapping.
 */

#include "size_classes.h"
#include <string.h>

/* --------------------------------------------------------------------------
 * Platform Abstraction
 * -------------------------------------------------------------------------- */

#ifdef _WIN32
  #include <windows.h>
  #define AX_PAGE_SIZE 4096

  static void* ax_os_alloc(size_t size) {
      return VirtualAlloc(NULL, size, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
  }

  static void ax_os_free(void* ptr, size_t size) {
      (void)size;
      VirtualFree(ptr, 0, MEM_RELEASE);
  }
#else
  #include <sys/mman.h>
  #define AX_PAGE_SIZE 4096

  static void* ax_os_alloc(size_t size) {
      void* p = mmap(NULL, size, PROT_READ | PROT_WRITE,
                      MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
      return (p == MAP_FAILED) ? NULL : p;
  }

  static void ax_os_free(void* ptr, size_t size) {
      munmap(ptr, size);
  }
#endif

/* --------------------------------------------------------------------------
 * Size Class Selection
 * -------------------------------------------------------------------------- */

SizeClass ax_size_class_for(size_t user_size) {
    size_t total = user_size + AX_HEADER_SIZE;

    if (total <= 8)    return SIZE_CLASS_8;
    if (total <= 16)   return SIZE_CLASS_16;
    if (total <= 32)   return SIZE_CLASS_32;
    if (total <= 64)   return SIZE_CLASS_64;
    if (total <= 128)  return SIZE_CLASS_128;
    if (total <= 256)  return SIZE_CLASS_256;
    if (total <= 512)  return SIZE_CLASS_512;
    if (total <= 1024) return SIZE_CLASS_1024;
    if (total <= 2048) return SIZE_CLASS_2048;
    if (total <= 4096) return SIZE_CLASS_4096;
    return SIZE_CLASS_LARGE;
}

/* --------------------------------------------------------------------------
 * Bump Allocator
 * -------------------------------------------------------------------------- */

void* ax_bump_alloc(char** bump, char* limit, SizeClass sc) {
    if (sc < 0 || sc >= NUM_SIZE_CLASSES) return NULL;
    size_t block_size = SIZE_CLASS_SIZES[sc];
    if (*bump + block_size > limit) return NULL;

    void* block = *bump;
    *bump += block_size;
    memset(block, 0, block_size);
    return block;
}

/* --------------------------------------------------------------------------
 * Size-Class Allocation
 * -------------------------------------------------------------------------- */

void* ax_size_class_alloc(FreeList* free_list, char** bump, char* limit,
                          size_t user_size) {
    SizeClass sc = ax_size_class_for(user_size);

    if (sc == SIZE_CLASS_LARGE) {
        return ax_large_alloc(user_size);
    }

    /* Try free list first */
    void* block = ax_free_list_pop(free_list);

    if (!block) {
        /* Fall back to bump allocator */
        block = ax_bump_alloc(bump, limit, sc);
        if (!block) return NULL;
    }

    /* Initialize header */
    AxHeader* hdr = (AxHeader*)block;
    hdr->gen_id = 1;                   /* gen_id=1: freshly allocated */
    hdr->flags  = (uint32_t)sc;        /* store size class in lower 4 bits */

    return ax_block_to_user(block);
}

/* --------------------------------------------------------------------------
 * Size-Class Deallocation
 * -------------------------------------------------------------------------- */

void ax_size_class_free(FreeList* free_lists, void* user_ptr) {
    if (!user_ptr) return;

    void* block = ax_user_to_block(user_ptr);
    AxHeader* hdr = (AxHeader*)block;

    /* Invalidate the generational ID */
    hdr->gen_id = 0;

    /* Extract size class from flags (lower 4 bits) */
    SizeClass sc = (SizeClass)(hdr->flags & 0xF);

    if (sc == SIZE_CLASS_LARGE || sc >= NUM_SIZE_CLASSES) {
        /* Large allocations are handled separately */
        return;
    }

    ax_free_list_push(&free_lists[sc], block);
}

/* --------------------------------------------------------------------------
 * Large Allocation (> 4096 bytes)
 * -------------------------------------------------------------------------- */

void* ax_large_alloc(size_t user_size) {
    size_t total = AX_HEADER_SIZE + user_size;
    total = (total + AX_PAGE_SIZE - 1) & ~(AX_PAGE_SIZE - 1);  /* page-align */

    void* block = ax_os_alloc(total);
    if (!block) return NULL;

    AxHeader* hdr = (AxHeader*)block;
    hdr->gen_id = 1;
    hdr->flags  = SIZE_CLASS_LARGE;

    return ax_block_to_user(block);
}

void ax_large_free(void* user_ptr, size_t user_size) {
    if (!user_ptr) return;

    void* block = ax_user_to_block(user_ptr);
    size_t total = AX_HEADER_SIZE + user_size;
    total = (total + AX_PAGE_SIZE - 1) & ~(AX_PAGE_SIZE - 1);

    ax_os_free(block, total);
}
