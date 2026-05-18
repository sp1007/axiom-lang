# p02-t06: Lexer Fuzz Testing

## Purpose
Write a Go fuzz target for the AXIOM lexer that verifies the lexer never panics or produces undefined behavior on any arbitrary byte sequence. Fuzz testing is a critical quality gate for the lexer: it systematically explores edge cases that human-written tests miss, such as truncated multi-byte sequences, random bytes that look like partial operators, and pathological indentation patterns. A fuzz-clean lexer is a prerequisite for shipping a reliable compiler.

## Context
Go 1.18+ includes native fuzzing support via `go test -fuzz`. The fuzz engine mutates the seed corpus (real `.ax` files) using coverage-guided mutation to find inputs that trigger new code paths. The invariants checked by the fuzz target are: (1) `Lex()` never panics, (2) the last token is always `TokenEOF`, (3) `TokenError` tokens have `Len=1`, (4) INDENT count equals DEDENT count. These invariants encode the contracts defined in p02-t02 through p02-t04. Fuzz testing runs continuously in CI as a nightly job (configured in p01-t04).

## Inputs
- `compiler/lexer/lexer.go` (fully implemented per p02-t02, p02-t03, p02-t04)
- `tests/lexer/*.ax` files from p02-t05 (seed corpus)

## Outputs
- `compiler/lexer/fuzz_test.go` — fuzz target `FuzzLexer`
- `compiler/lexer/testdata/fuzz/FuzzLexer/` — seed corpus directory (required by Go fuzzing)

## Dependencies
- p02-t02: lexer-core
- p02-t03: indent-dedent-handling
- p02-t04: lexer-error-recovery
- p02-t05: lexer-golden-tests — seed corpus `.ax` files

## Subsystems Affected
- `compiler/lexer/`: Fuzz target lives in the lexer package

## Detailed Requirements

1. **Fuzz target signature**:
   ```go
   func FuzzLexer(f *testing.F) {
       // seed corpus
       // fuzz body
   }
   ```
   File: `compiler/lexer/fuzz_test.go`, `package lexer` (white-box, can access unexported fields if needed).

