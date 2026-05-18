# p10-t11: Optimization Differential Tests

## Purpose
Verify that the optimization pipeline (O1, O2, O3) produces programs with identical runtime behavior to unoptimized (O0) programs, catching any optimization bugs that change program semantics.

## Context
Differential testing is the gold standard for optimization correctness: compile the same program with O0 (no optimization) and O2 (full optimization), run both binaries with the same inputs, compare outputs. Any divergence is a compiler bug. This is more reliable than unit testing individual passes because it catches interactions between passes.

## Inputs
- All existing `.ax` test programs
- `axc build` with `-O0` and `-O2` flags
- Runtime output of compiled binaries

## Outputs
- `tests/differential/diff_test.go` — differential test runner
- Test report: which programs diverge between O0 and O2

## Dependencies
- p10-t10: cbackend-v2-from-air — O2 uses the new AIR-based backend
- p08-t09: build-pipeline — O0 compilation path
- p08-t10: e2e-compliance-tests — test programs to differentially test

## Subsystems Affected
- Optimization correctness: primary validation mechanism
- CI: runs on every PR with new optimization changes

## Detailed Requirements

1. Differential test runner:
   ```go
   func TestDifferential(t *testing.T) {
       programs := collectTestPrograms("../../tests/")
       for _, prog := range programs {
           t.Run(prog, func(t *testing.T) {
               outO0 := compileAndRun(prog, "-O0")
               outO2 := compileAndRun(prog, "-O2")
               if outO0 != outO2 {
                   t.Errorf("divergence: O0=%q O2=%q", outO0, outO2)
               }
           })
       }
   }
   ```
2. `compileAndRun(prog, opt)`: `axc build {opt} prog.ax -o /tmp/test_bin && /tmp/test_bin`; capture stdout+stderr; return combined output.
3. Timeout: 10 seconds per program execution.
4. Test programs:
   - All compliance tests 001-100
   - All low-level tests sys_021-sys_030
   - Additional stress programs: fibonacci(40), sorting 10K elements, string manipulation
5. Also compare AIR structure: O0 vs O2 AIR should produce same instruction count per semantic block (via dump-air comparison of key blocks).
6. Report: for each divergence, print O0 binary output vs O2 binary output, and which optimization pass is likely responsible (use `--stats` to narrow down).

## Implementation Steps

1. Create `tests/differential/diff_test.go`.
2. Implement `collectTestPrograms()` — glob `tests/**/*.ax`.
3. Implement `compileAndRun()` with timeout.
4. Add comparison logic with clear divergence reporting.
5. Add to CI: `go test ./tests/differential/ -timeout 300s`.
6. Write "known-good" assertion: all existing programs pass differential test at all opt levels.

## Test Plan

- The differential tests ARE the test plan.
- Additional: `TestDifferentialStress` — programs with complex control flow and allocations.
- `TestDifferentialAllOptLevels`: test O0, O1, O2, O3 all produce same output.

## Validation Checklist

- [ ] All 100 compliance tests pass differential O0 vs O2
- [ ] All 30 low-level tests pass differential
- [ ] O3 results match O0 for all programs
- [ ] Stress programs (fibonacci, sorting) match
- [ ] CI runs differential tests on every PR touching optimization code

## Acceptance Criteria

- Zero divergences found across all test programs at O0 vs O2
- If a divergence is found, it is blocked from merging (CI failure)

## Definition of Done

- [ ] `tests/differential/diff_test.go` implemented
- [ ] All existing programs pass differential
- [ ] CI runs differential tests
- [ ] Documentation: how to investigate a divergence

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Timing-dependent programs diverge (race conditions) | Programs must be single-threaded and deterministic |
| Floating-point results differ at O3 (FP reordering) | Document O3 FP behavior; add --fast-math flag separate from O3 |

## Future Follow-up Tasks

- p11-t16: native-differential-tests (same idea, compare native vs C-backend)
- p18-t04: stage2-self-hosting-verify (ultimate differential: Go compiler vs self-hosted)
