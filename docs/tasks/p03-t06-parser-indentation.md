# p03-t06: Parser Indentation Block Handling

## Purpose
Implement indentation-aware block parsing so the parser correctly handles INDENT/DEDENT tokens produced by the lexer post-processing pass. This enables the Python-like block syntax of AXIOM where blocks are delimited by indentation rather than braces.

## Context
AXIOM uses 4-space indentation as its block delimiter. The lexer post-processor (p02-t03) emits INDENT and DEDENT tokens. The parser must consume these to form `NodeKind.Block` AST nodes. Every compound statement (if, for, while, fn body, struct body) ends with a colon `:` followed by INDENT...DEDENT. Mismatched INDENT/DEDENT must produce recoverable errors.

## Inputs
- `compiler/lexer/token.go` — TokenKind.Indent, TokenKind.Dedent definitions
- `compiler/parser/parser.go` — parser state struct
- `compiler/ast/node.go` — NodeKind.Block definition

## Outputs
- `compiler/parser/block.go` — `parseBlock()` function and helpers
- Block nodes correctly contain child statement nodes in the flat AST

## Dependencies
- p03-t04: parser-statements — statement parsing drives block parsing
- p02-t03: indent-dedent-handling — produces the INDENT/DEDENT tokens

## Subsystems Affected
- Parser: block parsing is called by every compound statement
- AST: Block nodes become parent of all contained statements

## Detailed Requirements

1. `parseBlock() uint32` returns the node index of the Block node.
2. Expect INDENT at start; if missing, emit error and return empty block.
3. Parse statements until DEDENT or EOF.
4. DEDENT must match the corresponding INDENT — track nesting depth.
5. Nested blocks (e.g., if inside for) must work correctly.
6. Empty blocks (INDENT immediately followed by DEDENT) are valid — used for forward declarations.
7. Block node's FirstChild points to first statement; statements linked via NextSibling.

```go
func (p *Parser) parseBlock() uint32 {
    blockIdx := p.tree.AddNode(NodeKind.Block, p.pos)
    if !p.expect(TokenKind.Indent) {
        return blockIdx // empty block with error
    }
    var prevStmt uint32 = 0
    for !p.check(TokenKind.Dedent) && !p.check(TokenKind.EOF) {
        stmtIdx := p.parseStatement()
        if prevStmt == 0 {
            p.tree.SetFirstChild(blockIdx, stmtIdx)
        } else {
            p.tree.SetNextSibling(prevStmt, stmtIdx)
        }
        prevStmt = stmtIdx
    }
    p.expect(TokenKind.Dedent)
    return blockIdx
}
```

## Implementation Steps

1. Create `compiler/parser/block.go` with `parseBlock()`.
2. In `parseBlock()`: call `p.expect(TokenKind.Indent)` — on failure, emit diagnostic and return empty block node.
3. Loop calling `parseStatement()` while not DEDENT/EOF, linking children in flat AST.
4. Call `p.expect(TokenKind.Dedent)` at end.
5. Update all compound statement parsers (if, for, while, fn, struct, interface) to call `parseBlock()` after the colon.
6. Add `parseColon()` helper: expects `:` then calls `parseBlock()`.
7. Handle the case where a block contains only a comment or blank lines (INDENT + DEDENT with no real statements).

## Test Plan

- `TestParseBlockSimple`: single statement in a function body
- `TestParseBlockNested`: if inside for, 3 levels deep
- `TestParseBlockEmpty`: `fn foo(): \n    pass` (empty block with pass stmt)
- `TestParseBlockMissingIndent`: `fn foo():` with no following INDENT — expect error node
- `TestParseBlockMissingDedent`: INDENT never closed — expect error at EOF
- Golden test: fibonacci function with nested if/else

## Validation Checklist

- [ ] parseBlock() never panics on any input
- [ ] Nested blocks produce correctly nested AST
- [ ] Empty blocks produce valid Block node with no children
- [ ] Missing INDENT emits exactly one diagnostic
- [ ] Missing DEDENT emits exactly one diagnostic
- [ ] All existing statement parsers call parseBlock() for their body

## Acceptance Criteria

- All parser golden tests from p03-t08 that involve blocks pass
- Fuzz target (p03-t09) finds no panics related to block parsing
- `axc dump-ast` correctly prints nested block structure

## Definition of Done

- [ ] parseBlock() implemented and tested
- [ ] All compound statements use parseBlock()
- [ ] Unit tests pass: `go test ./compiler/parser/ -run TestBlock`
- [ ] No panics in fuzz run of 10,000 iterations

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Off-by-one in INDENT/DEDENT matching causing infinite loops | Track depth counter, add max-depth guard (256) |
| Ambiguous grammar when else/elif follows dedent | Peek after DEDENT before consuming it |

## Future Follow-up Tasks

- p03-t07: error recovery builds on block parsing
- p03-t08: golden tests validate block output
- p09-t08: AIR builder control flow lowers Block nodes
