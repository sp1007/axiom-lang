# p04-t10: Semantic Analysis Golden Tests

## Purpose
Establish a comprehensive golden-file test suite for the semantic analysis passes (name resolution, type inference, type checking, effects system). These tests verify that valid programs produce zero errors and that invalid programs produce specific, correctly located diagnostic messages.

## Context
Unlike parser golden tests (which compare AST structure), sema golden tests compare `[]Diagnostic` output. Each test has a `.ax` source file and a `.diag` file containing expected diagnostics. The `.diag` format is one diagnostic per line: `file.ax:LINE:COL: SEVERITY: MESSAGE`. A program with zero expected errors has an empty `.diag` file.

## Inputs
- `tests/sema/` directory — test case files
- Full semantic pipeline: resolver + type inference + type checker + effects

## Outputs
- `tests/sema/*.ax` — source files (valid and invalid programs)
- `tests/sema/*.diag` — expected diagnostic output
- `compiler/sema/golden_test.go` — test runner

## Dependencies
- p04-t09: effects-system — last pass before golden tests
- p04-t06, p04-t07: type checker — main source of diagnostics
- p03-t08: parser golden tests — same pattern, reuse framework

## Subsystems Affected
- Testing infrastructure: sema golden tests run in CI on every PR
- Semantic analysis: all passes exercised

## Detailed Requirements

1. `.diag` file format:
   ```
   hello.ax:3:10: error: undefined: 'y'
   hello.ax:5:5: error: type mismatch: expected i32, found string
   ```
2. Empty `.diag` file = program is valid (0 errors expected).
3. Test runner: run full sema pipeline, collect diagnostics, sort by line/col, compare to `.diag`.
4. `--update` flag: regenerate `.diag` files.
5. Required test cases (at minimum):
   - `valid_hello.ax` — minimal valid program
   - `valid_fibonacci.ax` — recursive function
   - `valid_struct.ax` — struct with fields and methods
   - `valid_generics.ax` — generic function call
   - `valid_sum_type.ax` — match on sum type
   - `err_undefined.ax` — use of undefined variable
   - `err_type_mismatch.ax` — wrong type in assignment
   - `err_immutable.ax` — assign to immutable variable
   - `err_return_type.ax` — wrong return type
   - `err_bad_condition.ax` — non-bool condition in if
   - `err_missing_effect.ax` — unhandled raises effect
   - `err_call_args.ax` — wrong arg count in call
   - `err_no_field.ax` — field access on wrong type
   - `err_bad_cast.ax` — illegal cast
   - `err_use_after_move.ax` — use of moved value (after p06)
6. Add integration assertions: `TestNoFalsePositives` — all `valid_*.ax` files have empty `.diag`.

## Implementation Steps

1. Create `tests/sema/` directory.
2. Write all 15+ test files.
3. Implement `compiler/sema/golden_test.go` following the same pattern as `compiler/parser/golden_test.go`.
4. Run `go test ./compiler/sema/ -run TestSemaGolden -update` to generate initial `.diag` files.
5. Review all generated `.diag` files — verify error messages are clear and locations correct.
6. Commit all files.
7. Add to CI: `go test ./compiler/sema/ -run TestSemaGolden`.

## Test Plan

The test cases ARE the test plan. For each `err_*.ax` file, the golden test verifies:
- Correct error message
- Correct file:line:col location
- No extra spurious errors

## Validation Checklist

- [ ] All valid programs have empty `.diag` (zero errors)
- [ ] All error programs have matching `.diag` content
- [ ] Error locations are accurate (off-by-one line numbers are failures)
- [ ] CI runs sema golden tests on every PR
- [ ] `--update` correctly regenerates all `.diag` files

## Acceptance Criteria

- 15+ test cases all pass golden comparison
- Zero false positives on all `valid_*.ax` files
- Each error diagnostic is specific (mentions variable/type name, not generic)

## Definition of Done

- [ ] 15+ test pairs in `tests/sema/`
- [ ] Golden test runner implemented
- [ ] All tests pass: `go test ./compiler/sema/ -run TestSemaGolden`
- [ ] CI integrated

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Diagnostic message wording changes break all error tests | Use `--update` to regenerate; keep messages stable after v0.1 |
| Line number off-by-one in diagnostics | Test specifically with known line numbers |

## Future Follow-up Tasks

- p09-t12: AIR builder golden tests (same pattern)
- p10-t11: optimization differential tests
- p06-t08: ownership fuzz target (complements golden tests)
