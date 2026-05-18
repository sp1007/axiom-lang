# p03-t01: AST Node Definitions

## Purpose
Complete the AST module in `compiler/ast/` by implementing the `AstTree` container (which holds the flat `[]AstNode` array), the builder methods for constructing trees during parsing, and the extra-data storage for nodes that need more than what fits in the 24-byte `AstNode` struct. This module is the central data structure of the front end — every phase from the parser through to HIR lowering reads and writes nodes through this API.

## Context
The AXIOM AST uses a flat-array design: all nodes live in a single `[]AstNode` slice (not a pointer-based tree). Tree structure is encoded via `FirstChild` and `NextSibling` index fields. This design enables cache-friendly traversal, O(1) serialization, and avoids GC pressure from millions of small allocations. The `AstTree.Extras []uint32` array stores overflow data for nodes that need more than 4 fields (e.g., function declarations with many parameters). The root node is always index 0 (a `NodeProgram` node). The `Source []byte` field stores the raw source for zero-copy token text recovery.

Spec reference: `03. Thiết kế parser thực tế.md`.

## Inputs
- `compiler/ast/node.go` from p01-t03 — `AstNode`, `NodeKind`, `Flags` constants
- `compiler/lexer/token.go` from p01-t03 — `Token` struct
- `compiler/diagnostics/` from p01-t01

## Outputs
- `compiler/ast/tree.go` — `AstTree` struct and core navigation methods
- `compiler/ast/builder.go` — builder methods: `AddNode`, `SetFirstChild`, `SetNextSibling`, `SetPayload`, `AddExtra`, `SetFlags`
- `compiler/ast/visitor.go` — DFS traversal helpers (`WalkPreOrder`, `WalkPostOrder`, `WalkChildren`)
- `compiler/ast/tree_test.go` — unit tests
- `compiler/ast/visitor_test.go` — visitor tests

## Dependencies
- p01-t03: struct-layout-definitions — AstNode struct defined here

## Subsystems Affected
- `compiler/ast/`: Core data structure
- `compiler/parser/`: Parser creates nodes via builder methods
- `compiler/sema/`: Semantic analysis reads nodes via navigation methods
- `ir/`: IR builder reads typed nodes

## Detailed Requirements

1. **`AstTree` struct**:
   ```go
   // AstTree holds all AST nodes for a single compilation unit.
   // Nodes live in a flat slice; tree structure is encoded via index fields.
   // The zero value is not valid; use NewTree to create.
   type AstTree struct {
       Nodes  []AstNode // all nodes; index 0 is always NodeProgram
       Extras []uint32  // overflow storage for nodes with many sub-fields
       Source []byte    // original source bytes (zero-copy token text)
       Tokens []lexer.Token // token slice (for token text lookup)
   }
   ```

2. **`NewTree(source []byte, tokens []lexer.Token) *AstTree`**:
   ```go
   func NewTree(source []byte, tokens []lexer.Token) *AstTree {
       t := &AstTree{
           Nodes:  make([]AstNode, 0, 256),
           Extras: make([]uint32, 0, 64),
           Source: source,
           Tokens: tokens,
       }
       // Root node: NodeProgram at index 0
       t.Nodes = append(t.Nodes, AstNode{Kind: NodeProgram})
       return t
   }
   ```
   The root node (index 0) is always `NodeProgram`. Its `FirstChild` points to the first top-level declaration.

3. **`AddNode(kind NodeKind, tokenIdx uint32) uint32`** — append a node and return its index:
   ```go
   func (t *AstTree) AddNode(kind NodeKind, tokenIdx uint32) uint32 {
       idx := uint32(len(t.Nodes))
       t.Nodes = append(t.Nodes, AstNode{Kind: kind, TokenIdx: tokenIdx})
       return idx
   }
   ```

4. **`Node(idx uint32) *AstNode`** — get a pointer to a node by index:
   ```go
   func (t *AstTree) Node(idx uint32) *AstNode {
       return &t.Nodes[idx]
   }
   ```
   Note: returned pointer is invalidated if `t.Nodes` is reallocated. Callers must not cache pointers across `AddNode` calls; use indices instead.

5. **Tree structure builder methods**:
   ```go
   func (t *AstTree) SetFirstChild(parent, child uint32) {
       t.Nodes[parent].FirstChild = child
   }

   func (t *AstTree) SetNextSibling(node, sibling uint32) {
       t.Nodes[node].NextSibling = sibling
   }

   func (t *AstTree) SetPayload(node uint32, payload uint32) {
       t.Nodes[node].Payload = payload
   }

   func (t *AstTree) SetFlags(node uint32, flags uint16) {
       t.Nodes[node].Flags |= flags
   }

   func (t *AstTree) ClearFlags(node uint32, flags uint16) {
       t.Nodes[node].Flags &^= flags
   }
   ```

