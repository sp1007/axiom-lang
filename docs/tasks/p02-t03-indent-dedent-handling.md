# p02-t03: INDENT/DEDENT Handling

## Purpose
Implement an INDENT/DEDENT token post-processing pass in `compiler/lexer/indent.go` that transforms the raw token stream (which contains `NEWLINE` tokens but no structural indentation tokens) into a stream where block boundaries are explicitly marked. This pass is what makes AXIOM an indentation-based language: the parser then treats `INDENT`/`DEDENT` pairs exactly like `{`/`}` in brace-delimited languages, with no special indentation logic in the parser itself.

## Context
After p02-t02 produces the raw token stream, indentation is implicit in the source byte positions — the lexer does not look ahead to count spaces at the start of lines. This pass makes indentation structural. It reads each `NEWLINE` token, counts the spaces on the following line, compares to an indent stack, and emits `INDENT` or `DEDENT` tokens. This design cleanly separates concerns: the lexer handles character-level recognition; the indent pass handles indentation structure; the parser handles grammar structure. Indentation rule: exactly 4 spaces per level, no tabs.

Spec reference: `03. Thiết kế parser thực tế.md` — indentation-based block structure.

## Inputs
- Raw `[]Token` slice from `lexer.Lex()` (p02-t02) — contains NEWLINE tokens, no INDENT/DEDENT
- `[]byte` source (needed to count spaces after NEWLINE)
- `*LineTable` from `lexer.Lex()` (p02-t02)
- `compiler/diagnostics/` package (p01-t01)

## Outputs
- `compiler/lexer/indent.go` — `ProcessIndentation(src []byte, tokens []Token, lt *LineTable) ([]Token, []diagnostics.Diagnostic)` function
- `compiler/lexer/indent_test.go` — comprehensive tests

## Dependencies
- p02-t02: lexer-core — raw token stream with NEWLINE tokens
- p01-t03: struct-layout-definitions — Token struct
- p02-t01: token-kind-enum — TokenIndent, TokenDedent, TokenNewline constants

## Subsystems Affected
- `compiler/lexer/`: The public `Lex()` API in `lexer.go` should call `ProcessIndentation` internally so consumers get a fully-processed token stream
- `compiler/parser/`: Receives the processed stream with INDENT/DEDENT; no indentation logic needed in parser

## Detailed Requirements

1. **Algorithm overview**:
   ```
   indentStack = [0]   // stack of current indent levels (in spaces)
   output = []Token{}
   i = 0
   while i < len(tokens):
       tok = tokens[i]
       if tok.Kind != NEWLINE:
           output.append(tok)
           i++
           continue
       // Found a NEWLINE. Count spaces at start of next non-blank, non-comment line.
       nextIndent = countLeadingSpaces(src, nextNonBlankLineStart(tokens, i+1, src))
       currentIndent = indentStack.top()
       if nextIndent > currentIndent:
           if (nextIndent - currentIndent) != 4:
               emit E0010: "indentation must increase by exactly 4 spaces"
           indentStack.push(nextIndent)
           output.append(tok)         // keep the NEWLINE
           output.append(INDENT_tok)  // synthesize INDENT
       elif nextIndent < currentIndent:
           output.append(tok)         // keep the NEWLINE
           while indentStack.top() > nextIndent:
               indentStack.pop()
               output.append(DEDENT_tok)
           if indentStack.top() != nextIndent:
               emit E0011: "dedent does not match any outer indentation level"
       else:
           output.append(tok)  // same level: just keep NEWLINE
       i++
   // At EOF: emit DEDENT for every level still on stack except base 0
   while indentStack.top() > 0:
       indentStack.pop()
       output.append(DEDENT_tok)
   ```

2. **Synthesized token construction** — INDENT and DEDENT tokens have no corresponding bytes in source. Use the position of the NEWLINE token as their position:
   ```go
   func syntheticToken(kind TokenKind, afterTok Token) Token {
       return Token{Kind: kind, Offset: afterTok.Offset, Len: 0}
   }
   ```
   `Len: 0` marks synthesized tokens. Printer/diagnostic tools must handle `Len=0` gracefully.

3. **`countLeadingSpaces(src []byte, lineStart int) int`**: Count consecutive space bytes (`0x20`) at `src[lineStart:]`. Stop at the first non-space byte. Tabs at this position are an error (reported as E0001 by the raw lexer, but recheck here too).

