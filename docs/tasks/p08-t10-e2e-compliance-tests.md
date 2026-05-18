# p08-t10: End-to-End Compliance Test Suite

## Purpose
Run the AXIOM compliance test suite against the full compiler pipeline. Tests 001–030 (covering primitives, control flow, and functions) must pass. The compliance suite validates that the entire pipeline — from source text to runnable binary — produces correct output.

## Context
Unit tests validate individual compiler stages in isolation. The compliance test suite validates the entire pipeline end-to-end: an AXIOM source file is compiled with `axc build`, the resulting binary is executed, and its output (stdout, stderr, exit code) is compared against expected values. This is the primary correctness metric for the C-Backend.

The test runner is a Go test suite in `tests/e2e/` that shells out to `axc build` for each test case.

## Inputs
- `axc build` binary (from p08-t09) in PATH or at a known path
- `tests/axiom_compliance_suite/` directory with individual `.ax` test files
- `tests/axiom_compliance_suite/*.expected` files (expected stdout)
- `tests/axiom_compliance_suite/*.exit` files (expected exit code, default 0)

## Outputs
- `tests/e2e/compliance_test.go` — Go test file running all compliance tests
- `tests/axiom_compliance_suite/001-030/` — the first 30 test AXIOM programs

## Dependencies
- p08-t09 (build pipeline — the tool being tested)

## Subsystems Affected
- Testing infrastructure (compliance tests are the top-level integration test gate)
- CI (compliance tests run on every pull request)

## Detailed Requirements

### Test Directory Structure
```
tests/
  axiom_compliance_suite/
    001-int-arithmetic/
      test.ax
      expected_stdout
      expected_exit  (optional, default "0")
    002-string-literal/
      test.ax
      expected_stdout
    003-if-else/
      test.ax
      expected_stdout
    ...
    030-recursive-function/
      test.ax
      expected_stdout
  e2e/
    compliance_test.go
    helpers_test.go
```

### Test Cases 001–030 (Required Content)

**001: Integer arithmetic**
```
fn main() -> i32:
    let x: i32 = 10 + 20 * 3 - 5
    return x
```
Expected exit: 55

**002: String literal output** (requires print built-in or extern printf)
```
extern "C" fn puts(s: string) -> i32
fn main() -> i32:
    puts("Hello, AXIOM!")
    return 0
```
Expected stdout: `Hello, AXIOM!\n`

**003: If-else**
```
fn main() -> i32:
    let x: i32 = 7
    if x > 5:
        return 1
    else:
        return 0
```
Expected exit: 1

**004: While loop**
```
fn main() -> i32:
    mut i: i32 = 0
    while i < 10:
        i = i + 1
    return i
```
Expected exit: 10

**005: For-in range loop**
```
fn main() -> i32:
    mut sum: i32 = 0
    for i in 0..10:
        sum = sum + i
    return sum
```
Expected exit: 45

**006: Function call**
```
fn add(a: i32, b: i32) -> i32:
    return a + b
fn main() -> i32:
    return add(3, 4)
```
Expected exit: 7

**007: Nested function calls**
```
fn square(x: i32) -> i32: return x * x
fn sum_squares(a: i32, b: i32) -> i32: return square(a) + square(b)
fn main() -> i32: return sum_squares(3, 4)
```
Expected exit: 25

**008: Boolean logic**
```
fn main() -> i32:
    let a: bool = true
    let b: bool = false
    if a and not b: return 1
    return 0
```
Expected exit: 1

**009: Type casting**
```
fn main() -> i32:
    let x: f64 = 3.7
    let y: i32 = x as i32
    return y
```
Expected exit: 3

**010: Struct definition and access**
```
struct Point:
    x: i32
    y: i32
fn main() -> i32:
    let p = Point{x: 10, y: 20}
    return p.x + p.y
```
Expected exit: 30

**011–020**: Covers: multiple return values via struct, string operations, recursive functions (factorial), nested structs, arrays/slices (basic), variable shadowing, early return, multiple functions, defer statement.

**021–030**: Covers: match statement (basic), sum types (Option/Result), for-in over slice, mutable reference functions, nested loops with break/continue, global variables, integer overflow semantics, comparison operators all variants, bitwise operators, function pointers.

