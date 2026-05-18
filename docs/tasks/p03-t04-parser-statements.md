# p03-t04: Parser — Statements

## Purpose
Implement the recursive descent parser for all AXIOM statement forms in `compiler/parser/parser.go`. The parser consumes the post-processed token stream (including INDENT/DEDENT/NEWLINE) from p02-t03 and builds an `AstTree` (p03-t01). This task covers all statement-level grammar productions: function declarations, struct declarations, interface declarations, import declarations, variable declarations, assignments, control flow (if/elif/else, for, while, match, defer, unsafe, arena blocks), and return statements. Expression parsing is delegated to the Pratt parser (p03-t05); block parsing to p03-t06.

## Context
The parser is a single-pass recursive descent parser. It is driven by the EBNF grammar from p01-t02. Parser state is held in a `Parser` struct containing the token stream, current position, the `AstTree` being built, and a diagnostic slice. The parser never panics — on error it emits a diagnostic and uses synchronization tokens to resume. Error recovery is formalized in p03-t07 but the basic pattern (skip to sync token) is established here. The parser is the bridge between the token stream and the AST — it must be complete, deterministic, and testable in isolation.

## Inputs
- `[]Token` from `lexer.Lex()` (p02-t02/p02-t03) — post-processed with INDENT/DEDENT
- `compiler/ast/tree.go` (p03-t01) — `AstTree` and builder methods
- `compiler/ast/intern.go` (p03-t02) — `InternPool`
- `compiler/lexer/token_kind.go` (p02-t01) — all `TokenKind` constants
- `compiler/diagnostics/` (p01-t01)
- `docs/GRAMMAR.ebnf` (p01-t02) — authoritative grammar

## Outputs
- `compiler/parser/parser.go` — `Parser` struct and all statement parse methods
- `compiler/parser/parser_test.go` — unit tests for statement parsing

## Dependencies
- p03-t01: ast-node-definitions
- p03-t02: string-intern-pool
- p02-t03: indent-dedent-handling
- p01-t02: grammar-ebnf
- p01-t03: struct-layout-definitions

## Subsystems Affected
- `compiler/parser/`: Primary implementation
- `compiler/ast/`: Consumed by parser

## Detailed Requirements

1. **`Parser` struct**:
   ```go
   package parser

   // Parser converts a token stream into an AstTree.
   // It is a single-pass recursive descent parser following the EBNF grammar.
   // Parser never panics; errors produce diagnostics and NodeError nodes.
   type Parser struct {
       tokens []lexer.Token  // full post-processed token stream
       pos    int            // current position in tokens
       tree   *ast.AstTree
       pool   *ast.InternPool
       src    []byte         // original source (for error messages)
       diags  []diagnostics.Diagnostic
   }
   ```

2. **Public API**:
   ```go
   // Parse parses the token stream and returns an AstTree.
   // src is the original source bytes (for error reporting and token text recovery).
   // Always returns a non-nil tree; errors are in the returned diagnostics.
   func Parse(tokens []lexer.Token, src []byte, pool *ast.InternPool) (*ast.AstTree, []diagnostics.Diagnostic) {
       p := &Parser{
           tokens: tokens,
           tree:   ast.NewTree(src, tokens),
           pool:   pool,
           src:    src,
       }
       p.parseProgram()
       return p.tree, p.diags
   }
   ```

3. **Core helper methods**:
   ```go
   // peek returns the current token without advancing.
   func (p *Parser) peek() lexer.Token {
       for p.pos < len(p.tokens) && p.tokens[p.pos].Kind == lexer.TokenNewline {
           p.pos++ // skip bare NEWLINEs between statements
       }
       if p.pos >= len(p.tokens) {
           return lexer.Token{Kind: lexer.TokenEOF}
       }
       return p.tokens[p.pos]
   }

   // peekRaw returns the current token without skipping NEWLINEs.
   func (p *Parser) peekRaw() lexer.Token { ... }

   // consume advances past the current token and returns it.
   func (p *Parser) consume() lexer.Token {
       tok := p.peek()
       p.pos++
       return tok
   }

   // expect consumes a token of the given kind, or emits a diagnostic.
   func (p *Parser) expect(kind lexer.TokenKind) (lexer.Token, bool) {
       tok := p.peek()
       if tok.Kind != kind {
           p.errorf(tok, "expected %s, got %s", kind, tok.Kind)
           return tok, false
       }
       return p.consume(), true
   }

   // check returns true if the current token has the given kind (without consuming).
   func (p *Parser) check(kind lexer.TokenKind) bool { return p.peek().Kind == kind }
   ```

