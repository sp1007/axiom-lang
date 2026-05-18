# p15-t04: Isolated[T] Runtime Support

## Purpose
Implement the runtime mechanics for `Isolated[T]` — AXIOM's zero-copy ownership transfer type that enables safe data passing between actors without copying.

## Context
`Isolated[T]` guarantees that a value and all data it transitively owns has no external references. When actor A sends `Isolated[T]` to actor B, the entire subgraph of memory is transferred — segments move from A's heap to B's heap atomically. The compile-time verifier (p06-t03) ensures isolation at compile time; this task implements the runtime transfer mechanics.

## Inputs
- `ActorHeap` transfer API from p14-t04 (`ax_heap_transfer()`)
- `AxRef` generational reference from p07-t02
- Actor mailbox from p15-t05

## Outputs
- `runtime/isolated.h` — `AxIsolated` C struct
- `runtime/isolated.c` — transfer implementation
- `ax_isolated_send()`, `ax_isolated_recv()` API

## Dependencies
- p14-t04: axalloc-actor-heap — `ax_heap_transfer()`
- p06-t03: isolated-type-verification — compile-time guarantee relied upon at runtime
- p07-t02: generational-ref-runtime — `AxRef` for safe dereferencing

## Subsystems Affected
- Actor message queue (p15-t05): Isolated messages use special envelope
- Generational refs: after transfer, old AxRefs in source heap invalidated

## Detailed Requirements

```c
typedef struct AxIsolated {
    void*     root_ptr;      // pointer to root of isolated subgraph
    uint32_t  type_id;       // TypeID of T
    uint64_t  source_actor;  // actor that created this isolated value
    size_t    segment_count; // number of segments transferred
} AxIsolated;

// Wrap a value as Isolated[T] (verifier ensured it's safe)
AxIsolated ax_isolated_wrap(ActorHeap* heap, void* ptr, uint32_t type_id);

// Send Isolated[T] to target actor (transfers heap segments)
int ax_isolated_send(uint64_t target_actor_id, AxIsolated iso);

// Receive Isolated[T] from mailbox (called by handler)
void* ax_isolated_recv(AxIsolated iso, ActorHeap* my_heap);

// Invalidate all AxRefs in the source heap's transferred segments
void ax_invalidate_transferred_refs(ActorHeap* src, AxIsolated iso);
```

Transfer protocol:
1. `ax_isolated_wrap()`: snapshot which segments are reachable from `root_ptr`.
2. `ax_isolated_send()`: call `ax_heap_transfer(src, dst, root_ptr)` — moves segments.
3. Call `ax_invalidate_transferred_refs()` — bumps gen_id on all remaining refs in src heap that pointed into transferred segments.
4. Enqueue message (with AxIsolated) to target actor mailbox.
5. `ax_isolated_recv()` in target actor: root_ptr is now in target's heap — safe to use.

Zero-copy guarantee: no memcpy — only pointer/segment metadata updates.

## Implementation Steps

1. Create `runtime/isolated.h` and `runtime/isolated.c`.
2. Implement `ax_isolated_wrap()` — identify and count reachable segments.
3. Implement `ax_isolated_send()` — transfer segments, invalidate old refs.
4. Implement `ax_isolated_recv()` — unwrap in target heap context.
5. Wire into actor message queue as special message type.
6. Write correctness tests: verify old refs invalid after transfer.

## Test Plan
- `TestIsolatedWrap`: wrap a value → AxIsolated created with correct segment count
- `TestIsolatedSend`: send to actor B → segments appear in B's heap
- `TestIsolatedOldRefInvalid`: after send, old AxRef in A → deref panics
- `TestIsolatedZeroCopy`: no memcpy called during send (verified with mock)

## Validation Checklist
- [ ] Old AxRefs invalidated (gen_id bumped) in source heap after transfer
- [ ] No memcpy during transfer (only segment list manipulation)
- [ ] Type ID preserved through transfer
- [ ] Target actor receives valid pointer in its own heap

## Acceptance Criteria
- 1MB struct transferred from actor A to B in < 1µs (no copy)

## Definition of Done
- [ ] `runtime/isolated.c` implemented
- [ ] Transfer correctness and invalidation tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Segment reachability scan incomplete → dangling ptr | Use compile-time Connection Graph for precise reachability |
| Race: A reads value while transfer in progress | Transfer is atomic at segment level; A's access blocked by state check |

## Future Follow-up Tasks
- p15-t05: message queue handles AxIsolated envelope type
- Shared memory IPC: extend Isolated to cross-process transfer
