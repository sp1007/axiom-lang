# p14-t02: AxAlloc Segment Manager

## Purpose
Implement the 64KB segment management layer for AxAlloc in `runtime/axalloc/segment_manager.c`. Segments are the memory regions obtained from the OS that feed the bump allocators for each size class. The segment manager provides O(1) segment acquisition and O(1) release back to the OS.

## Context
AxAlloc uses 64KB segments as the fundamental unit of OS memory acquisition. Each size class maintains a list of segments, and within each segment, a bump pointer advances linearly until the segment is exhausted. Exhausted segments are added to a retired list; segments that still have free slots from the free list remain active.

Using 64KB as the segment size is a deliberate trade-off:
- Small enough to limit internal fragmentation for small size classes
- Large enough to amortize mmap syscall cost (typically 1-5 microseconds)
- Matches the typical huge page sub-division size on Linux (transparent huge pages are 2MB)
- On Windows, `VirtualAlloc` granularity is 64KB, making this a natural fit

## Inputs
- Size class definitions from p14-t01 (`size_classes.h`)
- Platform: Linux (mmap), Windows (VirtualAlloc), macOS (mmap with MAP_ANONYMOUS)
- Target segment size: 64KB (65536 bytes)

## Outputs
- `runtime/axalloc/segment_manager.h` — segment struct and API
- `runtime/axalloc/segment_manager.c` — implementation
- Test file: `runtime/axalloc/segment_manager_test.c`

## Dependencies
- p14-t01: Size class definitions (segment is carved up per size class)

## Subsystems Affected
- `runtime/axalloc/` — new files
- `runtime/axalloc/actor_heap.c` (p14-t04) uses segment lists

## Detailed Requirements

### Segment Structure
```c
#define SEGMENT_SIZE (64 * 1024)  // 65536 bytes

typedef struct Segment {
    char*           base;    // mmap base address
    char*           top;     // base + SEGMENT_SIZE
    char*           bump;    // current allocation frontier (base ≤ bump ≤ top)
    SizeClass       sclass;  // which size class this segment serves
    struct Segment* next;    // linked list pointer (for segment lists)
    uint32_t        magic;   // 0xAA55AA55 — detect corruption
} Segment;
```

The `Segment` struct itself is allocated from a small metadata slab (not from the segment it manages), to avoid contaminating the bump region.

### Segment Metadata Allocation
To avoid bootstrapping problems (we need memory to manage memory), maintain a static slab of `Segment` structs:
```c
#define MAX_SEGMENTS 4096  // supports up to 4096 * 64KB = 256MB per actor

static Segment segment_slab[MAX_SEGMENTS];
static uint32_t segment_slab_used = 0;

Segment* alloc_segment_meta(void) {
    if (segment_slab_used >= MAX_SEGMENTS) return NULL;
    return &segment_slab[segment_slab_used++];
}
```

Alternatively, use a separate small mmap for segment metadata. The static array approach is simpler for MVP.

### OS Memory Acquisition

#### POSIX (Linux, macOS)
```c
Segment* ax_segment_acquire(SizeClass sc) {
    void* mem = mmap(
        NULL, SEGMENT_SIZE,
        PROT_READ | PROT_WRITE,
        MAP_PRIVATE | MAP_ANONYMOUS,
        -1, 0
    );
    if (mem == MAP_FAILED) return NULL;
    
    Segment* seg = alloc_segment_meta();
    seg->base   = (char*)mem;
    seg->top    = (char*)mem + SEGMENT_SIZE;
    seg->bump   = (char*)mem;
    seg->sclass = sc;
    seg->next   = NULL;
    seg->magic  = 0xAA55AA55;
    return seg;
}
```

#### Windows
```c
#ifdef _WIN32
Segment* ax_segment_acquire(SizeClass sc) {
    void* mem = VirtualAlloc(NULL, SEGMENT_SIZE,
                             MEM_COMMIT | MEM_RESERVE,
                             PAGE_READWRITE);
    if (!mem) return NULL;
    // ... same init as POSIX
}
#endif
```

