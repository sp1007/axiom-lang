# p08-t04: C-Backend Expression Code Generation

## Purpose
Implement C code generation for all AXIOM expression forms in `codegen/cgen/exprs.go`. Every AXIOM expression is lowered to a C expression string that evaluates to the same value with the same semantics, including type coercions, name mangling for function calls, generational reference checks for heap dereferences, and runtime safety checks for array indexing.

## Context
Expressions in AXIOM are always typed: the type checker has annotated every node with its resolved type. The expression generator uses these type annotations to emit correct casts, choose the right slice field access pattern, and decide whether a dereference needs a generational check. The generator returns a `string` representing the C expression; the statement generator (p08-t03) is responsible for placing it in the right syntactic context.

## Inputs
- Typed AST expression nodes (from p04-t08)
- `codegen/cgen/types.go` (p08-t01) — for type names in casts
- `TypeTable` — for resolving expression types
- Name mangling rules (p12-t01)
- `IsUnsafe bool` context flag — suppresses safety checks in unsafe blocks

## Outputs
- `codegen/cgen/exprs.go` — expression code generator
- `codegen/cgen/exprs_test.go` — unit tests

## Dependencies
- p08-t01 (type mapping)
- p08-t03 (statement generator creates `ExprGen` and passes it context)

## Subsystems Affected
- C-Backend (all generated code passes through expression generation)
- Runtime integration (safety checks reference `ax_bounds_check`, `ax_deref`)

## Detailed Requirements

### Expression Mappings

**Integer and Float Literals**
```
42        → "42"
42u64     → "((ax_u64)42ULL)"
3.14f32   → "3.14f"
3.14      → "3.14"
true      → "AX_TRUE"
false     → "AX_FALSE"
```

**String Literals**
```
"hello"   → "(ax_string){.ptr=(const ax_u8*)\"hello\", .len=5}"
""        → "(ax_string){.ptr=(const ax_u8*)\"\", .len=0}"
```
String length is computed at compile time (byte count, not character count).

**Identifier**
```
x         → "x"             (local variable)
Mod.foo   → "ax_Mod_foo"    (module-qualified name — mangled)
```

**Binary Operators**
```
a + b     → "(a + b)"
a - b     → "(a - b)"
a * b     → "(a * b)"
a / b     → "(a / b)"       (integer div; no UB check in release, check in debug)
a % b     → "(a % b)"
a ** b    → "ax_pow(a, b)"  (runtime helper for integer exponentiation)
a == b    → "(a == b)"
a != b    → "(a != b)"
a < b     → "(a < b)"
a <= b    → "(a <= b)"
a > b     → "(a > b)"
a >= b    → "(a >= b)"
a and b   → "(a && b)"      (short-circuit; b is a C expression so short-circuit is preserved)
a or b    → "(a || b)"
a & b     → "(a & b)"       (bitwise and)
a | b     → "(a | b)"       (bitwise or)
a ^ b     → "(a ^ b)"       (bitwise xor)
a << b    → "(a << b)"
a >> b    → "(a >> b)"
```

**Unary Operators**
```
not a     → "(!a)"
-a        → "(-a)"
~a        → "(~a)"
```

**Function Call**
```
foo(a, b)         → "ax_module_foo(a, b)"
std.math.sqrt(x)  → "ax_std_math_sqrt(x)"
```

Name is mangled using the module path + function name. Arguments are emitted recursively.

**Method Call**
```
obj.method(arg)   → "ax_Module_TypeName_method(&obj, arg)"
```
The receiver is passed as a pointer to the first parameter.

**Field Access**
```
s.x               → "s.x"
s.inner.y         → "s.inner.y"
```

**Index Expression (array/slice)**
```
arr[i]
→ (safe mode):   "(ax_bounds_check((ax_u64)(i), arr.len), arr.ptr[i])"
→ (unsafe mode): "arr.ptr[i]"
```
The bounds check uses the C comma operator to evaluate `ax_bounds_check` for its side effect (panic on failure) and then evaluate the actual index.

**Cast Expression**
```
x as i32          → "((ax_i32)(x))"
x as f64          → "((ax_f64)(x))"
x as u8           → "((ax_u8)(x))"
```

**Heap Dereference**
```
ref.*             → "(*((TargetType*)ax_deref(ref)))"
```
The `ax_deref` call validates the generational ID; the result is cast to the concrete pointer type and then dereferenced.

**Address-of (for borrows)**
```
&x                → "(&x)"
```

**Struct Literal**
```
Point{x: 1, y: 2} → "((struct ax_Point){.x=1, .y=2})"
```

**Slice Literal**
```
[1, 2, 3]
→ (static/stack): "((ax_slice_ax_i32){.ptr=(ax_i32[]){1,2,3}, .len=3, .cap=3})"
```

**Tuple (not a first-class type in AXIOM v1; lowered to struct)**
Not applicable in v1.

**Match Expression (when used as expression)**
Lowered to a ternary chain or a statement block returning a value via a temp variable.

**`spawn` (MVP: synchronous call)**
```
spawn foo(x)      → "ax_module_foo(x)"  // MVP: ignore spawn, call directly
```
Document that this is a placeholder; Phase 15 implements actual concurrency.

**`await` (MVP: identity)**
```
await x           → "x"  // MVP: await is a no-op
```

### Context: Safe vs Unsafe Mode
The `ExprGen` carries a boolean `unsafe bool`. When true:
- Array indexing omits `ax_bounds_check`
- Heap dereference omits `ax_deref` (emits `*((T*)ref.ptr)` instead)

