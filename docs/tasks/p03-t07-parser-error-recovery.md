# p03-t07: Parser Error Recovery

## Purpose
Implement robust error recovery so the parser continues producing a useful (partial) AST after encountering syntax errors. A parser that stops at the first error is unusable in a language server or interactive development context. AXIOM's parser must collect all errors in a single pass.

## Context
Error recovery uses the "panic mode" strategy: when an unexpected token is encountered, the parser discards tokens until it finds a synchronization token that can safely restart parsing. Sync tokens in AXIOM are: `fn`, `struct`, `interface`, `let`, `mut`, `return`, `if`, `for`, `while`, `Newline`, `Dedent`, `EOF`. Error nodes (`NodeKind.Error`) are inserted into the AST to represent failed parse regions.

## Inputs
- `compiler/parser/parser.go` — parser state with `diags []Diagnostic`
- `compiler/ast/node.go` — NodeKind.Error definition
- `compiler/lexer/token.go` — sync token kinds

## Outputs
- `compiler/parser/recovery.go` — sync/recovery helpers
- Updated all parse functions to call recovery on failure
- `[]Diagnostic` accumulates all errors without stopping

## Dependencies
- p03-t04: parser-statements — all statement parsers need recovery hooks
- p03-t05: parser-expressions-pratt — expression parser needs recovery
- p03-t06: parser-indentation — block parser needs recovery

## Subsystems Affected
- Parser: every parse function affected
- AST: NodeKind.Error nodes appear in tree
- Diagnostics: all errors collected, not just first

## Detailed Requirements

1. `synchronize()` function: advance past tokens until a sync token is found.
2. Sync tokens: `fn`, `struct`, `interface`, `let`, `mut`, `return`, `if`, `for`, `while`, `Dedent`, `EOF`.
3. `errorNode(msg string) uint32`: create NodeKind.Error node, add diagnostic, return node index.
4. Every `expect()` call that fails must call `errorNode()` and then `synchronize()`.
5. Parser must never panic — all error paths produce error nodes + diagnostics.
6. Maximum error count: 50 errors before stopping (prevent error cascade).
7. Error messages must include the unexpected token text and expected token name.

```go
func (p *Parser) synchronize() {
    for !p.check(TokenKind.EOF) {
        switch p.peek().Kind {
        case TokenKind.KwFn, TokenKind.KwStruct, TokenKind.KwLet,
             TokenKind.KwMut, TokenKind.KwReturn, TokenKind.KwIf,
             TokenKind.KwFor, TokenKind.KwWhile, TokenKind.Dedent:
            return
        }
        p.advance()
    }
}

func (p *Parser) errorNode(msg string) uint32 {
    tok := p.peek()
    p.addDiagnostic(DiagError, tok, msg)
    idx := p.tree.AddNode(NodeKind.Error, p.pos)
    p.synchronize()
    return idx
}
```

## Implementation Steps

1. Create `compiler/parser/recovery.go` with `synchronize()` and `errorNode()`.
2. Add `maxErrors int = 50` field to Parser; check and stop emitting after limit.
3. Update `expect(kind)` to call `errorNode()` on mismatch instead of panicking.
4. Update `parseStatement()`: wrap each parse attempt in a recover path using `errorNode()`.
5. Update `parseExpression()`: on atom parse failure, return `errorNode()`.
6. Update `parseBlock()`: if statement parse returns error node, continue to next statement.
7. Add `HasErrors() bool` method on Parser for checking after full parse.

## Test Plan

- `TestRecoveryMissingColon`: `fn foo()` with no colon — parse continues, gets next fn
- `TestRecoveryBadExpr`: `let x = @@@` — error node for expr, parsing continues at next stmt
- `TestRecoveryMultipleErrors`: file with 5 syntax errors — all 5 diagnostics collected
- `TestRecoveryMaxErrors`: file with 100 syntax errors — stops at 50 diagnostics
- `TestRecoveryNoPanic`: fuzz input of random bytes — no panic

## Validation Checklist

- [ ] Parser never calls `panic()` (grep for panic in parser/)
- [ ] `errorNode()` always adds exactly one diagnostic
- [ ] `synchronize()` always terminates (no infinite loop)
- [ ] Error nodes appear in AST at correct positions
- [ ] MaxErrors limit works correctly

## Acceptance Criteria

- File with N syntax errors produces exactly min(N, 50) diagnostics
- Partial AST with error nodes is valid (no nil pointer issues in printer)
- `axc dump-ast` on error-containing files still prints partial tree

## Definition of Done

- [ ] Recovery functions implemented and tested
- [ ] All parse functions updated to use recovery
- [ ] `go test ./compiler/parser/ -run TestRecovery` passes
- [ ] Fuzz target runs 10K iterations without panic

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Infinite loop in synchronize() eating entire file | Max token consumption limit (1000 per synchronize call) |
| Error cascade generating thousands of misleading errors | MaxErrors=50 hard stop |

## Future Follow-up Tasks

- p03-t08: golden tests include error recovery cases
- p03-t09: fuzz target validates no panics
- p04-t04: name resolver must handle NodeKind.Error nodes gracefully
