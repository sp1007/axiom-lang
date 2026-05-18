# p04-t03: Lazy Field Analysis

## Purpose
Implement the lazy symbol resolution strategy for imported modules. When `import X` is encountered, only register the module name `X` as a symbol — do NOT resolve any of `X`'s exported members. Only when `X.field` is first accessed in the code should `field` be resolved and loaded into the symbol table. This avoids loading unused modules and enables fast incremental compilation.

## Context
AXIOM uses a Zig-inspired lazy analysis model. In traditional compilers, `import X` eagerly loads all of X's public symbols. In AXIOM, imports are lazy: `import math` registers only `math` as a `SymModule` symbol. When `math.sqrt` is first referenced, the resolver loads `sqrt` from `math`'s module declaration and registers it. This is critical for: (1) compile speed — unused imports have zero cost, (2) future incremental compilation — only recheck modules whose accessed symbols changed, (3) detecting unused imports at compile time.

**Spec references:** `docs/plan.md` Phase 2 — Lazy Field Analysis, cycle detection via `Resolving` flag

## Inputs
- `compiler/sema/symtable.go` — SymbolTable with PushScope/Define/Resolve (p04-t01)
- `compiler/ast/ast.go` — FlatAST with ImportDecl and FieldAccess nodes (p03-t01)
- `compiler/types/typetable.go` — TypeTable for registering imported types (p04-t02)

## Outputs
- `compiler/sema/lazy.go` — LazyResolver with on-demand field loading, cycle detection
- `compiler/sema/module.go` — ModuleInfo struct, module registry
- `compiler/sema/lazy_test.go` — ≥12 test cases

## Dependencies
- p04-t01: symbol-table — SymbolTable stores module and field symbols
- p04-t02: type-table-primitives — TypeTable for imported type entries
- p03-t04: parser-statements — ImportDecl AST nodes

## Subsystems Affected
- Name resolver (p04-t04): calls LazyResolver when encountering `X.field`
- Type checker (p04-t06): type-checks lazily resolved symbols
- Incremental compilation (p17-t05): leverages lazy analysis for cache invalidation

## Detailed Requirements

### ModuleInfo
```go
type ModuleStatus uint8
const (
    ModuleUnloaded ModuleStatus = iota  // import seen, not yet accessed
    ModuleLoading                        // currently resolving (cycle detection)
    ModuleLoaded                         // all accessed fields resolved
)

type ModuleInfo struct {
    NameID    uint32         // interned module name
    Status    ModuleStatus
    FilePath  string         // source file path (for multi-file projects)
    Exports   map[uint32]uint32  // nameID → symbolIdx of exported symbols
    AstRoot   uint32         // root AST node index of the module's file
}
```

### LazyResolver
```go
type LazyResolver struct {
    modules  map[uint32]*ModuleInfo  // nameID → module info
    symtable *SymbolTable
    types    *TypeTable
}

func NewLazyResolver(st *SymbolTable, tt *TypeTable) *LazyResolver

// RegisterImport: called when parser encounters `import X`
// Registers X as SymModule in current scope, creates ModuleInfo(Unloaded)
func (lr *LazyResolver) RegisterImport(nameID uint32, filePath string, astRoot uint32) (uint32, error)

// ResolveField: called when resolver encounters `X.field`
// If module X not yet loaded → load X's exports, mark ModuleLoading → ModuleLoaded
// Then look up field in module's exports
// Returns symbolIdx of the field, or error if not found
func (lr *LazyResolver) ResolveField(moduleNameID uint32, fieldNameID uint32) (uint32, *diagnostics.Diagnostic)

// Cycle detection: if module status is ModuleLoading when ResolveField is called,
// a circular import is detected → return error diagnostic
```

### Cycle Detection
- When `ResolveField` is called and `ModuleStatus == ModuleLoading`, a circular dependency is detected
- Error: `"circular import detected: module 'X' is already being resolved"`
- This prevents infinite recursion in cross-module references like `A imports B, B imports A`

### Unused Import Detection
- After type checking completes, scan all `ModuleInfo` entries
- Any module with `Status == ModuleUnloaded` → warning: `"unused import: 'X'"`
- Modules where `Status == ModuleLoaded` but only some exports accessed → no warning (partial use is fine)

## Implementation Steps

1. Create `compiler/sema/module.go` — ModuleStatus, ModuleInfo struct.
2. Create `compiler/sema/lazy.go`:
   - Implement `NewLazyResolver()`
   - Implement `RegisterImport()`: create ModuleInfo, define SymModule symbol
   - Implement `ResolveField()`: check status, load if needed, lookup field
   - Implement cycle detection via ModuleLoading status check
3. Create `compiler/sema/lazy_test.go`.

## Test Plan

1. `TestRegisterImport`: register import "math" → SymModule symbol created
2. `TestResolveField_Found`: register module with export "sqrt", resolve "math.sqrt" → found
3. `TestResolveField_NotFound`: resolve "math.nonexistent" → error diagnostic
4. `TestLazyLoading_OnDemand`: register import, verify ModuleUnloaded; resolve field → ModuleLoaded
5. `TestLazyLoading_NoEagerLoad`: register import, never access → stays ModuleUnloaded
6. `TestCycleDetection`: set module to ModuleLoading, call ResolveField → circular import error
7. `TestUnusedImportDetection`: register import never accessed → warning generated
8. `TestMultipleImports`: register 3 modules, access only 1 → others stay unloaded
9. `TestMultipleFields_SameModule`: resolve "math.sqrt" and "math.abs" → both found, loaded once
10. `TestResolveField_ModuleNotImported`: resolve field on unknown module → error
11. `TestImportShadowing`: import "X" in outer scope, import "X" in inner scope → inner shadows
12. `TestDeterminism`: same imports in same order → same symbol indices

## Validation Checklist
- [ ] Lazy loading: module not loaded until field accessed
- [ ] Cycle detection prevents infinite recursion
- [ ] Unused imports detected as warnings
- [ ] Module loaded exactly once (not re-loaded on second field access)
- [ ] `go test ./compiler/sema/ -run TestLazy` passes

## Acceptance Criteria
- All 12 tests pass
- Cycle detection catches A→B→A circular imports
- Unused imports generate warning diagnostics

## Definition of Done
- [ ] `compiler/sema/module.go` and `compiler/sema/lazy.go` implemented
- [ ] 12 tests passing
- [ ] No panics on any error path (all errors returned as Diagnostics)

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Cross-file resolution requires file I/O during type checking | Mock file loading in tests; real file loading added with multi-file support |
| Module re-entrancy during lazy resolution | ModuleLoading flag prevents re-entrant resolution |

## Future Follow-up Tasks
- p04-t04: name-resolver calls LazyResolver.ResolveField for dot-expressions
- p17-t05: incremental compilation uses module dependency graph