4. **`nextNonBlankLineStart(tokens []Token, startIdx int, src []byte) int`**: Scan forward from `startIdx` through the token stream to find the first token that is not NEWLINE, not whitespace-only-line. Skip blank lines (consecutive NEWLINE tokens) and comment-only lines. Return the byte offset of the first real token on the next non-blank line.
   - Blank line: a NEWLINE immediately following another NEWLINE (no tokens between them)
   - Comment-only line: a line where the only content between two NEWLINEs is a `//` comment — comments are not in the token stream after the raw lexer, so this is handled by checking if the source bytes between two NEWLINEs contain only spaces and `//`

5. **Indent stack implementation**:
   ```go
   type indentStack struct {
       levels []int
   }
   func (s *indentStack) top() int {
       return s.levels[len(s.levels)-1]
   }
   func (s *indentStack) push(n int) { s.levels = append(s.levels, n) }
   func (s *indentStack) pop()       { s.levels = s.levels[:len(s.levels)-1] }
   ```
   Initial state: `levels: []int{0}`. The base level 0 is never popped.

6. **Error codes**:
   - E0010: `indentation increase is not a multiple of 4 spaces (got %d, from level %d)`
   - E0011: `unindent does not match any enclosing indentation level`
   - E0012: `mixed indentation: tab character in indentation (tabs are not allowed)`

7. **EOF handling**: When the token stream ends (last token is EOF), emit DEDENT for each level on the stack above 0, then pass through the EOF token.

8. **Output slice pre-allocation**: The output may be larger than the input (INDENT/DEDENT tokens are added). Pre-allocate with `make([]Token, 0, len(tokens)+32)`.

9. **Integration with `Lex()`**: Update `compiler/lexer/lexer.go`'s `Lex()` function to call `ProcessIndentation` before returning:
   ```go
   func Lex(src []byte) ([]Token, *LineTable, []diagnostics.Diagnostic) {
       l := &lexer{ ... }
       l.run()
       processed, indentDiags := ProcessIndentation(src, l.tokens, &l.lt)
       allDiags := append(l.diags, indentDiags...)
       return processed, &l.lt, allDiags
   }
   ```

10. **NEWLINE tokens are preserved** in the output stream before INDENT and after DEDENT. This preserves positional information and makes the output stream fully round-trippable for debugging purposes.

11. **Token count constraint**: After processing, every INDENT must be matched by at least one DEDENT. Use a counter during processing to assert balance before returning (debug build only, via `internal/assert`).

## Implementation Steps

1. Create `compiler/lexer/indent.go` with `package lexer`.

2. Define `indentStack` struct with `push`, `pop`, `top` methods.

3. Implement `countLeadingSpaces(src []byte, offset int) int` — count spaces at `src[offset:]`.

4. Implement `nextNonBlankLineOffset(tokens []Token, startIdx int, src []byte) uint32` — returns the byte offset of the first non-blank, non-comment-only line after `startIdx`.

5. Implement `ProcessIndentation(src []byte, tokens []Token, lt *LineTable) ([]Token, []diagnostics.Diagnostic)`:
   - Initialize `indentStack{levels: []int{0}}`
   - Allocate output slice
   - Iterate tokens using the algorithm in Requirement 1
   - On each NEWLINE: compute next indent, compare, emit INDENT/DEDENTs
   - At EOF: emit remaining DEDENTs

6. Update `Lex()` in `lexer.go` to call `ProcessIndentation` (see Requirement 9).

7. Create `compiler/lexer/indent_test.go` with all tests below.

8. Run `go test ./compiler/lexer/` — all tests pass.

## Test Plan

Write `compiler/lexer/indent_test.go`:

