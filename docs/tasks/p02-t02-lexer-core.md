# p02-t02: Lexer Core

## Purpose
Implement the zero-copy AXIOM source lexer in `compiler/lexer/lexer.go`. The lexer converts raw UTF-8 source bytes into a flat `[]Token` slice without allocating strings — tokens store `(Offset, Len)` pairs into the original source buffer. It also builds a `LineTable` mapping byte offsets to line numbers for error reporting. The lexer is the first compiler pass and its correctness and performance set the baseline for everything downstream.

## Context
The AXIOM lexer processes UTF-8 source byte-by-byte (not rune-by-rune in most cases — only string literals and identifier bodies need Unicode handling). The zero-copy design means the `[]byte` source must remain alive for the duration of compilation — this is documented as a contract with the caller. The lexer emits a raw token stream including `NEWLINE` tokens but does NOT emit `INDENT`/`DEDENT` — that is handled by the post-processing pass (p02-t03). Error recovery is handled in p02-t04 but the core structure must support it. The output `[]Token` is pre-allocated with a capacity heuristic (`len(src)/4`) to avoid reallocation in most cases.

Spec reference: `03. Thiết kế parser thực tế.md`, `01.minimal core.md`.

## Inputs
- `compiler/lexer/token.go` (Token struct, TokenKind type) from p01-t03
- `compiler/lexer/token_kind.go` (all TokenKind constants, keywords map) from p02-t01
- `compiler/diagnostics/diagnostics.go` from p01-t01
- `docs/GRAMMAR.ebnf` — lexical terminal definitions

## Outputs
- `compiler/lexer/lexer.go` — `Lexer` struct, `Lex()` function, `LineTable`
- `compiler/lexer/lexer_test.go` — comprehensive unit tests

## Dependencies
- p01-t03: struct-layout-definitions — Token struct
- p02-t01: token-kind-enum — TokenKind constants and keywords map
- p01-t01: repository-bootstrap — diagnostics package

## Subsystems Affected
- `compiler/lexer/`: Core implementation lives here
- `compiler/parser/`: Consumes `[]Token` produced by this lexer

## Detailed Requirements

1. **Public API**:
   ```go
   // Lex tokenizes the UTF-8 source into a flat token slice.
   // src must remain alive for the lifetime of the returned tokens
   // (tokens reference into src by offset/length, no copying).
   // Returns the token slice, a line table for error reporting, and any diagnostics.
   // Lexing always completes — errors produce TokenError tokens and diagnostics,
   // but the returned slice is always non-nil.
   func Lex(src []byte) ([]Token, *LineTable, []diagnostics.Diagnostic)
   ```

2. **LineTable** — maps byte offsets to line/column:
   ```go
   // LineTable records the byte offset of each newline in the source.
   // newlineOffsets[i] is the offset of the '\n' at the end of line i+1.
   // Line numbers are 1-based; column numbers are 1-based.
   type LineTable struct {
       newlineOffsets []uint32 // sorted, one entry per \n in source
       srcLen         uint32
   }

   // LineCol returns the 1-based line and column for a byte offset.
   func (lt *LineTable) LineCol(offset uint32) (line, col uint32) {
       // binary search newlineOffsets for the largest offset <= given offset
       lo, hi := 0, len(lt.newlineOffsets)
       for lo < hi {
           mid := (lo + hi) / 2
           if lt.newlineOffsets[mid] < offset {
               lo = mid + 1
           } else {
               hi = mid
           }
       }
       line = uint32(lo) + 1
       lineStart := uint32(0)
       if lo > 0 {
           lineStart = lt.newlineOffsets[lo-1] + 1
       }
       col = offset - lineStart + 1
       return
   }
   ```

3. **Lexer state struct** (unexported):
   ```go
   type lexer struct {
       src    []byte
       pos    int            // current byte position
       tokens []Token        // output slice (pre-allocated)
       lt     LineTable      // being built
       diags  []diagnostics.Diagnostic
   }
   ```