4. **`parseProgram()`** — parse the root:
   ```go
   func (p *Parser) parseProgram() {
       for !p.check(lexer.TokenEOF) {
           // Skip leading NEWLINEs
           if p.check(lexer.TokenNewline) { p.consume(); continue }
           // Parse top-level declaration
           decl := p.parseTopLevelDecl()
           if decl != 0 {
               p.tree.AppendChild(0, decl)
           }
       }
   }
   ```

5. **`parseTopLevelDecl()`** — dispatch based on current token:
   ```go
   func (p *Parser) parseTopLevelDecl() uint32 {
       tok := p.peek()
       switch tok.Kind {
       case lexer.TokenImport:
           return p.parseImportDecl()
       case lexer.TokenPub:
           return p.parsePubDecl()
       case lexer.TokenFn:
           return p.parseFuncDecl(false)
       case lexer.TokenAsync:
           return p.parseAsyncFuncDecl()
       case lexer.TokenStruct:
           return p.parseStructDecl(false)
       case lexer.TokenInterface:
           return p.parseInterfaceDecl(false)
       case lexer.TokenType:
           return p.parseTypeAliasDecl(false)
       case lexer.TokenConst:
           return p.parseConstDecl(false)
       default:
           p.errorf(tok, "expected declaration, got %s", tok.Kind)
           p.syncToTopLevel()
           return 0
       }
   }
   ```

6. **`parseFuncDecl(isPub bool)`**:
   ```go
   func (p *Parser) parseFuncDecl(isPub bool) uint32 {
       fnTok, _ := p.expect(lexer.TokenFn)
       node := p.tree.AddNode(ast.NodeFuncDecl, p.tokenIdx(fnTok))
       if isPub { p.tree.SetFlags(node, ast.FlagIsPub) }

       nameTok, ok := p.expect(lexer.TokenIdent)
       if ok {
           nameID := p.pool.Intern(p.tokenText(nameTok))
           p.tree.SetPayload(node, nameID) // payload = intern ID of name
       }

       // Generic params: [T: Interface]
       if p.check(lexer.TokenLBracket) {
           p.tree.SetFlags(node, ast.FlagIsGeneric)
           genericNode := p.parseGenericParams()
           p.tree.AppendChild(node, genericNode)
       }

       // Parameter list
       p.expect(lexer.TokenLParen)
       for !p.check(lexer.TokenRParen) && !p.check(lexer.TokenEOF) {
           param := p.parseParam()
           if param != 0 { p.tree.AppendChild(node, param) }
           if !p.check(lexer.TokenRParen) { p.expect(lexer.TokenComma) }
       }
       p.expect(lexer.TokenRParen)

       // Return type: -> TypeExpr [{.effect.}]
       if p.check(lexer.TokenArrow) {
           p.consume()
           retType := p.parseTypeExpr()
           if retType != 0 { p.tree.AppendChild(node, retType) }
           if p.check(lexer.TokenLBrace) {
               effect := p.parseEffectAnnotation()
               if effect != 0 { p.tree.AppendChild(node, effect) }
           }
       }

       // Body: : Block
       p.expect(lexer.TokenColon)
       body := p.parseBlock()
       if body != 0 { p.tree.AppendChild(node, body) }

       return node
   }
   ```

7. **`parseParam()`**:
   ```go
   func (p *Parser) parseParam() uint32 {
       tok := p.peek()
       node := p.tree.AddNode(ast.NodeParamDecl, p.tokenIdx(tok))
       // Optional modifier: lent, !, mut
       switch tok.Kind {
       case lexer.TokenLent:
           p.consume()
           p.tree.SetFlags(node, ast.FlagIsLent)
           tok = p.peek()
       case lexer.TokenBang:
           p.consume()
           p.tree.SetFlags(node, ast.FlagIsSink)
           tok = p.peek()
       case lexer.TokenMut:
           p.consume()
           p.tree.SetFlags(node, ast.FlagIsMut)
           tok = p.peek()
       }
       nameTok, _ := p.expect(lexer.TokenIdent)
       nameID := p.pool.Intern(p.tokenText(nameTok))
       p.tree.SetPayload(node, nameID)
       p.expect(lexer.TokenColon)
       typeNode := p.parseTypeExpr()
       if typeNode != 0 { p.tree.AppendChild(node, typeNode) }
       return node
   }
   ```