6. **`AddExtra(values ...uint32) uint32`** — append overflow data and return start index:
   ```go
   func (t *AstTree) AddExtra(values ...uint32) uint32 {
       idx := uint32(len(t.Extras))
       t.Extras = append(t.Extras, values...)
       return idx
   }
   ```
   A node uses `ExtraIdx` to point to its extra data, and a convention for how many extra words it has (typically the first word is a count).

7. **`TokenText(tokenIdx uint32) []byte`** — recover token text from source:
   ```go
   func (t *AstTree) TokenText(tokenIdx uint32) []byte {
       tok := t.Tokens[tokenIdx]
       return t.Source[tok.Offset : tok.Offset+uint32(tok.Len)]
   }
   ```

8. **`Children(nodeIdx uint32) []uint32`** — collect all child indices (for debugging/traversal):
   ```go
   func (t *AstTree) Children(nodeIdx uint32) []uint32 {
       var children []uint32
       child := t.Nodes[nodeIdx].FirstChild
       for child != 0 {
           children = append(children, child)
           child = t.Nodes[child].NextSibling
       }
       return children
   }
   ```
   Note: index 0 is the root (NodeProgram), so `child != 0` is used as the "no child" sentinel. This means index 0 can only be the root and must never be a child of another node.

9. **`AppendChild(parent, child uint32)`** — convenience: append a child at the end of the parent's child list:
   ```go
   func (t *AstTree) AppendChild(parent, child uint32) {
       if t.Nodes[parent].FirstChild == 0 {
           t.Nodes[parent].FirstChild = child
           return
       }
       // Walk to last sibling
       cur := t.Nodes[parent].FirstChild
       for t.Nodes[cur].NextSibling != 0 {
           cur = t.Nodes[cur].NextSibling
       }
       t.Nodes[cur].NextSibling = child
   }
   ```
   Note: This is O(n) in the number of children. For performance-critical paths (function parameters, struct fields), the parser should track the last child explicitly. Document this tradeoff.

10. **`NodeCount()` and `ExtraCount()` accessors**:
    ```go
    func (t *AstTree) NodeCount() int  { return len(t.Nodes) }
    func (t *AstTree) ExtraCount() int { return len(t.Extras) }
    ```

11. **Index sentinel convention**: `0` is the root (NodeProgram) AND the null/no-child sentinel. This works because no node ever has index 0 as a child — the root is only referenced as `tree.Nodes[0]` directly. Document this clearly:
    ```go
    // NullIdx is the sentinel value for "no node".
    // It also happens to be the root node index, which is intentional:
    // no node may be the child of another node AND the root simultaneously.
    const NullIdx uint32 = 0
    ```

12. **`Validate()` method** for debug builds — check tree invariants:
    ```go
    func (t *AstTree) Validate() []string {
        var errors []string
        if len(t.Nodes) == 0 {
            return []string{"tree has no nodes (missing root)"}
        }
        if t.Nodes[0].Kind != NodeProgram {
            errors = append(errors, "node[0] is not NodeProgram")
        }
        // Check all child/sibling indices are within bounds
        for i, n := range t.Nodes {
            if n.FirstChild != 0 && int(n.FirstChild) >= len(t.Nodes) {
                errors = append(errors, fmt.Sprintf("node[%d].FirstChild=%d out of bounds", i, n.FirstChild))
            }
            if n.NextSibling != 0 && int(n.NextSibling) >= len(t.Nodes) {
                errors = append(errors, fmt.Sprintf("node[%d].NextSibling=%d out of bounds", i, n.NextSibling))
            }
        }
        return errors
    }
    ```

## Implementation Steps

1. Create `compiler/ast/tree.go` with `AstTree` struct, `NullIdx` constant, `NewTree()`, and all navigation methods (Requirements 2–10, 12).

2. Create `compiler/ast/builder.go` with all builder methods (Requirements 3, 5, 6). Alternatively, keep builder methods in `tree.go` if the file is short enough (<300 lines).

