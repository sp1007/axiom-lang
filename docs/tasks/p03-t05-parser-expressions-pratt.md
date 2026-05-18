# p03-t05: Parser — Expressions (Pratt)

## Purpose
Implement a Pratt (top-down operator precedence) parser for all AXIOM expression forms in `compiler/parser/pratt.go`. The Pratt parser elegantly handles operator precedence and right-associativity without left-recursion, and extends cleanly when new operators are added. This task implements the full `parseExpr()` function that was stubbed in p03-t04, covering all binary operators, unary operators, postfix operations (field access, indexing, function calls, casts, deref), and all atomic expression forms.

## Context
Pratt parsing uses two tables: a Null Denotation (NUD) table for prefix/atomic forms and a Left Denotation (LED) table for infix/postfix forms. Each token has a "binding power" that determines how tightly it binds with adjacent expressions. The algorithm is:
1. Call NUD for the current token (parses an atom or prefix expression)
2. While the next token's binding power > current minimum, call LED (parses infix/postfix)
This produces a right-deep tree for right-associative operators (`**`) and left-deep for left-associative (everything else). The Pratt table for AXIOM is defined explicitly per the EBNF grammar (p01-t02).

## Inputs
- `compiler/parser/parser.go` from p03-t04 — `Parser` struct and helpers
- `compiler/ast/tree.go` from p03-t01 — builder methods
- `compiler/lexer/token_kind.go` from p02-t01 — all token kinds
- `docs/GRAMMAR.ebnf` from p01-t02 — expression precedence table

## Outputs
- `compiler/parser/pratt.go` — `parseExpr()`, `parseExprWithPrec()`, NUD/LED dispatch
- `compiler/parser/pratt_test.go` — expression parsing tests

## Dependencies
- p03-t04: parser-statements — `Parser` struct defined here
- p03-t01: ast-node-definitions
- p01-t02: grammar-ebnf — precedence table

## Subsystems Affected
- `compiler/parser/`: Pratt parser lives alongside the statement parser

## Detailed Requirements

1. **Binding power table** — defined as constants:
   ```go
   // Binding powers (left binding power for infix, 0 for non-infix)
   const (
       bpNone    = 0   // non-operator
       bpOr      = 10  // or
       bpAnd     = 20  // and
       bpNot     = 30  // not (prefix — but used in comparisons too)
       bpCmp     = 40  // == != < > <= >=
       bpBitOr   = 50  // |
       bpBitXor  = 60  // ^
       bpBitAnd  = 70  // &
       bpShift   = 80  // << >>
       bpAdd     = 90  // + -
       bpMul     = 100 // * / %
       bpPower   = 110 // **  (right-associative: LED uses bpPower-1)
       bpUnary   = 120 // unary - ~ not
       bpPostfix = 130 // . .* [] () as spawn await
   )
   ```

2. **`leftBindingPower(kind lexer.TokenKind) int`** — lookup table:
   ```go
   func leftBindingPower(kind lexer.TokenKind) int {
       switch kind {
       case lexer.TokenOr:             return bpOr
       case lexer.TokenAnd:            return bpAnd
       case lexer.TokenEqEq, lexer.TokenBangEq,
            lexer.TokenLt, lexer.TokenGt,
            lexer.TokenLtEq, lexer.TokenGtEq: return bpCmp
       case lexer.TokenPipe:           return bpBitOr
       case lexer.TokenCaret:          return bpBitXor
       case lexer.TokenAmp:            return bpBitAnd
       case lexer.TokenLtLt, lexer.TokenGtGt: return bpShift
       case lexer.TokenPlus, lexer.TokenMinus: return bpAdd
       case lexer.TokenStar, lexer.TokenSlash, lexer.TokenPercent: return bpMul
       case lexer.TokenStarStar:       return bpPower
       // Postfix
       case lexer.TokenDot, lexer.TokenDotStar,
            lexer.TokenLBracket, lexer.TokenLParen,
            lexer.TokenAs:             return bpPostfix
       default:                        return bpNone
       }
   }
   ```
   Note: `lexer.TokenAs` needs to be added to the keyword list (p02-t01) — it's used for casts.

3. **`parseExpr()` entry point** — calls `parseExprWithPrec(0)`:
   ```go
   func (p *Parser) parseExpr() uint32 {
       return p.parseExprWithPrec(bpNone)
   }
   ```

