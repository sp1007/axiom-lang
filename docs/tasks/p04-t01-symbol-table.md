# p04-t01: Symbol Table & Scope Stack

## Purpose
Implement the symbol table and scope stack — the central data structures for name resolution, type checking, and all subsequent semantic analysis passes. Every identifier in an AXIOM program must be resolved to a `Symbol` entry via this table. This is one of the most performance-critical data structures in the compiler: every name lookup in every expression traverses the scope stack.

## Context
AXIOM uses lexical scoping with nested scopes. Scope boundaries occur at: module level (global), function bodies, `if`/`elif`/`else` blocks, `for`/`in` loop bodies, `match` arms, `lock` blocks, `in [Arena]` blocks, and closure bodies. The symbol table must support: (1) defining a symbol in the current scope, (2) resolving a symbol by searching from innermost to outermost scope, (3) detecting duplicate definitions within the same scope, and (4) efficiently popping all symbols when leaving a scope.

The implementation uses a flat `[]Symbol` array for cache-friendly sequential access (symbols are never deleted, only shadowed). Each scope uses an open-addressing hash map for O(1) lookup by interned name ID. The `ScopeStack` is a simple `[]uint32` of scope indices — pushed on block entry, popped on block exit.

**Spec references:**
- `AXIOM SPECIFICATION/04. Type checker.md` — scope rules, symbol resolution order
- `docs/plan.md` Section 2, Phase 2 — Symbol/SymbolTable/ScopeStack API design

## Inputs
- `compiler/ast/intern.go` — `InternPool` for `string → uint32` mapping (from p03-t02)
- `compiler/ast/node.go` — `NodeKind` definitions for mapping declarations to symbol kinds (from p03-t01)
- `compiler/diagnostics/diagnostics.go` — `Diagnostic`, `Pos` types for error reporting (from p01-t01)

## Outputs
- `compiler/sema/symbols.go` — `SymKind`, `Symbol`, `SymFlags` type definitions
- `compiler/sema/scope.go` — `Scope` (open-addressing hash map), `ScopeKind` enum
- `compiler/sema/symtable.go` — `SymbolTable` with `ScopeStack`, `Define`, `Resolve` methods
- `compiler/sema/symtable_test.go` — comprehensive unit tests (≥15 test cases)

## Dependencies
- p03-t02: string-intern-pool — symbols reference interned name IDs (`uint32`)
- p03-t01: ast-node-definitions — `Symbol.DeclNode` references AST node indices

## Subsystems Affected
- **Name resolver** (p04-t04): uses `SymbolTable.Resolve()` for every identifier
- **Type checker** (p04-t05 through p04-t07): resolves type names, function signatures via symbol table
- **Monomorphization** (p05-t02): creates new symbols for monomorphized instantiations
- **Ownership checker** (p06-t02): reads symbol flags (IsMut, IsSink, IsLent)
- **Code generation** (p08): uses mangled names from symbol table
- **AIR builder** (p09): maps symbols to virtual registers

## Detailed Requirements

### 1. SymKind Enum

```go
// SymKind classifies the kind of entity a symbol represents.
type SymKind uint8

const (
    SymVar         SymKind = iota // let/mut variable binding
    SymFunc                       // fn declaration
    SymStruct                     // struct declaration
    SymInterface                  // interface declaration
    SymTypeAlias                  // type X = Y declaration
    SymParam                      // function parameter
    SymField                      // struct field
    SymGenericParam               // generic type parameter [T]
    SymModule                     // import module name
    SymBuiltinType                // built-in primitive type (i32, bool, etc.)
    SymEnumVariant                // sum type variant
)

func (k SymKind) String() string // must return human-readable name
```

### 2. Symbol Struct