Platform detection via `#ifdef _WIN32` / `#else`.

### Segment Release
```c
void ax_segment_release(Segment* seg) {
    assert(seg->magic == 0xAA55AA55 && "segment_release: corrupted segment");
    seg->magic = 0xDEAD0000;  // poison on release
#ifdef _WIN32
    VirtualFree(seg->base, 0, MEM_RELEASE);
#else
    munmap(seg->base, SEGMENT_SIZE);
#endif
    // Return seg meta to slab (simple: just decrement if it was the last)
    // For MVP: leak the metadata slot (bounded by MAX_SEGMENTS)
}
```

### Segment List
Each per-actor heap maintains one segment list per size class:
```c
typedef struct SegmentList {
    Segment* active;   // current segment (has bump space remaining)
    Segment* retired;  // full segments (bump == top)
    uint32_t count;
} SegmentList;
```

```c
// Get or create an active segment for a size class
Segment* ax_segment_get_active(SegmentList* list, SizeClass sc) {
    if (list->active && list->active->bump < list->active->top) {
        return list->active;  // still has space
    }
    // Current segment is full; retire it
    if (list->active) {
        list->active->next = list->retired;
        list->retired = list->active;
        list->active = NULL;
    }
    // Acquire a new segment
    Segment* seg = ax_segment_acquire(sc);
    list->active = seg;
    list->count++;
    return seg;
}
```

### Bulk Release (for actor death)
When an actor dies, all its segments are released in O(N_segments) time:
```c
void ax_segment_list_release_all(SegmentList* list) {
    Segment* seg = list->active;
    while (seg) {
        Segment* next = seg->next;
        ax_segment_release(seg);
        seg = next;
    }
    seg = list->retired;
    while (seg) {
        Segment* next = seg->next;
        ax_segment_release(seg);
        seg = next;
    }
    list->active  = NULL;
    list->retired = NULL;
    list->count   = 0;
}
```

This is O(N_segments) = O(total_allocated / 64KB), which for typical actors is very small (< 100 segments).

### Bump Pointer Allocation within Segment
```c
void* ax_segment_bump_alloc(Segment* seg, size_t block_size) {
    assert(seg->magic == 0xAA55AA55);
    if (seg->bump + block_size > seg->top) return NULL;
    void* p = seg->bump;
    seg->bump += block_size;
    return p;
}
```

### Segment Utilization Tracking
```c
float ax_segment_utilization(Segment* seg) {
    return (float)(seg->bump - seg->base) / SEGMENT_SIZE;
}

size_t ax_segment_list_total_bytes(SegmentList* list) {
    // Walk all segments and sum (bump - base)
}
```

## Implementation Steps

### Step 1: Create Platform Abstraction
Create `runtime/axalloc/platform.h`:
```c
#ifdef _WIN32
  #include <windows.h>
  #define AX_OS_ALLOC(size) VirtualAlloc(NULL, (size), MEM_COMMIT|MEM_RESERVE, PAGE_READWRITE)
  #define AX_OS_FREE(ptr, size) VirtualFree((ptr), 0, MEM_RELEASE)
#else
  #include <sys/mman.h>
  #define AX_OS_ALLOC(size) mmap(NULL, (size), PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0)
  #define AX_OS_FREE(ptr, size) munmap((ptr), (size))
  #define MAP_FAILED_CHECK(p) ((p) == MAP_FAILED)
#endif
```

### Step 2: Implement Segment Structure and Slab
Implement `segment_manager.h` with struct definitions and `segment_manager.c` with the slab allocator for metadata.

### Step 3: Implement acquire/release
Wire `AX_OS_ALLOC`/`AX_OS_FREE` into `ax_segment_acquire` and `ax_segment_release`.

### Step 4: Implement SegmentList Operations
Implement `ax_segment_get_active`, `ax_segment_list_release_all`, and utilization reporting.