4. **`parseExprWithPrec(minBP int) uint32`** — core Pratt loop:
   ```go
   func (p *Parser) parseExprWithPrec(minBP int) uint32 {
       // NUD phase: parse prefix or atom
       left := p.parseNUD()
       if left == 0 { return 0 }

       // LED phase: parse infix/postfix while binding power allows
       for {
           tok := p.peek()
           bp := leftBindingPower(tok.Kind)
           if bp <= minBP { break }

           left = p.parseLED(left, tok, bp)
           if left == 0 { break }
       }
       return left
   }
   ```

5. **`parseNUD()` — null denotation (prefix/atom)**:
   ```go
   func (p *Parser) parseNUD() uint32 {
       tok := p.peek()
       switch tok.Kind {
       case lexer.TokenIdent:
           p.consume()
           node := p.tree.AddNode(ast.NodeIdent, p.tokenIdx(tok))
           nameID := p.pool.Intern(p.tokenText(tok))
           p.tree.SetPayload(node, nameID)
           return node
       case lexer.TokenIntLit:
           p.consume()
           return p.tree.AddNode(ast.NodeIntLit, p.tokenIdx(tok))
       case lexer.TokenFloatLit:
           p.consume()
           return p.tree.AddNode(ast.NodeFloatLit, p.tokenIdx(tok))
       case lexer.TokenStringLit:
           p.consume()
           return p.tree.AddNode(ast.NodeStringLit, p.tokenIdx(tok))
       case lexer.TokenCharLit:
           p.consume()
           return p.tree.AddNode(ast.NodeCharLit, p.tokenIdx(tok))
       case lexer.TokenTrue, lexer.TokenFalse:
           p.consume()
           return p.tree.AddNode(ast.NodeBoolLit, p.tokenIdx(tok))
       case lexer.TokenNil:
           p.consume()
           return p.tree.AddNode(ast.NodeNilLit, p.tokenIdx(tok))
       case lexer.TokenMinus, lexer.TokenTilde:
           p.consume()
           node := p.tree.AddNode(ast.NodeUnaryExpr, p.tokenIdx(tok))
           operand := p.parseExprWithPrec(bpUnary)
           if operand != 0 { p.tree.AppendChild(node, operand) }
           return node
       case lexer.TokenNot:
           p.consume()
           node := p.tree.AddNode(ast.NodeUnaryExpr, p.tokenIdx(tok))
           operand := p.parseExprWithPrec(bpNot)
           if operand != 0 { p.tree.AppendChild(node, operand) }
           return node
       case lexer.TokenLParen:
           return p.parseParenExpr()
       case lexer.TokenLBracket:
           return p.parseArrayLit()
       case lexer.TokenSpawn:
           return p.parseSpawnExpr()
       case lexer.TokenAwait:
           return p.parseAwaitExpr()
       case lexer.TokenPipe:
           return p.parseClosureExpr()
       default:
           p.errorf(tok, "expected expression, got %s", tok.Kind)
           return 0
       }
   }
   ```

6. **`parseLED(left uint32, opTok lexer.Token, bp int) uint32`** — left denotation (infix/postfix):
   ```go
   func (p *Parser) parseLED(left uint32, opTok lexer.Token, bp int) uint32 {
       p.consume() // consume the operator token
       switch opTok.Kind {
       // Binary operators
       case lexer.TokenOr, lexer.TokenAnd,
            lexer.TokenEqEq, lexer.TokenBangEq,
            lexer.TokenLt, lexer.TokenGt, lexer.TokenLtEq, lexer.TokenGtEq,
            lexer.TokenPipe, lexer.TokenCaret, lexer.TokenAmp,
            lexer.TokenLtLt, lexer.TokenGtGt,
            lexer.TokenPlus, lexer.TokenMinus,
            lexer.TokenStar, lexer.TokenSlash, lexer.TokenPercent:
           node := p.tree.AddNode(ast.NodeBinaryExpr, p.tokenIdx(opTok))
           p.tree.AppendChild(node, left)
           right := p.parseExprWithPrec(bp) // left-associative: same bp
           if right != 0 { p.tree.AppendChild(node, right) }
           return node
       case lexer.TokenStarStar: // right-associative
           node := p.tree.AddNode(ast.NodeBinaryExpr, p.tokenIdx(opTok))
           p.tree.AppendChild(node, left)
           right := p.parseExprWithPrec(bp - 1) // right-assoc: bp-1
           if right != 0 { p.tree.AppendChild(node, right) }
           return node
       // Postfix: field access
       case lexer.TokenDot:
           tok := p.peek()
           fieldTok, _ := p.expect(lexer.TokenIdent)
           node := p.tree.AddNode(ast.NodeFieldExpr, p.tokenIdx(tok))
           p.tree.AppendChild(node, left)
           fieldName := p.tree.AddNode(ast.NodeIdent, p.tokenIdx(fieldTok))
           p.tree.AppendChild(node, fieldName)
           return node
       case lexer.TokenDotStar: // deref: expr.*
           node := p.tree.AddNode(ast.NodeDerefExpr, p.tokenIdx(opTok))
           p.tree.AppendChild(node, left)
           return node
       // Postfix: index
       case lexer.TokenLBracket:
           node := p.tree.AddNode(ast.NodeIndexExpr, p.tokenIdx(opTok))
           p.tree.AppendChild(node, left)
           idx := p.parseExpr()
           if idx != 0 { p.tree.AppendChild(node, idx) }
           p.expect(lexer.TokenRBracket)
           return node
       // Postfix: function call
       case lexer.TokenLParen:
           return p.parseCallArgs(left, opTok)
       // Cast: expr as TypeExpr
       case lexer.TokenAs:
           node := p.tree.AddNode(ast.NodeCastExpr, p.tokenIdx(opTok))
           p.tree.AppendChild(node, left)
           typeNode := p.parseTypeExpr()
           if typeNode != 0 { p.tree.AppendChild(node, typeNode) }
           return node
       default:
           p.errorf(opTok, "unexpected infix operator %s", opTok.Kind)
           return left
       }
   }
   ```