2. **Seed corpus** — add all `.ax` files from `tests/lexer/` as seeds:
   ```go
   func FuzzLexer(f *testing.F) {
       // Add seeds from golden test files
       seedDir := "testdata"
       entries, _ := os.ReadDir(seedDir)
       for _, e := range entries {
           if !strings.HasSuffix(e.Name(), ".ax") { continue }
           data, err := os.ReadFile(filepath.Join(seedDir, e.Name()))
           if err != nil { continue }
           f.Add(data)
       }

       // Add targeted edge-case seeds
       f.Add([]byte{})
       f.Add([]byte{0x00})
       f.Add([]byte{0xFF})
       f.Add([]byte(`"`))
       f.Add([]byte("fn main():\n    x\n"))
       f.Add([]byte("0x"))
       f.Add([]byte("0b"))
       f.Add([]byte("0o"))
       f.Add([]byte("'"))
       f.Add([]byte("\n\n\n"))
       f.Add([]byte("    ")) // leading spaces only
       f.Add(bytes.Repeat([]byte("@"), 50))

       f.Fuzz(func(t *testing.T, src []byte) {
           fuzzLexer(t, src)
       })
   }
   ```

3. **Fuzz body invariants** — `fuzzLexer` checks all invariants:
   ```go
   func fuzzLexer(t *testing.T, src []byte) {
       t.Helper()

       // Invariant 0: Lex() must not panic
       // (enforced by Go's fuzz framework automatically)

       toks, lt, diags := Lex(src)

       // Invariant 1: output is never nil
       if toks == nil {
           t.Fatal("Lex returned nil token slice")
       }

       // Invariant 2: last token is always EOF
       if len(toks) == 0 {
           t.Fatal("Lex returned empty token slice (expected at least EOF)")
       }
       last := toks[len(toks)-1]
       if last.Kind != TokenEOF {
           t.Fatalf("last token is %s (offset=%d), want EOF", last.Kind, last.Offset)
       }

       // Invariant 3: all token offsets are within source bounds
       for i, tok := range toks {
           if tok.Kind == TokenEOF { continue }
           if tok.Kind == TokenIndent || tok.Kind == TokenDedent {
               // synthesized tokens have Len=0, offset may be at NEWLINE position
               continue
           }
           end := uint32(tok.Offset) + uint32(tok.Len)
           if end > uint32(len(src)) {
               t.Fatalf("token[%d] %s offset+len=%d exceeds source len=%d",
                   i, tok.Kind, end, len(src))
           }
       }

       // Invariant 4: TokenError tokens have Len=1
       for i, tok := range toks {
           if tok.Kind == TokenError && tok.Len != 1 {
               t.Fatalf("token[%d] TokenError.Len=%d, want 1", i, tok.Len)
           }
       }

       // Invariant 5: INDENT count == DEDENT count
       indents, dedents := 0, 0
       for _, tok := range toks {
           if tok.Kind == TokenIndent { indents++ }
           if tok.Kind == TokenDedent { dedents++ }
       }
       if indents != dedents {
           t.Fatalf("INDENT count (%d) != DEDENT count (%d)\nsrc: %q",
               indents, dedents, src)
       }

       // Invariant 6: LineTable is consistent (LineCol never out of bounds)
       if lt != nil && len(src) > 0 {
           // spot-check a few offsets
           lt.LineCol(0)
           lt.LineCol(uint32(len(src) - 1))
           if len(src) > 2 {
               lt.LineCol(uint32(len(src) / 2))
           }
       }

       // Invariant 7: diagnostic count is bounded
       if len(diags) > maxErrors+1 {
           t.Fatalf("diagnostic count %d exceeds maxErrors+1 (%d)", len(diags), maxErrors+1)
       }
   }
   ```

4. **Seed corpus directory** for Go's fuzz framework at `compiler/lexer/testdata/fuzz/FuzzLexer/`. Each file in this directory is a separate seed. The `f.Add()` calls above populate these automatically when you run `go test -fuzz=FuzzLexer` for the first time, but we also manually pre-populate them:
   ```
   compiler/lexer/testdata/fuzz/FuzzLexer/
       seed-empty          (empty file)
       seed-hello          ("fn main():\n    println(\"hello\")\n")
       seed-null-byte      (\x00)
       seed-all-operators  (+ - * / % ** == ...)
   ```
   File format: raw bytes (no header). Go's fuzz engine reads them directly.

5. **Running the fuzz target**:
   ```bash
   # Short run (developer check):
   go test -fuzz=FuzzLexer -fuzztime=30s ./compiler/lexer/

   # Long run (CI nightly):
   go test -fuzz=FuzzLexer -fuzztime=5m ./compiler/lexer/

   # Replay a specific crash:
   go test -run=FuzzLexer/CRASH_FILE ./compiler/lexer/
   ```

6. **Crash reproduction**: When the fuzz engine finds a crash, it saves the input to `compiler/lexer/testdata/fuzz/FuzzLexer/CRASH_XXXXXXXXX`. To reproduce:
   ```bash
   go test -run=FuzzLexer/CRASH_XXXXXXXXX ./compiler/lexer/
   ```
   Add the crashing input as a permanent seed so the regression is always tested:
   ```bash
   cp compiler/lexer/testdata/fuzz/FuzzLexer/CRASH_XXXXXXXXX \
      compiler/lexer/testdata/fuzz/FuzzLexer/regression-YYYYMMDD-description
   ```

7. **`go test` (non-fuzz) runs the seed corpus only**: When run without `-fuzz`, `go test -run=FuzzLexer` exercises all seeds as regular tests. This means the seed corpus is always tested in CI even when the full fuzz campaign is not running.

8. **Export `maxErrors`** from `lexer.go` so the fuzz test can reference it (or define it as an exported constant):
   ```go
   // MaxErrors is the maximum number of diagnostics Lex() will return.
   const MaxErrors = 100
   ```

## Implementation Steps

1. Create `compiler/lexer/testdata/fuzz/FuzzLexer/` directory.

2. Create seed files in the fuzz corpus directory:
   - `seed-empty`: empty file
   - `seed-hello`: `fn main():\n    println("hello")\n`
   - `seed-operators`: a line with all operators
   - `seed-null`: byte `\x00`
   - `seed-high-byte`: byte `\xFF`

3. Write `compiler/lexer/fuzz_test.go` with `FuzzLexer` function and `fuzzLexer` helper (see Requirements 2 and 3).

4. Export `MaxErrors` from `lexer.go` (rename from unexported `maxErrors`). Update `recovery_test.go` references.

5. Run `go test -run=FuzzLexer ./compiler/lexer/` — seed corpus exercises should all pass (no panics).

6. Run `go test -fuzz=FuzzLexer -fuzztime=30s ./compiler/lexer/` — should find no crashes.

7. If the fuzz run finds crashes, fix them and add the crashing input as a seed.

8. Verify `go test ./compiler/lexer/` still passes after all changes.

## Test Plan

The fuzz target IS the test. Additional tests:

- **TestFuzzSeedCorpus**: `go test -run=FuzzLexer` exercises all seeds as unit tests — all must pass.
- **TestFuzzLexerKnownEdgeCases**: Call `fuzzLexer(t, input)` directly for specific inputs:
  ```go
  func TestFuzzLexerKnownEdgeCases(t *testing.T) {
      cases := [][]byte{
          {},
          {0x00},
          {0xFF},
          []byte(`"`),
          []byte("fn:\n\n\n    x"),
          []byte("    "),         // leading spaces with no prior content
          []byte("\x09"),         // tab
          bytes.Repeat([]byte("@"), 101), // exceed max errors
      }
      for _, c := range cases {
          fuzzLexer(t, c)
      }
  }
  ```

## Validation Checklist
- [ ] `compiler/lexer/fuzz_test.go` compiles
- [ ] `go test -run=FuzzLexer ./compiler/lexer/` passes (seeds all pass)
- [ ] `go test -fuzz=FuzzLexer -fuzztime=30s ./compiler/lexer/` runs 30s with no crashes
- [ ] Seed corpus directory `testdata/fuzz/FuzzLexer/` has at least 5 seed files
- [ ] All 7 invariants checked in `fuzzLexer`
- [ ] `MaxErrors` exported from `lexer.go`
- [ ] CI nightly fuzz job configured (p01-t04)

## Acceptance Criteria
- Fuzz target runs without crashes on 30-second campaign with the seed corpus
- `go test -run=FuzzLexer` passes in CI (seed corpus regression test)
- Any fuzz crash found is reproducible with `go test -run=FuzzLexer/CRASH_FILE`
- Fuzz target checks all 7 invariants

## Definition of Done
- [ ] `compiler/lexer/fuzz_test.go` committed
- [ ] Seed corpus files committed
- [ ] 30-second local fuzz run finds no crashes
- [ ] `MaxErrors` exported
- [ ] CI nightly fuzz workflow (p01-t04) targets this fuzz function
- [ ] `docs/CONTRIBUTING.md` updated with fuzz run instructions

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Fuzz engine finds crash immediately on first run | Fix the bug, add crashing input as seed, rerun |
| Fuzz coverage plateaus with small seed corpus | Add more diverse seeds: unicode content, deeply nested, all escape sequences |
| Long fuzz run times block developers | Use `-fuzztime=30s` locally; full 5-minute run only in CI nightly |
| INDENT/DEDENT imbalance is a false positive | Verify invariant 5 is truly always expected; document exceptions if any |
| `testdata/fuzz/FuzzLexer/` files not committed | Must be committed; Go fuzz requires them for `go test -run=FuzzLexer` |

## Future Follow-up Tasks
- p03-t09: Parser fuzz target (`FuzzParser`) follows this same pattern
- Phase 4+: AIR builder fuzz target
- CI nightly job (p01-t04): Runs this fuzz target for 5 minutes
