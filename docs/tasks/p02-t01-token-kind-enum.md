# p02-t01: Token Kind Enum

## Purpose
Define the complete `TokenKind` enumeration in `compiler/lexer/token_kind.go`, covering every token that the AXIOM lexer can emit. This enum is the contract between the lexer and the parser — the lexer produces tokens with these kinds, and the parser switches on them. Getting the full set right up front prevents adding ad-hoc kinds later, which would require changing the parser and all existing tests. The enum must fit in `uint8` (max 255 values) to match the `Token.Kind` field layout established in p01-t03.

## Context
AXIOM's token kinds fall into seven categories: literals, identifiers, keywords, operators, punctuation, indentation/structure tokens, and control tokens (EOF, Error). The indentation tokens (INDENT, DEDENT, NEWLINE) are special — they are not present in the raw byte stream but are synthesized by the post-processing pass (p02-t03). Every keyword listed in the EBNF grammar (p01-t02) must have a corresponding `TokenKind` constant. Operators must cover all AXIOM operators including the unusual ones (`**` for power, `.*` for deref, `:=` for declaration). The `String()` method is required for error messages and debug dumps.

## Inputs
- `compiler/lexer/token.go` from p01-t03 (defines `TokenKind` type as `uint8`)
- `docs/GRAMMAR.ebnf` from p01-t02 (authoritative list of all terminals)
- `AXIOM LANGUAGE SPECIFICATION v1.0.md`

## Outputs
- `compiler/lexer/token_kind.go` — complete `TokenKind` enum with all constants and `String()` method
- `compiler/lexer/token_kind_test.go` — tests for `String()`, uniqueness, and count

## Dependencies
- p01-t03: struct-layout-definitions — `TokenKind` type declared in `token.go`
- p01-t02: grammar-ebnf — authoritative terminal list (use as checklist)

## Subsystems Affected
- `compiler/lexer/`: This enum is the primary output type of the lexer
- `compiler/parser/`: Parser switches on `TokenKind` values
- `compiler/diagnostics/`: Error messages use `tok.Kind.String()`

## Detailed Requirements

1. **Literal tokens**:
   ```go
   TokenIntLit    // 0xFF, 42, 0b1010, 0o77 (all integer literal forms)
   TokenFloatLit  // 3.14, 1.0e-6
   TokenStringLit // "hello\nworld"
   TokenCharLit   // 'a', '\n'
   ```

2. **Identifier and boolean atoms**:
   ```go
   TokenIdent    // any identifier not matching a keyword
   ```
   Note: `true`, `false`, and `nil` are keywords with their own token kinds, not `TokenIdent`.

3. **Keyword tokens** — one constant per keyword, in alphabetical order for readability:
   ```go
   TokenAnd       // and
   TokenAsync     // async
   TokenAwait     // await
   TokenConst     // const
   TokenDefer     // defer
   TokenElif      // elif
   TokenElse      // else
   TokenExtern    // extern
   TokenFalse     // false
   TokenFn        // fn
   TokenFor       // for
   TokenIf        // if
   TokenImport    // import
   TokenIn        // in
   TokenInterface // interface
   TokenIsolated  // Isolated (used in type position)
   TokenFuture    // Future (used in type position)
   TokenLent      // lent
   TokenLet       // let
   TokenMatch     // match
   TokenMut       // mut
   TokenNil       // nil
   TokenNot       // not
   TokenOr        // or
   TokenPacked    // packed
   TokenPub       // pub
   TokenReturn    // return
   TokenSpawn     // spawn
   TokenStruct    // struct
   TokenTrue      // true
   TokenType      // type
   TokenUnsafe    // unsafe
   TokenWhile     // while
   ```

4. **Arithmetic operators**:
   ```go
   TokenPlus      // +
   TokenMinus     // -
   TokenStar      // *
   TokenSlash     // /
   TokenPercent   // %
   TokenStarStar  // ** (power)
   ```

5. **Comparison operators**:
   ```go
   TokenEqEq      // ==
   TokenBangEq    // !=
   TokenLt        // <
   TokenGt        // >
   TokenLtEq      // <=
   TokenGtEq      // >=
   ```

6. **Bitwise operators**:
   ```go
   TokenAmp       // & (bitwise and)
   TokenPipe      // | (bitwise or / sum type)
   TokenCaret     // ^ (bitwise xor)
   TokenTilde     // ~ (bitwise not)
   TokenLtLt     // << (left shift)
   TokenGtGt     // >> (right shift)
   ```

7. **Assignment operators**:
   ```go
   TokenEq        // =
   TokenColonEq   // := (declare-and-assign, used in mut x := expr)
   TokenPlusEq    // +=
   TokenMinusEq   // -=
   TokenStarEq    // *=
   TokenSlashEq   // /=
   TokenPercentEq // %=
   ```