### Step 5: Tests
```c
void test_segment_acquire_release(void) {
    Segment* seg = ax_segment_acquire(SIZE_CLASS_64);
    assert(seg != NULL);
    assert(seg->base != NULL);
    assert(seg->bump == seg->base);
    assert(seg->top == seg->base + SEGMENT_SIZE);
    ax_segment_release(seg);
}

void test_bump_alloc(void) {
    Segment* seg = ax_segment_acquire(SIZE_CLASS_64);
    void* p1 = ax_segment_bump_alloc(seg, 64);
    void* p2 = ax_segment_bump_alloc(seg, 64);
    assert(p1 != NULL);
    assert(p2 != NULL);
    assert((char*)p2 == (char*)p1 + 64);
    ax_segment_release(seg);
}

void test_segment_exhaustion(void) {
    Segment* seg = ax_segment_acquire(SIZE_CLASS_4096);
    int count = 0;
    while (ax_segment_bump_alloc(seg, 4096) != NULL) count++;
    assert(count == SEGMENT_SIZE / 4096);  // 16 blocks per 64KB segment
    ax_segment_release(seg);
}

void test_bulk_release(void) {
    SegmentList list = {0};
    for (int i = 0; i < 5; i++) {
        ax_segment_get_active(&list, SIZE_CLASS_64);
        // exhaust it
        while (ax_segment_bump_alloc(list.active, 64)) {}
    }
    ax_segment_list_release_all(&list);
    assert(list.active == NULL);
    assert(list.retired == NULL);
}
```

## Test Plan

### Unit Tests
1. `test_segment_acquire_release` — acquire, verify fields, release
2. `test_bump_alloc` — sequential bumps produce non-overlapping addresses
3. `test_bump_alloc_alignment` — all returned pointers are 8-byte aligned
4. `test_segment_exhaustion` — bump returns NULL when full
5. `test_segment_list_get_active` — returns same segment if not full, new segment if full
6. `test_segment_list_retires_on_full` — full segment moved to retired list
7. `test_bulk_release` — all segments released, list zeroed
8. `test_magic_corruption_detection` — modifying magic → assert fails in debug build
9. `test_multiple_size_classes` — segments for different size classes don't overlap
10. `test_windows_virtual_alloc` (Windows only) — VirtualAlloc path works

### Stress Tests
1. Acquire 1000 segments across 10 size classes, release all — check no OS resource leak
2. Interleave acquire/bump/release in random order for 10K operations

## Validation Checklist
- [ ] `SEGMENT_SIZE` is exactly 65536 bytes
- [ ] Segment base is always page-aligned (guaranteed by mmap/VirtualAlloc)
- [ ] Bump pointer never exceeds `top`
- [ ] Magic byte set on acquire, poisoned on release
- [ ] `ax_segment_list_release_all` walks both `active` and `retired` lists
- [ ] Platform abstraction compiles on Linux, macOS, and Windows
- [ ] `segment_slab` is statically allocated (no malloc dependency)

## Acceptance Criteria
1. All unit tests pass on Linux and Windows
2. Stress test: 1000 segment acquire/release cycles with no OS resource leak (verify with `/proc/self/maps` on Linux or `HeapValidate` on Windows)
3. `test_segment_exhaustion` shows correct block count per size class
4. `ax_segment_list_release_all` verified clean with valgrind on Linux

## Definition of Done
- `segment_manager.h` and `segment_manager.c` implemented and reviewed
- Platform abstraction in `platform.h` works on Linux, macOS, and Windows
- All unit tests pass
- No memory leaks in stress test

## Risks & Mitigations
| Risk | Mitigation |
|---|---|
| Static slab of 4096 segment metas may be too small | Document limit; add assertion; make configurable via build flag |
| VirtualAlloc granularity on Windows is 64KB (already aligned) | Verify with `GetSystemInfo`; 64KB segment size matches granularity |
| mmap MAP_ANONYMOUS not available on older kernels | Require Linux ≥ 2.4 (MAP_ANONYMOUS available since 2.0) |

## Future Follow-up Tasks
- p14-t03: Free-list sharding uses segment lists from this module
- p14-t04: Per-actor heap holds one `SegmentList` per size class
- p14-t05: NUMA-aware allocation extends `ax_segment_acquire` with `mbind`