```go
// SymFlags encodes boolean properties of a symbol.
type SymFlags uint16

const (
    SymFlagPub    SymFlags = 1 << iota // pub visibility
    SymFlagMut                          // mutable (mut keyword)
    SymFlagExtern                       // extern "C" declaration
    SymFlagSink                         // sink parameter (!T)
    SymFlagLent                         // lent (borrowed) parameter
    SymFlagAsync                        // async fn
    SymFlagPure                         // verified pure function
    SymFlagMoved                        // symbol has been moved (invalidated)
    SymFlagUsed                         // symbol has been referenced at least once
    SymFlagComptime                     // compile-time constant (#run result)
)

// Symbol represents a named entity in the program.
// Stored in a flat array for cache-friendly access.
type Symbol struct {
    NameID   uint32   // interned name (index into InternPool)
    Kind     SymKind  // what kind of entity
    Flags    SymFlags // boolean properties
    TypeID   uint32   // index into TypeTable (0 = unresolved)
    DeclNode uint32   // index into AstTree.Nodes (source location)
    ScopeID  uint32   // which scope this symbol belongs to
    // Total: 20 bytes (no padding needed for alignment)
}
```

**Invariants:**
- `NameID` is always a valid index into `InternPool` (never 0 for user-defined symbols)
- `DeclNode` is always a valid index into `Ast.Nodes` (except for built-in types where `DeclNode = 0`)
- `TypeID = 0` means "type not yet resolved" — valid only during the name resolution pass; after type checking, every symbol must have `TypeID > 0` or an error diagnostic
- Once `SymFlagMoved` is set, it is never cleared — use-after-move is a one-way transition
- `ScopeID` is always a valid index into `SymbolTable.Scopes`

### 3. Scope (Open-Addressing Hash Map)

```go
// ScopeKind identifies what language construct created this scope.
type ScopeKind uint8

const (
    ScopeGlobal   ScopeKind = iota // module-level scope
    ScopeFunction                   // function body
    ScopeBlock                      // if/elif/else/for/match/lock/arena block
    ScopeClosure                    // closure body
    ScopeLoop                       // for/in loop (owns loop variable)
)

// scopeEntry is a slot in the open-addressing hash map.
type scopeEntry struct {
    nameID    uint32 // 0 = empty slot
    symbolIdx uint32 // index into SymbolTable.Symbols
}

// Scope is a single lexical scope with O(1) name lookup.
type Scope struct {
    Kind     ScopeKind
    ParentID uint32       // index of parent scope (0 for global)
    Depth    uint32       // nesting depth (0 = global)
    entries  []scopeEntry // open-addressing hash table
    count    uint32       // number of occupied slots
    capacity uint32       // length of entries slice (always power of 2)
}
```

**Hash map implementation rules:**
- Initial capacity: 8 slots (power of 2)
- Hash function: FNV-1a of the `uint32` nameID (fast, good distribution for sequential IDs)
- Probing: linear probing (cache-friendly for small tables)
- Load factor threshold: 75% — when `count > capacity * 3 / 4`, resize to `capacity * 2`
- Lookup: `hash(nameID) & (capacity - 1)` → probe until `nameID` matches or empty slot found
- No deletion: symbols are never removed from a scope (they persist until the scope is popped)

### 4. SymbolTable

