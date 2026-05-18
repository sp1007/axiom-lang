# p17-t07: Performance Profiler

## Purpose
Implement a sampling-based performance profiler for AXIOM programs that produces flame graphs and function-level timing data, enabling identification of hot spots without source modification.

## Context
Profiling is essential for optimizing AXIOM programs. The profiler uses OS-level sampling (perf events or signal-based) to record stack samples at 1000Hz, symbolizes them using DWARF debug info, and produces a flame graph SVG or text report.

## Inputs
- Running AXIOM program (as subprocess)
- DWARF debug info from p11-t13
- `.axmeta` for symbol demangling from p11-t14
- OS profiling APIs: perf_event_open (Linux), DTrace (macOS), ETW (Windows)

## Outputs
- `tools/prof/prof.go` — profiler driver
- `profile.json` — sampling data (stackcollapse format)
- `flame.svg` — FlameGraph SVG
- `axc profile [program] [args...]` subcommand

## Dependencies
- p11-t13: dwarf-line-info — symbolization of PC addresses
- p12-t01: symbol-mangling — demangling profiled symbols
- p16-t10: std-time — wall-clock timing

## Detailed Requirements

```go
type Profiler struct {
    SampleHz   int    // default 1000
    Duration   time.Duration
    OutputPath string
}

type StackSample struct {
    Timestamp uint64
    Frames    []Frame
    Count     int
}

type Frame struct {
    PC       uint64
    Symbol   string  // demangled
    File     string
    Line     uint32
}

type ProfileReport struct {
    TotalSamples  int
    Duration      time.Duration
    Functions     []FunctionProfile
}

type FunctionProfile struct {
    Name         string
    SelfSamples  int
    TotalSamples int
    SelfPct      float64
    TotalPct     float64
}

func Run(program string, args []string, opts ProfileOptions) (ProfileReport, error)
func WriteFlameGraph(report ProfileReport, path string) error
func WriteJSON(report ProfileReport, path string) error
```

Sampling implementation (Linux):
1. `fork()` the AXIOM program.
2. `perf_event_open(PERF_TYPE_SOFTWARE, PERF_COUNT_SW_CPU_CLOCK)`.
3. At each overflow (1000Hz): collect stack via `perf_event` mmap buffer.
4. Symbolize each PC using DWARF lookup table from .debug_line.
5. Aggregate samples into function profiles.

Alternative (signal-based, all platforms):
- Send SIGPROF to child at 1000Hz.
- Child's signal handler writes stack trace to shared buffer.
- Profiler process reads and symbolizes.

Flame graph output: Brendan Gregg's folded stackcollapse format → FlameGraph.pl SVG.

## Implementation Steps

1. Create `tools/prof/prof.go`.
2. Implement SIGPROF-based sampling (cross-platform).
3. Implement DWARF symbolization — map PC to function:file:line.
4. Implement demangling via `axc demangle`.
5. Implement stackcollapse format output.
6. Implement FlameGraph SVG generation.
7. Add `axc profile` subcommand.

## Test Plan
- `TestProfilerBasic`: profile simple loop → top function is loop body
- `TestProfilerSymbolization`: DWARF lookup resolves PC to function name
- `TestProfilerFlamegraph`: output is valid SVG
- `TestProfilerOutput`: JSON output parseable
- `TestProfilerOverhead`: profiling overhead < 5% (timed)

## Validation Checklist
- [ ] Sampling at 1000Hz (verify with timing)
- [ ] Demangled names in output (not _AX_... mangled)
- [ ] FlameGraph SVG renderable in browser
- [ ] < 5% overhead when profiling

## Acceptance Criteria
- Profile fibonacci(40) → recursive calls clearly visible in flame graph

## Definition of Done
- [ ] `tools/prof/prof.go` implemented
- [ ] Flame graph SVG produced for test program

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| SIGPROF handler needs to be async-signal-safe | Use `write()` only in handler; symbolize in profiler process |
| DWARF lookup slow (linear scan) | Build PC → (file,line) sorted array for binary search |

## Future Follow-up Tasks
- Memory allocation profiler (track alloc site + size)
- Integrated profiling in `axc bench`