7. **`parseCallArgs(callee uint32, lparen lexer.Token) uint32`**:
   ```go
   func (p *Parser) parseCallArgs(callee uint32, lparen lexer.Token) uint32 {
       node := p.tree.AddNode(ast.NodeCallExpr, p.tokenIdx(lparen))
       p.tree.AppendChild(node, callee) // first child is the callee
       for !p.check(lexer.TokenRParen) && !p.check(lexer.TokenEOF) {
           // Named arg: ident: expr
           if p.check(lexer.TokenIdent) && p.peekAt(1).Kind == lexer.TokenColon {
               argNode := p.tree.AddNode(ast.NodeNamedArg, p.tokenIdx(p.peek()))
               nameTok := p.consume()
               p.consume() // :
               nameID := p.pool.Intern(p.tokenText(nameTok))
               p.tree.SetPayload(argNode, nameID)
               expr := p.parseExpr()
               if expr != 0 { p.tree.AppendChild(argNode, expr) }
               p.tree.AppendChild(node, argNode)
           } else {
               expr := p.parseExpr()
               if expr != 0 { p.tree.AppendChild(node, expr) }
           }
           if !p.check(lexer.TokenRParen) { p.expect(lexer.TokenComma) }
       }
       p.expect(lexer.TokenRParen)
       return node
   }
   ```

8. **`parseArrayLit()`**:
   ```go
   func (p *Parser) parseArrayLit() uint32 {
       tok := p.consume() // [
       node := p.tree.AddNode(ast.NodeArrayLit, p.tokenIdx(tok))
       for !p.check(lexer.TokenRBracket) && !p.check(lexer.TokenEOF) {
           elem := p.parseExpr()
           if elem != 0 { p.tree.AppendChild(node, elem) }
           if !p.check(lexer.TokenRBracket) { p.expect(lexer.TokenComma) }
       }
       p.expect(lexer.TokenRBracket)
       return node
   }
   ```

9. **`parseParenExpr()`** — handles `(expr)` and struct literals (if followed by `{`):
   ```go
   func (p *Parser) parseParenExpr() uint32 {
       p.consume() // (
       expr := p.parseExpr()
       p.expect(lexer.TokenRParen)
       return expr // no node wrapping: paren is just for grouping
   }
   ```

10. **`parseSpawnExpr()` and `parseAwaitExpr()`**:
    ```go
    func (p *Parser) parseSpawnExpr() uint32 {
        tok, _ := p.expect(lexer.TokenSpawn)
        node := p.tree.AddNode(ast.NodeSpawnExpr, p.tokenIdx(tok))
        expr := p.parseExprWithPrec(bpPostfix)
        if expr != 0 { p.tree.AppendChild(node, expr) }
        return node
    }

    func (p *Parser) parseAwaitExpr() uint32 {
        tok, _ := p.expect(lexer.TokenAwait)
        node := p.tree.AddNode(ast.NodeAwaitExpr, p.tokenIdx(tok))
        expr := p.parseExprWithPrec(bpPostfix)
        if expr != 0 { p.tree.AppendChild(node, expr) }
        return node
    }
    ```