4. **Core scanning loop**: Process bytes in a `for` loop calling `next()` to get the current byte, dispatching based on the byte value:
   ```go
   func (l *lexer) run() {
       for l.pos < len(l.src) {
           b := l.src[l.pos]
           switch {
           case b == ' ' || b == '\t' || b == '\r':
               // whitespace (not newline): skip but track tabs as error
               if b == '\t' {
                   l.emitError("tab character not allowed; use 4 spaces for indentation")
               }
               l.pos++
           case b == '\n':
               l.emitNewline()
           case b == '/' && l.peek1() == '/':
               l.scanLineComment()
           case b >= '0' && b <= '9':
               l.scanNumber()
           case b == '"':
               l.scanString()
           case b == '\'':
               l.scanChar()
           case isIdentStart(b):
               l.scanIdent()
           default:
               l.scanOperatorOrPunct()
           }
       }
       l.emit(TokenEOF, uint32(l.pos), 0)
   }
   ```

5. **`scanNumber()`** — handle all integer and float literal forms:
   - Decimal: `[0-9][0-9_]*`
   - Hex: `0x[0-9a-fA-F_]+`
   - Octal: `0o[0-7_]+`
   - Binary: `0b[01_]+`
   - Float: `[0-9][0-9]*\.[0-9][0-9]*([eE][+-]?[0-9]+)?`
   - Disambiguation: after scanning `0`, peek at next byte to determine prefix
   - Underscore separators are included in the token but validated later (semantic pass)
   - Emit `TokenIntLit` or `TokenFloatLit`

6. **`scanString()`** — handle double-quoted string literals:
   - Scan until closing `"`, processing escape sequences
   - Escape sequences: `\n`, `\t`, `\\`, `\"`, `\u{XXXXXX}`
   - Do NOT allocate a new string — the token spans the whole `"..."` including quotes
   - Multi-line strings: error — AXIOM strings must be single-line (use `\n`)
   - Emit `TokenStringLit`

7. **`scanChar()`** — handle single-quoted char literals:
   - Must contain exactly one character or one escape sequence
   - Emit `TokenCharLit`

8. **`scanIdent()`** — handle identifiers and keywords:
   ```go
   func (l *lexer) scanIdent() {
       start := l.pos
       for l.pos < len(l.src) && isIdentContinue(l.src[l.pos]) {
           l.pos++
       }
       text := l.src[start:l.pos]
       kind := TokenIdent
       if kw, ok := keywords[string(text)]; ok {
           kind = kw
       }
       l.emit(kind, uint32(start), uint16(l.pos-start))
   }

   func isIdentStart(b byte) bool {
       return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
   }

   func isIdentContinue(b byte) bool {
       return isIdentStart(b) || (b >= '0' && b <= '9')
   }
   ```
   Note: AXIOM identifiers are ASCII-only in the MVP. Unicode identifiers may be added via RFC later.

9. **`scanOperatorOrPunct()`** — scan multi-character operators before single-character ones:
   - Check 2-character operators first: `==`, `!=`, `<=`, `>=`, `**`, `<<`, `>>`, `->`, `:=`, `+=`, `-=`, `*=`, `/=`, `%=`, `.*`
   - Then 1-character operators/punctuation
   - Unknown characters: emit `TokenError` and advance 1 byte
   ```go
   func (l *lexer) scanOperatorOrPunct() {
       start := l.pos
       b := l.src[l.pos]
       if l.pos+1 < len(l.src) {
           two := string(l.src[l.pos : l.pos+2])
           if kind, ok := twoCharOps[two]; ok {
               l.pos += 2
               l.emit(kind, uint32(start), 2)
               return
           }
       }
       if kind, ok := oneCharOps[b]; ok {
           l.pos++
           l.emit(kind, uint32(start), 1)
           return
       }
       // Unknown character
       l.emitError(fmt.Sprintf("unexpected character %q (0x%02x)", rune(b), b))
       l.pos++
   }
   ```