### Test Runner Implementation
```go
// tests/e2e/compliance_test.go
package e2e_test

import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

func TestCompliance(t *testing.T) {
    axcPath := findAxc(t)
    suiteDir := filepath.Join("..", "axiom_compliance_suite")

    entries, err := os.ReadDir(suiteDir)
    if err != nil { t.Fatalf("ReadDir: %v", err) }

    for _, entry := range entries {
        if !entry.IsDir() { continue }
        entry := entry  // capture
        t.Run(entry.Name(), func(t *testing.T) {
            t.Parallel()
            runCompliance(t, axcPath, filepath.Join(suiteDir, entry.Name()))
        })
    }
}

func runCompliance(t *testing.T, axcPath, dir string) {
    t.Helper()
    testFile := filepath.Join(dir, "test.ax")
    outBin := filepath.Join(t.TempDir(), "test_bin")

    // Build
    buildCmd := exec.Command(axcPath, "build", testFile, "-o", outBin)
    buildOut, err := buildCmd.CombinedOutput()
    if err != nil {
        t.Fatalf("axc build failed:\n%s", buildOut)
    }

    // Run
    runCmd := exec.Command(outBin)
    var stdout, stderr strings.Builder
    runCmd.Stdout = &stdout
    runCmd.Stderr = &stderr
    runErr := runCmd.Run()

    // Check exit code
    expectedExit := 0
    if data, err := os.ReadFile(filepath.Join(dir, "expected_exit")); err == nil {
        fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &expectedExit)
    }
    gotExit := 0
    if runErr != nil {
        if exitErr, ok := runErr.(*exec.ExitError); ok {
            gotExit = exitErr.ExitCode()
        }
    }
    if gotExit != expectedExit {
        t.Errorf("exit code: got %d, want %d\nstderr: %s", gotExit, expectedExit, stderr.String())
    }

    // Check stdout
    if data, err := os.ReadFile(filepath.Join(dir, "expected_stdout")); err == nil {
        if stdout.String() != string(data) {
            t.Errorf("stdout mismatch:\ngot:  %q\nwant: %q", stdout.String(), string(data))
        }
    }
}

func findAxc(t *testing.T) string {
    t.Helper()
    if p := os.Getenv("AXC_PATH"); p != "" { return p }
    path, err := exec.LookPath("axc")
    if err != nil { t.Fatalf("axc not found in PATH; set AXC_PATH") }
    return path
}
```

### Acceptance Threshold
- **Phase 08 target**: Tests 001–030 all pass
- Tests 031–100 are defined but allowed to fail in this phase
- A test is a pass if: exit code matches AND stdout matches (if expected_stdout file exists)

### CI Integration
```yaml
# .github/workflows/e2e.yml
- name: Build axc
  run: go build -o axc ./cmd/axc
- name: Run compliance tests
  run: AXC_PATH=./axc go test ./tests/e2e/ -run TestCompliance -timeout 120s -v
```

## Implementation Steps

### Step 1: Create `tests/axiom_compliance_suite/001-xxx/` through `030-xxx/`
Write all 30 test AXIOM source files as described above. Each directory contains:
- `test.ax` — the AXIOM source
- `expected_exit` — the expected exit code (integer)
- `expected_stdout` — the expected stdout output (if any)

### Step 2: Create `tests/e2e/compliance_test.go`
As shown above.

### Step 3: Create `tests/e2e/helpers_test.go`
Helper functions: `findAxc`, `readExpected`, temp directory management.

### Step 4: Verify all 30 tests pass locally
```
go build -o axc ./cmd/axc && AXC_PATH=./axc go test ./tests/e2e/ -run "TestCompliance/00[12][0-9]" -v
```

### Step 5: Add CI job

## Test Plan
The compliance test suite IS the test plan. Each of the 30 test cases must:
1. Compile without errors (`axc build` exits 0)
2. Produce a runnable binary
3. Binary exits with the expected exit code
4. Binary produces the expected stdout (if specified)

## Validation Checklist
- [ ] All 30 test AXIOM programs are syntactically valid
- [ ] All 30 compile with `axc build` without errors
- [ ] All 30 binaries run and produce expected exit codes
- [ ] Tests are parallelized (no shared state between test cases)
- [ ] `AXC_PATH` env variable allows CI to specify the exact binary
- [ ] Test timeout set to 120 seconds (generous for first run)

## Acceptance Criteria
- `go test ./tests/e2e/ -run TestCompliance` passes for tests 001–030
- CI job passes on Linux
- No flaky tests (deterministic output from all 30 programs)

## Definition of Done
- 30 test directories exist in `tests/axiom_compliance_suite/`
- `tests/e2e/compliance_test.go` exists and runs
- All 30 tests pass locally and in CI
- CI workflow file updated

## Risks & Mitigations
- **Risk**: Some tests require features not yet complete (e.g., `puts` via `extern "C"`). **Mitigation**: Implement `extern "C"` support (p08-t07) before attempting string output tests; use exit code–only tests for tests that don't need output.
- **Risk**: Parallel test compilation races on temp directories. **Mitigation**: Each test uses `t.TempDir()` which creates a unique directory per test.
- **Risk**: Exit code vs signal: process killed by signal returns -1, not a specific code. **Mitigation**: Check `ExitError.ExitCode() < 0` and report as "killed by signal".

## Future Follow-up Tasks
- After Phase 09: run tests 031–060 (AIR-based features)
- After Phase 10: run tests 061–080 (optimized code correctness)
- After Phase 11: run tests 081–100 (native backend features)
- Long-term: expand suite to 1000+ tests covering the full language
