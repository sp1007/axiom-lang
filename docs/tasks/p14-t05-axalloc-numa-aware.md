# p14-t05: NUMA-Aware Allocation

## Purpose
Implement NUMA (Non-Uniform Memory Access) awareness in AxAlloc so that actor heaps preferentially allocate memory from the NUMA node local to the CPU running the actor, reducing memory latency for multi-socket servers.

## Context
On multi-socket machines (common in cloud/HPC), memory access to a remote NUMA node is 2-4x slower. AXIOM actors pinned to specific CPUs should allocate from the local NUMA node. This is optional for correctness but important for performance at scale.

## Inputs
- `ActorHeap` from p14-t04 — per-actor heap with CPU affinity
- OS NUMA APIs: `libnuma` (Linux), `numa_alloc_onnode()`, `/sys/devices/system/node/`
- Actor CPU affinity from scheduler (p15)

## Outputs
- `runtime/axalloc_numa.c` — NUMA-aware segment allocator
- `ax_alloc_numa_segment(node_id)` — allocate 64KB segment from specific NUMA node

## Dependencies
- p14-t02: axalloc-segment-manager — segment allocation hook
- p14-t04: axalloc-actor-heap — per-actor heap
- p15-t02: scheduler — actor-to-CPU mapping (for NUMA node lookup)

## Subsystems Affected
- Segment manager: when allocating new 64KB segment, query current thread's NUMA node
- Actor scheduler: NUMA-aware actor placement hints

## Detailed Requirements

```c
// NUMA node detection
int ax_numa_node_count(void);          // number of NUMA nodes (1 if no NUMA)
int ax_current_numa_node(void);        // NUMA node of current thread's CPU
int ax_cpu_to_numa_node(int cpu_id);   // lookup table

// NUMA-aware segment allocation
AxSegment* ax_alloc_segment_on_node(size_t size, int numa_node);
void ax_free_segment_on_node(AxSegment* seg, int numa_node);

// Hint for actor placement
typedef struct NUMAHint {
    int preferred_node;  // -1 = no preference
    int fallback_node;
} NUMAHint;
```

Platform detection:
- Linux: `numa_available()` from libnuma; fallback to `mmap(MAP_ANONYMOUS)` if no NUMA.
- Windows: `GetNumaProcessorNodeEx()`.
- macOS: no NUMA (single node, no-op).

Compile-time guard: `#ifdef AX_NUMA_SUPPORT` — disable if libnuma not found.

Fallback: if `ax_numa_node_count() == 1`, behave identically to non-NUMA path.

## Implementation Steps

1. Create `runtime/axalloc_numa.c`.
2. Implement `ax_numa_node_count()` — detect NUMA via libnuma or return 1.
3. Implement `ax_current_numa_node()` — getcpu() + cpuset → NUMA node.
4. Implement `ax_alloc_segment_on_node()` — `numa_alloc_onnode()` or fallback to mmap.
5. Wire into segment manager: when allocating new segment, call `ax_alloc_segment_on_node(ax_current_numa_node())`.
6. Build with `#ifdef AX_NUMA_SUPPORT`; skip if libnuma absent.
7. Benchmark: NUMA-local vs NUMA-remote allocation latency.

## Test Plan
- `TestNUMANodeCount`: returns ≥ 1
- `TestNUMASegmentAllocation`: segment allocated on specified node
- `TestNUMAFallback`: single-node system → no performance difference
- `TestNUMABenchmark`: NUMA-local alloc latency vs NUMA-remote on 2-socket system

## Validation Checklist
- [ ] Graceful fallback when libnuma not available
- [ ] No NUMA calls on single-node systems (ax_numa_node_count == 1)
- [ ] Segment NUMA node matches requesting thread's NUMA node

## Acceptance Criteria
- 20%+ lower memory latency on 2-socket system (measured via benchmark)

## Definition of Done
- [ ] `runtime/axalloc_numa.c` implemented
- [ ] Builds with and without libnuma
- [ ] Benchmark shows improvement on NUMA hardware

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| libnuma not available on all Linux distros | Compile-time detection; fallback to regular mmap |
| NUMA node changes as thread migrates CPUs | Re-check NUMA node per segment allocation, not per object |

## Future Follow-up Tasks
- NUMA-aware work-stealing scheduler (prefer stealing from same NUMA node)
