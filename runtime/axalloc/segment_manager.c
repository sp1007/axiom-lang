/*
 * p14-t02: AxAlloc Segment Manager — Implementation
 *
 * OS memory management layer for 64KB segments.
 * Uses mmap/VirtualAlloc for OS-level allocation.
 */

#include "segment_manager.h"
#include <string.h>

/* --------------------------------------------------------------------------
 * Platform Abstraction
 * -------------------------------------------------------------------------- */

#ifdef _WIN32
  #include <windows.h>
  static void* os_alloc(size_t size) {
      return VirtualAlloc(NULL, size, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
  }
  static void os_free(void* ptr, size_t size) {
      (void)size;
      VirtualFree(ptr, 0, MEM_RELEASE);
  }
#else
  #include <sys/mman.h>
  static void* os_alloc(size_t size) {
      void* p = mmap(NULL, size, PROT_READ | PROT_WRITE,
                      MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
      return (p == MAP_FAILED) ? NULL : p;
  }
  static void os_free(void* ptr, size_t size) {
      munmap(ptr, size);
  }
#endif

/* --------------------------------------------------------------------------
 * Segment Pool
 * -------------------------------------------------------------------------- */

static Segment  segment_slab[MAX_SEGMENTS];
static int      segment_slab_used = 0;
static Segment* free_segment_pool = NULL;

static Segment* alloc_segment_meta(void) {
    if (free_segment_pool) {
        Segment* seg = free_segment_pool;
        free_segment_pool = seg->next;
        return seg;
    }
    if (segment_slab_used >= MAX_SEGMENTS) return NULL;
    return &segment_slab[segment_slab_used++];
}

static void free_segment_meta(Segment* seg) {
    memset(seg, 0, sizeof(Segment));
    seg->next = free_segment_pool;
    free_segment_pool = seg;
}

/* --------------------------------------------------------------------------
 * Segment Lifecycle
 * -------------------------------------------------------------------------- */

Segment* ax_segment_acquire(SizeClass sc) {
    Segment* seg = alloc_segment_meta();
    if (!seg) return NULL;

    char* mem = (char*)os_alloc(SEGMENT_SIZE);
    if (!mem) {
        free_segment_meta(seg);
        return NULL;
    }

    seg->base   = mem;
    seg->bump   = mem;
    seg->limit  = mem + SEGMENT_SIZE;
    seg->sclass = sc;
    seg->next   = NULL;
    seg->magic  = SEGMENT_MAGIC;

    return seg;
}

void ax_segment_release(Segment* seg) {
    if (!seg || seg->magic != SEGMENT_MAGIC) return;

    os_free(seg->base, SEGMENT_SIZE);
    seg->magic = 0;
    free_segment_meta(seg);
}

Segment* ax_segment_get_active(SegmentList* list, SizeClass sc) {
    size_t block_size = SIZE_CLASS_SIZES[sc];
    if (list->active && (size_t)(list->active->limit - list->active->bump) >= block_size) {
        return list->active;
    }

    /* Retire current active segment */
    if (list->active) {
        list->active->next = list->retired;
        list->retired = list->active;
        list->count++;
    }

    /* Acquire a new segment */
    Segment* seg = ax_segment_acquire(sc);
    list->active = seg;
    return seg;
}

void ax_segment_list_release_all(SegmentList* list) {
    if (list->active) {
        ax_segment_release(list->active);
        list->active = NULL;
    }

    Segment* seg = list->retired;
    while (seg) {
        Segment* next = seg->next;
        ax_segment_release(seg);
        seg = next;
    }
    list->retired = NULL;
    list->count = 0;
}

/* --------------------------------------------------------------------------
 * Bump Allocation from Segment
 * -------------------------------------------------------------------------- */

void* ax_segment_bump_alloc(Segment* seg, SizeClass sc) {
    if (!seg || sc >= NUM_SIZE_CLASSES) return NULL;

    size_t block_size = SIZE_CLASS_SIZES[sc];
    if (seg->bump + block_size > seg->limit) return NULL;

    void* block = seg->bump;
    seg->bump += block_size;
    return block;
}

/* --------------------------------------------------------------------------
 * Diagnostics
 * -------------------------------------------------------------------------- */

double ax_segment_utilization(const Segment* seg) {
    if (!seg || seg->magic != SEGMENT_MAGIC) return 0.0;
    size_t used = (size_t)(seg->bump - seg->base);
    return (double)used / (double)SEGMENT_SIZE;
}

/* --------------------------------------------------------------------------
 * Init / Shutdown
 * -------------------------------------------------------------------------- */

void ax_segment_manager_init(void) {
    segment_slab_used = 0;
    free_segment_pool = NULL;
    memset(segment_slab, 0, sizeof(segment_slab));
}

void ax_segment_manager_shutdown(void) {
    /* Release all segments still in the pool */
    for (int i = 0; i < segment_slab_used; i++) {
        if (segment_slab[i].magic == SEGMENT_MAGIC) {
            os_free(segment_slab[i].base, SEGMENT_SIZE);
            segment_slab[i].magic = 0;
        }
    }
    segment_slab_used = 0;
    free_segment_pool = NULL;
}