### API
```go
type ExprGen struct {
    table  *typecheck.TypeTable
    queue  *TypeDeclQueue
    unsafe bool
}

// EmitExpr returns the C expression string for the given typed expression.
func (g *ExprGen) EmitExpr(expr ast.TypedExpr) string

// WithUnsafe returns a new ExprGen with unsafe=true.
func (g *ExprGen) WithUnsafe() *ExprGen
```

## Implementation Steps

### Step 1: Skeleton `emitExpr` with type switch
```go
func (g *ExprGen) EmitExpr(expr ast.TypedExpr) string {
    switch e := expr.(type) {
    case *ast.IntLit:    return g.emitIntLit(e)
    case *ast.FloatLit:  return g.emitFloatLit(e)
    case *ast.BoolLit:   return g.emitBoolLit(e)
    case *ast.StringLit: return g.emitStringLit(e)
    case *ast.Ident:     return g.emitIdent(e)
    case *ast.BinaryExpr:return g.emitBinary(e)
    case *ast.UnaryExpr: return g.emitUnary(e)
    case *ast.CallExpr:  return g.emitCall(e)
    case *ast.FieldExpr: return g.emitField(e)
    case *ast.IndexExpr: return g.emitIndex(e)
    case *ast.CastExpr:  return g.emitCast(e)
    case *ast.DerefExpr: return g.emitDeref(e)
    case *ast.AddrExpr:  return g.emitAddr(e)
    case *ast.StructLit: return g.emitStructLit(e)
    case *ast.SliceLit:  return g.emitSliceLit(e)
    case *ast.SpawnExpr: return g.emitSpawn(e) // MVP: sync call
    case *ast.AwaitExpr: return g.emitAwait(e) // MVP: identity
    default:
        panic(fmt.Sprintf("EmitExpr: unknown expr type %T", expr))
    }
}
```

### Step 2: Implement string literal emitter
```go
func (g *ExprGen) emitStringLit(e *ast.StringLit) string {
    escaped := strings.ReplaceAll(e.Value, `"`, `\"`)
    escaped = strings.ReplaceAll(escaped, "\n", `\n`)
    return fmt.Sprintf(`(ax_string){.ptr=(const ax_u8*)"%s", .len=%d}`,
        escaped, len(e.Value))
}
```

### Step 3: Implement index expression with bounds check
```go
func (g *ExprGen) emitIndex(e *ast.IndexExpr) string {
    arr := g.EmitExpr(e.Array)
    idx := g.EmitExpr(e.Index)
    if g.unsafe {
        return fmt.Sprintf("(%s).ptr[%s]", arr, idx)
    }
    return fmt.Sprintf("(ax_bounds_check((ax_u64)(%s), (%s).len), (%s).ptr[%s])",
        idx, arr, arr, idx)
}
```

### Step 4: Implement heap deref
```go
func (g *ExprGen) emitDeref(e *ast.DerefExpr) string {
    innerType := CTypeName(e.InnerTypeID, g.table, g.queue)
    ref := g.EmitExpr(e.Ref)
    if g.unsafe {
        return fmt.Sprintf("(*((%s*)(%s).ptr))", innerType, ref)
    }
    return fmt.Sprintf("(*((%s*)ax_deref(%s)))", innerType, ref)
}
```

### Step 5: Write `exprs_test.go`
Test all expression forms against expected C output strings.

## Test Plan
1. Integer literal `42` → `"42"`
2. String literal `"hello"` → correct compound literal with len=5
3. `a + b` → `"(a + b)"`
4. `a and b` → `"(a && b)"`
5. `not a` → `"(!a)"`
6. `foo(a, b)` → `"ax_module_foo(a, b)"`
7. `arr[i]` (safe) → bounds check comma expression
8. `arr[i]` (unsafe) → no bounds check
9. `ref.*` (safe) → `ax_deref` call
10. `ref.*` (unsafe) → direct `.ptr` access
11. `x as i32` → `((ax_i32)(x))`
12. `"hello"` string literal length is computed from byte count, not character count
13. `Point{x:1, y:2}` struct literal → compound literal
14. `spawn foo(x)` → synchronous call in MVP

## Validation Checklist
- [ ] All expression forms have corresponding case in type switch
- [ ] String literals with special characters are correctly escaped
- [ ] Bounds check uses `ax_u64` cast for the index (to handle signed/unsigned)
- [ ] `WithUnsafe()` returns a new generator (does not mutate original)
- [ ] All tests pass

## Acceptance Criteria
- Generated expressions are valid C11 expressions
- Bounds checks are emitted in safe mode for all array indexing
- Heap dereferences go through `ax_deref` in safe mode
- All test cases pass

## Definition of Done
- `codegen/cgen/exprs.go` exists with all expression forms handled
- `codegen/cgen/exprs_test.go` exists and passes
- `go test ./codegen/cgen/` passes

## Risks & Mitigations
- **Risk**: The comma-operator pattern `(check(), access)` can confuse some static analyzers. **Mitigation**: Document the pattern; it is valid C11. For `-O2` GCC/Clang, the check will be inlined and optimized away if provably in-range.
- **Risk**: Binary operator precedence: wrapping every binary expression in parentheses avoids precedence bugs. **Mitigation**: Always wrap in `(...)` — already required by the format strings above.

## Future Follow-up Tasks
- p08-t05: Ownership-aware expression generation (moves, borrows)
- p08-t06: Generational check emission at allocation sites
- p09-t06: AIR builder for expressions mirrors this structure
