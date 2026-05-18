# p04-t04: Name Resolver

## Purpose
Implement the name resolution pass that walks the FlatAST and resolves every identifier to its corresponding Symbol in the SymbolTable. After this pass, every `IdentExpr` and type reference AST node has its `Payload` field set to a valid symbol index. Unresolved identifiers produce diagnostics.

## Context
Name resolution is the bridge between parsing and type checking. The parser produces AST nodes with raw token indices; the name resolver converts those to symbol table references. Resolution follows AXIOM's lexical scoping rules: inner scopes shadow outer scopes, and function parameters are visible within the function body. Import resolution uses the LazyResolver for `X.field` dot-access patterns.

**Spec references:** `04. Type checker.md`, `docs/plan.md` Phase 2 — NameResolver

## Inputs
- `compiler/sema/symtable.go` — SymbolTable (p04-t01)
- `compiler/sema/lazy.go` — LazyResolver for imports (p04-t03)
- `compiler/types/typetable.go` — TypeTable (p04-t02)
- `compiler/ast/ast.go` — FlatAST with all node kinds (p03-t01)
- `compiler/ast/intern.go` — InternPool (p03-t02)

## Outputs
- `compiler/sema/resolver.go` — NameResolver pass implementation
- `compiler/sema/resolver_test.go` — ≥15 test cases

## Dependencies
- p04-t01: symbol-table
- p04-t02: type-table-primitives
- p04-t03: lazy-field-analysis
- p03-t04: parser-statements (AST structure)

## Subsystems Affected
- Type checker (p04-t05–t07): operates on resolved AST nodes
- Ownership checker (p06-t02): reads symbol flags set during resolution

## Detailed Requirements

### NameResolver Struct
```go
type NameResolver struct {
    ast      *ast.AstTree
    intern   *ast.InternPool
    symtable *SymbolTable
    types    *TypeTable
    lazy     *LazyResolver
    errors   []diagnostics.Diagnostic
}

func NewNameResolver(tree *ast.AstTree, intern *ast.InternPool,
    st *SymbolTable, tt *TypeTable, lr *LazyResolver) *NameResolver

// Resolve walks the entire AST and resolves all names.
// Returns collected diagnostics (errors + warnings).
func (nr *NameResolver) Resolve() []diagnostics.Diagnostic
```

### Resolution Rules

1. **Declarations** — register in current scope:
   - `FnDecl`: define function name as `SymFunc`, push `ScopeFunction`, define params as `SymParam`, resolve body, pop scope
   - `StructDecl`: define struct name as `SymStruct`, register in TypeTable, define fields as `SymField`
   - `InterfaceDecl`: define as `SymInterface`
   - `LetDecl`/`MutDecl`: define variable as `SymVar` (with `SymFlagMut` for mut)
   - `ImportDecl`: call `LazyResolver.RegisterImport()`
   - `TypeDecl`: define as `SymTypeAlias`

2. **References** — resolve and annotate AST node:
   - `IdentExpr`: call `SymbolTable.Resolve(nameID)` → set node.Payload = symbolIdx. Error if not found: `"undefined: 'X'"`
   - `FieldAccessExpr` (`X.field`): if X is a module → call `LazyResolver.ResolveField()`. If X is a struct instance → resolve field name against struct type's fields.
   - `TypeRef` (in type annotations like `: i32`): resolve via `SymbolTable.Resolve()` → must be `SymStruct`, `SymBuiltinType`, `SymTypeAlias`, or `SymGenericParam`

3. **Scope management**:
   - `FnDecl` → push `ScopeFunction`, pop after body
   - `IfStmt`/`ElifStmt`/`ElseStmt` → push `ScopeBlock`, pop after block
   - `ForStmt` → push `ScopeLoop` (loop variable defined in this scope), pop after body
   - `MatchArm` → push `ScopeBlock` (pattern bindings visible in arm body), pop
   - `ClosureExpr` → push `ScopeClosure`, pop
   - `LockStmt` → push `ScopeBlock` (bound variable visible in body), pop
   - `ArenaBlock` → push `ScopeBlock`, pop

