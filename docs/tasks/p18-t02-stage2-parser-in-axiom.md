# p18-t02: Stage 2 — AXIOM Parser in AXIOM

## Purpose
Implement the AXIOM parser in AXIOM (Stage 2 of self-hosting) — a recursive descent + Pratt parser producing FlatAST from token streams, compiled by the Go bootstrap compiler.

## Context
With the Stage 1 lexer working, Stage 2 implements the parser in AXIOM. The AXIOM parser produces the same `AstNode` flat array representation as the Go parser. Stage 2 output is compared against the Go parser for all test files.

## Inputs
- `bootstrap/stage1/lexer.ax` — token stream from Stage 1
- AstNode definitions from p03-t01
- Go parser algorithm from p03-t04 and p03-t05

## Outputs
- `bootstrap/stage2/ast.ax` — AstNode, NodeKind in AXIOM
- `bootstrap/stage2/parser.ax` — recursive descent parser in AXIOM

## Dependencies
- p18-t01: stage1-compiler-in-axiom — token stream input
- p03-t01: ast-node-definitions — AstNode layout to replicate
- p03-t04 through p03-t07: parser algorithms to port

## Detailed Requirements

```axiom
# bootstrap/stage2/ast.ax
type NodeKind: u16
const NK_FUNC_DECL: NodeKind = 1
const NK_VAR_DECL:  NodeKind = 2
const NK_BLOCK:     NodeKind = 3
# ... all node kinds

type AstNode:
    var kind:        NodeKind
    var flags:       u16
    var token_idx:   u32
    var first_child: u32
    var next_sibling: u32
    var payload:     u32
    var extra_idx:   u32

# bootstrap/stage2/parser.ax
type Parser:
    var tokens:  Array[Token]
    var nodes:   Array[AstNode]
    var pos:     u32
    var errors:  Array[ParseError]

    fn new(tokens: Array[Token]) -> Parser
    fn parse_module(mut self) -> u32      # returns root node index
    fn parse_stmt(mut self) -> u32
    fn parse_expr(mut self, min_bp: u8) -> u32  # Pratt
    fn parse_func_decl(mut self) -> u32
    fn parse_var_decl(mut self) -> u32
    fn parse_if(mut self) -> u32
    fn parse_while(mut self) -> u32
    fn parse_for(mut self) -> u32
    fn peek(self) -> Token
    fn consume(mut self) -> Token
    fn expect(mut self, kind: TokenKind) -> Token
```

Validation: compare AstNode arrays from Go parser and AXIOM parser for all test files. Node kinds and tree structure must be identical.

## Implementation Steps

1. Create `bootstrap/stage2/ast.ax` — AstNode struct and NodeKind constants.
2. Create `bootstrap/stage2/parser.ax` — full recursive descent parser.
3. Port Pratt expression parser from Go.
4. Implement parse_func_decl, parse_var_decl, parse_if, etc.
5. Write comparison harness: Go AST vs AXIOM AST.
6. Run on repository corpus.

## Test Plan
- `TestStage2ParserCorpus`: all .ax files produce identical AST structure
- `TestStage2ParserExpr`: all expression forms parse correctly
- `TestStage2ParserError`: invalid syntax → error with correct location

## Validation Checklist
- [ ] AST structure matches Go parser for all 200+ test cases
- [ ] Error recovery produces same synchronization points as Go parser
- [ ] Indentation-based blocks parsed correctly

## Acceptance Criteria
- AXIOM parser in AXIOM correctly parses all of stdlib source

## Definition of Done
- [ ] `bootstrap/stage2/parser.ax` implemented
- [ ] Corpus comparison passes

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Pratt parser precedence table errors | Copy exact binding-power table from Go parser |

## Future Follow-up Tasks
- p18-t03: Stage 3 — type checker in AXIOM
