# p03-t03: AST Printer

## Purpose
Implement an AST pretty-printer in `compiler/ast/printer.go` that renders the flat `AstTree` as a human-readable indented tree for debugging, testing, and the `axc dump-ast` command. The printer is a critical debugging tool — compiler engineers use it to verify the parser produced the correct tree structure before proceeding to semantic analysis. It is also used by golden tests (p03-t08) to produce the `.ast` expected output files.

## Context
The AST printer traverses the flat `[]AstNode` array using the `FirstChild`/`NextSibling` index links, recursively rendering each node with indentation. The output format shows node kind, flags, and the token text for nodes that have a primary token. The printer writes to an `io.Writer` for flexibility (stdout, string builder, file). It uses the `InternPool` to resolve identifier names. The output format must be stable across runs (deterministic) — all fields printed in a fixed order.

## Inputs
- `compiler/ast/tree.go` from p03-t01 — `AstTree`
- `compiler/ast/node.go` from p01-t03 — `NodeKind`, `Flags` constants
- `compiler/ast/intern.go` from p03-t02 — `InternPool`
- `compiler/lexer/token.go` from p01-t03 — `Token`

## Outputs
- `compiler/ast/printer.go` — `Printer` struct and `Print()` function
- `compiler/ast/printer_test.go` — unit tests with expected output strings

## Dependencies
- p03-t01: ast-node-definitions
- p03-t02: string-intern-pool
- p01-t03: struct-layout-definitions

## Subsystems Affected
- `compiler/ast/`: Printer lives here
- `cmd/axc/`: `axc dump-ast` command uses the printer (p03-t10)
- `compiler/parser/`: Golden tests (p03-t08) use the printer for expected output

## Detailed Requirements

1. **Output format**: Indented tree, one node per line. Example for `fn main(): println("hello")`:
   ```
   Program
     FuncDecl [pub=false async=false] name="main"
       ParamDecl name="args" type=
         SliceType
           Ident name="string"
       Block
         ExprStmt
           CallExpr
             Ident name="println"
             StringLit value="hello"
   ```
   - Indentation: 2 spaces per level (not 4, to keep output compact)
   - Node kind: from `NodeKind.String()` (implement this method)
   - Flags: printed as `[pub=true async=false ...]` for all flag fields
   - Token text: printed as `name="..."` for ident nodes, `value="..."` for literals

2. **`NodeKind.String()` method** in `compiler/ast/node.go`:
   ```go
   func (k NodeKind) String() string {
       if int(k) < len(nodeKindNames) {
           return nodeKindNames[k]
       }
       return fmt.Sprintf("NodeKind(%d)", k)
   }

   var nodeKindNames = [NodeKindCount]string{
       NodeInvalid:      "Invalid",
       NodeProgram:      "Program",
       NodeFuncDecl:     "FuncDecl",
       NodeStructDecl:   "StructDecl",
       // ... all kinds
   }
   ```

3. **`Printer` struct**:
   ```go
   type Printer struct {
       w      io.Writer
       tree   *AstTree
       pool   *InternPool // may be nil; identifiers shown as hex ID if nil
       indent int
   }
   ```

4. **Public `Print` function**:
   ```go
   // Print writes the AST rooted at nodeIdx to w.
   // pool may be nil (identifiers shown as raw token text without intern resolution).
   func Print(w io.Writer, tree *AstTree, pool *InternPool) {
       p := &Printer{w: w, tree: tree, pool: pool}
       p.printNode(0) // start from root (NodeProgram, index 0)
   }

   // PrintNode prints a single subtree rooted at nodeIdx.
   func PrintNode(w io.Writer, tree *AstTree, nodeIdx uint32, pool *InternPool) {
       p := &Printer{w: w, tree: tree, pool: pool}
       p.printNode(nodeIdx)
   }
   ```

5. **`printNode(idx uint32)` recursive method**:
   ```go
   func (p *Printer) printNode(idx uint32) {
       node := p.tree.Node(idx)
       p.writeIndent()
       p.writef("%s", node.Kind)
       p.writeFlags(node)
       p.writeTokenInfo(node)
       fmt.Fprintln(p.w)
       // Recurse into children
       p.indent++
       child := node.FirstChild
       for child != 0 {
           p.printNode(child)
           child = p.tree.Node(child).NextSibling
       }
       p.indent--
   }
   ```

