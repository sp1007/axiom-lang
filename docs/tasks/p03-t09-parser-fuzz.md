# p03-t09: Parser Fuzz Target

## Purpose
Write a Go fuzz target for the parser that ensures no input — however malformed — can cause a panic, infinite loop, or crash. The fuzz target feeds random byte sequences through the full lex+parse pipeline and validates that it always terminates with a result (even if full of errors).

## Context
Fuzz testing is essential for parser robustness. Real-world source files contain encoding issues, unexpected characters, and edge-case syntax. The AXIOM parser must handle all inputs gracefully, producing diagnostics rather than panics. The Go fuzzer (introduced in Go 1.18) natively supports coverage-guided fuzzing.

## Inputs
- `compiler/lexer/lexer.go` — lexer from p02-t02
- `compiler/parser/parser.go` — parser from p03-t04 through p03-t07
- `tests/parser/*.ax` — seed corpus from p03-t08

## Outputs
- `compiler/parser/fuzz_test.go` — fuzz target
- `testdata/fuzz/FuzzParser/` — seed corpus directory

## Dependencies
- p03-t07: parser-error-recovery — must handle all error cases without panic
- p03-t08: parser-golden-tests — provides seed corpus files

## Subsystems Affected
- Parser: exercises all code paths
- Lexer: also exercised as part of the pipeline

## Detailed Requirements

1. Fuzz target signature:
   ```go
   func FuzzParser(f *testing.F) {
       // Add seed corpus
       seedDir := "../../tests/parser"
       entries, _ := os.ReadDir(seedDir)
       for _, e := range entries {
           if strings.HasSuffix(e.Name(), ".ax") {
               data, _ := os.ReadFile(filepath.Join(seedDir, e.Name()))
               f.Add(data)
           }
       }
       f.Fuzz(func(t *testing.T, src []byte) {
           defer func() {
               if r := recover(); r != nil {
                   t.Fatalf("panic on input: %v", r)
               }
           }()
           tokens, lineTable, _ := Lex(src)
           tree := NewTree(src)
           parser := NewParser(tokens, lineTable, tree)
           parser.ParseProgram()
           // No panic = success. Errors are expected.
       })
   }
   ```
2. The fuzz function must always return (never hang): add context with timeout.
3. Validate post-conditions: `tree.Nodes` length > 0 (root Program node always exists).
4. Run with: `go test -fuzz=FuzzParser -fuzztime=60s ./compiler/parser/`
5. CI runs fuzz for 10 seconds per PR: `go test -fuzz=FuzzParser -fuzztime=10s`.

## Implementation Steps

1. Create `compiler/parser/fuzz_test.go` with the fuzz target above.
2. Copy all `tests/parser/*.ax` files as seed corpus entries using `f.Add()`.
3. Add a 5-second timeout via `context.WithTimeout` to catch infinite loops.
4. Add panic recovery inside the fuzz function.
5. Run locally: `go test -fuzz=FuzzParser -fuzztime=300s ./compiler/parser/`
6. Any found crash: save to `testdata/fuzz/FuzzParser/` and fix the bug.
7. Add fuzz run to CI with short duration (10s).

## Test Plan

- The fuzz target itself is the test.
- Additionally: `TestFuzzParserSeeds` — runs all seed corpus entries as unit tests (fast, no fuzzing).
- `TestFuzzParserKnownCrashers` — runs any crash inputs found during fuzzing.

## Validation Checklist

- [ ] Fuzz target compiles and runs with `go test -fuzz=FuzzParser`
- [ ] All seed corpus inputs succeed (no panics)
- [ ] 5-minute local fuzz run finds no panics
- [ ] CI runs 10s fuzz as part of test suite
- [ ] Any found crash is filed as a bug and fixed before merge

## Acceptance Criteria

- Zero panics found in 60-second fuzz run on development machine
- Zero panics found in 5-minute CI fuzz run
- Fuzz target terminates within 5s for any single input (timeout guard)

## Definition of Done

- [ ] `compiler/parser/fuzz_test.go` created
- [ ] Seed corpus loaded from `tests/parser/*.ax`
- [ ] No panics in 60s local fuzz run
- [ ] CI fuzz step added to `.github/workflows/ci.yml`

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Fuzz finds real panics requiring significant parser fixes | Fix panics as they are found; fuzz is expected to find bugs |
| Fuzz runs too slowly in CI | Limit CI fuzz to 10s; longer runs done manually |

## Future Follow-up Tasks

- p02-t06: lexer fuzz (already done)
- p06-t08: ownership checker fuzz
- p04-t10: sema golden tests (not fuzz but related)