10. **`emit()` helper**:
    ```go
    func (l *lexer) emit(kind TokenKind, offset uint32, length uint16) {
        l.tokens = append(l.tokens, Token{Kind: kind, Offset: offset, Len: length})
    }
    ```

11. **NEWLINE emission** — track newlines in LineTable AND emit a NEWLINE token:
    ```go
    func (l *lexer) emitNewline() {
        offset := uint32(l.pos)
        l.lt.newlineOffsets = append(l.lt.newlineOffsets, offset)
        l.emit(TokenNewline, offset, 1)
        l.pos++
    }
    ```

12. **Pre-allocation heuristic**: `l.tokens = make([]Token, 0, len(src)/4+16)`. The `/4` comes from the average token size being ~4 bytes in AXIOM source; the `+16` handles tiny files.

13. **`peek1()` helper** — look at the next byte without advancing:
    ```go
    func (l *lexer) peek1() byte {
        if l.pos+1 < len(l.src) {
            return l.src[l.pos+1]
        }
        return 0
    }
    ```

14. **Tab handling**: Tabs (`\t`) in indentation position must emit a diagnostic `E0001: tab character in indentation; AXIOM requires 4-space indentation`. Tabs in string literals are allowed.

15. **One-char and two-char operator maps** (package-level vars, initialized at startup):
    ```go
    var twoCharOps = map[string]TokenKind{
        "==": TokenEqEq, "!=": TokenBangEq, "<=": TokenLtEq, ">=": TokenGtEq,
        "**": TokenStarStar, "<<": TokenLtLt, ">>": TokenGtGt,
        "->": TokenArrow, ":=": TokenColonEq,
        "+=": TokenPlusEq, "-=": TokenMinusEq, "*=": TokenStarEq,
        "/=": TokenSlashEq, "%=": TokenPercentEq, ".*": TokenDotStar,
    }

    var oneCharOps = map[byte]TokenKind{
        '+': TokenPlus, '-': TokenMinus, '*': TokenStar, '/': TokenSlash,
        '%': TokenPercent, '=': TokenEq, '<': TokenLt, '>': TokenGt,
        '&': TokenAmp, '|': TokenPipe, '^': TokenCaret, '~': TokenTilde,
        '.': TokenDot, ',': TokenComma, ':': TokenColon, ';': TokenSemicolon,
        '!': TokenBang, '(': TokenLParen, ')': TokenRParen,
        '[': TokenLBracket, ']': TokenRBracket, '{': TokenLBrace, '}': TokenRBrace,
    }
    ```

## Implementation Steps

1. Create `compiler/lexer/lexer.go` with package declaration and imports.

2. Define `LineTable` struct with `newlineOffsets []uint32` and `srcLen uint32`. Implement `LineCol(offset uint32) (line, col uint32)` using binary search.

3. Define unexported `lexer` struct with `src`, `pos`, `tokens`, `lt`, `diags` fields.

4. Implement `Lex(src []byte)` as the public entry point:
   ```go
   func Lex(src []byte) ([]Token, *LineTable, []diagnostics.Diagnostic) {
       l := &lexer{
           src:    src,
           tokens: make([]Token, 0, len(src)/4+16),
           lt:     LineTable{srcLen: uint32(len(src))},
       }
       l.run()
       return l.tokens, &l.lt, l.diags
   }
   ```

5. Implement `run()` — the main dispatch loop (see Requirement 4).

6. Implement `scanNumber()` with decimal/hex/octal/binary/float disambiguation.

7. Implement `scanString()` with escape sequence processing. No allocation — token covers raw bytes including quotes.

8. Implement `scanChar()` similarly.

9. Implement `scanIdent()` with keyword lookup (see Requirement 8).

10. Implement `scanLineComment()` — advance until `\n` or EOF, emit nothing.

11. Define `twoCharOps` and `oneCharOps` maps (see Requirement 15).