```go
func TestIndentDedentBasic(t *testing.T) {
    // fn main():\n    let x = 1
    src := "fn main():\n    let x = 1\n"
    toks, _, diags := Lex([]byte(src))
    requireNoErrors(t, diags)
    kinds := tokenKinds(toks)
    // expect: fn main ( ) : NEWLINE INDENT let x = 1 NEWLINE DEDENT EOF
    requireContains(t, kinds, TokenIndent)
    requireContains(t, kinds, TokenDedent)
    requireIndentBeforeDedent(t, kinds)
}

func TestIndentNestedBlocks(t *testing.T) {
    src := "if x:\n    if y:\n        let z = 1\n"
    toks, _, diags := Lex([]byte(src))
    requireNoErrors(t, diags)
    // Should have 2 INDENTs and 2 DEDENTs
    indents := countKind(toks, TokenIndent)
    dedents := countKind(toks, TokenDedent)
    if indents != 2 { t.Errorf("expected 2 INDENTs, got %d", indents) }
    if dedents != 2 { t.Errorf("expected 2 DEDENTs, got %d", dedents) }
}

func TestIndentBlankLinesIgnored(t *testing.T) {
    src := "fn main():\n\n    let x = 1\n"
    toks, _, diags := Lex([]byte(src))
    requireNoErrors(t, diags)
    // blank line should not affect indentation
    requireContains(t, tokenKinds(toks), TokenIndent)
}

func TestIndentBadAmount(t *testing.T) {
    // 2-space indent is invalid
    src := "fn main():\n  let x = 1\n"
    _, _, diags := Lex([]byte(src))
    requireError(t, diags, 10) // E0010
}

func TestDedentMismatch(t *testing.T) {
    // Dedenting to an unknown level
    src := "if x:\n    if y:\n      let z = 1\n"
    _, _, diags := Lex([]byte(src))
    requireError(t, diags, 11) // E0011
}

func TestIndentEOFDedentsEmitted(t *testing.T) {
    src := "fn main():\n    if x:\n        y()\n"
    toks, _, diags := Lex([]byte(src))
    requireNoErrors(t, diags)
    // Both blocks must be closed at EOF
    dedents := countKind(toks, TokenDedent)
    if dedents != 2 { t.Errorf("expected 2 EOF DEDENTs, got %d", dedents) }
}

func TestIndentCommentOnlyLineIgnored(t *testing.T) {
    src := "fn main():\n    // comment\n    let x = 1\n"
    toks, _, diags := Lex([]byte(src))
    requireNoErrors(t, diags)
    // comment line should not cause additional INDENT/DEDENT
    indents := countKind(toks, TokenIndent)
    if indents != 1 { t.Errorf("expected 1 INDENT, got %d", indents) }
}

func TestIndentSynthesizedTokensHaveZeroLen(t *testing.T) {
    src := "fn main():\n    x()\n"
    toks, _, _ := Lex([]byte(src))
    for _, tok := range toks {
        if tok.Kind == TokenIndent || tok.Kind == TokenDedent {
            if tok.Len != 0 {
                t.Errorf("synthesized %s token has Len=%d, want 0", tok.Kind, tok.Len)
            }
        }
    }
}
```

## Validation Checklist
- [ ] INDENT emitted after NEWLINE when next line has 4 more spaces
- [ ] Multiple DEDENTs emitted when indentation decreases multiple levels at once
- [ ] Blank lines (NEWLINE immediately after NEWLINE) do not trigger INDENT/DEDENT changes
- [ ] Comment-only lines do not trigger INDENT/DEDENT changes
- [ ] EOF triggers DEDENT for every open block
- [ ] Synthesized INDENT/DEDENT tokens have `Len=0`
- [ ] Non-multiple-of-4 indentation increase produces E0010
- [ ] Dedent to unknown level produces E0011
- [ ] Indent stack base level 0 is never popped
- [ ] `Lex()` now calls `ProcessIndentation` — output includes INDENT/DEDENT
- [ ] `go test ./compiler/lexer/` passes all tests

## Acceptance Criteria
- Single-level block: exactly 1 INDENT + 1 DEDENT in output
- Double-nested block: exactly 2 INDENTs + 2 DEDENTs
- Every INDENT is matched by exactly one DEDENT (count equality)
- Bad indentation amount emits diagnostic with code E0010, lexing continues
- EOF always emits all remaining DEDENTs before EOF token
- `go test -race ./compiler/lexer/` passes

## Definition of Done
- [ ] `compiler/lexer/indent.go` committed
- [ ] `compiler/lexer/indent_test.go` committed with all test cases
- [ ] All tests pass including race detector
- [ ] `Lex()` updated to call `ProcessIndentation`
- [ ] Golangci-lint passes

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| `nextNonBlankLineOffset` skips too many lines | Unit test with consecutive blank lines; step through with debug print |
| Synthesized token offset confuses diagnostics | Diagnostics use the NEWLINE token's offset as the INDENT/DEDENT position; document this |
| Mismatched INDENT/DEDENT count in complex programs | Add debug-build assertion checking count balance before returning |
| Comment detection requires re-scanning source bytes | `ProcessIndentation` receives `src []byte` for this purpose; avoid full re-lex |
| Indentation in string literals affects algorithm | Raw lexer emits string content as a single token; only NEWLINEs after real newlines affect indent algorithm |

## Future Follow-up Tasks
- p02-t05: Golden tests include indentation cases
- p03-t06: Parser uses INDENT/DEDENT from this pass to parse blocks
- p03-t04: Parser never needs to count spaces — it relies on INDENT/DEDENT from here
