# p08-t11: Compilation Speed Benchmark

## Purpose
Establish a baseline compilation speed benchmark for the AXIOM compiler. The benchmark measures the time to compile a synthetic 1000-line AXIOM program through the full pipeline. The target is under 100ms on a modern laptop. Results are saved as a baseline JSON file to track regressions over time.

## Context
Compilation speed is a first-class engineering metric. A slow compiler degrades the edit-compile-test loop and signals architectural inefficiencies (excessive copying, redundant traversals, poor data locality). Establishing a baseline early makes regressions detectable before they accumulate. The benchmark covers the full pipeline through GCC invocation, making it an end-to-end wall-clock measurement.

## Inputs
- `axc build` binary (from p08-t09)
- A synthetic 1000-line AXIOM program (`benchmarks/synthetic/synth1000.ax`)
- Go benchmark framework (`testing.B`)

## Outputs
- `benchmarks/synthetic/synth1000.ax` — the synthetic benchmark program
- `cmd/axc/bench_test.go` — Go benchmark test
- `docs/benchmarks/baseline.json` — saved baseline results

## Dependencies
- p08-t09 (build pipeline — must be complete and functional)

## Subsystems Affected
- Build pipeline (measured end-to-end)
- CI (baseline checked on each PR; regression alerts if >10% slower)

## Detailed Requirements

### Synthetic Program `synth1000.ax`
The program must be representative of real AXIOM code. It should include:
- 20 struct definitions with 3–6 fields each
- 40 functions: a mix of pure functions, functions with loops, functions calling other functions
- 10 recursive functions (fibonacci variants, tree traversal stubs)
- Slice creation and iteration (5 instances)
- Match expressions with 3–4 arms (5 instances)
- Nested control flow (if inside for inside while)
- FFI declarations (5 extern "C" functions)
- Total: approximately 1000 lines (±50)

The program must be valid AXIOM (type-correct) and the compiled binary must exit with code 0.

### Benchmark Structure
```go
// cmd/axc/bench_test.go
package main_test

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

func BenchmarkCompile1000LOC(b *testing.B) {
    axcPath := buildAxc(b)
    synthFile := filepath.Join("..", "..", "benchmarks", "synthetic", "synth1000.ax")
    outDir := b.TempDir()
    outBin := filepath.Join(outDir, "synth1000")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cmd := exec.Command(axcPath, "build", synthFile, "-o", outBin, "-O0")
        if out, err := cmd.CombinedOutput(); err != nil {
            b.Fatalf("compile failed: %v\n%s", err, out)
        }
    }
}

func BenchmarkCompile1000LOC_O2(b *testing.B) {
    axcPath := buildAxc(b)
    synthFile := filepath.Join("..", "..", "benchmarks", "synthetic", "synth1000.ax")
    outDir := b.TempDir()
    outBin := filepath.Join(outDir, "synth1000_O2")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cmd := exec.Command(axcPath, "build", synthFile, "-o", outBin, "-O2")
        if out, err := cmd.CombinedOutput(); err != nil {
            b.Fatalf("compile failed: %v\n%s", err, out)
        }
    }
}

func buildAxc(b *testing.B) string {
    b.Helper()
    if p := os.Getenv("AXC_PATH"); p != "" { return p }
    tmpDir := b.TempDir()
    out := filepath.Join(tmpDir, "axc")
    cmd := exec.Command("go", "build", "-o", out, "./cmd/axc")
    if err := cmd.Run(); err != nil {
        b.Fatalf("go build axc: %v", err)
    }
    return out
}
```

### Stage-Level Benchmarks
In addition to the end-to-end benchmark, add per-stage benchmarks to identify the slow stages:

```go
func BenchmarkLex1000LOC(b *testing.B) { /* lex only */ }
func BenchmarkParse1000LOC(b *testing.B) { /* parse only */ }
func BenchmarkTypeCheck1000LOC(b *testing.B) { /* typecheck only */ }
func BenchmarkCGen1000LOC(b *testing.B) { /* cgen only, no gcc */ }
func BenchmarkGCC1000LOC(b *testing.B) { /* gcc invocation only */ }
```