12. Implement `scanOperatorOrPunct()` (see Requirement 9).

13. Implement `emit()`, `emitNewline()`, `emitError()` helpers.

14. Implement `peek1()` and `isIdentStart()`, `isIdentContinue()` helpers.

## Test Plan

Write `compiler/lexer/lexer_test.go` in `package lexer`:

```go
func TestLexerEmpty(t *testing.T) {
    toks, _, diags := Lex([]byte{})
    requireNoErrors(t, diags)
    require(t, len(toks) == 1, "expected 1 token (EOF)")
    require(t, toks[0].Kind == TokenEOF, "expected EOF")
}

func TestLexerIntLitDecimal(t *testing.T) {
    toks, _, diags := Lex([]byte("42"))
    requireNoErrors(t, diags)
    require(t, toks[0].Kind == TokenIntLit)
    require(t, toks[0].Offset == 0)
    require(t, toks[0].Len == 2)
}

func TestLexerIntLitHex(t *testing.T) {
    toks, _, diags := Lex([]byte("0xFF"))
    requireNoErrors(t, diags)
    require(t, toks[0].Kind == TokenIntLit)
    require(t, toks[0].Offset == 0)
    require(t, toks[0].Len == 4)
}

func TestLexerIntLitBinary(t *testing.T) {
    toks, _, _ := Lex([]byte("0b1010_1010"))
    require(t, toks[0].Kind == TokenIntLit)
    require(t, toks[0].Len == 11)
}

func TestLexerIntLitOctal(t *testing.T) {
    toks, _, _ := Lex([]byte("0o755"))
    require(t, toks[0].Kind == TokenIntLit)
}

func TestLexerFloatLit(t *testing.T) {
    toks, _, _ := Lex([]byte("3.14"))
    require(t, toks[0].Kind == TokenFloatLit)
}

func TestLexerFloatLitExponent(t *testing.T) {
    toks, _, _ := Lex([]byte("1.0e-6"))
    require(t, toks[0].Kind == TokenFloatLit)
}

func TestLexerStringLit(t *testing.T) {
    src := []byte(`"hello"`)
    toks, _, diags := Lex(src)
    requireNoErrors(t, diags)
    require(t, toks[0].Kind == TokenStringLit)
    require(t, toks[0].Offset == 0)
    require(t, toks[0].Len == 7)
}

func TestLexerStringEscapes(t *testing.T) {
    toks, _, diags := Lex([]byte(`"\n\t\\\""`))
    requireNoErrors(t, diags)
    require(t, toks[0].Kind == TokenStringLit)
}

func TestLexerAllKeywords(t *testing.T) {
    for text, expectedKind := range keywords {
        toks, _, _ := Lex([]byte(text))
        if toks[0].Kind != expectedKind {
            t.Errorf("keyword %q: got %s, want %s", text, toks[0].Kind, expectedKind)
        }
    }
}

func TestLexerIdentNotKeyword(t *testing.T) {
    toks, _, _ := Lex([]byte("foobar"))
    require(t, toks[0].Kind == TokenIdent)
}

func TestLexerAllOperators(t *testing.T) {
    cases := []struct{ src string; want TokenKind }{
        {"==", TokenEqEq}, {"!=", TokenBangEq}, {"<=", TokenLtEq},
        {">=", TokenGtEq}, {"**", TokenStarStar}, {"<<", TokenLtLt},
        {">>", TokenGtGt}, {"->", TokenArrow}, {":=", TokenColonEq},
        {"+=", TokenPlusEq}, {"-=", TokenMinusEq}, {"*=", TokenStarEq},
        {"/=", TokenSlashEq}, {"%=", TokenPercentEq}, {".*", TokenDotStar},
        {"+", TokenPlus}, {"-", TokenMinus}, {"*", TokenStar}, {"/", TokenSlash},
    }
    for _, c := range cases {
        toks, _, _ := Lex([]byte(c.src))
        if toks[0].Kind != c.want {
            t.Errorf("%q: got %s, want %s", c.src, toks[0].Kind, c.want)
        }
    }
}

func TestLexerLineComment(t *testing.T) {
    toks, _, _ := Lex([]byte("// this is a comment\nfoo"))
    // comment should be skipped; first token after newline is ident
    require(t, toks[0].Kind == TokenNewline)
    require(t, toks[1].Kind == TokenIdent)
}

func TestLexerLineTable(t *testing.T) {
    src := []byte("foo\nbar\nbaz")
    _, lt, _ := Lex(src)
    line, col := lt.LineCol(4) // 'b' of 'bar'
    require(t, line == 2, "expected line 2, got %d", line)
    require(t, col == 1, "expected col 1, got %d", col)
}

func TestLexerNewlineTokens(t *testing.T) {
    toks, _, _ := Lex([]byte("a\nb"))
    require(t, toks[1].Kind == TokenNewline)
}

func TestLexerTabError(t *testing.T) {
    _, _, diags := Lex([]byte("\t"))
    require(t, len(diags) > 0, "expected diagnostic for tab")
    require(t, diags[0].Code == 1, "expected E0001")
}
```

