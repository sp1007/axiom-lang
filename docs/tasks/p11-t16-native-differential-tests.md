# p11-t16: Native Backend Differential Tests

## Purpose
Validate native backend correctness by comparing program outputs between the C backend (known-correct reference) and the x86-64 native backend across a suite of AXIOM programs.

## Context
Differential testing is the gold standard for compiler backend validation. If the C backend and native backend produce identical outputs for the same AXIOM program, the native backend is almost certainly correct. This catches register allocation bugs, ABI errors, stack frame mistakes, and encoding errors that unit tests might miss.

## Inputs
- AXIOM test programs in `tests/codegen/` (from p08 golden test suite)
- C backend compiler pipeline (p08/p10)
- Native backend pipeline (p11-t15)

## Outputs
- `tests/codegen/native_diff_test.go` — differential test harness
- Test report: pass/fail per test case, diff of stdout/stderr on failure

## Dependencies
- p11-t15: native-backend-integration — end-to-end native backend
- p08-t12: axc-emit-c-flag — C backend produces runnable C
- p10-t11: opt-differential-tests — reference for differential test patterns

## Subsystems Affected
- Test infrastructure: adds new test category
- CI pipeline: native diff tests must pass before native backend is merged

## Detailed Requirements

```go
type DiffTest struct {
    Name    string
    Source  string  // AXIOM source
    Stdin   string  // optional stdin
    Args    []string
}

func RunDiffTest(t *testing.T, dt DiffTest) {
    // 1. Compile with C backend → run → capture stdout/stderr/exitcode
    // 2. Compile with native backend → run → capture stdout/stderr/exitcode
    // 3. Assert outputs identical
}
```

Test programs (minimum):
1. `arith.ax` — arithmetic: `1 + 2 * 3`, `(5 - 1) / 2`
2. `loop.ax` — for loop 0..100, sum
3. `fib.ax` — recursive fibonacci(10)
4. `string.ax` — print string literal
5. `if_else.ax` — conditional branches
6. `func_call.ax` — multiple function calls
7. `spill.ax` — function with >16 locals (forces spill)
8. `float.ax` — float arithmetic + return

For each: compile both, execute both with same stdin/args, compare stdout+stderr+exit_code.

Failure mode: print diff with line numbers, dump generated assembly for native backend.

## Implementation Steps

1. Create `tests/codegen/native_diff_test.go`.
2. Implement `compileCBackend()` — run axc → gcc → executable.
3. Implement `compileNativeBackend()` — run axc --target=x86_64-linux → ld → executable.
4. Implement `runProgram()` — exec with timeout, capture output.
5. Implement diff comparison with helpful failure messages.
6. Add test programs in `tests/codegen/programs/`.
7. Run CI gate: all diff tests must pass.

## Test Plan
- `TestDiffArith`: arithmetic program identical output
- `TestDiffFib`: fibonacci identical output
- `TestDiffSpill`: forced spill program identical output
- `TestDiffFloat`: float arithmetic identical output
- `TestDiffLargeFunc`: function with many locals → no spill failures

## Validation Checklist
- [ ] All 8+ test programs produce identical stdout
- [ ] Exit codes match between backends
- [ ] No segfaults in native backend output
- [ ] Test failure message includes assembly dump

## Acceptance Criteria
- All differential tests pass before Phase 12 work begins

## Definition of Done
- [ ] `tests/codegen/native_diff_test.go` implemented
- [ ] All 8 differential tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Float rounding differences | Use integer-only tests initially; add float tests after confirming IEEE754 compliance |
| Linking differences (C runtime vs AXIOM runtime) | Both use same libc; minimize difference |

## Future Follow-up Tasks
- Add property-based (fuzz) differential testing (generate random programs, compare outputs)
- p13: ARM64 backend gets same differential test suite