8. **`parseImportDecl()`**:
   ```go
   func (p *Parser) parseImportDecl() uint32 {
       tok, _ := p.expect(lexer.TokenImport)
       node := p.tree.AddNode(ast.NodeImportDecl, p.tokenIdx(tok))
       // Parse dotted path: std.fs.path
       pathTok, _ := p.expect(lexer.TokenIdent)
       pathID := p.pool.Intern(p.tokenText(pathTok))
       for p.check(lexer.TokenDot) {
           p.consume()
           seg, ok := p.expect(lexer.TokenIdent)
           if !ok { break }
           // Concatenate: store full path as "std.fs.path"
           full := p.pool.Get(pathID) + "." + string(p.tokenText(seg))
           pathID = p.pool.InternString(full)
       }
       p.tree.SetPayload(node, pathID)
       // Optional selective import: { read, write }
       if p.check(lexer.TokenLBrace) {
           p.consume()
           for !p.check(lexer.TokenRBrace) && !p.check(lexer.TokenEOF) {
               name, _ := p.expect(lexer.TokenIdent)
               nameNode := p.tree.AddNode(ast.NodeIdent, p.tokenIdx(name))
               p.tree.AppendChild(node, nameNode)
               if !p.check(lexer.TokenRBrace) { p.expect(lexer.TokenComma) }
           }
           p.expect(lexer.TokenRBrace)
       }
       p.expectNewline()
       return node
   }
   ```

9. **`parseVarDecl()`** — handles both `let` and `mut`:
   ```go
   func (p *Parser) parseVarDecl() uint32 {
       tok := p.consume() // let or mut
       node := p.tree.AddNode(ast.NodeVarDecl, p.tokenIdx(tok))
       if tok.Kind == lexer.TokenMut { p.tree.SetFlags(node, ast.FlagIsMut) }
       nameTok, _ := p.expect(lexer.TokenIdent)
       p.tree.SetPayload(node, p.pool.Intern(p.tokenText(nameTok)))
       // Optional type annotation: : TypeExpr
       if p.check(lexer.TokenColon) {
           p.consume()
           typeNode := p.parseTypeExpr()
           if typeNode != 0 { p.tree.AppendChild(node, typeNode) }
       }
       // Initializer: = expr or := expr
       if p.check(lexer.TokenEq) || p.check(lexer.TokenColonEq) {
           p.consume()
           expr := p.parseExpr()
           if expr != 0 { p.tree.AppendChild(node, expr) }
       }
       p.expectNewline()
       return node
   }
   ```

10. **`parseIfStmt()`**:
    ```go
    func (p *Parser) parseIfStmt() uint32 {
        tok, _ := p.expect(lexer.TokenIf)
        node := p.tree.AddNode(ast.NodeIfStmt, p.tokenIdx(tok))
        cond := p.parseExpr()
        if cond != 0 { p.tree.AppendChild(node, cond) }
        p.expect(lexer.TokenColon)
        body := p.parseBlock()
        if body != 0 { p.tree.AppendChild(node, body) }
        for p.check(lexer.TokenElif) {
            elifTok := p.consume()
            elifNode := p.tree.AddNode(ast.NodeElifClause, p.tokenIdx(elifTok))
            elifCond := p.parseExpr()
            if elifCond != 0 { p.tree.AppendChild(elifNode, elifCond) }
            p.expect(lexer.TokenColon)
            elifBody := p.parseBlock()
            if elifBody != 0 { p.tree.AppendChild(elifNode, elifBody) }
            p.tree.AppendChild(node, elifNode)
        }
        if p.check(lexer.TokenElse) {
            p.consume()
            elseTok := p.tokens[p.pos-1]
            elseNode := p.tree.AddNode(ast.NodeElseClause, p.tokenIdx(elseTok))
            p.expect(lexer.TokenColon)
            elseBody := p.parseBlock()
            if elseBody != 0 { p.tree.AppendChild(elseNode, elseBody) }
            p.tree.AppendChild(node, elseNode)
        }
        return node
    }
    ```