6. **`writeFlags(node *AstNode)`**: Print non-zero flags in `[flag1 flag2 ...]` format:
   ```go
   func (p *Printer) writeFlags(node *AstNode) {
       type flagInfo struct { mask uint16; name string }
       flags := []flagInfo{
           {FlagIsPub, "pub"}, {FlagIsMut, "mut"}, {FlagIsAsync, "async"},
           {FlagIsExtern, "extern"}, {FlagIsSink, "sink"}, {FlagIsLent, "lent"},
           {FlagIsPacked, "packed"}, {FlagEscapesToHeap, "heap"},
           {FlagUsesArena, "arena"}, {FlagIsGeneric, "generic"}, {FlagIsMoved, "moved"},
       }
       var active []string
       for _, f := range flags {
           if node.Flags&f.mask != 0 {
               active = append(active, f.name)
           }
       }
       if len(active) > 0 {
           fmt.Fprintf(p.w, " [%s]", strings.Join(active, " "))
       }
   }
   ```

7. **`writeTokenInfo(node *AstNode)`**: Print token-based info per node kind:
   ```go
   func (p *Printer) writeTokenInfo(node *AstNode) {
       if node.TokenIdx == 0 && node.Kind != NodeProgram { return }
       text := p.tree.TokenText(node.TokenIdx)
       switch node.Kind {
       case NodeIdent, NodeFuncDecl, NodeStructDecl, NodeInterfaceDecl,
            NodeParamDecl, NodeFieldDecl, NodeVariantDecl, NodeTypeAliasDecl:
           fmt.Fprintf(p.w, " name=%q", text)
       case NodeIntLit, NodeFloatLit, NodeBoolLit, NodeNilLit:
           fmt.Fprintf(p.w, " value=%q", text)
       case NodeStringLit, NodeCharLit:
           // Show up to 40 chars to keep output readable
           s := string(text)
           if len(s) > 40 { s = s[:40] + "..." }
           fmt.Fprintf(p.w, " value=%q", s)
       case NodeBinaryExpr, NodeUnaryExpr, NodeAssignStmt:
           fmt.Fprintf(p.w, " op=%q", text)
       }
   }
   ```

8. **`writeIndent()`**:
   ```go
   func (p *Printer) writeIndent() {
       for i := 0; i < p.indent; i++ {
           fmt.Fprint(p.w, "  ")
       }
   }
   ```

9. **`writef()` helper**:
   ```go
   func (p *Printer) writef(format string, args ...any) {
       fmt.Fprintf(p.w, format, args...)
   }
   ```

10. **PrintToString convenience function**:
    ```go
    func PrintToString(tree *AstTree, pool *InternPool) string {
        var sb strings.Builder
        Print(&sb, tree, pool)
        return sb.String()
    }
    ```

11. **Payload display**: For nodes where `Payload` carries semantic information (TypeID, SymbolIdx), print it as a comment if non-zero:
    ```go
    if node.Payload != 0 {
        fmt.Fprintf(p.w, " @%d", node.Payload) // e.g., @42 for TypeID 42
    }
    ```
    This makes the output useful after type-checking.

12. **Determinism**: The printer must produce identical output for identical input across runs, platforms, and Go versions. No map iteration, no runtime addresses, no timestamps.

## Implementation Steps

1. Add `NodeKind.String()` method and `nodeKindNames` array to `compiler/ast/node.go`.

2. Create `compiler/ast/printer.go` with:
   - `Printer` struct
   - `Print()` public function
   - `PrintNode()` public function
   - `PrintToString()` convenience function
   - All private methods: `printNode`, `writeIndent`, `writeFlags`, `writeTokenInfo`, `writef`