8. **Punctuation**:
   ```go
   TokenDot       // .
   TokenDotStar   // .* (deref)
   TokenComma     // ,
   TokenColon     // :
   TokenSemicolon // ;
   TokenArrow     // -> (return type arrow)
   TokenBang      // ! (sink/consume prefix)
   TokenLParen    // (
   TokenRParen    // )
   TokenLBracket  // [
   TokenRBracket  // ]
   TokenLBrace    // {
   TokenRBrace    // }
   TokenPipe2     // || (used in closure: |params|)
   ```
   Note: `TokenPipe` (`|`) is reused in closure syntax. The parser uses context to disambiguate `|params|` from bitwise-or.

9. **Indentation/structure tokens** (synthesized by the post-processor, not the raw lexer):
   ```go
   TokenIndent  // increase in indentation level
   TokenDedent  // decrease in indentation level
   TokenNewline // end of logical line
   ```

10. **Control tokens**:
    ```go
    TokenEOF   // end of file
    TokenError // lexer error: bad character or malformed literal
    ```

11. **Sentinel**:
    ```go
    TokenKindCount // total count; must remain last
    ```

12. **`String()` method** — must return a human-readable name for use in error messages:
    ```go
    func (k TokenKind) String() string {
        if int(k) < len(tokenKindNames) {
            return tokenKindNames[k]
        }
        return fmt.Sprintf("TokenKind(%d)", k)
    }

    var tokenKindNames = [TokenKindCount]string{
        TokenIntLit:    "integer literal",
        TokenFloatLit:  "float literal",
        TokenStringLit: "string literal",
        TokenCharLit:   "char literal",
        TokenIdent:     "identifier",
        TokenAnd:       "'and'",
        TokenAsync:     "'async'",
        // ... all keywords show with quotes
        TokenPlus:      "'+'",
        // ... etc
        TokenIndent:    "INDENT",
        TokenDedent:    "DEDENT",
        TokenNewline:   "NEWLINE",
        TokenEOF:       "EOF",
        TokenError:     "ERROR",
    }
    ```

13. **Keyword lookup map** — for the lexer to convert identifier bytes to keyword tokens:
    ```go
    // keywords maps identifier text to the corresponding keyword TokenKind.
    // The lexer uses this after scanning an identifier.
    var keywords = map[string]TokenKind{
        "and":       TokenAnd,
        "async":     TokenAsync,
        "await":     TokenAwait,
        "const":     TokenConst,
        "defer":     TokenDefer,
        "elif":      TokenElif,
        "else":      TokenElse,
        "extern":    TokenExtern,
        "false":     TokenFalse,
        "fn":        TokenFn,
        "for":       TokenFor,
        "if":        TokenIf,
        "import":    TokenImport,
        "in":        TokenIn,
        "interface": TokenInterface,
        "Isolated":  TokenIsolated,
        "Future":    TokenFuture,
        "lent":      TokenLent,
        "let":       TokenLet,
        "match":     TokenMatch,
        "mut":       TokenMut,
        "nil":       TokenNil,
        "not":       TokenNot,
        "or":        TokenOr,
        "packed":    TokenPacked,
        "pub":       TokenPub,
        "return":    TokenReturn,
        "spawn":     TokenSpawn,
        "struct":    TokenStruct,
        "true":      TokenTrue,
        "type":      TokenType,
        "unsafe":    TokenUnsafe,
        "while":     TokenWhile,
    }
    ```
    Note: `Isolated` and `Future` are PascalCase — this is intentional, they are type-level names, not regular keywords.

14. **Count constraint**: `TokenKindCount` must be ≤ 255 at all times. Currently ~80 tokens — well within uint8 range. If the count ever approaches 200, file an RFC to evaluate if uint8 should become uint16 in the Token struct.

## Implementation Steps

1. Create `compiler/lexer/token_kind.go`:
   ```go
   package lexer

   import "fmt"

   // TokenKind is the discriminant for a Token. Values are defined in
   // the iota block below. Must fit in uint8 (max 255).
   // See token.go for the Token struct definition.
   ```

2. Define the `const` iota block with all tokens in this order:
   - Literals (IntLit, FloatLit, StringLit, CharLit)
   - Ident
   - Keywords (alphabetical)
   - Arithmetic operators
   - Comparison operators
   - Bitwise operators
   - Assignment operators
   - Punctuation
   - Indentation tokens
   - Control tokens (EOF, Error)
   - TokenKindCount sentinel

3. Define the `tokenKindNames` array of size `TokenKindCount` with human-readable names. Initialize all entries — blank entries produce empty string which means a token's String() returns `TokenKind(N)`.

4. Implement `func (k TokenKind) String() string` using the lookup array with bounds check.

5. Define the `keywords` map as shown in Requirement 13.

6. Add a comment at the top of the `const` block:
   ```go
   // IMPORTANT: TokenKind values must fit in uint8 (max 255).
   // TokenKindCount is the sentinel that enforces this in token_kind_test.go.
   // Do not reorder existing constants — doing so breaks serialized token streams.
   ```

7. Create `compiler/lexer/token_kind_test.go`:
   - `TestTokenKindFitsUint8`: assert `TokenKindCount <= 255`
   - `TestTokenKindStringNonEmpty`: assert all defined kinds have non-empty String()
   - `TestKeywordsMapComplete`: assert every keyword token kind appears in the `keywords` map
   - `TestTokenKindUnique`: assert no two constant values are equal (iota guarantees this but make it explicit)

