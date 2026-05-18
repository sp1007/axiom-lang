# p02-t05: Lexer Golden Tests

## Purpose
Write a comprehensive golden-file test suite for the AXIOM lexer that validates the token stream produced for a wide range of `.ax` source inputs. Each test case pairs a `.ax` input file with a `.tokens` expected output file. When the expected output is correct, tests pass by comparison. When the lexer changes, developers run `go test -update` to regenerate golden files and inspect the diff before committing. This creates a stable regression baseline for the lexer.

## Context
Golden tests are the most effective way to test a lexer: they capture the complete output for realistic inputs rather than cherry-picking individual tokens. The `.tokens` format is human-readable (one token per line: `KIND offset len "text"`) so diffs are easy to review. The test runner reads both files, runs the lexer on the `.ax` file, and compares the output to the `.tokens` file character by character. The `-update` flag regenerates `.tokens` from the current lexer output. Tests MUST NOT use `-update` in CI — CI always compares against committed golden files.

## Inputs
- `compiler/lexer/lexer.go` (fully implemented per p02-t02, p02-t03, p02-t04)
- `tests/lexer/` directory (empty, created in p01-t01)
- `.ax` input files (created as part of this task)

## Outputs
- `tests/lexer/*.ax` — input source files (one per test case)
- `tests/lexer/*.tokens` — expected token output (one per test case)
- `compiler/lexer/golden_test.go` — test runner using golden file comparison
- `compiler/lexer/testdata/` — symbolic link or copy of `tests/lexer/` accessible to `go test`

## Dependencies
- p02-t02: lexer-core
- p02-t03: indent-dedent-handling
- p02-t04: lexer-error-recovery

## Subsystems Affected
- `compiler/lexer/`: Golden tests live alongside this package
- `tests/lexer/`: Test data files

## Detailed Requirements

1. **`.tokens` file format** — one token per line:
   ```
   KIND offset:len "text"
   ```
   Example for `fn main():`:
   ```
   'fn' 0:2 "fn"
   identifier 3:4 "main"
   '(' 7:1 "("
   ')' 8:1 ")"
   ':' 9:1 ":"
   NEWLINE 10:1 "\n"
   EOF 11:0 ""
   ```
   - `KIND` is the result of `tok.Kind.String()`
   - `offset:len` — decimal offset and length
   - `"text"` — the raw source text of the token (empty for synthesized INDENT/DEDENT)
   - INDENT and DEDENT appear as: `INDENT 5:0 ""`

2. **Golden test runner** in `compiler/lexer/golden_test.go`:
   ```go
   package lexer_test

   import (
       "flag"
       "os"
       "path/filepath"
       "strings"
       "testing"

       "github.com/axiom-lang/axiom/compiler/lexer"
   )

   var update = flag.Bool("update", false, "update golden files")

   func TestLexerGolden(t *testing.T) {
       inputs, err := filepath.Glob("testdata/*.ax")
       if err != nil || len(inputs) == 0 {
           t.Fatal("no .ax test files found in testdata/")
       }
       for _, axFile := range inputs {
           t.Run(filepath.Base(axFile), func(t *testing.T) {
               src, err := os.ReadFile(axFile)
               if err != nil { t.Fatal(err) }

               toks, _, _ := lexer.Lex(src)
               got := formatTokens(toks, src)

               goldenFile := strings.TrimSuffix(axFile, ".ax") + ".tokens"
               if *update {
                   os.WriteFile(goldenFile, []byte(got), 0644)
                   return
               }
               want, err := os.ReadFile(goldenFile)
               if err != nil {
                   t.Fatalf("missing golden file %s; run with -update to create", goldenFile)
               }
               if got != string(want) {
                   t.Errorf("token mismatch for %s\n--- want ---\n%s\n--- got ---\n%s",
                       axFile, want, got)
               }
           })
       }
   }

   func formatTokens(toks []lexer.Token, src []byte) string {
       var sb strings.Builder
       for _, tok := range toks {
           text := ""
           if tok.Len > 0 {
               text = string(src[tok.Offset : tok.Offset+uint32(tok.Len)])
           }
           fmt.Fprintf(&sb, "%s %d:%d %q\n", tok.Kind, tok.Offset, tok.Len, text)
       }
       return sb.String()
   }
   ```