11. **Other statement parsers** — implement all remaining statement forms:
    - `parseForStmt()`: `for IDENT in Expr: Block`
    - `parseWhileStmt()`: `while Expr: Block`
    - `parseMatchStmt()`: `match Expr: INDENT MatchArm* DEDENT`
    - `parseDeferStmt()`: `defer Expr NEWLINE`
    - `parseUnsafeBlock()`: `unsafe: Block`
    - `parseArenaBlock()`: `in [IDENT]: Block`
    - `parseReturnStmt()`: `return [Expr] NEWLINE`
    - `parseAssignOrExprStmt()`: parse Expr, then check for AssignOp

12. **`parseStructDecl(isPub bool)`**:
    - Parse `struct IDENT [GenericParams]: INDENT fields methods DEDENT`
    - Fields: `[pub] [mut] IDENT: TypeExpr NEWLINE`
    - Methods: `[pub] fn IDENT(...): Block`

13. **`parseInterfaceDecl(isPub bool)`**:
    - Parse `interface IDENT [GenericParams]: INDENT MethodSig* DEDENT`
    - MethodSig: `[async] fn IDENT(...) [-> TypeExpr] NEWLINE`

14. **`tokenIdx(tok lexer.Token) uint32`** — find token index in the token slice:
    ```go
    func (p *Parser) tokenIdx(tok lexer.Token) uint32 {
        // Binary search tokens by offset for O(log n) lookup
        lo, hi := 0, len(p.tokens)
        for lo < hi {
            mid := (lo + hi) / 2
            if p.tokens[mid].Offset < tok.Offset { lo = mid + 1 } else { hi = mid }
        }
        return uint32(lo)
    }
    ```

15. **`errorf()` helper**:
    ```go
    func (p *Parser) errorf(tok lexer.Token, format string, args ...any) {
        // Build a diagnostic from the token position
        p.diags = append(p.diags, diagnostics.Diagnostic{
            Severity: diagnostics.SeverityError,
            Code:     1000, // parser error base
            Pos:      diagnostics.Pos{Offset: tok.Offset},
            Message:  fmt.Sprintf(format, args...),
        })
    }
    ```

## Implementation Steps

1. Create `compiler/parser/parser.go` with `Parser` struct and `Parse()` public function.

2. Implement all core helpers: `peek()`, `peekRaw()`, `consume()`, `expect()`, `check()`, `errorf()`, `tokenIdx()`, `tokenText()`, `expectNewline()`.

3. Implement `parseProgram()` and `parseTopLevelDecl()`.

4. Implement `parseFuncDecl()`, `parseParam()`, `parseAsyncFuncDecl()`, `parsePubDecl()`.

5. Implement `parseImportDecl()`.

6. Implement `parseStructDecl()` and `parseInterfaceDecl()`.

7. Implement `parseTypeAliasDecl()` and `parseConstDecl()`.

8. Implement `parseVarDecl()`, `parseReturnStmt()`, `parseAssignOrExprStmt()`.

9. Implement `parseIfStmt()` with elif/else chains.

10. Implement `parseForStmt()`, `parseWhileStmt()`.

11. Implement `parseMatchStmt()` and `parseMatchArm()`.

12. Implement `parseDeferStmt()`, `parseUnsafeBlock()`, `parseArenaBlock()`.

13. Stub `parseExpr()` and `parseTypeExpr()` to return 0 (implemented in p03-t05 and p03-t06).

14. Create `compiler/parser/parser_test.go` with basic parsing tests.

## Test Plan

Write `compiler/parser/parser_test.go`:

```go
func TestParseFuncDecl(t *testing.T) {
    toks, _, _ := lexer.Lex([]byte("fn main():\n    return\n"))
    pool := ast.NewInternPool(16)
    tree, diags := parser.Parse(toks, nil, pool)
    requireNoErrors(t, diags)
    children := tree.Children(0) // root's children
    if len(children) != 1 { t.Fatalf("expected 1 top-level decl, got %d", len(children)) }
    if tree.Node(children[0]).Kind != ast.NodeFuncDecl {
        t.Errorf("expected FuncDecl, got %s", tree.Node(children[0]).Kind)
    }
}

func TestParsePubFuncDecl(t *testing.T) {
    toks, _, _ := lexer.Lex([]byte("pub fn foo():\n    return\n"))
    pool := ast.NewInternPool(16)
    tree, diags := parser.Parse(toks, nil, pool)
    requireNoErrors(t, diags)
    fn := tree.Node(tree.Children(0)[0])
    if fn.Flags&ast.FlagIsPub == 0 { t.Error("expected FlagIsPub") }
}

func TestParseImportDecl(t *testing.T) {
    toks, _, _ := lexer.Lex([]byte("import std.io\n"))
    pool := ast.NewInternPool(16)
    tree, diags := parser.Parse(toks, nil, pool)
    requireNoErrors(t, diags)
    node := tree.Node(tree.Children(0)[0])
    if node.Kind != ast.NodeImportDecl { t.Error("expected ImportDecl") }
}

func TestParseVarDecl(t *testing.T) {
    src := "fn main():\n    let x: i32 = 42\n"
    toks, _, _ := lexer.Lex([]byte(src))
    pool := ast.NewInternPool(16)
    tree, diags := parser.Parse(toks, []byte(src), pool)
    requireNoErrors(t, diags)
    // Navigate to the VarDecl inside the function body
    fn := tree.Children(0)[0]
    block := tree.Children(fn)[0] // first child should be Block (after return type)
    // ... verify VarDecl node
}

func TestParseIfStmt(t *testing.T) {
    src := "fn f():\n    if x:\n        y()\n    else:\n        z()\n"
    toks, _, _ := lexer.Lex([]byte(src))
    pool := ast.NewInternPool(16)
    tree, diags := parser.Parse(toks, []byte(src), pool)
    requireNoErrors(t, diags)
    // Verify IfStmt with ElseClause is in tree
    _ = tree
}

func TestParseStructDecl(t *testing.T) {
    src := "struct Point:\n    x: f64\n    y: f64\n"
    toks, _, _ := lexer.Lex([]byte(src))
    pool := ast.NewInternPool(16)
    tree, diags := parser.Parse(toks, []byte(src), pool)
    requireNoErrors(t, diags)
    node := tree.Node(tree.Children(0)[0])
    if node.Kind != ast.NodeStructDecl { t.Error("expected StructDecl") }
    // Verify 2 FieldDecl children
}
```

## Validation Checklist
- [ ] `Parse()` always returns non-nil tree
- [ ] Root node is always NodeProgram (index 0)
- [ ] FuncDecl parsed with name intern ID in Payload
- [ ] `pub` modifier sets FlagIsPub
- [ ] `async` modifier sets FlagIsAsync
- [ ] Import path stored as concatenated intern string
- [ ] VarDecl differentiates `let` (no FlagIsMut) from `mut` (FlagIsMut)
- [ ] IfStmt includes elif and else chains as children
- [ ] StructDecl includes FieldDecl and method children
- [ ] `parseExpr()` stub returns 0 (will be replaced in p03-t05)
- [ ] `go test ./compiler/parser/` passes

## Acceptance Criteria
- All 6 test functions pass
- `Parse()` never panics on any valid grammar input
- FuncDecl, ImportDecl, StructDecl, VarDecl, IfStmt all produce correct node kinds
- `go test -race ./compiler/parser/` passes

## Definition of Done
- [ ] `compiler/parser/parser.go` committed with all statement parse methods
- [ ] Stub for `parseExpr()` returns 0 (documented as stub)
- [ ] `compiler/parser/parser_test.go` committed
- [ ] All tests pass
- [ ] Lint passes

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| `parseExpr()` stub causes tests to produce empty expression nodes | Acceptable in this phase; mark as TODO in code |
| `peek()` skipping NEWLINEs incorrectly in some contexts | Use `peekRaw()` when NEWLINEs are significant (inside block) |
| Infinite loop if parser forgets to consume a token | Add a progress check: if `p.pos` hasn't advanced after 1000 iterations, break |
| `parsePubDecl()` needs to look ahead past `pub` | Peek at token after `pub` to dispatch to fn/struct/interface/type/const |

## Future Follow-up Tasks
- p03-t05: Implement `parseExpr()` via Pratt parser
- p03-t06: Implement `parseBlock()` with INDENT/DEDENT handling
- p03-t07: Error recovery via sync tokens
- p03-t08: Parser golden tests validate full parse output