## Test Plan

Write `compiler/lexer/token_kind_test.go` in `package lexer`:

```go
package lexer

import "testing"

func TestTokenKindFitsUint8(t *testing.T) {
    if TokenKindCount > 255 {
        t.Fatalf("TokenKindCount = %d > 255; TokenKind must fit in uint8", TokenKindCount)
    }
}

func TestTokenKindStringNonEmpty(t *testing.T) {
    // Every defined kind should have a meaningful String() representation
    for i := TokenKind(0); i < TokenKindCount; i++ {
        s := i.String()
        if s == "" {
            t.Errorf("TokenKind(%d).String() is empty; add to tokenKindNames", i)
        }
    }
}

func TestKeywordsMapComplete(t *testing.T) {
    // Every keyword token must appear in the keywords map
    kwTokens := []TokenKind{
        TokenAnd, TokenAsync, TokenAwait, TokenConst, TokenDefer,
        TokenElif, TokenElse, TokenExtern, TokenFalse, TokenFn,
        TokenFor, TokenIf, TokenImport, TokenIn, TokenInterface,
        TokenIsolated, TokenFuture, TokenLent, TokenLet, TokenMatch,
        TokenMut, TokenNil, TokenNot, TokenOr, TokenPacked,
        TokenPub, TokenReturn, TokenSpawn, TokenStruct, TokenTrue,
        TokenType, TokenUnsafe, TokenWhile,
    }
    inMap := map[TokenKind]bool{}
    for _, v := range keywords {
        inMap[v] = true
    }
    for _, kw := range kwTokens {
        if !inMap[kw] {
            t.Errorf("keyword token %s (kind=%d) not found in keywords map", kw, kw)
        }
    }
}

func TestKeywordsMapNoDuplicateValues(t *testing.T) {
    seen := map[TokenKind]string{}
    for text, kind := range keywords {
        if prev, ok := seen[kind]; ok {
            t.Errorf("TokenKind %d mapped to both %q and %q", kind, prev, text)
        }
        seen[kind] = text
    }
}

func TestKeywordIsNotIdent(t *testing.T) {
    // All keywords must NOT be TokenIdent
    for text, kind := range keywords {
        if kind == TokenIdent {
            t.Errorf("keyword %q maps to TokenIdent; must have its own kind", text)
        }
    }
}
```

## Validation Checklist
- [ ] All literals defined: IntLit, FloatLit, StringLit, CharLit
- [ ] All 33 keywords defined (and, async, await, const, defer, elif, else, extern, false, fn, for, if, import, in, interface, Isolated, Future, lent, let, match, mut, nil, not, or, packed, pub, return, spawn, struct, true, type, unsafe, while)
- [ ] All arithmetic operators: +, -, *, /, %, **
- [ ] All comparison operators: ==, !=, <, >, <=, >=
- [ ] All bitwise operators: &, |, ^, ~, <<, >>
- [ ] All assignment operators: =, :=, +=, -=, *=, /=, %=
- [ ] All punctuation: ., .*, ,, :, ;, ->, !, (, ), [, ], {, }, ||
- [ ] INDENT, DEDENT, NEWLINE tokens defined
- [ ] EOF and Error tokens defined
- [ ] TokenKindCount sentinel is last
- [ ] `String()` returns non-empty for every kind
- [ ] `keywords` map covers all keyword token kinds
- [ ] `TokenKindCount <= 255` verified by test
- [ ] `go test ./compiler/lexer/` passes

## Acceptance Criteria
- `TestTokenKindFitsUint8` passes (TokenKindCount ≤ 255)
- `TestTokenKindStringNonEmpty` passes (every kind has a String())
- `TestKeywordsMapComplete` passes (all keywords in map)
- `TestKeywordsMapNoDuplicateValues` passes (no two texts share a kind)
- `golangci-lint run` passes on the new file
- `gofmt` produces no diff

## Definition of Done
- [ ] `compiler/lexer/token_kind.go` committed
- [ ] `compiler/lexer/token_kind_test.go` committed
- [ ] All 5 test functions pass
- [ ] Lint passes
- [ ] Cross-referenced with `docs/GRAMMAR.ebnf` — every terminal in grammar has a token kind

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Forgetting a keyword or operator | Use GRAMMAR.ebnf as a checklist; `TestKeywordsMapComplete` catches missing keywords |
| Reordering iota breaks serialized tokens | Add comment warning; never reorder; only append |
| `Isolated` and `Future` being PascalCase causes confusion | Document clearly: these are type-level reserved names, not regular identifiers |
| Operator precedence info in enum | Do NOT store precedence in the enum; that belongs in the parser's Pratt table |
| Token kinds exceeding 200 as language grows | Document the limit; file RFC if approaching 200; current count ~80 |

## Future Follow-up Tasks
- p02-t02: Lexer core uses `TokenKind` constants and `keywords` map defined here
- p03-t04: Parser switches on `TokenKind` values; must cover all relevant kinds
- p01-t02: GRAMMAR.ebnf terminals must map 1:1 to token kinds (cross-reference)