11. **`parseClosureExpr()`** — `|params| body`:
    ```go
    func (p *Parser) parseClosureExpr() uint32 {
        tok, _ := p.expect(lexer.TokenPipe)
        node := p.tree.AddNode(ast.NodeClosureExpr, p.tokenIdx(tok))
        for !p.check(lexer.TokenPipe) && !p.check(lexer.TokenEOF) {
            param := p.parseParam()
            if param != 0 { p.tree.AppendChild(node, param) }
            if !p.check(lexer.TokenPipe) { p.expect(lexer.TokenComma) }
        }
        p.expect(lexer.TokenPipe)
        // Body: block or expression
        if p.check(lexer.TokenColon) {
            body := p.parseBlock()
            if body != 0 { p.tree.AppendChild(node, body) }
        } else {
            body := p.parseExpr()
            if body != 0 { p.tree.AppendChild(node, body) }
        }
        return node
    }
    ```

12. **`peekAt(offset int) lexer.Token`** — look ahead by more than 1:
    ```go
    func (p *Parser) peekAt(offset int) lexer.Token {
        pos := p.pos + offset
        for pos < len(p.tokens) && p.tokens[pos].Kind == lexer.TokenNewline {
            pos++
        }
        if pos >= len(p.tokens) { return lexer.Token{Kind: lexer.TokenEOF} }
        return p.tokens[pos]
    }
    ```

13. **Struct literal disambiguation**: After parsing an Ident, if the next token is `{`, parse as struct literal (only in expression context, not statement context). Use a boolean flag `p.allowStructLit` to distinguish:
    - In `if cond:` context: `{` starts a dict/struct literal → allowed
    - At statement level: `Foo { ... }` as a statement → struct literal expression statement
    This is a known parsing challenge; document the resolution in code.

## Implementation Steps

1. Create `compiler/parser/pratt.go` with all constants, functions, and methods.

2. Wire `parseExpr()` into `compiler/parser/parser.go` — replace the stub with the real implementation.

3. Add `TokenAs` to `compiler/lexer/token_kind.go` if not already present (p02-t01 should have it; verify).

4. Create `compiler/parser/pratt_test.go` with all tests below.

5. Run `go test ./compiler/parser/` — all tests pass.

## Test Plan

Write `compiler/parser/pratt_test.go`:

