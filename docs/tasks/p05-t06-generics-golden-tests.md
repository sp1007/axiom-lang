# p05-t06: Generics & Advanced Types Golden Tests

## Purpose
Validate the complete generics, sum types, interfaces, and async type annotation subsystems through golden tests and end-to-end compilation, ensuring compliance tests 041-070 all pass.

## Context
This task consolidates testing for the entire Phase 05 feature set. It follows the same golden test pattern as p03-t08 (parser) and p04-t10 (sema), but focuses specifically on generic instantiation, sum type pattern matching, structural interface satisfaction, and async function types. These tests are the acceptance gate for Phase 05.

## Inputs
- Full sema pipeline (Phases 04+05)
- `tests/axiom_compliance_suite.ax` — compliance tests 041-070
- `tests/generics/` — golden test files for this phase

## Outputs
- `tests/generics/*.ax` — golden source files
- `tests/generics/*.diag` — expected diagnostic output
- `compiler/sema/generics_test.go` — golden test runner

## Dependencies
- p05-t05: async-type-annotation — last feature in Phase 05
- p05-t03: sum-types — sum type tests
- p05-t04: structural-duck-typing — interface tests
- p04-t10: sema-golden-tests — reuse test runner framework

## Subsystems Affected
- Testing: comprehensive coverage of Phase 05 features
- CI: all generics tests run on every PR

## Detailed Requirements

Required test cases:
1. `generic_sort.ax` — generic sort function, instantiated with i32 and string
2. `generic_stack.ax` — generic Stack struct, push/pop operations
3. `generic_constrained.ax` — `fn max[T: Ord](a: T, b: T) -> T`
4. `generic_multiple_params.ax` — `fn pair[A, B](a: A, b: B) -> (A, B)`
5. `sum_type_result.ax` — `type Result[T,E] = Ok(T) | Err(E)` with match
6. `sum_type_color.ax` — unit variants, match exhaustiveness
7. `sum_type_nonexhaustive.ax` — should produce exhaustiveness error
8. `interface_basic.ax` — struct implementing Printable without explicit declaration
9. `interface_missing.ax` — struct missing required method → error
10. `async_basic.ax` — `async fn`, `await`, return type inference
11. `async_nested.ax` — async fn calling another async fn
12. `async_await_outside.ax` — await outside async fn → error
13. `valid_compliance_041_050.ax` — all interface/generics compliance tests
14. `valid_compliance_051_060.ax` — all error handling compliance tests
15. `valid_compliance_061_070.ax` — all async/actor compliance tests (synchronous MVP)

Additional validation:
- `TestGenericsCaching`: verify monomorphization cache — same instantiation used multiple times
- `TestNoCodeBloat`: verify generic function instantiated with N types → N copies (not N²)

## Implementation Steps

1. Create `tests/generics/` directory.
2. Write all 15 source files above.
3. Add golden test runner in `compiler/sema/generics_test.go`.
4. Run `go test ./compiler/sema/ -run TestGenericsGolden -update` to generate `.diag` files.
5. Review `.diag` files — verify all error messages are specific and accurate.
6. Run compliance test subset: `axc build tests/axiom_compliance_suite.ax` focusing on tests 041-070.
7. Fix any failures in Phase 05 implementation.
8. Commit all files.

## Test Plan

- Golden comparison: all 15 test files match expected diagnostics
- Compliance: tests 041-070 compile and run successfully
- Regression: all Phase 04 golden tests still pass (no regressions)

## Validation Checklist

- [ ] All 15 golden test files pass comparison
- [ ] Compliance tests 041-070 pass
- [ ] No regressions in Phase 04 tests (041 golden tests still pass)
- [ ] Monomorphization cache hit verified (log output)
- [ ] CI runs all generics tests

## Acceptance Criteria

- Zero unexpected failures in `go test ./compiler/sema/ -run TestGenericsGolden`
- Compliance tests 001-070 all pass (cumulative from previous phases)
- Error test files produce exactly the right diagnostics

## Definition of Done

- [ ] 15+ golden test pairs in `tests/generics/`
- [ ] Golden test runner added
- [ ] Compliance tests 041-070 pass
- [ ] All previous compliance tests still pass (no regression)

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Async tests fail because Phase 15 not complete | MVP: async is synchronous, so tests only check types not execution |
| Compliance tests reference stdlib not yet implemented | Only test type-checking pass, not execution in Phase 05 |

## Future Follow-up Tasks

- p06-t01: ownership analysis tests build on this foundation
- p08-t10: e2e compliance tests include these in compiled execution
