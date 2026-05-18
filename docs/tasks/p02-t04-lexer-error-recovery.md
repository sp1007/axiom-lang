# p02-t04: Lexer Error Recovery

## Purpose
Implement robust error recovery in the AXIOM lexer so that it continues lexing after encountering invalid input and collects all errors in a single pass. Rather than stopping at the first bad character, the lexer emits a `TokenError` token for each problematic byte or sequence, records a `Diagnostic`, and advances past the bad input to continue. This enables the parser and IDE tooling to receive a partial token stream and report all errors in a file simultaneously, rather than forcing the user to fix one error at a time.

## Context
Lexer error recovery is a product feature: users expect to see all their errors at once, not one per compile. The recovery strategy is "skip and continue" — the simplest approach that keeps downstream passes working. Each `TokenError` token is a placeholder that the parser can handle gracefully (producing an error AST node rather than crashing). The error types handled here are: unknown/unexpected bytes, malformed numeric literals, unterminated string/char literals, invalid escape sequences, tab characters in indentation. This task refines and formalizes what was partially described in p02-t02 (the `emitError` helper) into a complete, well-tested system.

## Inputs
- `compiler/lexer/lexer.go` from p02-t02 — the `emitError()` helper and `TokenError` handling
- `compiler/lexer/token_kind.go` from p02-t01 — `TokenError` constant
- `compiler/diagnostics/diagnostics.go` from p01-t01
- `compiler/lexer/indent.go` from p02-t03 — indent errors

## Outputs
- Updates to `compiler/lexer/lexer.go` — formalized `emitError()`, error context tracking
- `compiler/lexer/errors.go` — error code constants and diagnostic constructors
- `compiler/lexer/recovery_test.go` — tests for error recovery behavior

## Dependencies
- p02-t02: lexer-core — lexer implementation to extend
- p02-t03: indent-dedent-handling — indent errors integrated here
- p01-t01: repository-bootstrap — diagnostics package

## Subsystems Affected
- `compiler/lexer/`: All error paths formalized
- `compiler/parser/`: Parser must handle `TokenError` in the token stream (addressed in p03-t07)
- `compiler/diagnostics/`: Error codes defined here

## Detailed Requirements

1. **Error recovery principle**: The lexer MUST NOT panic or return early on bad input. Every call to `Lex()` must return a valid `[]Token` slice with `TokenEOF` as the last element, regardless of how malformed the input is.

2. **`emitError()` formalization**:
   ```go
   // emitError records a diagnostic and emits a TokenError token at the current position.
   // After calling emitError, the caller must advance pos by at least 1 byte.
   func (l *lexer) emitError(code uint32, msg string) {
       pos := uint32(l.pos)
       line, col := l.lt.LineCol(pos)
       l.diags = append(l.diags, diagnostics.Diagnostic{
           Severity: diagnostics.SeverityError,
           Code:     code,
           Pos:      diagnostics.Pos{Offset: pos, Line: line, Col: col},
           Message:  msg,
       })
       l.emit(TokenError, pos, 1) // always emit exactly 1 byte of ErrorToken
   }
   ```

