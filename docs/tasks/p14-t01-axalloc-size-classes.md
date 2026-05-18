# p14-t01: AxAlloc Size-Classed Allocation

## Purpose
Implement the foundational size-class allocation system for AxAlloc in `runtime/axalloc/size_classes.c`. Size classes provide O(1) allocation and deallocation by maintaining separate free lists for fixed-size memory buckets, minimizing fragmentation and avoiding the overhead of general-purpose allocators.

## Context
AxAlloc is AXIOM's custom allocator designed around the actor model. Each actor has an isolated heap, eliminating cross-actor lock contention. The size-class system is the core of this design: instead of tracking arbitrary-size allocations, memory is rounded up to the nearest size class, and free lists for each class make reclamation O(1).

Every heap allocation in AXIOM carries an 8-byte `AxHeader{gen_id: u32, flags: u32}` for generational reference validation. The size-class system must account for this header in its size calculations.

Size classes: 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096 bytes. Allocations larger than 4096 bytes bypass the size-class system and use direct `mmap` (handled by the segment manager in p14-t02).

## Inputs
- AXIOM memory model specification (from `09. Runtime architecture production-grade.md`)
- `AxHeader` struct definition from the generational reference system (p07-t01)
- C11 standard library (`stdint.h`, `stddef.h`, `stdalign.h`)

## Outputs
- `runtime/axalloc/size_classes.h` — size class definitions and API declarations
- `runtime/axalloc/size_classes.c` — implementation
- Test file: `runtime/axalloc/size_classes_test.c` (using a simple C test harness)

## Dependencies
- p07-t01: Generational reference system (defines `AxHeader` structure)

## Subsystems Affected
- `runtime/axalloc/` — new files
- `runtime/axalloc/segment_manager.c` (p14-t02) will use size class definitions

## Detailed Requirements

### Size Class Table
```c
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
    SIZE_CLASS_LARGE = 10,  // > 4096: direct mmap
} SizeClass;

static const size_t SIZE_CLASS_SIZES[NUM_SIZE_CLASSES] = {
    8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096
};
```

### AxHeader Integration
Every allocation includes an 8-byte header before the user-visible pointer:
```c
typedef struct {
    uint32_t gen_id;   // generational ID for reference validation
    uint32_t flags;    // GC flags, type tag (future use)
} AxHeader;

// Total allocation block size = sizeof(AxHeader) + user_size
// Round up to size class
```

Size class selection:
```c
SizeClass ax_size_class_for(size_t user_size) {
    size_t total = user_size + sizeof(AxHeader);  // 8 bytes header
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
```

User pointer = block base + sizeof(AxHeader):
```c
static inline void* ax_block_to_user(void* block) {
    return (char*)block + sizeof(AxHeader);
}

static inline void* ax_user_to_block(void* user_ptr) {
    return (char*)user_ptr - sizeof(AxHeader);
}

static inline AxHeader* ax_get_header(void* user_ptr) {
    return (AxHeader*)ax_user_to_block(user_ptr);
}
```

### Free List Structure
Each size class maintains a free list of recycled blocks. The free list is overlaid onto the freed memory itself (after the AxHeader):
```c
typedef struct FreeSlot {
    struct FreeSlot* next;  // stored in bytes [8..15] of the block (after 8-byte header)
} FreeSlot;

typedef struct {
    FreeSlot* head;
    size_t    count;  // for diagnostics
} FreeList;
```

The `FreeSlot.next` pointer is placed at offset 8 from the block base (i.e., where the user data was), since the AxHeader occupies bytes 0-7. This works because:
- A free block is not accessible to user code (the type system enforces this)
- We need only 8 bytes to store the next pointer

### Bump Allocator Integration
Each size class has a bump pointer within a 64KB segment (managed by p14-t02). The size-class module exposes:
```c
// Allocate from a bump pointer region
// Returns NULL if bump region is exhausted (caller must get new segment)
void* ax_bump_alloc(char** bump, char* limit, SizeClass sc);
```

