# p08-t06: C-Backend Generational Reference Check Emission

## Purpose
Implement the generational reference check emission layer in `codegen/cgen/gencheck.go`. Every heap pointer allocation in generated C code must create an `AxRef` via `ax_make_ref`, and every dereference of a heap reference must go through `ax_deref`. This layer determines when these calls are necessary and emits them correctly, based on escape analysis results that distinguish heap allocations from stack allocations and arena allocations.

## Context
AXIOM's use-after-free detection (the generational reference system, p07-t02) only applies to heap-allocated values. Stack-allocated values are not heap-managed and do not carry `AxHeader`s, so `ax_deref` must not be called for them. Arena-allocated values are bulk-freed and also bypass per-object generational checks.

The escape analysis pass (Phase 06) annotates each allocation site with one of three allocation strategies:
- `AllocStack` — variable lives on the stack; no generational check needed
- `AllocHeap` — variable lives on the heap via `ax_alloc`; generational check required
- `AllocArena` — variable lives in an arena; arena-level lifetime management applies

This task extends the expression and statement generators to emit the correct allocation and dereference patterns based on these annotations.

## Inputs
- `codegen/cgen/exprs.go` (p08-t04) — extended by this task
- `codegen/cgen/stmts.go` (p08-t03) — extended by this task
- Escape analysis annotations on the typed AST (from p06-t05 or equivalent)
- `runtime/axalloc/genref.h` — the `ax_make_ref`, `ax_deref` API

## Outputs
- `codegen/cgen/gencheck.go` — generational check emission helpers
- `codegen/cgen/gencheck_test.go` — unit tests

## Dependencies
- p08-t05 (ownership-aware code gen — establishes context for heap vs stack)
- p07-t02 (generational ref runtime API consumed by emitted code)

## Subsystems Affected
- C-Backend (all heap allocation and dereference sites)
- Runtime (emitted code calls into ax_make_ref and ax_deref)

## Detailed Requirements

### Allocation Emission Patterns

**Stack Allocation (AllocStack)**
```c
// AXIOM: let x: Foo = Foo{...}   (escape analysis: stays on stack)
struct ax_Foo x = (struct ax_Foo){.field = value};
// No AxRef, no ax_make_ref — accessed directly as 'x'
```

**Heap Allocation (AllocHeap)**
```c
// AXIOM: let x: Foo = Foo{...}   (escape analysis: escapes to heap)
struct ax_Foo* _ax_raw_x = (struct ax_Foo*)ax_alloc(sizeof(struct ax_Foo));
*_ax_raw_x = (struct ax_Foo){.field = value};
AxRef x = ax_make_ref(_ax_raw_x);
// Subsequent access: (struct ax_Foo*)ax_deref(x)
```

The variable name in generated C is still `x`, but its C type is `AxRef` rather than `struct ax_Foo`. Field accesses through `x` are emitted as `((struct ax_Foo*)ax_deref(x))->field`.

**Arena Allocation (AllocArena)**
```c
// AXIOM: in [arena]: let x: Foo = Foo{...}
struct ax_Foo* x = (struct ax_Foo*)ax_arena_alloc(arena, sizeof(struct ax_Foo));
*x = (struct ax_Foo){.field = value};
// No AxRef — arena manages lifetime; access via plain pointer
```

### Dereference Emission Patterns

The expression generator must check the allocation strategy of the variable being dereferenced:

```go
func (g *ExprGen) emitField(e *ast.FieldExpr) string {
    ident, ok := e.Object.(*ast.Ident)
    if ok {
        switch g.allocMode[ident.Name] {
        case AllocHeap:
            // Must deref through generational check
            innerType := CTypeName(e.ObjectTypeID, g.table, g.queue)
            return fmt.Sprintf("(((%s*)ax_deref(%s))->%s)",
                innerType, ident.Name, e.Field)
        case AllocArena:
            // Plain pointer dereference (arena-managed)
            return fmt.Sprintf("(%s->%s)", ident.Name, e.Field)
        default: // AllocStack
            return fmt.Sprintf("(%s.%s)", ident.Name, e.Field)
        }
    }
    // Non-identifier: recurse
    return fmt.Sprintf("(%s.%s)", g.EmitExpr(e.Object), e.Field)
}
```

### CTGC Inject Pattern: Reuse Allocation
When the optimizer (p10-t05) produces an `OpReuseAlloc` node (reuse a freed allocation), the C-Backend emits:
```c
// Instead of: ax_free(old_ptr); raw_new = ax_alloc(sizeof(T));
// Emit: increment gen_id manually and reuse the same memory
AxHeader* _ax_hdr = ((AxHeader*)ax_deref(old_ref)) - 1;
_ax_hdr->gen_id++;
AxRef new_ref = ax_make_ref((void*)(_ax_hdr + 1));
// new_ref points to the same memory with a new generation ID
```

This is an advanced optimization; emit a TODO comment in MVP and implement in Phase 10.

### Variable Allocation Mode Map
The `ExprGen` carries a map from variable name to allocation mode:
```go
type AllocMode int
const (
    AllocStack AllocMode = iota
    AllocHeap
    AllocArena
)

// Populated by StmtGen when processing variable declarations
allocMode map[string]AllocMode
```