3. **Error codes** in `compiler/lexer/errors.go`:
   ```go
   package lexer

   const (
       // ErrTabChar: tab character in indentation or source.
       ErrTabChar = 1
       // ErrUnexpectedChar: byte not recognized as start of any token.
       ErrUnexpectedChar = 2
       // ErrUnterminatedString: string literal not closed before end of line.
       ErrUnterminatedString = 3
       // ErrUnterminatedChar: char literal not closed.
       ErrUnterminatedChar = 4
       // ErrInvalidEscape: unrecognized escape sequence in string or char.
       ErrInvalidEscape = 5
       // ErrEmptyCharLit: char literal with no character inside.
       ErrEmptyCharLit = 6
       // ErrMultiCharLit: char literal with more than one character.
       ErrMultiCharLit = 7
       // ErrInvalidHexDigit: non-hex digit in hex literal.
       ErrInvalidHexDigit = 8
       // ErrInvalidBinDigit: non-binary digit in binary literal.
       ErrInvalidBinDigit = 9
       // ErrInvalidOctDigit: non-octal digit in octal literal.
       ErrInvalidOctDigit = 10
       // ErrBadIndentAmount: indentation not a multiple of 4 spaces.
       ErrBadIndentAmount = 10 // same range as lexer; indent.go uses this
       // ErrDedentMismatch: dedent to unknown level.
       ErrDedentMismatch = 11
       // ErrUnicodeEscapeTooLong: \u{...} with more than 6 hex digits.
       ErrUnicodeEscapeTooLong = 12
       // ErrUnicodeEscapeInvalid: \u{...} is not a valid Unicode scalar.
       ErrUnicodeEscapeInvalid = 13
   )
   ```

4. **Unterminated string recovery**: When scanning a string and hitting `\n` or EOF without a closing `"`:
   ```go
   func (l *lexer) scanString() {
       start := l.pos
       l.pos++ // consume opening "
       for l.pos < len(l.src) {
           b := l.src[l.pos]
           if b == '"' {
               l.pos++
               l.emit(TokenStringLit, uint32(start), uint16(l.pos-start))
               return
           }
           if b == '\n' || b == '\r' {
               l.emitError(ErrUnterminatedString,
                   "unterminated string literal: string must be closed on the same line")
               // emit what we have as an error token; do NOT advance past \n
               // so the newline is processed normally for indentation
               return
           }
           if b == '\\' {
               l.scanEscape()
               continue
           }
           l.pos++
       }
       // EOF inside string
       l.emitError(ErrUnterminatedString, "unterminated string literal at end of file")
   }
   ```

5. **Invalid escape recovery**: When scanning `\` followed by an unknown character:
   ```go
   func (l *lexer) scanEscape() {
       l.pos++ // consume '\'
       if l.pos >= len(l.src) {
           l.emitError(ErrInvalidEscape, "escape sequence at end of file")
           return
       }
       switch l.src[l.pos] {
       case 'n', 't', '\\', '"', '\'', 'r', '0':
           l.pos++
       case 'u':
           l.pos++
           l.scanUnicodeEscape()
       default:
           l.emitError(ErrInvalidEscape,
               fmt.Sprintf("invalid escape sequence '\\%c'", rune(l.src[l.pos])))
           l.pos++ // skip the bad byte and continue
       }
   }
   ```

6. **Char literal recovery**:
   - Empty: `''` → `ErrEmptyCharLit`
   - Multi-char: `'ab'` → emit valid first char, emit `ErrMultiCharLit`, skip remaining
   - Unterminated: `'a` at EOF → `ErrUnterminatedChar`

7. **Unexpected character recovery**: For bytes that don't start any valid token:
   ```go
   default:
       l.emitError(ErrUnexpectedChar,
           fmt.Sprintf("unexpected character %q (U+%04X)", rune(b), b))
       l.pos++ // advance exactly 1 byte; do not get stuck
   ```
   This guarantees lexer termination on any input.

8. **Maximum error limit**: To prevent diagnostic flood on pathological input, stop emitting diagnostics after 100 errors. Continue emitting `TokenError` tokens but suppress diagnostics:
   ```go
   const maxErrors = 100

   func (l *lexer) emitError(code uint32, msg string) {
       if len(l.diags) < maxErrors {
           // ... append diagnostic ...
       }
       l.emit(TokenError, uint32(l.pos), 1)
   }
   ```
   After 100 errors, append one final diagnostic: `E0099: too many errors; stopping error reporting`.

9. **Error continuation guarantee**: After any call to `emitError()`, the lexer's `l.pos` must advance by at least 1. Callers are responsible for advancing. `emitError` itself does NOT advance. Document this invariant:
   ```go
   // INVARIANT: caller must advance l.pos by at least 1 after calling emitError.
   ```