```go
// SymbolTable is the central symbol storage for the entire compilation unit.
type SymbolTable struct {
    Symbols []Symbol   // flat array of all symbols across all scopes
    Scopes  []Scope    // all scopes (index 0 = global scope)
    stack   []uint32   // active scope stack (indices into Scopes)
}

// NewSymbolTable creates a SymbolTable with the global scope pre-populated
// with built-in primitive types.
func NewSymbolTable(intern *InternPool) *SymbolTable

// PushScope creates a new child scope and pushes it onto the stack.
// Returns the new scope's index.
func (st *SymbolTable) PushScope(kind ScopeKind) uint32

// PopScope pops the current scope from the stack.
// Panics if attempting to pop the global scope.
func (st *SymbolTable) PopScope()

// CurrentScope returns the index of the innermost active scope.
func (st *SymbolTable) CurrentScope() uint32

// CurrentDepth returns the current nesting depth.
func (st *SymbolTable) CurrentDepth() uint32

// Define adds a new symbol to the current scope.
// Returns the symbol index and nil error on success.
// Returns 0 and a Diagnostic if the name is already defined in the current scope.
func (st *SymbolTable) Define(nameID uint32, kind SymKind, flags SymFlags, declNode uint32) (uint32, *diagnostics.Diagnostic)

// Resolve searches for a symbol from the innermost scope outward.
// Returns the symbol index and true if found, (0, false) if not found.
func (st *SymbolTable) Resolve(nameID uint32) (uint32, bool)

// ResolveInScope searches for a symbol only in a specific scope.
func (st *SymbolTable) ResolveInScope(nameID uint32, scopeID uint32) (uint32, bool)

// ResolveGlobal searches only the global scope (index 0).
func (st *SymbolTable) ResolveGlobal(nameID uint32) (uint32, bool)

// SymbolAt returns a pointer to the symbol at the given index.
// The pointer is invalidated if Symbols slice grows.
func (st *SymbolTable) SymbolAt(idx uint32) *Symbol

// MarkMoved sets the SymFlagMoved flag on a symbol.
func (st *SymbolTable) MarkMoved(idx uint32)

// IsMoved returns true if the symbol has been moved.
func (st *SymbolTable) IsMoved(idx uint32) bool

// MarkUsed sets the SymFlagUsed flag on a symbol.
func (st *SymbolTable) MarkUsed(idx uint32)
```

### 5. Built-in Types Pre-population

`NewSymbolTable()` must pre-populate the global scope (scope 0) with these built-in type symbols:

| Name | SymKind | TypeID | Notes |
|------|---------|--------|-------|
| `i8` | SymBuiltinType | 1 | Signed 8-bit integer |
| `i16` | SymBuiltinType | 2 | Signed 16-bit integer |
| `i32` | SymBuiltinType | 3 | Signed 32-bit integer |
| `i64` | SymBuiltinType | 4 | Signed 64-bit integer |
| `u8` | SymBuiltinType | 5 | Unsigned 8-bit integer |
| `u16` | SymBuiltinType | 6 | Unsigned 16-bit integer |
| `u32` | SymBuiltinType | 7 | Unsigned 32-bit integer |
| `u64` | SymBuiltinType | 8 | Unsigned 64-bit integer |
| `f32` | SymBuiltinType | 9 | 32-bit float |
| `f64` | SymBuiltinType | 10 | 64-bit float |
| `bool` | SymBuiltinType | 11 | Boolean |
| `string` | SymBuiltinType | 12 | UTF-8 string |
| `char8` | SymBuiltinType | 13 | ASCII character |
| `void` | SymBuiltinType | 14 | Void / unit type |
| `isize` | SymBuiltinType | 15 | Platform-sized signed int |
| `usize` | SymBuiltinType | 16 | Platform-sized unsigned int |

The TypeID values here must be coordinated with the TypeTable (p04-t02). Built-in type symbols have `DeclNode = 0` (no source location).

### 6. Error Handling

- `Define()` with a duplicate name in the current scope returns a `Diagnostic` with:
  - `Severity: SeverityError`
  - `Code: 2001` (duplicate definition)
  - `Message: "symbol 'X' already defined in this scope"`
  - `Hint: "previous definition was at <file>:<line>:<col>"`
- `Define()` shadowing an outer scope name is **allowed** (not an error) — AXIOM permits shadowing
- `Resolve()` returning `false` does NOT emit a diagnostic — the caller (name resolver) handles "undefined" errors

### 7. Performance Requirements

- `Resolve()` must be O(D) where D is the scope depth — no full table scan
- `Define()` must be O(1) amortized — hash table insert
- No heap allocations in the hot `Resolve()` path after initial setup
- `ScopeStack` uses a pre-allocated slice (initial capacity 32) — not a linked list
- `Symbols` slice initial capacity: 1024 — avoids frequent reallocation for small programs

### 8. Determinism Requirements

