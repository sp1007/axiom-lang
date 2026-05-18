# p08-t05: C-Backend Ownership-Aware Code Generation

## Purpose
Implement ownership-aware C code emission in `codegen/cgen/ownership.go`. This extends the base code generators (p08-t03, p08-t04) to correctly represent AXIOM's ownership semantics in C: moves poison the source location in debug mode, sink parameters pass by value, borrows pass as const pointers, and `Isolated[T]` values are annotated. Ownership is a compile-time guarantee; this layer makes the runtime behavior match.

## Context
AXIOM's ownership system is verified at compile time by the ownership checker (Phase 06). The C-Backend is not responsible for re-checking ownership rules, but it is responsible for ensuring that the generated C code enforces the runtime consequences:
1. After a move, the source location is unusable (debug: poisoned with zero bytes)
2. Sink parameters consume the value (C pass-by-value achieves this for value types)
3. Borrow parameters cannot be written through (const pointer)
4. `Isolated[T]` carries a documentation comment indicating no external references

This task extends `StmtGen` and `ExprGen` by adding ownership-aware variants.

## Inputs
- `codegen/cgen/stmts.go` (p08-t03)
- `codegen/cgen/exprs.go` (p08-t04)
- Ownership information in the typed AST (p06-t02 annotates each variable with ownership mode)
- Typed AST parameter annotations: `SinkParam`, `LentParam`, `IsolatedParam`

## Outputs
- `codegen/cgen/ownership.go` — ownership code generation helpers
- `codegen/cgen/ownership_test.go` — unit tests

## Dependencies
- p08-t04 (expression generator — extended here)
- p08-t03 (statement generator — extended here)
- p06-t02 (ownership checker annotates AST with move/borrow/sink information)

## Subsystems Affected
- C-Backend (all function calls and variable uses go through this)
- Runtime (move poisoning calls `memset`)

## Detailed Requirements

### Move Semantics
When the ownership checker marks a value as "moved", the source variable becomes inaccessible. In generated C:

```c
// AXIOM: let y = x  (move x into y)
struct ax_Foo y = x;           // C struct copy (by value) — this IS the move
#if AX_DEBUG
memset(&x, 0, sizeof(x));      // poison source in debug mode
#endif
```

The `#if AX_DEBUG` guard ensures the poisoning is a zero-cost release-mode operation. The poisoning turns use-after-move into a detectable null/zero read rather than silent reuse of stale data.

For heap types (`AxRef`):
```c
// AXIOM: let y = ref_x  (move heap ref)
AxRef y = ref_x;
#if AX_DEBUG
ref_x = (AxRef){.ptr = NULL, .gen_id = 0};  // poison ref
#endif
```

### Sink Parameters (`!T`)
A sink parameter indicates the function takes ownership of the argument. In C, this is pass-by-value — the caller's copy is logically consumed. The C-Backend emits:
```c
// Function signature: fn consume(x: !Foo) -> void
void ax_module_consume(struct ax_Foo x) {   // pass by value = C copy
    // ...
    // x is freed at end of function body (if heap-allocated, CTGC injects =destroy)
}
```

No special caller-side code is needed beyond the move emission (the value is passed, and the source is optionally poisoned).

### Borrow Parameters (`lent T`)
A borrow parameter gives read-only access to a value owned elsewhere. In C, this is a `const` pointer to the value:
```c
// AXIOM: fn inspect(s: lent Foo) -> i32
ax_i32 ax_module_inspect(const struct ax_Foo* s) {
    return s->x;
}
```

At call sites, the argument must have its address taken:
```c
ax_i32 result = ax_module_inspect(&my_foo);
```

The `ExprGen` must detect when a function parameter is `lent` and automatically emit `&expr` for the corresponding argument.

### Mutable Borrow Parameters (`mut lent T`)
A mutable borrow gives write access. In C, this is a non-const pointer:
```c
// AXIOM: fn increment(s: mut lent Foo) -> void
void ax_module_increment(struct ax_Foo* s) {
    s->x += 1;
}
```

At call sites, the argument must also have its address taken.

### `Isolated[T]` Parameters
An `Isolated[T]` value guarantees no external references to the heap data. In C:
```c
// AXIOM: fn process(data: Isolated[Foo]) -> void
void ax_module_process(struct ax_Foo* data /* Isolated - no external refs */) {
    // ...
}
```
The parameter type is a pointer (not a value copy, since Isolated wraps heap data), with a comment.

### Field Access Through Borrows
When accessing a field through a `lent T` parameter (which is a `const T*`), field access uses `->` instead of `.`:
```c
// s is 'const struct ax_Foo*' (borrow parameter)
ax_i32 val = s->x;  // arrow operator for pointer access
```

The expression generator must track whether an identifier refers to a value or a pointer-to-value and emit `.` vs `->` accordingly.

### Ownership Mode Tracking
Extend `ExprGen` with a map from variable name to ownership mode:
```go
type OwnershipMode int
const (
    ModeValue    OwnershipMode = iota // local value, pass by value
    ModeRef                          // lent reference (const pointer)
    ModeMutRef                       // mutable borrow (non-const pointer)
    ModeIsolated                     // isolated heap value
)

type ExprGen struct {
    // ... existing fields ...
    ownerModes map[string]OwnershipMode
}
```

The statement generator populates `ownerModes` when processing function parameters and variable declarations.