10. **Partial token stream validity**: Even with errors, the output token stream must satisfy:
    - Last token is always `TokenEOF`
    - `TokenError` tokens have `Len=1` (exactly one bad byte)
    - INDENT/DEDENT balance is maintained on a best-effort basis

## Implementation Steps

1. Create `compiler/lexer/errors.go` with all error code constants (Requirement 3).

2. Update `compiler/lexer/lexer.go`:
   - Replace the informal `emitError` helper with the formalized version (Requirement 2)
   - Add the `maxErrors` limit (Requirement 8)
   - Update `scanString()` with unterminated string recovery (Requirement 4)
   - Add `scanEscape()` method (Requirement 5)
   - Update `scanChar()` with multi-char and empty char recovery (Requirement 6)
   - Update `scanOperatorOrPunct()` default case with proper error (Requirement 7)

3. Create `compiler/lexer/recovery_test.go` with all test cases below.

4. Run `go test ./compiler/lexer/` — all tests pass including new recovery tests.

5. Verify all previous tests from p02-t02 and p02-t03 still pass (no regressions).

## Test Plan

Write `compiler/lexer/recovery_test.go` in `package lexer`:

```go
func TestLexerNeverPanics(t *testing.T) {
    // A handful of pathological inputs that must not panic
    inputs := [][]byte{
        {},
        {0x00},
        {0xFF},
        []byte(`"`),               // unterminated string
        []byte(`"unclosed`),       // unterminated string
        []byte(`'\''`),            // escape in char
        []byte(`''`),              // empty char
        []byte(`'ab'`),            // multi-char literal
        []byte(`\`),               // lone backslash
        []byte("0x"),              // incomplete hex
        []byte("0b"),              // incomplete binary
        []byte("@#$%"),            // all invalid chars
        bytes.Repeat([]byte("@"), 200), // many errors
    }
    for _, in := range inputs {
        func() {
            defer func() {
                if r := recover(); r != nil {
                    t.Errorf("Lex panicked on input %q: %v", in, r)
                }
            }()
            toks, _, _ := Lex(in)
            if len(toks) == 0 || toks[len(toks)-1].Kind != TokenEOF {
                t.Errorf("last token is not EOF for input %q", in)
            }
        }()
    }
}

func TestLexerUnterminatedString(t *testing.T) {
    _, _, diags := Lex([]byte(`"hello`))
    requireError(t, diags, ErrUnterminatedString)
}

func TestLexerUnterminatedStringNewline(t *testing.T) {
    _, _, diags := Lex([]byte("\"hello\nworld\""))
    requireError(t, diags, ErrUnterminatedString)
}

func TestLexerInvalidEscapeSequence(t *testing.T) {
    _, _, diags := Lex([]byte(`"\q"`))
    requireError(t, diags, ErrInvalidEscape)
}

func TestLexerEmptyCharLiteral(t *testing.T) {
    _, _, diags := Lex([]byte(`''`))
    requireError(t, diags, ErrEmptyCharLit)
}

func TestLexerMultiCharLiteral(t *testing.T) {
    _, _, diags := Lex([]byte(`'ab'`))
    requireError(t, diags, ErrMultiCharLit)
}

func TestLexerUnexpectedCharacter(t *testing.T) {
    _, _, diags := Lex([]byte("@"))
    requireError(t, diags, ErrUnexpectedChar)
}

func TestLexerContinuesAfterError(t *testing.T) {
    // Error in middle; tokens after the error should still be lexed
    toks, _, diags := Lex([]byte("fn @ main"))
    requireError(t, diags, ErrUnexpectedChar) // @ is bad
    // 'fn' and 'main' should still be in token stream
    kinds := tokenKinds(toks)
    requireContains(t, kinds, TokenFn)
    requireContains(t, kinds, TokenIdent) // 'main'
}

