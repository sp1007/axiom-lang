# p06-t08: Ownership Checker Fuzz Target

## Purpose
Fuzz the ownership analysis, escape analysis, and CTGC passes to ensure they never panic on any valid or invalid input, and to find edge cases where ownership rules are incorrectly applied.

## Context
The ownership subsystem is complex — it involves multiple interacting passes (ConnectionGraph, ownership rules, escape analysis, CTGC). Fuzz testing both ensures robustness (no panics) and finds correctness bugs (use-after-free accepted, valid move rejected, etc.). The fuzz target runs the full semantic pipeline on random programs.

## Inputs
- Random byte sequences (fuzz corpus)
- Existing ownership test programs as seed corpus
- Full semantic pipeline (Phases 04-06)

## Outputs
- `compiler/sema/ownership_fuzz_test.go` — fuzz target
- Any found crash inputs stored in `testdata/fuzz/FuzzOwnership/`

## Dependencies
- p06-t05: ctgc-destroy-injection — last ownership pass
- p06-t07: arena-block-handling — exercise arena paths
- p03-t09: parser-fuzz — reuse the fuzzing infrastructure

## Subsystems Affected
- Ownership analysis: all code paths exercised
- Safety: ensures no panic on adversarial input

## Detailed Requirements

1. Fuzz target:
   ```go
   func FuzzOwnershipChecker(f *testing.F) {
       // Load seed corpus from ownership test files
       f.Add([]byte(`fn main(): let x = Foo{}; return x`))
       f.Add([]byte(`fn main(): mut x = 5; x = 10`))
       // ... more seeds ...
       f.Fuzz(func(t *testing.T, src []byte) {
           defer func() {
               if r := recover(); r != nil {
                   t.Fatalf("ownership checker panic: %v\n%s", r, debug.Stack())
               }
           }()
           runFullSemaPipeline(src) // no panic expected
       })
   }
   ```
2. `runFullSemaPipeline(src)`: lex → parse → resolve → type-infer → type-check → ownership → escape → CTGC. Ignore returned diagnostics; just verify no panic.
3. Seed corpus: all existing `.ax` test files from `tests/parser/` and `tests/sema/`.
4. Correctness assertions (not just no-panic):
   - If a program contains `use_after_move`, at least one diagnostic must be present.
   - If a program has no syntax errors, ownership pass must not emit spurious errors for valid ownership patterns.
5. `TestOwnershipCorrectness`: 50 known-valid programs → 0 ownership errors each.
6. `TestOwnershipErrors`: 20 known-invalid programs → at least 1 ownership error each.

## Implementation Steps

1. Create `compiler/sema/ownership_fuzz_test.go`.
2. Load seed corpus from `tests/parser/` and `tests/sema/`.
3. Implement `runFullSemaPipeline()` helper.
4. Add panic recovery.
5. Run locally for 5 minutes: `go test -fuzz=FuzzOwnershipChecker -fuzztime=5m`.
6. Fix any found panics.
7. Add correctness assertions (with separate non-fuzz tests).

## Test Plan

- Fuzz: run for 60 seconds on CI, 5+ minutes locally
- Correctness: 50 valid programs each produce 0 ownership errors
- Correctness: 20 invalid programs each produce ≥ 1 ownership error
- Regression: all crash inputs from fuzzing added as regression test cases

## Validation Checklist

- [ ] Fuzz target compiles
- [ ] Seed corpus loaded (at minimum 10 seeds)
- [ ] Panic recovery in place
- [ ] 60-second CI fuzz run finds no panics
- [ ] Correctness assertions pass

## Acceptance Criteria

- Zero panics in 60-second fuzz run
- All 50 valid programs pass with 0 ownership errors
- All 20 invalid programs fail with ≥ 1 ownership error

## Definition of Done

- [ ] `ownership_fuzz_test.go` implemented
- [ ] Seed corpus populated
- [ ] No panics in 5-minute local fuzz run
- [ ] CI fuzz step added

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Fuzz finds real panic bugs requiring significant fixes | Fix them; this is the purpose of fuzzing |
| Slow fuzzing due to heavy pipeline | Limit pipeline to sema only (skip codegen) for fuzz speed |

## Future Follow-up Tasks

- p03-t09: parser fuzz (already done)
- p09-t04: AIR verifier provides correctness guarantees for the next pipeline stage