### API
```go
// EmitMove emits the move of `src` variable into the expression position.
// In debug mode, also emits source poisoning.
func (g *StmtGen) EmitMove(src string, typeID uint32) string

// AdaptArgForParam adapts an argument expression for a parameter with the given ownership mode.
// For lent params: wraps expr in "&expr" if expr is a value.
// For sink params: ensures expr is emitted as a copy.
func (g *ExprGen) AdaptArgForParam(expr ast.TypedExpr, paramMode OwnershipMode) string
```

## Implementation Steps

### Step 1: Define `OwnershipMode` and extend `ExprGen`
Add `ownerModes map[string]OwnershipMode` to `ExprGen`. Populate it in `StmtGen` when entering a function body.

### Step 2: Implement move poisoning in `StmtGen.emitVarDecl`
```go
func (g *StmtGen) emitVarDecl(s *ast.VarDeclStmt) {
    ctype := CTypeName(s.TypeID, g.table, g.queue)
    val := g.exprGen.EmitExpr(s.Init)
    g.w.Line(fmt.Sprintf("%s %s = %s;", ctype, s.Name, val))

    // If this is a move, poison the source
    if move, ok := s.Init.(*ast.MoveExpr); ok && g.debugMode {
        src := move.Source.(*ast.Ident).Name
        ty := g.table.Get(s.TypeID)
        if ty.IsHeapRef() {
            g.w.Line(fmt.Sprintf("#if AX_DEBUG"))
            g.w.Line(fmt.Sprintf(`%s = (AxRef){.ptr=NULL, .gen_id=0};`, src))
            g.w.Line(fmt.Sprintf("#endif"))
        } else {
            g.w.Line(fmt.Sprintf("#if AX_DEBUG"))
            g.w.Line(fmt.Sprintf(`memset(&%s, 0, sizeof(%s));`, src, src))
            g.w.Line(fmt.Sprintf("#endif"))
        }
    }
}
```

### Step 3: Implement borrow parameter emission in `DeclEmitter`
When emitting function prototypes (p08-t02) and definitions, check each parameter's ownership mode:
```go
func emitParamC(p *ast.TypedParam, table *typecheck.TypeTable, queue *TypeDeclQueue) string {
    ctype := CTypeName(p.TypeID, table, queue)
    switch p.Mode {
    case ast.ParamLent:
        return fmt.Sprintf("const %s* %s", ctype, p.Name)
    case ast.ParamMutLent:
        return fmt.Sprintf("%s* %s", ctype, p.Name)
    case ast.ParamIsolated:
        return fmt.Sprintf("%s* %s /* Isolated */", ctype, p.Name)
    default: // sink or value
        return fmt.Sprintf("%s %s", ctype, p.Name)
    }
}
```

### Step 4: Implement `AdaptArgForParam` in `ExprGen`
```go
func (g *ExprGen) AdaptArgForParam(expr ast.TypedExpr, mode ast.ParamMode) string {
    inner := g.EmitExpr(expr)
    switch mode {
    case ast.ParamLent, ast.ParamMutLent, ast.ParamIsolated:
        return "&(" + inner + ")"
    default:
        return inner
    }
}
```

### Step 5: Update field access in `emitField` to use `->` for pointer types
```go
func (g *ExprGen) emitField(e *ast.FieldExpr) string {
    obj := g.EmitExpr(e.Object)
    mode := g.ownerModes[e.Object.(*ast.Ident).Name]
    if mode == ModeRef || mode == ModeMutRef || mode == ModeIsolated {
        return fmt.Sprintf("(%s)->%s", obj, e.Field)
    }
    return fmt.Sprintf("(%s).%s", obj, e.Field)
}
```

## Test Plan
1. Move of a value type: source is zeroed in debug mode
2. Move of an `AxRef`: source ref is nulled in debug mode
3. `lent T` parameter: emitted as `const T*`, call site wraps in `&`
4. `mut lent T` parameter: emitted as `T*`
5. `Isolated[T]` parameter: emitted as `T*` with comment
6. Field access on a lent param uses `->` not `.`
7. `!T` (sink) parameter: emitted as plain value type (C pass-by-value)
8. Debug mode disabled: no `memset` poisoning emitted
9. Call to a function with mixed param modes: each arg adapted correctly

## Validation Checklist
- [ ] Move poisoning is guarded with `#if AX_DEBUG`
- [ ] `const` correctness: lent params cannot be written through in generated C
- [ ] All parameter modes have corresponding test cases
- [ ] Field access uses `->` for all pointer-mode parameters
- [ ] All tests pass with `go test ./codegen/cgen/`

## Acceptance Criteria
- Generated C compiles without warnings for all ownership modes
- Debug builds poison moved-from locations
- Release builds have zero overhead from move semantics
- `const` parameters are correctly typed

## Definition of Done
- `codegen/cgen/ownership.go` exists with all ownership modes implemented
- `codegen/cgen/ownership_test.go` exists and passes
- `go test ./codegen/cgen/` passes

## Risks & Mitigations
- **Risk**: `memset` on a stack struct is a valid C operation but may be removed by the optimizer. **Mitigation**: Use `volatile` in debug mode or `memset` with a compiler barrier — acceptable for debug builds only.
- **Risk**: Adapting arguments for borrow params may produce invalid C if the expression is not an lvalue. **Mitigation**: The ownership checker guarantees that lent args are always lvalues; add an assertion in `AdaptArgForParam` to verify this.

## Future Follow-up Tasks
- p08-t06: Generational checks are complementary to ownership: heap borrows still go through `ax_deref`
- p09-t10: AIR builder for ownership operations mirrors the patterns established here
- p10-t05: CTGC optimizer decides when to emit moves vs borrows at the AIR level