### Baseline JSON Format
```json
{
  "date": "2026-05-16",
  "git_commit": "abc123",
  "machine": "Linux x86_64, 8-core, 16GB RAM",
  "go_version": "go1.23",
  "gcc_version": "gcc 13.2",
  "benchmarks": {
    "BenchmarkCompile1000LOC": {
      "ns_per_op": 87500000,
      "ms_per_op": 87.5,
      "target_ms": 100,
      "pass": true
    },
    "BenchmarkCompile1000LOC_O2": {
      "ns_per_op": 210000000,
      "ms_per_op": 210,
      "target_ms": 500,
      "pass": true
    },
    "BenchmarkLex1000LOC": { "ms_per_op": 1.2 },
    "BenchmarkParse1000LOC": { "ms_per_op": 3.5 },
    "BenchmarkTypeCheck1000LOC": { "ms_per_op": 8.1 },
    "BenchmarkCGen1000LOC": { "ms_per_op": 12.4 },
    "BenchmarkGCC1000LOC": { "ms_per_op": 62.3 }
  }
}
```

### Baseline Save Script
```bash
#!/bin/bash
# scripts/save_benchmark_baseline.sh
go test -bench=BenchmarkCompile ./cmd/axc/ \
    -benchtime=5x \
    -benchmem \
    -json 2>/dev/null | \
    scripts/parse_bench.go > docs/benchmarks/baseline.json
echo "Baseline saved to docs/benchmarks/baseline.json"
```

### Performance Targets
| Benchmark | Target |
|-----------|--------|
| End-to-end 1000 LOC (`-O0`) | < 100ms |
| End-to-end 1000 LOC (`-O2`) | < 500ms (GCC takes longer) |
| Lex 1000 LOC | < 2ms |
| Parse 1000 LOC | < 5ms |
| Type check 1000 LOC | < 15ms |
| C codegen 1000 LOC | < 20ms |
| GCC 1000 LOC | < 80ms |

Note: GCC invocation time dominates the end-to-end time. The AXIOM frontend (lex+parse+typecheck+cgen) should be < 30ms total.

### Regression Detection
Add a CI step that compares the current benchmark against the baseline:
```yaml
- name: Check benchmark regression
  run: |
    go test -bench=BenchmarkCompile1000LOC ./cmd/axc/ -benchtime=3x -json \
        | scripts/check_regression.py --baseline docs/benchmarks/baseline.json --max-regression 10
```
If the new time is > 10% slower than the baseline, the CI step fails and requires a justification.

## Implementation Steps

### Step 1: Generate `benchmarks/synthetic/synth1000.ax`
Write a script `scripts/gen_synthetic.go` that generates the synthetic file programmatically. Verify it is valid AXIOM by compiling it.

### Step 2: Write `cmd/axc/bench_test.go`
Implement all benchmarks as described.

### Step 3: Run benchmarks and save baseline
```
go test -bench=. ./cmd/axc/ -benchtime=5x > bench_raw.txt
scripts/save_benchmark_baseline.sh
```

### Step 4: Commit baseline
Commit `docs/benchmarks/baseline.json` to the repository as the reference point.

### Step 5: Add CI step

## Test Plan
1. `go test -bench=BenchmarkCompile1000LOC ./cmd/axc/` runs without error
2. Reported time is under 100ms for `-O0`
3. `synth1000.ax` compiles to a binary that exits 0
4. Per-stage benchmarks all run and report plausible times
5. Baseline JSON file is generated and can be parsed

## Validation Checklist
- [ ] `synth1000.ax` is valid AXIOM (compiles without errors)
- [ ] End-to-end benchmark passes the 100ms target
- [ ] Per-stage benchmarks are individually runnable
- [ ] Baseline JSON is generated and committed
- [ ] Regression detection script works

## Acceptance Criteria
- End-to-end compilation of 1000 LOC completes in < 100ms
- Baseline JSON file exists in `docs/benchmarks/`
- All benchmarks run without errors

## Definition of Done
- `benchmarks/synthetic/synth1000.ax` exists and is valid
- `cmd/axc/bench_test.go` exists and all benchmarks run
- `docs/benchmarks/baseline.json` exists and is committed
- CI regression check is configured

## Risks & Mitigations
- **Risk**: GCC startup time dominates the benchmark, masking frontend regressions. **Mitigation**: The per-stage benchmarks (especially `BenchmarkCGen1000LOC`) isolate the AXIOM frontend. Track these separately.
- **Risk**: Benchmark noise (system load, thermal throttling) causes false regression alerts. **Mitigation**: Run benchmarks with `-benchtime=5x` and report the median. Set a 10% regression threshold, not 5%.
- **Risk**: The 100ms target is not achievable on the first implementation. **Mitigation**: The target is aspirational. If the initial baseline exceeds 100ms, document the gap and create follow-up tasks for performance work.

## Future Follow-up Tasks
- p10-t11: After Phase 10 optimizations, re-run benchmarks and compare against baseline
- p11-t15: Native backend benchmarks (should eliminate GCC startup overhead)
- Future: incremental compilation benchmarks (only recompile changed functions)