4. **Error diagnostics**:
   - `"undefined: 'X'"` (code 2010) — identifier not found in any scope
   - `"'X' is not a type"` (code 2011) — identifier used as type but is SymVar/SymFunc
   - `"'X' is not a module"` (code 2012) — dot-access on non-module symbol
   - `"duplicate definition: 'X'"` (code 2001) — from SymbolTable.Define()
   - `"unused import: 'X'"` (code 2020, warning) — from LazyResolver

5. **AST mutation**: After resolution, each `IdentExpr` node has `Payload = symbolIdx`. Each `TypeRef` node has `Payload = TypeID`. The AST is modified in-place (payload field written).

## Implementation Steps

1. Create `compiler/sema/resolver.go`.
2. Implement `Resolve()` — top-level DFS walk of AST from root.
3. Implement `resolveNode(nodeIdx uint32)` — dispatch on NodeKind:
   - Declaration nodes → define in SymbolTable, recurse into children
   - Expression nodes → resolve identifiers, recurse into children
   - Block nodes → push/pop scopes around children
4. Implement identifier resolution: `resolveIdent(nodeIdx)` — lookup in SymbolTable, set Payload.
5. Implement type reference resolution: `resolveTypeRef(nodeIdx)` — lookup, validate is type symbol.
6. Implement field access resolution: `resolveFieldAccess(nodeIdx)` — check if base is module or struct.
7. Implement unused import detection: after full walk, check LazyResolver modules.
8. Create `compiler/sema/resolver_test.go`.

## Test Plan

1. `TestResolve_Variable`: `let x = 42; x` → x resolved to symbol
2. `TestResolve_Function`: `fn foo(): ...` → foo registered, callable
3. `TestResolve_Param`: `fn foo(a: i32): a` → a resolved within body
4. `TestResolve_Undefined`: `x` (never defined) → error "undefined: 'x'"
5. `TestResolve_Shadowing`: outer `x`, inner `x` → inner resolved in inner scope
6. `TestResolve_ScopeExit`: `if cond: let y = 1` then `y` outside → undefined
7. `TestResolve_TypeRef`: `let x: i32` → i32 resolved to TypeBool=11... TypeI32=3
8. `TestResolve_TypeRefInvalid`: `let x: foo` where foo is a variable → "not a type"
9. `TestResolve_StructFields`: `struct Point { x: i32 }` → fields registered
10. `TestResolve_FieldAccess`: `p.x` where p is Point → field resolved
11. `TestResolve_ImportLazy`: `import math; math.sqrt` → lazy resolution triggered
12. `TestResolve_DuplicateError`: `let x = 1; let x = 2` in same scope → error
13. `TestResolve_ForLoopVar`: `for i in 0..10: i` → i visible in loop body, not outside
14. `TestResolve_MatchBinding`: `match x: i32(v) => v` → v visible in arm
15. `TestResolve_UnusedImport`: `import math` never accessed → warning

## Validation Checklist
- [ ] Every IdentExpr has Payload set after resolution (or error emitted)
- [ ] Scopes pushed/popped correctly for all block constructs
- [ ] Shadowing works: inner overrides outer, restored after pop
- [ ] Undefined identifiers produce diagnostic, not panic
- [ ] Type references validated (must be type symbol)
- [ ] `go test ./compiler/sema/ -run TestResolve` passes

## Acceptance Criteria
- All 15 tests pass
- Parse+resolve `axiom_compliance_suite.ax` groups 1–3: zero "undefined" errors on valid code
- All error paths return diagnostics (no panics)

## Definition of Done
- [ ] `compiler/sema/resolver.go` implemented
- [ ] 15 tests passing
- [ ] Compliance suite groups 1–3 resolve without errors

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| AST DFS traversal misses node kinds | Test with full compliance suite; add catch-all for unknown NodeKind |
| Field access resolution ambiguous (module vs struct) | Check symbol kind first: SymModule → lazy resolve, SymVar → struct field |

## Future Follow-up Tasks
- p04-t05: type inference operates on resolved AST
- p04-t06: type checker validates types of resolved symbols
