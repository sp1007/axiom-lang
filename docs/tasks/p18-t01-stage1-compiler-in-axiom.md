# p18-t01: Stage 1 — AXIOM Lexer in AXIOM

## Purpose
Implement the AXIOM lexer in AXIOM itself (Stage 1 of self-hosting), producing a lexer that can tokenize AXIOM source code and is compiled by the Go-based bootstrap compiler.

## Context
Self-hosting is the ultimate validation: the compiler can compile itself. Stage 1 starts with the lexer — the simplest component. The AXIOM lexer written in AXIOM is compiled by the existing Go compiler, and its output is compared to the Go lexer's output for correctness verification.

## Inputs
- AXIOM lexer specification from p02 tasks (TokenKind, token rules)
- Bootstrap compiler (Go-based) from phases 01-17
- Test corpus: all .ax files in the repository

## Outputs
- `bootstrap/stage1/lexer.ax` — AXIOM lexer written in AXIOM
- `bootstrap/stage1/token.ax` — Token, TokenKind in AXIOM

## Dependencies
- All of phases 01-17 (bootstrap compiler must be complete)
- p02-t01: token-kind-enum — token definitions to replicate
- p02-t02: lexer-core — algorithm to replicate in AXIOM

## Subsystems Affected
- Self-hosting pipeline: stage1 lexer output feeds stage2 parser
- Testing: lexer output compared to Go reference implementation

## Detailed Requirements

```axiom
# bootstrap/stage1/token.ax

type TokenKind: u8
const TK_EOF:     TokenKind = 0
const TK_IDENT:   TokenKind = 1
const TK_INT:     TokenKind = 2
const TK_FLOAT:   TokenKind = 3
const TK_STRING:  TokenKind = 4
const TK_INDENT:  TokenKind = 5
const TK_DEDENT:  TokenKind = 6
const TK_NEWLINE: TokenKind = 7
# ... all token kinds

type Token:
    var kind:   TokenKind
    var offset: u32
    var len:    u16
    var flags:  u8

# bootstrap/stage1/lexer.ax
type Lexer:
    var src:    str
    var pos:    u32
    var indent: Array[u32]   # indent stack
    var tokens: Array[Token]

    fn new(src: str) -> Lexer
    fn tokenize(mut self) -> Array[Token]
    fn next_token(mut self) -> Token
    fn skip_whitespace(mut self)
    fn lex_ident(mut self) -> Token
    fn lex_number(mut self) -> Token
    fn lex_string(mut self) -> Token
    fn lex_indent(mut self) -> Array[Token]   # may emit INDENT/DEDENT
```

Correctness validation:
1. Tokenize 1000 .ax files with both Go lexer and AXIOM lexer.
2. Assert: for every file, `go_tokens == axiom_tokens` (same kind, offset, len).
3. Any mismatch is a bug in the AXIOM lexer.

## Implementation Steps

1. Create `bootstrap/stage1/token.ax` with all TokenKind constants.
2. Create `bootstrap/stage1/lexer.ax` — complete lexer implementation.
3. Port Go lexer algorithm to AXIOM (mechanical translation).
4. Write comparison harness: compile both, compare outputs.
5. Run on entire repository .ax corpus.
6. Fix all discrepancies.

## Test Plan
- `TestStage1LexerCorpus`: all .ax files tokenized identically by Go and AXIOM lexers
- `TestStage1LexerSpeed`: AXIOM lexer within 3x of Go lexer speed (expected overhead)
- `TestStage1IndentDedent`: indent/dedent tokens correct on AXIOM source with deep nesting

## Validation Checklist
- [ ] All .ax files produce identical token streams
- [ ] INDENT/DEDENT counts match Go lexer
- [ ] Token offsets match (not just kinds)
- [ ] EOF token always present as last token

## Acceptance Criteria
- AXIOM lexer tokenizes AXIOM's own source correctly (no discrepancies)

## Definition of Done
- [ ] `bootstrap/stage1/lexer.ax` implemented
- [ ] Corpus comparison test passes on 100% of repository .ax files

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Indent handling edge cases differ between implementations | Focus testing on files with deep nesting and mixed blank lines |
| AXIOM lexer runs out of memory on large files | Use streaming tokenization, not load-all-at-once |

## Future Follow-up Tasks
- p18-t02: Stage 2 — AXIOM parser in AXIOM