3. **Symlink `testdata/` to `tests/lexer/`**: Go's `go test` looks for test data relative to the package directory. Since test files are in `tests/lexer/`, either:
   - Create `compiler/lexer/testdata/` as an actual directory (duplicate files — bad)
   - Symlink: `compiler/lexer/testdata -> ../../tests/lexer/` (preferred on Unix)
   - On Windows: use `mklink /J` junction. Document in `docs/CONTRIBUTING.md`.
   - Alternative: set `testdata` path relative to project root using `os.Getenv("AXIOM_ROOT")` or `runtime.Caller(0)` to find the test file location.

4. **Required test cases** (`.ax` + `.tokens` pairs):

   **empty.ax** — empty file:
   ```
   (empty)
   ```
   Expected: just `EOF 0:0 ""`

   **hello_world.ax**:
   ```
   fn main():
       println("hello, world")
   ```
   Expected: `'fn'`, `identifier "main"`, `(`, `)`, `:`, `NEWLINE`, `INDENT`, `identifier "println"`, `(`, `string "hello, world"`, `)`, `NEWLINE`, `DEDENT`, `EOF`

   **all_int_literals.ax**:
   ```
   42
   0xFF
   0o777
   0b1010_1010
   1_000_000
   ```
   Expected: 5 IntLit tokens + NEWLINEs + EOF

   **all_float_literals.ax**:
   ```
   3.14
   1.0e6
   2.5e-3
   0.001
   ```
   Expected: 4 FloatLit tokens

   **all_operators.ax** — one operator per line:
   ```
   + - * / % **
   == != < > <= >=
   & | ^ ~ << >>
   = := += -= *= /= %=
   . .* , : ; ->
   ! ( ) [ ] { }
   ```
   Expected: all operator tokens correctly identified, no errors

   **string_escapes.ax**:
   ```
   "\n\t\\\""
   "\u{1F600}"
   ```
   Expected: 2 StringLit tokens, no errors

   **all_keywords.ax** — one keyword per line:
   ```
   and async await const defer elif else extern false fn
   for if import in interface let match mut nil not
   or packed pub return spawn struct true type unsafe while
   ```
   Expected: all keyword tokens (no identifiers)

   **nested_indent.ax**:
   ```
   fn outer():
       fn inner():
           let x = 1
   ```
   Expected: 2 INDENT tokens and 2 DEDENT tokens, properly nested

   **line_comments.ax**:
   ```
   // top-level comment
   fn main(): // inline comment
       // body comment
       x()
   ```
   Expected: comments produce no tokens; NEWLINE, INDENT, DEDENT properly handled

   **bad_indent.ax** — 2-space indent (error case):
   ```
   fn main():
     let x = 1
   ```
   Expected: error diagnostic E0010; partial token stream still valid

   **unknown_chars.ax** — error case:
   ```
   fn main():
       @ # $
   ```
   Expected: 3 `TokenError` tokens for `@`, `#`, `$`; other tokens (fn, main, etc.) present

   **unterminated_string.ax** — error case:
   ```
   let x = "hello
   let y = 1
   ```
   Expected: ErrUnterminatedString for first line; `let y = 1` still lexed correctly

5. **Error case `.tokens` format**: Include diagnostics in a separate section:
   ```
   TOKENS:
   'fn' 0:2 "fn"
   ...
   DIAGNOSTICS:
   E0002 3:5 "unterminated string literal"
   ```
   Or simpler: keep separate `.diags` golden files for error cases.

6. **Update procedure documentation**: Add to `docs/CONTRIBUTING.md`:
   ```
   To update lexer golden files after a valid lexer change:
   cd compiler/lexer && go test -run TestLexerGolden -update
   Review the diff: git diff tests/lexer/
   Only commit if the changes are intentional.
   ```

## Implementation Steps

1. Create `tests/lexer/` subdirectory (should already exist from p01-t01).

2. Write all `.ax` input files listed in Requirement 4.