func TestLexerMaxErrorsLimit(t *testing.T) {
    // 200 '@' signs should not produce 200 diagnostics
    input := bytes.Repeat([]byte("@"), 200)
    _, _, diags := Lex(input)
    if len(diags) > maxErrors+1 {
        t.Errorf("expected at most %d diagnostics, got %d", maxErrors+1, len(diags))
    }
}

func TestLexerErrorTokenHasLenOne(t *testing.T) {
    toks, _, _ := Lex([]byte("@"))
    found := false
    for _, tok := range toks {
        if tok.Kind == TokenError {
            if tok.Len != 1 {
                t.Errorf("TokenError.Len = %d, want 1", tok.Len)
            }
            found = true
        }
    }
    if !found { t.Error("expected TokenError in output") }
}

func TestLexerEOFAlwaysLast(t *testing.T) {
    inputs := []string{"@", `"unclosed`, "fn @@ x"}
    for _, in := range inputs {
        toks, _, _ := Lex([]byte(in))
        last := toks[len(toks)-1]
        if last.Kind != TokenEOF {
            t.Errorf("input %q: last token is %s, want EOF", in, last.Kind)
        }
    }
}
```

## Validation Checklist
- [ ] `Lex()` never panics on any input (verified by fuzz and pathological test cases)
- [ ] Last token is always `TokenEOF`
- [ ] `TokenError` tokens have `Len=1`
- [ ] Unterminated string produces `ErrUnterminatedString` diagnostic
- [ ] Multi-line string produces `ErrUnterminatedString` diagnostic
- [ ] Invalid escape produces `ErrInvalidEscape` diagnostic
- [ ] Empty char literal produces `ErrEmptyCharLit` diagnostic
- [ ] Multi-char literal produces `ErrMultiCharLit` diagnostic
- [ ] Unknown byte produces `ErrUnexpectedChar` diagnostic
- [ ] Lexer continues after each error (subsequent tokens are correct)
- [ ] Max 100 diagnostics returned for any input
- [ ] `go test ./compiler/lexer/` passes including all p02-t02 and p02-t03 tests

## Acceptance Criteria
- `Lex([]byte{0x00})` returns `[TokenError, TokenEOF]` with 1 diagnostic, no panic
- `Lex([]byte(`"hello`))` returns `[TokenError, TokenEOF]` with `ErrUnterminatedString`
- `Lex(bytes.Repeat([]byte("@"), 200))` returns ≤ 101 diagnostics (100 + the limit notice)
- `Lex([]byte("fn @ main"))` returns tokens including `TokenFn` and `TokenIdent("main")`
- `go test -race -count=3 ./compiler/lexer/` passes (deterministic)

## Definition of Done
- [ ] `compiler/lexer/errors.go` committed with all error codes
- [ ] `compiler/lexer/lexer.go` updated with formalized error recovery
- [ ] `compiler/lexer/recovery_test.go` committed
- [ ] All tests pass (p02-t02, p02-t03, and p02-t04 tests)
- [ ] Lint passes
- [ ] Verified by fuzz testing (p02-t06) that no crashes exist

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Infinite loop if `emitError` caller forgets to advance `l.pos` | Document the invariant clearly; add a debug-build check using `internal/assert` |
| Error recovery produces so many tokens that parser runs out of memory | MaxErrors=100 limits diagnostic count; parser also has error limits (p03-t07) |
| Unterminated string that spans multiple lines confuses indent pass | Emit error and return from `scanString()` without consuming `\n`; indent pass sees the `\n` normally |
| Error code numbering conflicts with indent.go error codes | Keep all lexer error codes in `errors.go`; indent.go imports from here |

## Future Follow-up Tasks
- p02-t05: Golden tests include error recovery cases
- p02-t06: Fuzz target verifies no panics on arbitrary input
- p03-t07: Parser handles `TokenError` tokens gracefully