## Validation Checklist
- [ ] `Lex()` returns `([]Token, *LineTable, []Diagnostic)` — never panics
- [ ] Zero-copy: no string allocation in the lexer hot path
- [ ] EOF token always last in the output slice
- [ ] `LineTable.LineCol()` returns 1-based line and column
- [ ] All integer literal forms: decimal, hex (0x), octal (0o), binary (0b), with `_` separators
- [ ] Float literals with optional exponent (`e`/`E`, `+`/`-`)
- [ ] String literals with all escape sequences
- [ ] Char literals
- [ ] Line comments (`//`) produce no tokens
- [ ] Tab characters produce E0001 diagnostic
- [ ] All two-character operators recognized before single-character
- [ ] Keyword lookup correct (all 33 keywords)
- [ ] Unknown character produces `TokenError` + diagnostic, lexer continues
- [ ] `go test ./compiler/lexer/` passes all tests

## Acceptance Criteria
- `Lex([]byte{})` returns `[TokenEOF]` with zero diagnostics
- `Lex([]byte("0xFF"))` returns token with `Kind=TokenIntLit, Offset=0, Len=4`
- All keyword strings produce their keyword token, not `TokenIdent`
- `LineTable.LineCol(offset)` returns correct line/col for multi-line input
- No heap allocations in `Lex()` per allocation profiling (except pre-allocated slice)
- `go test -race ./compiler/lexer/` passes

## Definition of Done
- [ ] `compiler/lexer/lexer.go` committed
- [ ] `compiler/lexer/lexer_test.go` with all test functions committed
- [ ] All tests pass including race detector
- [ ] No allocations in lexer hot path (verified with `go test -bench=. -benchmem`)
- [ ] `golangci-lint run` passes
- [ ] INDENT/DEDENT NOT emitted here (that's p02-t03)

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| UTF-8 multi-byte sequences cause out-of-bounds reads | Bound-check `l.pos+1` before reading second byte; use `peek1()` safely |
| String literal scanning runs past EOF | Check `l.pos < len(l.src)` in every scan loop; emit E0002 on EOF-in-string |
| Integer literal underscore validation | Don't validate in lexer — emit the full token and let semantic analysis validate |
| `twoCharOps` map lookup on every byte is slow | Map lookup is O(1); for performance-critical path, replace with a switch table in future |
| `keywords` map allocation at package init | Acceptable for compiler bootstrap; profile if startup is slow |

## Future Follow-up Tasks
- p02-t03: Post-processing pass converts NEWLINE tokens + indentation into INDENT/DEDENT
- p02-t04: Error recovery layer on top of this core
- p02-t05: Golden tests use this lexer
- p02-t06: Fuzz target wraps this Lex() function