- Symbol indices must be deterministic: same source → same symbol table layout
- Scope traversal order in `Resolve()` must always be innermost-first, outermost-last
- Hash map iteration order is NOT required to be deterministic (hash maps are only for lookup, not iteration)

## Implementation Steps

1. Create `compiler/sema/symbols.go`:
   - Define `SymKind` enum with `String()` method
   - Define `SymFlags` constants
   - Define `Symbol` struct

2. Create `compiler/sema/scope.go`:
   - Define `ScopeKind` enum
   - Define `scopeEntry` and `Scope` struct
   - Implement `Scope.init(capacity)` — allocate entries slice
   - Implement `Scope.put(nameID, symbolIdx)` — insert with linear probing
   - Implement `Scope.get(nameID) (uint32, bool)` — lookup with linear probing
   - Implement `Scope.grow()` — double capacity, rehash all entries

3. Create `compiler/sema/symtable.go`:
   - Implement `NewSymbolTable(intern *InternPool)`:
     - Allocate `Symbols` with capacity 1024
     - Create global scope (index 0, `ScopeGlobal`, depth 0)
     - Pre-populate built-in types by interning their names and calling `Define()`
     - Push global scope onto stack
   - Implement `PushScope(kind)`:
     - Append new `Scope` to `Scopes` with `ParentID = CurrentScope()`, `Depth = CurrentDepth() + 1`
     - Push new scope index onto `stack`
   - Implement `PopScope()`:
     - Assert `len(stack) > 1` (cannot pop global)
     - Pop from `stack`
   - Implement `Define(nameID, kind, flags, declNode)`:
     - Get current scope from `stack[len(stack)-1]`
     - Check if `nameID` already exists in current scope via `scope.get(nameID)` → error if duplicate
     - Append new `Symbol` to `Symbols`, get its index
     - Call `scope.put(nameID, symbolIndex)`
     - Return symbol index
   - Implement `Resolve(nameID)`:
     - Walk `stack` from top (innermost) to bottom (outermost)
     - At each scope, call `scope.get(nameID)`
     - Return first match found
   - Implement `ResolveGlobal(nameID)`: directly call `Scopes[0].get(nameID)`

4. Create `compiler/sema/symtable_test.go` — all tests listed in Test Plan below.

## Test Plan

### Unit Tests

1. **TestNewSymbolTable_BuiltinsPresent**: Create new table, verify `Resolve("i32")`, `Resolve("bool")`, `Resolve("string")` all succeed and return `SymBuiltinType` kind.

2. **TestDefine_SingleScope**: Define `x` in global scope, resolve `x` → found, correct index.

3. **TestDefine_DuplicateError**: Define `x` twice in same scope → second `Define()` returns non-nil Diagnostic with code 2001.

4. **TestResolve_InnerToOuter**: Define `x` in global, push scope, resolve `x` → found (searches upward).

5. **TestResolve_Shadowing**: Define `x` in outer scope, define `x` in inner scope → resolve returns inner `x`, not outer. After pop, resolve returns outer `x`.

6. **TestResolve_NotFound**: Resolve `y` (never defined) → returns `(0, false)`.

7. **TestPopScope_HidesInnerSymbols**: Define `x` in inner scope, pop scope → resolve `x` fails.

8. **TestPopScope_GlobalPanics**: Attempt to pop global scope → panics.

9. **TestScopeKind_Preserved**: Push `ScopeFunction`, verify `CurrentScope()` returns scope with `Kind == ScopeFunction`.

10. **TestScopeDepth_Increments**: Push 5 scopes → `CurrentDepth() == 5`. Pop 3 → `CurrentDepth() == 2`.

11. **TestSymbolFlags_MoveTracking**: Define symbol, call `MarkMoved()`, verify `IsMoved()` returns true.

12. **TestSymbolFlags_Preserved**: Define symbol with `SymFlagPub | SymFlagMut`, verify flags preserved on retrieval.

