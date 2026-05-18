# p09-t12: AIR Builder Golden Tests

## Purpose
Establish golden-file tests for the AIR builder that validate correct AIR output for representative AXIOM programs. These tests catch regressions in the AIR builder and provide documentation of expected AIR structure.

## Context
AIR golden tests follow the same pattern as parser golden tests (p03-t08) and sema golden tests (p04-t10), but compare AIR text output. Each test has a `.ax` source file and a `.air` expected output file. The `axc dump-air` command (p09-t11) produces the text to compare.

## Inputs
- `tests/air/` directory — test case files
- `axc dump-air` command — produces actual AIR output
- AIR printer (p09-t05) — text format

## Outputs
- `tests/air/*.ax` — source files
- `tests/air/*.air` — expected AIR text output
- `ir/air/golden_test.go` — test runner

## Dependencies
- p09-t11: axc-dump-air-command — produces AIR for comparison
- p09-t05: air-text-printer — must be stable before generating goldens
- p09-t04: air-verifier — all golden AIR must pass verification

## Subsystems Affected
- Testing: AIR builder regression protection
- CI: golden tests run on every PR

## Detailed Requirements

Required test cases:
1. `hello_world.ax` — minimal main function with println
2. `arithmetic.ax` — basic arithmetic expressions (+,-,*,/,%) with all types
3. `fibonacci.ax` — recursive function with if/else and return
4. `while_loop.ax` — while loop with counter, loop region markers visible
5. `for_range.ax` — for loop with range iterator
6. `struct_access.ax` — struct field read/write with OpGEP
7. `heap_alloc.ax` — function that allocates on heap → OpAlloc + OpMakeRef
8. `destroy_inject.ax` — function where CTGC injects destroy → OpFree visible
9. `alias_reuse.ax` — pattern where alias reuse applies → OpAliasReuse visible
10. `match_sum.ax` — match on Result type → tag extraction + branch chain
11. `generics_mono.ax` — generic function after monomorphization
12. `async_basic.ax` — async function in MVP mode (synchronous execution)
13. `ownership_move.ax` — move operation → OpMove visible
14. `arena_block.ax` — arena-allocated variables → OpArenaAlloc, no OpMakeRef

Test runner:
```go
func TestAIRGolden(t *testing.T) {
    entries, _ := os.ReadDir("../../tests/air")
    for _, e := range entries {
        if !strings.HasSuffix(e.Name(), ".ax") { continue }
        t.Run(e.Name(), func(t *testing.T) {
            actual := runDumpAIR(e.Name())
            golden := readGolden(e.Name() + ".air")
            if actual != golden { t.Fail() }
        })
    }
}
```

## Implementation Steps

1. Create `tests/air/` directory.
2. Write all 14 source files.
3. Implement `ir/air/golden_test.go` following established pattern.
4. Run `go test ./ir/air/ -run TestAIRGolden -update` to generate initial `.air` files.
5. Review all `.air` files manually — verify AIR structure is correct.
6. Verify AIR verifier passes on all golden AIR.
7. Commit all files.

## Test Plan

- Golden comparison: 14 test files match expected AIR
- Verification: all golden AIR passes `air.Verify()`
- Regression: all previous golden tests (parser, sema) still pass

## Validation Checklist

- [ ] All 14 test pairs created
- [ ] All golden AIR passes verifier
- [ ] Loop regions correctly marked (OpLoopBegin/OpLoopEnd visible)
- [ ] Ownership operations visible in relevant golden files
- [ ] CI runs golden tests

## Acceptance Criteria

- Zero diffs on `go test ./ir/air/ -run TestAIRGolden`
- All 14 golden AIR files verified by `air.Verify()`

## Definition of Done

- [ ] 14 test pairs in `tests/air/`
- [ ] Golden test runner implemented
- [ ] All tests pass in CI (ubuntu, windows, macos)

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| AIR printer format changes invalidate all goldens | Use `--update` to regenerate; version the format |
| Platform differences in virtual register numbering | Register numbers are deterministic (monotonic counter); no platform variation |

## Future Follow-up Tasks

- p10-t11: opt-differential-tests compare O0 vs O2 AIR output
- p11-t16: native-differential-tests compare native vs C-backend output
