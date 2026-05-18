# p14-t04: AxAlloc Actor-Local Heap

## Purpose
Implement per-actor isolated heap allocators so that each actor has its own AxAlloc instance, enabling actor-local bump allocation without locks and zero-copy `Isolated[T]` transfers between actors.

## Context
AXIOM's actor model requires heap isolation: actor A's allocations are invisible to actor B. When actor A sends `Isolated[T]` to actor B, ownership transfers with zero copying. This requires each actor to own its allocator instance; after transfer, the receiver allocator takes ownership of the sender's allocated segments.

## Inputs
- `AxAlloc` base implementation from p14-t01 through p14-t03
- Actor struct definition from p15-t01
- `Isolated[T]` transfer protocol from p06-t03

## Outputs
- `runtime/axalloc_actor.go` — `ActorHeap` type wrapping AxAlloc with transfer support
- `ax_heap_transfer()` C function for segment ownership transfer

## Dependencies
- p14-t01: axalloc-size-classes — base allocator
- p14-t02: axalloc-segment-manager — segment ownership model
- p14-t03: axalloc-free-list-sharding — free list integration

> **Note:** This task does NOT depend on p15-t01 (actor struct). The `ActorHeap` uses a forward-declared `uint64_t actor_id` field. The full Actor struct (p15-t01) integrates `ActorHeap` as a member — the dependency flows p14-t04 → p15-t01, not the reverse.

## Subsystems Affected
- Actor runtime: each actor initialized with its own `ActorHeap`
- `Isolated[T]` sends: segment ownership transferred between `ActorHeap`s

## Detailed Requirements

```c
typedef struct ActorHeap {
    AxAlloc    alloc;           // base allocator (owns segments)
    uint64_t   actor_id;
    size_t     bytes_allocated;
    size_t     bytes_freed;
    AxSegment* transfer_pending; // segments being transferred out
} ActorHeap;

// Initialize actor heap (called on actor spawn)
void ax_actor_heap_init(ActorHeap* heap, uint64_t actor_id);

// Allocate within this actor's heap (no locks)
void* ax_actor_alloc(ActorHeap* heap, size_t size);

// Free within this actor's heap (no locks)
void ax_actor_free(ActorHeap* heap, void* ptr);

// Transfer segments containing ptr to target heap (for Isolated[T] send)
// Returns number of segments transferred
int ax_heap_transfer(ActorHeap* src, ActorHeap* dst, void* root_ptr);

// Destroy actor heap on actor exit (free all remaining segments)
void ax_actor_heap_destroy(ActorHeap* heap);
```

Transfer algorithm for `Isolated[T]` send:
1. Identify all segments reachable from `root_ptr` via pointer scanning (conservative).
2. Detach those segments from `src->alloc`.
3. Attach them to `dst->alloc`.
4. Invalidate generation IDs of remaining pointers in transferred segments if not transferred (dangling ref protection).

Actor heap stats: track `bytes_allocated`, `bytes_freed` per actor for monitoring.

## Implementation Steps

1. Create `runtime/axalloc_actor.c` (C implementation).
2. Implement `ax_actor_heap_init()` — initialize base AxAlloc with actor_id tag.
3. Implement `ax_actor_alloc()` / `ax_actor_free()` — delegate to base AxAlloc.
4. Implement `ax_heap_transfer()` — segment reachability scan + transfer.
5. Implement `ax_actor_heap_destroy()` — free all segments on actor exit.
6. Wire into actor spawn/exit in p15.
7. Write unit tests for transfer.

## Test Plan
- `TestActorHeapIsolation`: alloc in actor A, verify not accessible from actor B heap
- `TestActorHeapTransfer`: send Isolated[T] from A to B → segments moved to B's heap
- `TestActorHeapDestroy`: actor exit → all segments freed, verified by segment manager
- `TestActorHeapNoLock`: concurrent actors allocate simultaneously without contention

## Validation Checklist
- [ ] No cross-actor pointer sharing after transfer
- [ ] Transferred segments removed from source heap
- [ ] Actor exit frees all remaining segments
- [ ] No allocator locks needed for intra-actor alloc

## Acceptance Criteria
- 1000 actors each allocating 64KB independently, no corruption

## Definition of Done
- [ ] `runtime/axalloc_actor.c` implemented
- [ ] Transfer tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Conservative pointer scan misses pointers in packed structs | Use explicit ownership graph (Connection Graph) for precise scan |
| Transfer while actor B is concurrently allocating | Transfer uses atomic segment list swap |

## Future Follow-up Tasks
- p15-t05: actor message queue uses actor heap for message allocation
- NUMA-aware actor placement (p14-t05)
