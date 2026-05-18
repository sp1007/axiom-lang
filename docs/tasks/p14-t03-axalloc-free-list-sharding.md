# p14-t03: AxAlloc Free-List Sharding

## Purpose
Implement sharded free lists for each size class to reduce contention in multi-actor scenarios where multiple threads may return memory to the same actor heap concurrently. Sharding distributes the free list across N CPU-local shards, so that threads rarely contend on the same shard.

## Context
In AXIOM's actor model, each actor owns its heap exclusively during normal operation. However, when an actor is migrated between scheduler threads (work stealing) or when memory is returned from a previous actor incarnation, concurrent access to the free list can occur. Sharding eliminates most of this contention by partitioning the free list per CPU shard.

Each shard has its own `head` pointer and a spinlock. A thread picks its shard as `thread_id % N_SHARDS`. When a shard's local free list runs dry, the thread batch-steals 16 entries from the global (unsharded) pool, amortizing lock acquisitions.

This is a standard technique used by `jemalloc` (tcache) and `mimalloc` (thread-local free lists).

## Inputs
- Free list definitions from p14-t01 (`size_classes.h` — `FreeList`, `FreeSlot`)
- Segment manager from p14-t02 (provides fresh blocks when free lists are exhausted)
- C11 atomics (`<stdatomic.h>`) for the spinlock implementation
- Thread ID: available via `pthread_self()` on POSIX, `GetCurrentThreadId()` on Windows

## Outputs
- `runtime/axalloc/free_list_shard.h` — sharded free list structure and API
- `runtime/axalloc/free_list_shard.c` — implementation
- Test file: `runtime/axalloc/free_list_shard_test.c`

## Dependencies
- p14-t01: Size class and free list definitions
- p14-t02: Segment manager (for batch refill)

## Subsystems Affected
- `runtime/axalloc/actor_heap.c` (p14-t04) uses sharded free lists

## Detailed Requirements

### Sharding Strategy
```c
#define N_SHARDS 8  // number of shards per size class per actor heap
                    // should be >= max concurrent threads accessing this heap

typedef struct {
    FreeSlot*    head;          // head of the shard's free list
    size_t       count;         // number of free slots in this shard
    _Atomic bool lock;          // spinlock: false=unlocked, true=locked
    char         _pad[7];       // padding to 64-byte cache line alignment
} __attribute__((aligned(64))) FreeListShard;
```

Each shard is padded to exactly one cache line (64 bytes on x86-64/ARM64) to prevent false sharing between shards.

### Spinlock Implementation
```c
static inline void shard_lock(FreeListShard* shard) {
    bool expected = false;
    while (!atomic_compare_exchange_weak_explicit(
        &shard->lock, &expected, true,
        memory_order_acquire, memory_order_relaxed
    )) {
        expected = false;
        // Spin with pause/yield to reduce CPU pressure
        #if defined(__x86_64__)
            __asm__ volatile("pause" ::: "memory");
        #elif defined(__aarch64__)
            __asm__ volatile("yield" ::: "memory");
        #elif defined(__riscv)
            __asm__ volatile("" ::: "memory");  // No pause on RISC-V, just memory barrier
        #endif
    }
}

static inline void shard_unlock(FreeListShard* shard) {
    atomic_store_explicit(&shard->lock, false, memory_order_release);
}
```

### Sharded Free List Per Size Class
```c
typedef struct {
    FreeListShard shards[N_SHARDS];  // N_SHARDS × 64 bytes = 512 bytes per size class
} ShardedFreeList;

// Full sharded free list for all size classes:
// NUM_SIZE_CLASSES × ShardedFreeList = 10 × 512 = 5120 bytes per actor heap
```

### Thread-to-Shard Mapping
```c
static inline uint32_t get_shard_index(void) {
#ifdef _WIN32
    return GetCurrentThreadId() % N_SHARDS;
#else
    return (uint32_t)pthread_self() % N_SHARDS;
#endif
}
```

In practice, `pthread_self()` returns a pointer-sized value that varies per thread. Using `% N_SHARDS` gives a roughly uniform distribution.

### Allocation from Shard
```c
void* ax_shard_alloc(ShardedFreeList* sfl, SizeClass sc,
                     SegmentList* segs, size_t user_size)
{
    uint32_t si = get_shard_index();
    FreeListShard* shard = &sfl->shards[si];

    shard_lock(shard);
    void* block = free_list_shard_pop(shard);
    shard_unlock(shard);

    if (block) {
        goto init_and_return;
    }

    // Shard empty: batch-refill from segment bump allocator
    {
        size_t block_size = SIZE_CLASS_SIZES[sc];
        shard_lock(shard);
        // Batch: take up to 16 blocks from bump allocator
        int refilled = 0;
        for (int i = 0; i < 16 && refilled < 16; i++) {
            Segment* seg = ax_segment_get_active(segs, sc);
            void* b = ax_segment_bump_alloc(seg, block_size);
            if (!b) break;
            free_list_shard_push(shard, b);
            refilled++;
        }
        block = free_list_shard_pop(shard);
        shard_unlock(shard);
    }

    if (!block) return NULL;  // OOM

init_and_return:
    AxHeader* hdr = (AxHeader*)block;
    hdr->gen_id = 1;
    hdr->flags  = (uint32_t)sc;
    return ax_block_to_user(block);
}
```

