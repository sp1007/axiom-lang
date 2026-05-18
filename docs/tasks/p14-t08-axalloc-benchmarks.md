# p14-t08: AxAlloc Performance Benchmarks

## Purpose
Establish baseline performance measurements for AxAlloc across all size classes and usage patterns, validate that allocator meets throughput targets, and detect regressions in CI.

## Context
AxAlloc must outperform system malloc for AXIOM's typical allocation patterns (many small short-lived objects from actors). Benchmarks quantify throughput (allocs/sec), latency percentiles (p50/p99), and fragmentation ratio, enabling informed optimization decisions.

## Inputs
- All AxAlloc components from p14-t01 through p14-t07
- Benchmark harness from Go `testing.B` or custom C microbenchmark
- Target baselines: `malloc` (glibc), `jemalloc`, `mimalloc`

## Outputs
- `benchmarks/axalloc/alloc_bench_test.go` — Go benchmark wrappers
- `benchmarks/axalloc/alloc_bench.c` — C microbenchmarks
- Benchmark report: CSV with alloc size × thread count × throughput

## Dependencies
- p14-t01 through p14-t07: all AxAlloc components

## Subsystems Affected
- CI: benchmark gate — fail if regression > 15% vs baseline
- Optimization: benchmark results guide allocator tuning

## Detailed Requirements

Benchmark matrix:
```
Sizes: 8, 16, 32, 64, 128, 256, 512, 1024, 4096, 65536 bytes
Threads: 1, 2, 4, 8, 16, 32
Patterns:
  - alloc-free pairs (deallocation-heavy)
  - alloc N, free all (arena-like)
  - mixed: 70% alloc, 30% free
  - cross-thread: alloc in thread A, free in thread B
```

Target metrics:
- Single-thread alloc: ≥ 500M allocs/sec for 8-byte objects
- P99 latency: < 100ns for 8-64 byte objects
- Fragmentation: < 10% for steady-state mixed workload
- Scale: throughput scales to at least 4 threads before lock contention

```go
func BenchmarkAxAllocSmall(b *testing.B) {
    for i := 0; i < b.N; i++ {
        p := AxAlloc(8)
        AxFree(p)
    }
}

func BenchmarkAxAllocMixed(b *testing.B) {
    // 70% alloc, 30% free pattern
}

func BenchmarkAxAllocVsMalloc(b *testing.B) {
    // Compare using b.Run sub-benchmarks
}
```

Fragmentation measurement:
```c
double ax_fragmentation_ratio(AxAlloc* alloc) {
    return (double)alloc->total_allocated_bytes /
           (double)alloc->total_segment_bytes;
}
```

CI regression gate:
- Store baseline throughput JSON in `benchmarks/axalloc/baseline.json`.
- On PR: compare new run vs baseline; fail if >15% regression.

## Implementation Steps

1. Create `benchmarks/axalloc/alloc_bench.c` — C microbenchmark with rdtsc timing.
2. Create `benchmarks/axalloc/alloc_bench_test.go` — Go wrapper calling C benchmarks.
3. Implement all benchmark patterns (pairs, arena, mixed, cross-thread).
4. Implement `ax_fragmentation_ratio()`.
5. Implement baseline comparison script (`scripts/bench_compare.py`).
6. Add CI step: run benchmarks, compare to baseline, fail on regression.
7. Run initial benchmarks, establish baseline.json.

## Test Plan
- `BenchmarkAxAllocSmall8`: 8-byte alloc throughput
- `BenchmarkAxAllocSmall64`: 64-byte alloc throughput
- `BenchmarkAxAllocLarge4096`: 4KB alloc throughput
- `BenchmarkAxAllocVsMalloc`: compare side-by-side
- `BenchmarkAxAllocParallel`: parallel allocs scale with thread count

## Validation Checklist
- [ ] 8-byte alloc ≥ 500M/sec single-thread
- [ ] P99 < 100ns for small allocs
- [ ] Fragmentation < 10% on mixed workload
- [ ] Benchmark results deterministic (< 5% variance between runs)

## Acceptance Criteria
- AxAlloc ≥ 2x faster than glibc malloc for 8-byte single-thread alloc

## Definition of Done
- [ ] All benchmark programs implemented
- [ ] baseline.json committed
- [ ] CI regression gate active

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Benchmark noise on CI machines | Run 5 iterations, take median; use isolated CI worker |
| Compiler optimizes away benchmarked allocs | Use `volatile` pointer in benchmark loop |

## Future Follow-up Tasks
- Continuous benchmark tracking with time-series graph (Grafana)
- NUMA-aware benchmark showing latency improvement on multi-socket