```c
void* ax_bump_alloc(char** bump, char* limit, SizeClass sc) {
    size_t block_size = SIZE_CLASS_SIZES[sc];
    if (*bump + block_size > limit) return NULL;
    void* block = *bump;
    *bump += block_size;
    return block;
}
```

### Free List Operations
```c
// Push a block onto the free list for its size class
static inline void ax_free_list_push(FreeList* list, void* block) {
    FreeSlot* slot = (FreeSlot*)((char*)block + sizeof(AxHeader));
    slot->next = list->head;
    list->head = slot;
    list->count++;
}

// Pop a block from the free list (returns NULL if empty)
static inline void* ax_free_list_pop(FreeList* list) {
    if (!list->head) return NULL;
    FreeSlot* slot = list->head;
    list->head = slot->next;
    list->count--;
    // Convert FreeSlot* (which points to byte 8) back to block base
    return (char*)slot - sizeof(AxHeader);
}
```

### Allocation Path
```c
// ax_alloc: allocate user_size bytes from the given free list + bump region
// Returns user pointer (header is hidden before the pointer)
void* ax_size_class_alloc(FreeList* free_list, char** bump, char* limit, size_t user_size) {
    // Try free list first
    void* block = ax_free_list_pop(free_list);
    if (!block) {
        // Fall back to bump allocator
        SizeClass sc = ax_size_class_for(user_size);
        block = ax_bump_alloc(bump, limit, sc);
        if (!block) return NULL;  // caller must get new segment
    }
    // Initialize header
    AxHeader* hdr = (AxHeader*)block;
    hdr->gen_id = 1;  // generation 1 = freshly allocated
    hdr->flags  = 0;
    return ax_block_to_user(block);
}
```

### Deallocation Path
```c
void ax_size_class_free(FreeList* free_lists, void* user_ptr) {
    void* block = ax_user_to_block(user_ptr);
    AxHeader* hdr = (AxHeader*)block;
    hdr->gen_id = 0;  // invalidate: gen_id=0 means freed
    // Determine size class from block (requires caller to know, or store in header)
    // NOTE: store size class in hdr->flags for O(1) class lookup
    SizeClass sc = (SizeClass)(hdr->flags & 0xF);
    ax_free_list_push(&free_lists[sc], block);
}
```

Store the size class in the lower 4 bits of `hdr->flags` at allocation time so free can be O(1).

### Large Allocation (> 4096 bytes)
Bypass size classes, use direct mmap:
```c
void* ax_large_alloc(size_t user_size) {
    size_t total = sizeof(AxHeader) + user_size;
    total = (total + 4095) & ~4095;  // round up to page
    void* block = mmap(NULL, total, PROT_READ|PROT_WRITE,
                       MAP_PRIVATE|MAP_ANONYMOUS, -1, 0);
    if (block == MAP_FAILED) return NULL;
    AxHeader* hdr = (AxHeader*)block;
    hdr->gen_id = 1;
    hdr->flags  = SIZE_CLASS_LARGE;
    return ax_block_to_user(block);
}

void ax_large_free(void* user_ptr, size_t user_size) {
    void* block = ax_user_to_block(user_ptr);
    size_t total = (sizeof(AxHeader) + user_size + 4095) & ~4095;
    munmap(block, total);
}
```

## Implementation Steps

### Step 1: Create Header File
Create `runtime/axalloc/size_classes.h` with all type definitions and function declarations.

### Step 2: Implement Size Class Selection
Implement `ax_size_class_for()` as a branchless lookup or simple if-chain. Consider a lookup table for user sizes 0-4096.

### Step 3: Implement Free List Operations
Implement `ax_free_list_push()` and `ax_free_list_pop()` as inline functions in the header.

### Step 4: Implement Allocation/Deallocation
Implement `ax_size_class_alloc()` and `ax_size_class_free()` in the .c file.

### Step 5: Implement Large Allocation
Implement `ax_large_alloc()` and `ax_large_free()` using mmap/munmap.