13. **TestResolveInScope_SpecificScope**: Define `x` in scope A and `y` in scope B → `ResolveInScope("x", scopeA)` succeeds, `ResolveInScope("x", scopeB)` fails.

14. **TestHashMapGrowth**: Define 100 symbols in one scope (forces hash map resize) → all resolvable.

15. **TestDeterminism**: Create two symbol tables with identical inputs → symbol indices identical.

### Property Tests

16. **TestResolve_NeverPanics**: Random sequences of `PushScope/PopScope/Define/Resolve` — never panics (only returns errors).

### Benchmark

17. **BenchmarkResolve_Depth10**: 10 nested scopes, resolve from innermost — measure ns/op.
18. **BenchmarkDefine_1000Symbols**: Define 1000 symbols in one scope — measure ns/op.

## Validation Checklist

- [ ] `Resolve()` searches inner-to-outer correctly (test 4, 5)
- [ ] Duplicate definition in same scope returns error, not panic (test 3)
- [ ] `PopScope()` makes inner symbols invisible (test 7)
- [ ] `PopScope()` on global scope panics (test 8)
- [ ] All 16 built-in types pre-populated in global scope (test 1)
- [ ] Open-addressing hash map handles collisions correctly (test 14)
- [ ] Shadowing: inner symbol overrides outer, restored after pop (test 5)
- [ ] `SymFlagMoved` one-way transition works (test 11)
- [ ] Symbol struct size is ≤ 24 bytes (verify with `unsafe.Sizeof`)
- [ ] No heap allocations in hot `Resolve()` path (verify with benchmark + `testing.AllocsPerRun`)
- [ ] `go test ./compiler/sema/ -run TestSymbol` — all pass
- [ ] `go vet ./compiler/sema/` — zero warnings

## Acceptance Criteria

- All 18 unit tests pass
- `Resolve()` is O(scope_depth) — verified by benchmark showing linear scaling with depth
- `Define()` is O(1) amortized — verified by benchmark for 1000 symbols
- Zero allocations in `Resolve()` hot path (verified with `testing.AllocsPerRun`)
- `go vet` and `golangci-lint` clean

## Definition of Done

- [ ] `compiler/sema/symbols.go` implemented with all types
- [ ] `compiler/sema/scope.go` implemented with hash map
- [ ] `compiler/sema/symtable.go` implemented with all API methods
- [ ] `compiler/sema/symtable_test.go` with all 18 tests passing
- [ ] `go test ./compiler/sema/ -v` — zero failures
- [ ] `go vet ./compiler/sema/` — zero warnings
- [ ] Benchmark results recorded: `BenchmarkResolve_Depth10`, `BenchmarkDefine_1000Symbols`
- [ ] No circular imports with `compiler/ast/` package

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Hash collisions degrade lookup performance for large scopes | Use Robin Hood probing (swap entries to reduce max probe length); resize at 75% load factor |
| Scope stack overflow for deeply nested programs | Use `[]uint32` slice (auto-grows); set initial capacity to 32 which handles 99% of programs |
| Circular import: `sema` imports `ast` imports `sema` | `sema` depends on `ast` (one-way); `ast` never imports `sema`; types referenced only by uint32 ID |
| Symbol indices invalidated when `Symbols` slice grows | Document that `SymbolAt()` returns pointer invalidated by next `Define()` call; callers must copy if needed |
| Built-in type TypeIDs hardcoded — must match TypeTable | Coordinate with p04-t02; define shared constants in a `compiler/types/builtin_ids.go` file |

## Future Follow-up Tasks

- **p04-t02**: Type table primitives — uses TypeIDs assigned to built-in symbols here
- **p04-t04**: Name resolver — primary consumer of `SymbolTable.Resolve()`
- **p05-t01**: Generic type representation — extends `Symbol` with type parameters
- **p05-t02**: Monomorphization — creates new symbols for concrete instantiations
- **p06-t02**: Ownership rules — reads `SymFlagMut`, `SymFlagSink`, `SymFlagLent`, `SymFlagMoved`