### Deallocation to Shard
```c
void ax_shard_free(ShardedFreeList* sfl, void* user_ptr) {
    void* block = ax_user_to_block(user_ptr);
    AxHeader* hdr = (AxHeader*)block;
    SizeClass sc = (SizeClass)(hdr->flags & 0xF);
    hdr->gen_id = 0;  // invalidate

    if (sc == SIZE_CLASS_LARGE) {
        // Large allocations bypass shards
        ax_large_free(user_ptr, /* size unknown — need to store in header */);
        return;
    }

    uint32_t si = get_shard_index();
    FreeListShard* shard = &sfl->shards[sc * N_SHARDS + si];
    // Note: shards indexed as [sc][shard_index] or flat [sc * N_SHARDS + si]

    shard_lock(shard);
    free_list_shard_push(shard, block);
    shard_unlock(shard);
}
```

### Batch Reclaim
When a shard's count exceeds a high-water mark (e.g., 64 slots), return a batch of 16 slots to the global segment pool (not yet released to OS — that happens at actor death).

```c
#define SHARD_HIGH_WATER 64
#define SHARD_BATCH_RETURN 16

void ax_shard_maybe_flush(FreeListShard* shard, GlobalFreePool* pool) {
    if (shard->count < SHARD_HIGH_WATER) return;
    shard_lock(shard);
    // Move 16 slots from shard to pool
    for (int i = 0; i < SHARD_BATCH_RETURN && shard->head; i++) {
        void* block = free_list_shard_pop(shard);
        global_pool_push(pool, block);
    }
    shard_unlock(shard);
}
```

(The `GlobalFreePool` is a future optimization; for MVP, simply keep all freed blocks in shards until actor death.)

### FreeSlot Push/Pop Helpers
```c
static inline void free_list_shard_push(FreeListShard* shard, void* block) {
    FreeSlot* slot = (FreeSlot*)((char*)block + sizeof(AxHeader));
    slot->next = shard->head;
    shard->head = slot;
    shard->count++;
}

static inline void* free_list_shard_pop(FreeListShard* shard) {
    if (!shard->head) return NULL;
    FreeSlot* slot = shard->head;
    shard->head = slot->next;
    shard->count--;
    return (char*)slot - sizeof(AxHeader);
}
```

## Implementation Steps

### Step 1: Define Structures
Create `runtime/axalloc/free_list_shard.h` with all struct definitions, constants, and inline functions.

### Step 2: Implement Spinlock
Implement `shard_lock` and `shard_unlock` with platform-specific `pause`/`yield` hints.

### Step 3: Implement Alloc/Free
Implement `ax_shard_alloc` and `ax_shard_free` in `free_list_shard.c`.

### Step 4: Implement Batch Refill
Implement the batch-refill loop in `ax_shard_alloc` that pulls from the segment bump allocator.

### Step 5: Tests
Write `free_list_shard_test.c` covering single-threaded correctness and basic multi-threaded safety.

## Test Plan

### Unit Tests (Single-Threaded)
1. `test_shard_push_pop` — push 5 blocks, pop 5 blocks (LIFO)
2. `test_shard_empty_pop` — pop from empty shard returns NULL
3. `test_shard_index` — `get_shard_index()` returns value in [0, N_SHARDS)
4. `test_shard_alloc_uses_free_list` — free then alloc same size → free list hit (no new segment)
5. `test_shard_alloc_refill` — drain shard, alloc → triggers batch refill of 16 blocks
6. `test_cache_line_alignment` — `sizeof(FreeListShard) == 64`

### Multi-Threaded Tests
1. `test_concurrent_alloc_free` — 4 threads each alloc/free 10K blocks of size 64 concurrently; no crash, no TSAN report
2. `test_shard_distribution` — verify that different threads hit different shards (reduce lock contention)

### Correctness Tests
1. Allocate 100 blocks, free all, allocate 100 again — all 100 come from free list (verify no new segment bump allocations)
2. Verify gen_id=1 on fresh alloc, gen_id=0 after free

## Validation Checklist
- [ ] `FreeListShard` is exactly 64 bytes (fits in one cache line)
- [ ] Spinlock uses `_Atomic bool` with acquire/release memory ordering
- [ ] Batch refill acquires 16 blocks per refill (not 1 at a time)
- [ ] `shard_free` invalidates gen_id=0 before pushing to free list
- [ ] Large allocations (SIZE_CLASS_LARGE) bypass shard
- [ ] `get_shard_index()` works on POSIX and Windows
- [ ] Platform-specific `pause`/`yield` hints present in spinlock spin loop

## Acceptance Criteria
1. Single-threaded alloc/free of 10K blocks of each size class: no corruption
2. Multi-threaded test with 4 threads passes with Thread Sanitizer (TSAN) enabled
3. `sizeof(FreeListShard) == 64` verified via static assert
4. Batch refill verified: after exhausting a shard, next alloc refills with exactly min(16, available) blocks

## Definition of Done
- `free_list_shard.h` and `free_list_shard.c` implemented and reviewed
- All unit tests pass
- Multi-threaded test passes with TSAN
- Cache-line alignment verified

## Risks & Mitigations
| Risk | Mitigation |
|---|---|
| Spinlock causes priority inversion | For MVP, spinlock is acceptable; replace with futex-based lock if needed |
| N_SHARDS too small for high-core-count systems | Make N_SHARDS configurable at runtime (from `nproc`) |
| TSAN false positives for lock-free patterns | Use explicit acquire/release memory orders |

## Future Follow-up Tasks
- p14-t04: Per-actor heap integrates sharded free lists
- Lock-free free list (Treiber stack) as alternative to spinlock — optimization task