### Step 6: Write Tests
Create `runtime/axalloc/size_classes_test.c`:
```c
void test_size_class_selection(void) {
    assert(ax_size_class_for(0)    == SIZE_CLASS_8);    // 0+8=8
    assert(ax_size_class_for(1)    == SIZE_CLASS_16);   // 1+8=9, rounds to 16
    assert(ax_size_class_for(8)    == SIZE_CLASS_16);   // 8+8=16
    assert(ax_size_class_for(9)    == SIZE_CLASS_32);   // 9+8=17, rounds to 32
    assert(ax_size_class_for(56)   == SIZE_CLASS_64);   // 56+8=64
    assert(ax_size_class_for(57)   == SIZE_CLASS_128);
    assert(ax_size_class_for(4089) == SIZE_CLASS_4096); // 4089+8=4097, rounds to... LARGE
    assert(ax_size_class_for(4088) == SIZE_CLASS_4096); // 4088+8=4096
    assert(ax_size_class_for(4089) == SIZE_CLASS_LARGE);
}
```

## Test Plan

### Unit Tests
1. `test_size_class_selection` — all boundary values (0, 1, 7, 8, 9, 55, 56, 57, ...)
2. `test_free_list_push_pop` — push 5 blocks, pop 5 blocks (LIFO order)
3. `test_free_list_empty` — pop from empty list returns NULL
4. `test_header_gen_id` — freshly allocated block has gen_id=1
5. `test_header_invalidated_on_free` — freed block has gen_id=0
6. `test_size_class_stored_in_flags` — flags lower 4 bits == size class index
7. `test_alloc_free_cycle` — alloc 100 blocks, free all, alloc 100 again (from free list)
8. `test_large_alloc` — allocate 8192 bytes, verify user pointer is valid, free it
9. `test_user_header_conversion` — `ax_block_to_user` and `ax_user_to_block` are inverses
10. `test_bump_alloc_exhaustion` — bump alloc returns NULL when region is full

### Stress Tests
1. Allocate 10000 blocks of size 1 → all size class 16 → free list should handle recycling
2. Alternating alloc/free of different sizes → verify no corruption

## Validation Checklist
- [ ] All 10 size classes defined correctly
- [ ] `AxHeader` (8 bytes) included in size class total
- [ ] Size class stored in `hdr->flags` lower 4 bits for O(1) free
- [ ] gen_id=1 on alloc, gen_id=0 on free
- [ ] Free list is a singly-linked list overlaid on freed memory
- [ ] Large allocations (> 4096 bytes user) use mmap directly
- [ ] No use of malloc/free (axalloc replaces the system allocator for managed heap)
- [ ] Thread safety: these functions are NOT thread-safe (thread safety is in p14-t03 sharding)

## Acceptance Criteria
1. All unit tests pass
2. `ax_size_class_for(user_size)` correct for all boundary values 0-4097
3. Alloc/free cycle for 10K blocks in all 10 size classes works without corruption
4. Large alloc/free (4097 bytes) uses mmap, verified by `valgrind --tool=massif`
5. No memory leaks detected by `valgrind --leak-check=full`

## Definition of Done
- `size_classes.h` and `size_classes.c` implemented and reviewed
- All unit tests pass
- Valgrind clean (no leaks, no invalid reads/writes)
- Functions documented with parameter/return value contracts in header

## Risks & Mitigations
| Risk | Mitigation |
|---|---|
| FreeSlot pointer alignment issues | Ensure size classes are always ≥ 16 bytes (header 8 + next 8); min class is 8 but effective min is 16 due to header |
| gen_id overflow (wrap around) | gen_id is u32, wraps at 2^32; document this; generational validation in p07 handles it |
| Large alloc page size on non-Linux | Abstract mmap behind `ax_os_alloc(size)` for portability |

## Future Follow-up Tasks
- p14-t02: Segment manager provides the 64KB bump regions used here
- p14-t03: Free-list sharding adds per-shard thread safety
- p14-t04: Per-actor heap wraps all of this in an isolated heap structure
