# p16-t01: std.testing — Assert and Test Runner

## Purpose
Implement the AXIOM standard testing library (`std.testing`) with assert functions and a test runner, enabling AXIOM programs to write and run unit tests using native AXIOM syntax.

## Context
AXIOM needs a first-class testing story. `std.testing` provides `assert_eq`, `assert_ne`, `assert_true`, and a test discovery/runner so that `axc test` can find and run all `#[test]` annotated functions. This mirrors Go's `testing` package but with AXIOM syntax.

## Inputs
- AXIOM `#[test]` attribute from semantic layer (p04)
- Panic handler from p07-t03 (failed assert calls panic)
- Compiler driver `axc test` subcommand

## Outputs
- `stdlib/testing/testing.ax` — test assertion API
- `stdlib/testing/runner.ax` — test runner (discovers and runs tests)
- `axc test` driver (compiler-side test discovery)

## Dependencies
- p07-t03: panic-handler — assert failure calls ax_panic
- p04: type checker — validates `#[test]` attribute

## Subsystems Affected
- Compiler driver: `axc test` collects and runs tests
- All stdlib tests use std.testing

## Detailed Requirements

```axiom
# std/testing/testing.ax

fn assert_eq[T: Eq](actual: T, expected: T, msg: str = ""):
    if actual != expected:
        panic("assert_eq failed: expected {expected}, got {actual}. {msg}")

fn assert_ne[T: Eq](a: T, b: T, msg: str = ""):
    if a == b:
        panic("assert_ne failed: {a} == {b}. {msg}")

fn assert_true(cond: bool, msg: str = ""):
    if !cond:
        panic("assert_true failed: condition is false. {msg}")

fn assert_false(cond: bool, msg: str = ""):
    if cond:
        panic("assert_false failed: condition is true. {msg}")

fn assert_panics(f: fn() -> void, msg: str = "") -> bool:
    # Run f, return true if it panicked, false if not
    ...
```

Test runner:
```axiom
# Compiler synthesizes main() that calls:
fn run_tests(tests: []TestCase):
    var passed = 0
    var failed = 0
    for tc in tests:
        print("--- RUN {tc.name}")
        let ok = tc.run_fn()   # catches panic
        if ok:
            passed += 1
            print("--- PASS {tc.name}")
        else:
            failed += 1
            print("--- FAIL {tc.name}")
    print("Passed: {passed}, Failed: {failed}")
    if failed > 0:
        exit(1)
```

`#[test]` discovery: compiler collects all functions annotated `#[test]` and generates `run_tests([...])` call in test binary `main`.

## Implementation Steps

1. Create `stdlib/testing/testing.ax`.
2. Implement `assert_eq`, `assert_ne`, `assert_true`, `assert_false`, `assert_panics`.
3. Create `stdlib/testing/runner.ax` — `TestCase` struct + `run_tests()`.
4. Wire `axc test` command: collect `#[test]` functions, generate test `main`.
5. Write tests for the testing library itself (meta-tests).

## Test Plan
- `TestAssertEqPass`: assert_eq(1, 1) → no panic
- `TestAssertEqFail`: assert_eq(1, 2) → panic with message
- `TestAssertPanics`: assert_panics catches panic correctly
- `TestRunnerOutput`: axc test produces PASS/FAIL output in correct format

## Validation Checklist
- [ ] assert_eq message includes both actual and expected
- [ ] Test runner exit code 0 on all pass, 1 on any fail
- [ ] #[test] functions discovered without explicit registration
- [ ] assert_panics catches panic without crashing test runner

## Acceptance Criteria
- `axc test stdlib/testing/` runs all tests and reports results

## Definition of Done
- [ ] `stdlib/testing/testing.ax` implemented
- [ ] `axc test` command implemented
- [ ] Meta-tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| assert_panics implementation requires setjmp/longjmp | Use C shim: `ax_try_catch()` wrapper around panic |

## Future Follow-up Tasks
- Test coverage reporting
- Benchmark support (`#[bench]` attribute)
