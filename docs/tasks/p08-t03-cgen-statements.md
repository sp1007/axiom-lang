# p08-t03: C-Backend Statement Code Generation

## Purpose
Implement C code generation for all AXIOM statement forms in `codegen/cgen/stmts.go`. Each AXIOM statement is lowered to equivalent C11 code. This includes variable declarations, assignments, control flow, defer blocks, unsafe blocks, and CTGC-injected destroy nodes.

## Context
Statement generation is the core of function body emission in the C-Backend. It consumes the typed AST statement nodes produced by the type checker (Phase 04) and outputs C code that preserves AXIOM's semantics. The `defer` statement requires special handling: deferred expressions must execute at every exit point of the containing scope, including early returns and implicit scope exits.

## Inputs
- Typed AST statement nodes (from p04-t08)
- `codegen/cgen/types.go` (p08-t01) — for variable type names
- `codegen/cgen/decls.go` (p08-t02) — for context about declared names
- Type checker output — resolved types for all expressions

## Outputs
- `codegen/cgen/stmts.go` — statement code generator
- `codegen/cgen/stmts_test.go` — unit tests

## Dependencies
- p08-t01 (type mapping)
- p08-t02 (declaration context)

## Subsystems Affected
- C-Backend (function body emission)
- Ownership system (p08-t05 extends this for move/borrow semantics)

## Detailed Requirements

### Statement Mappings

**Variable Declaration**
```
AXIOM: let x: i32 = 5
C:     ax_i32 x = 5;

AXIOM: mut x := 5          (type inferred as i32)
C:     ax_i32 x = 5;

AXIOM: let x: i32          (no initializer, default zero)
C:     ax_i32 x = 0;
```

**Assignment**
```
AXIOM: x = 10
C:     x = 10;

AXIOM: arr[i] = v
C:     ax_bounds_check(i, arr.len); arr.ptr[i] = v;

AXIOM: s.field = v
C:     s.field = v;
```

**Return**
```
AXIOM: return x
C:     <emit defer stack in reverse>; return x;

AXIOM: return           (void function)
C:     <emit defer stack in reverse>; return;
```

**If Statement**
```
AXIOM:
if cond:
    body
else:
    else_body

C:
if (cond) {
    body
} else {
    else_body
}
```

**While Loop**
```
AXIOM:
while cond:
    body

C:
while (cond) {
    body
}
```

**For-In Loop (range over slice)**
```
AXIOM:
for x in list:
    body

C:
for (ax_u64 _ax_i = 0; _ax_i < list.len; _ax_i++) {
    <element_type> x = list.ptr[_ax_i];
    body
}
```

**For-In Loop (range over integer range `0..n`)**
```
AXIOM:
for i in 0..n:
    body

C:
for (ax_i64 i = 0; i < n; i++) {
    body
}
```

**Break / Continue**
```
AXIOM: break    → C: break;
AXIOM: continue → C: continue;
```

**Defer Statement**
```
AXIOM:
defer foo(x)

C: (accumulated; emitted at every scope exit)
// At return point:
foo(x);
return ...;
```

Defer is stack-ordered (LIFO). The statement generator maintains a `DeferStack` per function scope. When a return is encountered (or the function body ends), all deferred expressions are emitted in reverse push order before the `return` statement.

**Unsafe Block**
```
AXIOM:
unsafe:
    raw_ptr.* = 42

C:
{
    /* unsafe */
    *ax_deref_unsafe(raw_ptr) = 42;
}
```

In unsafe blocks, generational checks are skipped; raw pointer arithmetic is allowed.

**Destroy (CTGC-Injected Node)**
```
AXIOM: =destroy(x)    (injected by CTGC analysis, not user-written)
C:     ax_free(x);
```

**Block Statement**
```
AXIOM:
block:
    s1
    s2

C:
{
    s1
    s2
}
```

### Defer Implementation
The `DeferStack` is a per-scope slice of AST expression nodes. When a `defer expr` statement is encountered, the expression is pushed. When a scope exit is reached:
1. Pop all deferred expressions in LIFO order
2. Emit each as a statement
3. Then emit the actual exit (return, break, or block end)

Nested scopes each have their own defer stack. On `break` or `continue`, only the defers for the current loop scope are emitted (not the function-level defers).

```go
type DeferStack struct {
    scopes [][]ast.Expr  // stack of scopes
}
func (d *DeferStack) Push(expr ast.Expr) { d.scopes[len(d.scopes)-1] = append(..., expr) }
func (d *DeferStack) PushScope()         { d.scopes = append(d.scopes, nil) }
func (d *DeferStack) PopScope() []ast.Expr { ... }  // returns exprs in LIFO order
```

### Indentation
The statement generator tracks an indentation level (integer, incremented per block, decremented on close). Output each statement with `strings.Repeat("    ", indent)` prefix.

### API
```go
type StmtGen struct {
    w      *IndentWriter  // indentation-aware writer
    exprGen *ExprGen      // expression generator (p08-t04)
    defers  DeferStack
    table   *typecheck.TypeTable
    queue   *TypeDeclQueue
}

func (g *StmtGen) EmitStmt(stmt ast.TypedStmt)
func (g *StmtGen) EmitBlock(stmts []ast.TypedStmt)
```