3. Create `compiler/ast/printer_test.go`:
   ```go
   func TestPrinterEmptyProgram(t *testing.T) {
       tree := NewTree(nil, nil)
       got := PrintToString(tree, nil)
       if !strings.Contains(got, "Program") {
           t.Errorf("expected 'Program' in output, got:\n%s", got)
       }
   }

   func TestPrinterFuncDecl(t *testing.T) {
       tree := NewTree([]byte("fn main():"), nil)
       fnIdx := tree.AddNode(NodeFuncDecl, 0)
       tree.AppendChild(0, fnIdx) // attach to root
       tree.SetFlags(fnIdx, FlagIsPub)
       got := PrintToString(tree, nil)
       if !strings.Contains(got, "FuncDecl") { t.Error("expected FuncDecl") }
       if !strings.Contains(got, "[pub]") { t.Error("expected [pub] flag") }
   }

   func TestPrinterNestedBlocks(t *testing.T) {
       tree := NewTree(nil, nil)
       fn := tree.AddNode(NodeFuncDecl, 0)
       block := tree.AddNode(NodeBlock, 0)
       stmt := tree.AddNode(NodeReturnStmt, 0)
       tree.AppendChild(0, fn)
       tree.AppendChild(fn, block)
       tree.AppendChild(block, stmt)
       got := PrintToString(tree, nil)
       lines := strings.Split(strings.TrimSpace(got), "\n")
       // Program, FuncDecl (indent 1), Block (indent 2), ReturnStmt (indent 3)
       if len(lines) < 4 { t.Fatalf("expected ≥4 lines, got %d:\n%s", len(lines), got) }
       if !strings.HasPrefix(lines[1], "  FuncDecl") { t.Errorf("line 1: %q", lines[1]) }
       if !strings.HasPrefix(lines[2], "    Block")  { t.Errorf("line 2: %q", lines[2]) }
       if !strings.HasPrefix(lines[3], "      Ret")  { t.Errorf("line 3: %q", lines[3]) }
   }

   func TestPrinterDeterministic(t *testing.T) {
       tree := NewTree(nil, nil)
       for i := 0; i < 10; i++ {
           tree.AddNode(NodeFuncDecl, 0)
           tree.AppendChild(0, uint32(i+1))
       }
       out1 := PrintToString(tree, nil)
       out2 := PrintToString(tree, nil)
       if out1 != out2 { t.Error("printer is not deterministic") }
   }

   func TestNodeKindString(t *testing.T) {
       if NodeProgram.String() != "Program" {
           t.Errorf("NodeProgram.String() = %q, want %q", NodeProgram.String(), "Program")
       }
       if NodeFuncDecl.String() != "FuncDecl" {
           t.Errorf("NodeFuncDecl.String() = %q", NodeFuncDecl.String())
       }
   }
   ```

4. Run `go test ./compiler/ast/` — all tests pass.

## Test Plan
All tests in Implementation Step 3. Additionally:
- **TestPrinterAllNodeKinds**: Create one node of each kind, verify `String()` is non-empty
- **TestPrinterPayloadDisplay**: Set `Payload=42`, verify `@42` appears in output

## Validation Checklist
- [ ] `NodeKind.String()` returns non-empty for all defined kinds
- [ ] `Print()` produces indented output starting with "Program"
- [ ] Flags shown as `[pub async]` for active flags only
- [ ] Token text shown for Ident, Literal, and operator nodes
- [ ] Nested nodes indented correctly (2 spaces per level)
- [ ] `PrintToString()` returns same result as `Print(&strings.Builder{}, ...)`
- [ ] Output is deterministic across multiple calls
- [ ] `go test ./compiler/ast/` passes

## Acceptance Criteria
- `PrintToString` on a tree with 3-level nesting produces correct 2-space indentation
- All NodeKind constants have non-empty `String()` output
- Output is byte-for-byte identical on repeat calls (deterministic)

## Definition of Done
- [ ] `compiler/ast/printer.go` committed
- [ ] `NodeKind.String()` added to `compiler/ast/node.go`
- [ ] `compiler/ast/printer_test.go` committed
- [ ] All tests pass
- [ ] Lint passes

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Circular references in AST cause infinite loop | `Validate()` (p03-t01) detects cycles; add depth limit to printer (max 100 levels) |
| String literal text too long floods output | Truncate at 40 characters with `...` suffix |
| `nodeKindNames` array not updated when new kinds added | `TestPrinterAllNodeKinds` catches empty strings for undefined names |

## Future Follow-up Tasks
- p03-t08: Parser golden tests use `PrintToString` for `.ast` golden files
- p03-t10: `axc dump-ast` calls `Print(os.Stdout, tree, pool)`
- p04-t10: `axc dump-ast` extended to show type annotations after type checking