### Variable Declaration with Alloc Mode
Extend `emitVarDecl` in `StmtGen`:
```go
func (g *StmtGen) emitVarDecl(s *ast.VarDeclStmt) {
    mode := s.AllocMode  // from escape analysis annotation
    switch mode {
    case AllocStack:
        g.emitStackDecl(s)
    case AllocHeap:
        g.emitHeapDecl(s)
    case AllocArena:
        g.emitArenaDecl(s, g.currentArena)
    }
    g.exprGen.allocMode[s.Name] = mode
}
```

### Unsafe Block Suppression
In unsafe blocks (`g.unsafe == true`), generational checks are suppressed. Heap variable field access in unsafe mode uses direct pointer arithmetic:
```c
// unsafe: x.field  (x is heap-allocated)
((struct ax_Foo*)x.ptr)->field  // skip ax_deref, access ptr directly
```

## Implementation Steps

### Step 1: Add `AllocMode` type and `allocMode` map to `ExprGen`
```go
type AllocMode int
const (AllocStack AllocMode = iota; AllocHeap; AllocArena)

// In ExprGen:
allocMode map[string]AllocMode
```

### Step 2: Implement `emitHeapDecl` in `StmtGen`
```go
func (g *StmtGen) emitHeapDecl(s *ast.VarDeclStmt) {
    ctype := CTypeName(s.TypeID, g.table, g.queue)
    rawName := "_ax_raw_" + s.Name
    g.w.Line(fmt.Sprintf("%s* %s = (%s*)ax_alloc(sizeof(%s));",
        ctype, rawName, ctype, ctype))
    if s.Init != nil {
        initVal := g.exprGen.EmitExpr(s.Init)
        g.w.Line(fmt.Sprintf("*%s = %s;", rawName, initVal))
    }
    g.w.Line(fmt.Sprintf("AxRef %s = ax_make_ref(%s);", s.Name, rawName))
}
```

### Step 3: Update `emitField` to route through `allocMode`
See the code snippet in Detailed Requirements above.

### Step 4: Update index expression for heap slices
```go
func (g *ExprGen) emitIndex(e *ast.IndexExpr) string {
    arr := g.emitArrayExpr(e.Array) // returns the slice struct, dereferencing if heap
    idx := g.EmitExpr(e.Index)
    if g.unsafe {
        return fmt.Sprintf("(%s).ptr[%s]", arr, idx)
    }
    return fmt.Sprintf("(ax_bounds_check((ax_u64)(%s), (%s).len), (%s).ptr[%s])",
        idx, arr, arr, idx)
}

func (g *ExprGen) emitArrayExpr(e ast.TypedExpr) string {
    if id, ok := e.(*ast.Ident); ok && g.allocMode[id.Name] == AllocHeap {
        innerType := CTypeName(e.TypeID(), g.table, g.queue)
        return fmt.Sprintf("(*((const %s*)ax_deref(%s)))", innerType, id.Name)
    }
    return g.EmitExpr(e)
}
```

### Step 5: Write `gencheck_test.go`
Test heap allocation pattern, stack allocation pattern, field access routing, and unsafe suppression.

## Test Plan
1. Stack allocation: no `ax_alloc`, no `ax_make_ref`, no `ax_deref` in output
2. Heap allocation: `ax_alloc` + `ax_make_ref` emitted, variable has type `AxRef`
3. Heap field access: `ax_deref` called before field access
4. Stack field access: direct `.field` access
5. Arena allocation: `ax_arena_alloc` + plain pointer, no `AxRef`
6. Unsafe heap field access: `x.ptr->field` without `ax_deref`
7. Index into heap slice: `ax_deref` on the slice before indexing
8. Multiple heap variables: each has independent `AxRef`

## Validation Checklist
- [ ] Stack-allocated variables never produce `ax_make_ref` calls
- [ ] Heap-allocated variables always produce `ax_make_ref` calls
- [ ] Arena-allocated variables use `ax_arena_alloc` (not `ax_alloc`)
- [ ] All generated C compiles without type errors
- [ ] Unsafe mode correctly suppresses `ax_deref` calls
- [ ] All tests pass

## Acceptance Criteria
- Heap allocations are always wrapped in `AxRef`
- Stack allocations have zero generational overhead
- Generated C is type-correct for all allocation patterns

## Definition of Done
- `codegen/cgen/gencheck.go` exists
- `codegen/cgen/gencheck_test.go` exists and passes
- `go test ./codegen/cgen/` passes

## Risks & Mitigations
- **Risk**: The escape analysis annotations may be incomplete in early phases. **Mitigation**: Default conservatively to `AllocHeap` for any unannotated variable, producing correct (though potentially suboptimal) code.
- **Risk**: The `_ax_raw_` prefix for the raw pointer local variable could clash with user variables. **Mitigation**: The AXIOM identifier grammar forbids identifiers starting with `_ax_`; enforce in the parser.

## Future Follow-up Tasks
- p10-t05: CTGC pass introduces `OpReuseAlloc` nodes that this layer handles
- p08-t08: Arena allocations use `ax_arena_alloc` from `runtime/axalloc/arena.c`
- p11-t15: Native backend implements the same allocation mode distinction at the machine level
