# p16-t10: std.time — Time and Duration

## Purpose
Implement time measurement, duration arithmetic, and wall-clock access for AXIOM programs, enabling timing, benchmarking, and time-based logic.

## Context
`std.time` provides monotonic clock for timing (via `clock_gettime(CLOCK_MONOTONIC)`), wall clock for display (via `clock_gettime(CLOCK_REALTIME)`), and `Duration` arithmetic. It's essential for benchmarking, timeouts, and scheduling.

## Inputs
- OS time APIs: `clock_gettime` (POSIX), `QueryPerformanceCounter` (Windows)
- AXIOM arithmetic operators for Duration

## Outputs
- `stdlib/time/time.ax` — Instant, SystemTime, Duration
- `stdlib/time/timer.ax` — sleep, Ticker, Timeout

## Dependencies
- p15-t06: async-await-runtime — async sleep via future + event loop timer
- p16-t02: std-string — Duration formatting

## Detailed Requirements

```axiom
# stdlib/time/time.ax

type Duration:
    var nanos: u64    # total nanoseconds

    fn from_nanos(n: u64) -> Duration
    fn from_micros(n: u64) -> Duration
    fn from_millis(n: u64) -> Duration
    fn from_secs(n: u64) -> Duration
    fn as_nanos(self) -> u64
    fn as_micros(self) -> u64
    fn as_millis(self) -> u64
    fn as_secs(self) -> f64
    fn add(self, other: Duration) -> Duration
    fn sub(self, other: Duration) -> Duration
    fn mul(self, factor: u64) -> Duration

type Instant:  # monotonic
    var nanos_since_start: u64

    fn now() -> Instant
    fn elapsed(self) -> Duration
    fn duration_since(self, earlier: Instant) -> Duration

type SystemTime:  # wall clock, may go backward
    var secs_since_epoch: i64
    var nanos: u32

    fn now() -> SystemTime
    fn unix_timestamp(self) -> i64
    fn to_str(self) -> str  # RFC 3339 format

# stdlib/time/timer.ax
fn sleep(d: Duration)          # synchronous (blocks worker thread)
async fn sleep_async(d: Duration)  # non-blocking sleep via timer event

type Ticker:
    fn new(interval: Duration) -> Ticker
    async fn tick(mut self)    # await next tick
    fn stop(mut self)
```

Implementation of `Instant::now()`:
```c
Instant ax_instant_now(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (Instant){ .nanos = (uint64_t)ts.tv_sec * 1000000000ULL + ts.tv_nsec };
}
```

`sleep_async`: register timer in I/O event loop with `timerfd_create` (Linux) or `SetWaitableTimer` (Windows); future resolved on expiry.

## Implementation Steps

1. Create `stdlib/time/time.ax` with Duration, Instant, SystemTime.
2. Implement `Instant::now()` via clock_gettime C shim.
3. Implement `SystemTime::now()` and `to_str()` (RFC 3339).
4. Create `stdlib/time/timer.ax` — sleep + Ticker.
5. Implement `sleep_async` using timerfd + I/O event loop.
6. Write tests verifying duration accuracy.

## Test Plan
- `TestInstantElapsed`: elapsed after 10ms sleep ≥ 10ms
- `TestDurationArith`: 1s + 500ms = 1500ms
- `TestSystemTimeNow`: unix_timestamp() within 1 second of OS time
- `TestSleepAsync`: async sleep 10ms → task resumes in 10-20ms
- `TestTicker`: tick every 10ms → 10 ticks in ~100ms

## Validation Checklist
- [ ] Monotonic clock never goes backward
- [ ] Duration overflow handled (checked arithmetic or saturating)
- [ ] async sleep does not block worker thread
- [ ] SystemTime handles leap seconds gracefully (POSIX CLOCK_REALTIME)

## Acceptance Criteria
- `Instant::now().elapsed()` resolution < 100ns

## Definition of Done
- [ ] `stdlib/time/time.ax` implemented
- [ ] All tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| CLOCK_MONOTONIC not available on all targets | Fallback to CLOCK_REALTIME with warning |

## Future Follow-up Tasks
- Time zone support
- Date/time formatting (strftime wrapper)