```go
func TestParseBinaryAdd(t *testing.T) {
    tree, diags := parseExprHelper(t, "1 + 2")
    requireNoErrors(t, diags)
    root := tree.Children(0)[0] // single expression at root
    n := tree.Node(root)
    if n.Kind != ast.NodeBinaryExpr { t.Fatalf("expected BinaryExpr, got %s", n.Kind) }
    if len(tree.Children(root)) != 2 { t.Fatal("expected 2 children") }
}

func TestParsePrecedenceAddVsMul(t *testing.T) {
    // 1 + 2 * 3 should parse as 1 + (2 * 3)
    tree, _ := parseExprHelper(t, "1 + 2 * 3")
    root := tree.Children(0)[0]
    // root is BinaryExpr(+)
    // right child of + is BinaryExpr(*)
    addNode := tree.Node(root)
    if addNode.Kind != ast.NodeBinaryExpr { t.Fatal("outer is not BinaryExpr") }
    children := tree.Children(root)
    rightChild := tree.Node(children[1])
    if rightChild.Kind != ast.NodeBinaryExpr { t.Fatal("right child is not BinaryExpr (mul)") }
}

func TestParsePowerRightAssoc(t *testing.T) {
    // 2 ** 3 ** 4 should parse as 2 ** (3 ** 4)
    tree, _ := parseExprHelper(t, "2 ** 3 ** 4")
    root := tree.Children(0)[0]
    children := tree.Children(root)
    if len(children) != 2 { t.Fatal("expected 2 children for **") }
    rightChild := tree.Node(children[1])
    if rightChild.Kind != ast.NodeBinaryExpr { t.Fatal("** is not right-associative") }
}

func TestParseUnaryNeg(t *testing.T) {
    tree, _ := parseExprHelper(t, "-x")
    root := tree.Children(0)[0]
    if tree.Node(root).Kind != ast.NodeUnaryExpr { t.Fatal("expected UnaryExpr") }
}

func TestParseFieldExpr(t *testing.T) {
    tree, _ := parseExprHelper(t, "x.y.z")
    // x.y.z → FieldExpr(FieldExpr(x, y), z)
    root := tree.Children(0)[0]
    if tree.Node(root).Kind != ast.NodeFieldExpr { t.Fatal("expected FieldExpr") }
}

func TestParseCallExpr(t *testing.T) {
    tree, diags := parseExprHelper(t, "f(1, 2, 3)")
    requireNoErrors(t, diags)
    root := tree.Children(0)[0]
    if tree.Node(root).Kind != ast.NodeCallExpr { t.Fatal("expected CallExpr") }
    children := tree.Children(root)
    if len(children) != 4 { t.Fatalf("expected 4 children (callee+3 args), got %d", len(children)) }
}

func TestParseNamedArgCall(t *testing.T) {
    tree, diags := parseExprHelper(t, "f(x: 1, y: 2)")
    requireNoErrors(t, diags)
    root := tree.Children(0)[0]
    children := tree.Children(root)
    // callee + 2 NamedArg nodes
    if tree.Node(children[1]).Kind != ast.NodeNamedArg { t.Fatal("expected NamedArg") }
}

func TestParseIndexExpr(t *testing.T) {
    tree, _ := parseExprHelper(t, "arr[i]")
    root := tree.Children(0)[0]
    if tree.Node(root).Kind != ast.NodeIndexExpr { t.Fatal("expected IndexExpr") }
}

func TestParseCastExpr(t *testing.T) {
    tree, _ := parseExprHelper(t, "x as i64")
    root := tree.Children(0)[0]
    if tree.Node(root).Kind != ast.NodeCastExpr { t.Fatal("expected CastExpr") }
}

func TestParseSpawnExpr(t *testing.T) {
    tree, _ := parseExprHelper(t, "spawn f()")
    root := tree.Children(0)[0]
    if tree.Node(root).Kind != ast.NodeSpawnExpr { t.Fatal("expected SpawnExpr") }
}

func TestParseArrayLit(t *testing.T) {
    tree, _ := parseExprHelper(t, "[1, 2, 3]")
    root := tree.Children(0)[0]
    if tree.Node(root).Kind != ast.NodeArrayLit { t.Fatal("expected ArrayLit") }
    if len(tree.Children(root)) != 3 { t.Fatal("expected 3 elements") }
}

func TestParseComplexExpr(t *testing.T) {
    // not x and y or z == 1
    tree, _ := parseExprHelper(t, "not x and y or z == 1")
    // Should parse as: (not x and y) or (z == 1)
    root := tree.Children(0)[0]
    if tree.Node(root).Kind != ast.NodeBinaryExpr { t.Fatal("expected BinaryExpr at top") }
}
```

Helper `parseExprHelper` wraps a full parse of `fn f():\n    <expr>\n`.

## Validation Checklist
- [ ] `leftBindingPower` returns correct values for all operators
- [ ] `**` is right-associative (2**3**4 parses as 2**(3**4))
- [ ] `or` has lower precedence than `and`
- [ ] `not` has higher precedence than `and`
- [ ] Field access chains left-associate: `a.b.c` → `(a.b).c`
- [ ] Function call parses callee as first child
- [ ] Named args produce `NodeNamedArg` children
- [ ] `as` cast produces `NodeCastExpr`
- [ ] Array literal `[1,2,3]` produces `NodeArrayLit` with 3 children
- [ ] `spawn f()` produces `NodeSpawnExpr` with `NodeCallExpr` child
- [ ] `go test ./compiler/parser/` passes all tests

## Acceptance Criteria
- All 12 test functions pass
- `2 ** 3 ** 4` parses right-associatively
- `1 + 2 * 3` parses with `*` binding tighter than `+`
- `not x and y` parses as `(not x) and y` (not applied to x, then and)
- `go test -race ./compiler/parser/` passes

## Definition of Done
- [ ] `compiler/parser/pratt.go` committed
- [ ] `compiler/parser/pratt_test.go` committed
- [ ] Stub `parseExpr()` in parser.go replaced with real call
- [ ] All tests pass
- [ ] Lint passes

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| `as` keyword conflicts with identifier | Add `TokenAs` to keyword list in p02-t01; verify no programs use `as` as variable name |
| Struct literal `Foo{...}` ambiguous after `if` | Use `allowStructLit` flag; document disambiguation strategy |
| Closure `|params| expr` conflicts with `|` bitwise-or | Parse `|` as closure NUD only at expression start; as LED it's bitwise-or |
| Stack overflow on deeply nested expressions | Go default goroutine stack grows; practical programs won't hit Go's limit |

## Future Follow-up Tasks
- p03-t06: `parseBlock()` uses `parseExpr()` for expression statements
- p03-t07: Error recovery applies in `parseNUD()` and `parseLED()` on bad tokens
- p03-t08: Parser golden tests validate expression trees for all operators