3. Create `compiler/ast/tree_test.go` with tests:
   ```go
   func TestNewTreeRootNode(t *testing.T) {
       tree := NewTree(nil, nil)
       if tree.NodeCount() != 1 { t.Fatal("expected 1 node") }
       if tree.Nodes[0].Kind != NodeProgram { t.Fatal("root must be NodeProgram") }
   }

   func TestAddNode(t *testing.T) {
       tree := NewTree(nil, nil)
       idx := tree.AddNode(NodeFuncDecl, 0)
       if idx != 1 { t.Fatalf("expected idx=1, got %d", idx) }
       if tree.NodeCount() != 2 { t.Fatal("expected 2 nodes") }
   }

   func TestAppendChild(t *testing.T) {
       tree := NewTree(nil, nil)
       child1 := tree.AddNode(NodeFuncDecl, 0)
       child2 := tree.AddNode(NodeStructDecl, 0)
       tree.AppendChild(0, child1)
       tree.AppendChild(0, child2)
       children := tree.Children(0)
       if len(children) != 2 { t.Fatalf("expected 2 children, got %d", len(children)) }
       if children[0] != child1 { t.Errorf("first child = %d, want %d", children[0], child1) }
       if children[1] != child2 { t.Errorf("second child = %d, want %d", children[1], child2) }
   }

   func TestSetFlags(t *testing.T) {
       tree := NewTree(nil, nil)
       idx := tree.AddNode(NodeFuncDecl, 0)
       tree.SetFlags(idx, FlagIsPub|FlagIsAsync)
       n := tree.Node(idx)
       if n.Flags&FlagIsPub == 0 { t.Error("expected FlagIsPub set") }
       if n.Flags&FlagIsAsync == 0 { t.Error("expected FlagIsAsync set") }
       tree.ClearFlags(idx, FlagIsPub)
       if tree.Node(idx).Flags&FlagIsPub != 0 { t.Error("expected FlagIsPub cleared") }
   }

   func TestAddExtra(t *testing.T) {
       tree := NewTree(nil, nil)
       idx := tree.AddExtra(10, 20, 30)
       if idx != 0 { t.Fatalf("expected Extra idx=0, got %d", idx) }
       if tree.Extras[0] != 10 || tree.Extras[2] != 30 {
           t.Error("extra values not stored correctly")
       }
   }

   func TestValidate(t *testing.T) {
       tree := NewTree(nil, nil)
       errs := tree.Validate()
       if len(errs) != 0 { t.Fatalf("fresh tree has validation errors: %v", errs) }

       // Corrupt a child index
       child := tree.AddNode(NodeFuncDecl, 0)
       tree.Nodes[child].FirstChild = 9999 // out of bounds
       errs = tree.Validate()
       if len(errs) == 0 { t.Error("expected validation error for out-of-bounds child") }
   }

   func TestNullIdxIsSentinel(t *testing.T) {
       if NullIdx != 0 { t.Fatal("NullIdx must be 0") }
   }
   ```

4. Create `compiler/ast/visitor.go` with DFS traversal helpers:
   ```go
   // VisitFn is called for each node. Return false to stop traversal.
   type VisitFn func(tree *AstTree, nodeIdx uint32) bool

   // WalkPreOrder visits every reachable node in pre-order (parent before children).
   func WalkPreOrder(tree *AstTree, root uint32, visit VisitFn) { ... }

   // WalkPostOrder visits every reachable node in post-order (children before parent).
   func WalkPostOrder(tree *AstTree, root uint32, visit VisitFn) { ... }

   // WalkChildren visits only the direct children of the given node.
   func WalkChildren(tree *AstTree, parent uint32, visit VisitFn) { ... }

   // NodeCount returns the number of reachable nodes from root (for validation).
   func ReachableCount(tree *AstTree, root uint32) int { ... }
   ```

5. Create `compiler/ast/visitor_test.go` — verify traversal order, early termination, cycle safety.

6. Run `go test ./compiler/ast/` — all tests pass.

7. Verify `go build ./...` still passes.

## Test Plan
All tests described in Implementation Step 3. Additionally:
- **TestChildrenOrder**: verify `Children()` returns children in the order they were appended
- **TestTokenText**: verify `TokenText(idx)` returns correct source slice
- **TestLargeTree**: create 10,000 nodes, verify `Validate()` passes and no panics

## Validation Checklist
- [ ] `AstTree` struct has `Nodes`, `Extras`, `Source`, `Tokens` fields
- [ ] `NewTree()` creates root NodeProgram at index 0
- [ ] `AddNode()` returns correct index
- [ ] `AppendChild()` maintains child order
- [ ] `Children()` returns all children in order
- [ ] `SetFlags()` and `ClearFlags()` work correctly
- [ ] `AddExtra()` returns start index
- [ ] `NullIdx = 0` documented and used as sentinel
- [ ] `Validate()` catches out-of-bounds indices
- [ ] `go test ./compiler/ast/` passes

## Acceptance Criteria
- All tests pass
- `NewTree()` always creates a valid tree (Validate() returns no errors)
- AppendChild with 100 children maintains correct order
- `go test -race ./compiler/ast/` passes

## Definition of Done
- [ ] `compiler/ast/tree.go` committed
- [ ] `compiler/ast/builder.go` committed (or merged into tree.go)
- [ ] `compiler/ast/visitor.go` committed
- [ ] `compiler/ast/tree_test.go` committed
- [ ] `compiler/ast/visitor_test.go` committed
- [ ] All tests pass
- [ ] Lint passes

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| `Node()` pointer invalidated by slice growth | Document clearly; parser should use indices only, call Node() only at read time |
| `AppendChild()` O(n) for many children | Document tradeoff; parser tracks last child for hot paths |
| NullIdx=0 confusion (root vs sentinel) | Comment everywhere; `Validate()` enforces invariant |
| Extras slice growing unpredictably | Pre-allocate with `make([]uint32, 0, 64)`; grow automatically |

## Future Follow-up Tasks
- p03-t02: String intern pool uses AstTree source bytes
- p03-t03: AST printer reads AstTree nodes
- p03-t04: Parser creates nodes via AstTree builder methods