## Implementation Steps

### Step 1: Implement `IndentWriter`
```go
type IndentWriter struct {
    w      io.Writer
    indent int
    buf    bytes.Buffer
}
func (iw *IndentWriter) Indent()   { iw.indent++ }
func (iw *IndentWriter) Dedent()   { iw.indent-- }
func (iw *IndentWriter) Line(s string) {
    fmt.Fprintf(iw.w, "%s%s\n", strings.Repeat("    ", iw.indent), s)
}
```

### Step 2: Implement `EmitStmt` with a type switch
```go
func (g *StmtGen) EmitStmt(stmt ast.TypedStmt) {
    switch s := stmt.(type) {
    case *ast.VarDeclStmt:   g.emitVarDecl(s)
    case *ast.AssignStmt:    g.emitAssign(s)
    case *ast.ReturnStmt:    g.emitReturn(s)
    case *ast.IfStmt:        g.emitIf(s)
    case *ast.WhileStmt:     g.emitWhile(s)
    case *ast.ForInStmt:     g.emitForIn(s)
    case *ast.DeferStmt:     g.emitDefer(s)
    case *ast.UnsafeStmt:    g.emitUnsafe(s)
    case *ast.DestroyStmt:   g.emitDestroy(s)
    case *ast.BlockStmt:     g.emitBlock(s)
    case *ast.BreakStmt:     g.emitBreak(s)
    case *ast.ContinueStmt:  g.w.Line("continue;")
    case *ast.ExprStmt:      g.emitExprStmt(s)
    default:
        panic(fmt.Sprintf("EmitStmt: unknown stmt type %T", stmt))
    }
}
```

### Step 3: Implement defer emission at return
```go
func (g *StmtGen) emitReturn(s *ast.ReturnStmt) {
    // Emit all pending defers in LIFO order
    deferred := g.defers.PopScope()
    for i := len(deferred)-1; i >= 0; i-- {
        g.w.Line(g.exprGen.EmitExpr(deferred[i]) + ";")
    }
    if s.Value != nil {
        g.w.Line("return " + g.exprGen.EmitExpr(s.Value) + ";")
    } else {
        g.w.Line("return;")
    }
}
```

### Step 4: For-in loop with bounds check
```go
func (g *StmtGen) emitForIn(s *ast.ForInStmt) {
    iter := g.exprGen.EmitExpr(s.Iterable)
    elemType := CTypeName(s.ElemTypeID, g.table, g.queue)
    idx := freshName("_ax_i")
    g.w.Line(fmt.Sprintf("for (ax_u64 %s = 0; %s < (%s).len; %s++) {",
        idx, idx, iter, idx))
    g.w.Indent()
    g.w.Line(fmt.Sprintf("%s %s = (%s).ptr[%s];", elemType, s.VarName, iter, idx))
    g.EmitBlock(s.Body)
    g.w.Dedent()
    g.w.Line("}")
}
```

## Test Plan
1. `let x: i32 = 5` → `ax_i32 x = 5;`
2. `mut x := 5` (inferred) → `ax_i32 x = 5;`
3. `x = 10` → `x = 10;`
4. `return x` without defer → `return x;`
5. `return x` with one defer `defer foo()` → `foo(); return x;`
6. Two defers: emitted LIFO
7. `if cond: body` → correct braces and indentation
8. `if cond: body else: else_body` → correct else branch
9. `while cond: body` → `while (cond) { body }`
10. `for x in list: body` → iterator with `_ax_i` index variable
11. `unsafe: block` → block with `/* unsafe */` comment
12. `=destroy(x)` → `ax_free(x);`
13. Nested blocks: indentation level increases/decreases correctly
14. Array assignment `arr[i] = v` → bounds check emitted

## Validation Checklist
- [ ] All statement forms produce syntactically valid C
- [ ] Defer stack is LIFO
- [ ] Return in nested scope emits only the local scope's defers, not parent scope defers
- [ ] `unsafe` block suppresses generational checks (flag passed to ExprGen)
- [ ] Loop body has its own defer scope
- [ ] `continue` in a for loop emits loop-scope defers before continuing
- [ ] All tests pass

## Acceptance Criteria
- Generated C compiles without warnings for all test cases
- Defer ordering is correct (verified by observable output in a C program)
- `for x in list` generates a bounds-safe C loop

## Definition of Done
- `codegen/cgen/stmts.go` exists with all statement forms handled
- `codegen/cgen/stmts_test.go` exists and passes
- `go test ./codegen/cgen/` passes

## Risks & Mitigations
- **Risk**: Defer in a loop may emit defers for every loop iteration. **Mitigation**: Defer is per-scope; a loop body creates a new scope. Defers inside the loop body are emitted at each iteration's end (before the implicit `continue`), matching Go semantics.
- **Risk**: `continue` in a for loop with defers: must emit loop-scope defers before the `continue` statement. **Mitigation**: `emitContinue` calls `PopScope()` and emits defers before emitting `continue;`, then pushes a new scope for the next iteration.

## Future Follow-up Tasks
- p08-t04: Expression generator used by `StmtGen.exprGen`
- p08-t05: Ownership-aware statements (move poisoning, sink params)
- p09-t07: AIR builder for statements mirrors this structure