3. Create `compiler/lexer/testdata/` directory (or symlink to `../../tests/lexer/`).

4. Write `compiler/lexer/golden_test.go` with the test runner from Requirement 2.

5. Run `go test -run TestLexerGolden -update ./compiler/lexer/` to generate initial `.tokens` files.

6. Review each generated `.tokens` file manually — verify every token is correct.

7. Correct any lexer bugs discovered during review (re-run with `-update` after fixes).

8. Commit both `.ax` files and `.tokens` files.

9. Run `go test ./compiler/lexer/` without `-update` — must pass (golden files match).

10. Add a check in the golden test to ensure no `.ax` file exists without a corresponding `.tokens` file:
    ```go
    axFiles, _ := filepath.Glob("testdata/*.ax")
    for _, f := range axFiles {
        golden := strings.TrimSuffix(f, ".ax") + ".tokens"
        if _, err := os.Stat(golden); os.IsNotExist(err) {
            t.Errorf("missing golden file for %s", f)
        }
    }
    ```

## Test Plan

The golden test framework IS the test plan. Specific verification:

- **TestLexerGolden/empty.ax**: zero tokens except EOF
- **TestLexerGolden/hello_world.ax**: verifies complete pipeline including INDENT/DEDENT
- **TestLexerGolden/all_int_literals.ax**: all 4 integer literal forms lexed
- **TestLexerGolden/all_operators.ax**: all operators, multi-char before single-char
- **TestLexerGolden/all_keywords.ax**: every keyword produces its own token kind
- **TestLexerGolden/nested_indent.ax**: 2 INDENTs, 2 DEDENTs in order
- **TestLexerGolden/line_comments.ax**: comments invisible in token stream
- **TestLexerGolden/bad_indent.ax**: diagnostics present AND partial token stream valid
- **TestLexerGolden/unknown_chars.ax**: TokenError tokens present, other tokens correct
- **TestLexerGolden/unterminated_string.ax**: error for first line, recovery for second

Additionally: `TestGoldenFilesComplete` — every `.ax` file has a `.tokens` file.

## Validation Checklist
- [ ] All 12 test case `.ax` files created
- [ ] All 12 `.tokens` golden files generated and manually verified
- [ ] `compiler/lexer/golden_test.go` compiles and runs
- [ ] `go test -run TestLexerGolden ./compiler/lexer/` passes (no `-update`)
- [ ] `-update` flag regenerates golden files correctly
- [ ] Error case golden files include diagnostic information
- [ ] `TestGoldenFilesComplete` ensures no orphan `.ax` files
- [ ] `testdata/` accessible from `go test` working directory

## Acceptance Criteria
- All 12 golden tests pass without `-update` flag after initial setup
- Intentional lexer change → golden test fails → `-update` generates correct new output → re-running passes
- Unintentional lexer regression → golden test fails with a clear diff showing what changed
- CI runs golden tests (no `-update`) and fails on regression

## Definition of Done
- [ ] All `.ax` + `.tokens` pairs committed to `tests/lexer/`
- [ ] `compiler/lexer/golden_test.go` committed
- [ ] All tests pass in CI
- [ ] Update procedure documented in `docs/CONTRIBUTING.md`
- [ ] No orphan `.ax` files (every input has a golden file)

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Golden file path resolution fails on CI | Use `runtime.Caller(0)` to find the test file's directory; derive `testdata/` path from there |
| Windows symlinks require admin rights | Use actual file copy or `os.DirFS` with a relative path from `AXIOM_ROOT` env var |
| Golden file committed with wrong content | Manual review step before commit; CI fails if golden doesn't match |
| Too many test cases slow down `go test` | Golden tests are fast (just lexing + string comparison); 12 cases = ~1ms |
| `-update` run in CI accidentally modifies golden files | CI must not pass `-update`; document clearly |

## Future Follow-up Tasks
- p02-t06: Fuzz testing uses the `.ax` files here as seed corpus
- p03-t08: Parser golden tests follow the same pattern (`.ax` + `.ast` golden files)
- Phase 5+: Add more complex `.ax` golden test cases as language features are added
