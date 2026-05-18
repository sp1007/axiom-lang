# p15-t10: Actor Runtime Stress Tests

## Purpose
Validate the entire actor runtime (scheduler, message queues, supervisor trees, async/await, I/O loop) under high-concurrency stress conditions to detect races, deadlocks, and memory corruption before production use.

## Context
Actor systems are notoriously hard to test because bugs manifest only under specific timing conditions. Stress tests run thousands of actors with randomized message patterns, fault injection, and long durations to expose races, queue corruption, scheduler starvation, and supervisor failures.

## Inputs
- All Phase 15 runtime components
- Randomized workloads with configurable concurrency and duration
- Fault injection: random actor panics at random intervals

## Outputs
- `tests/runtime/actor_stress_test.go` — stress test suite
- `tests/runtime/actor_stress_bench_test.go` — throughput benchmarks

## Dependencies
- All p15-t01 through p15-t09 components

## Subsystems Affected
- CI: stress tests run nightly with 60-second duration
- Memory: stress tests verify no leaks under sustained load

## Detailed Requirements

Stress test categories:

**Ping-Pong:**
- N actor pairs, each sending M messages back and forth
- Verify: all M×N×2 messages delivered, no deadlock
- Target: 1M messages/sec on 4 cores

**Fan-Out:**
- 1 producer → N consumers
- Verify: each consumer receives exactly M messages (no duplication/loss)

**Fan-In:**
- N producers → 1 aggregator
- Verify: aggregator receives exactly N×M messages

**Supervisor Fault Injection:**
- Spawn supervisor with 10 children
- Randomly panic children at rate 10/sec
- Run for 60 seconds
- Verify: system remains responsive, no orphaned actors

**Async I/O Stress:**
- 1000 actors each doing 100 async reads from socketpairs
- Verify: all reads complete correctly

**Memory Stress:**
- Each actor allocates + frees 1MB over lifetime
- 10K actors created and destroyed
- Verify: no memory leak after all actors dead (check segment manager stats)

```go
func TestActorPingPong(t *testing.T)
func TestActorFanOut(t *testing.T)
func TestActorFanIn(t *testing.T)
func TestActorSupervisorFaultInjection(t *testing.T)
func TestActorAsyncIOStress(t *testing.T)
func TestActorMemoryLeak(t *testing.T)
```

Race detection: run all stress tests with `go test -race`.

Leak detection: check `ax_segment_count_live()` before and after each test; assert 0 after cleanup.

## Implementation Steps

1. Create `tests/runtime/actor_stress_test.go`.
2. Implement `PingPong` actor pair using AXIOM actor runtime directly.
3. Implement `FaultInjectActor` — panics randomly based on a `fault_rate` parameter.
4. Implement fault injection supervisor test.
5. Implement memory leak check helper using segment manager stats.
6. Add nightly CI job with `-count=1 -timeout=300s`.
7. Add `-race` flag to CI stress test run.

## Test Plan
- All test functions listed in Detailed Requirements
- Duration: 30 seconds each in CI; 300 seconds locally for release validation

## Validation Checklist
- [ ] Zero messages lost in ping-pong test
- [ ] Zero actor leaks after supervisor fault injection test
- [ ] Zero segment leaks in memory stress test
- [ ] No data races detected by Go race detector

## Acceptance Criteria
- All stress tests pass with 60-second duration, no failures

## Definition of Done
- [ ] `tests/runtime/actor_stress_test.go` implemented
- [ ] All stress tests pass without race/leak
- [ ] Nightly CI job added

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Stress tests too slow for CI | Configurable duration: 10s in CI, 60s nightly |
| Non-deterministic failures hard to reproduce | Seed PRNG with timestamp; log seed on failure for replay |

## Future Follow-up Tasks
- Distributed stress test: actors across multiple processes/machines
- Fuzzing actor message patterns with random message generators
