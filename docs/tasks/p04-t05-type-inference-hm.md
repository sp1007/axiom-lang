# p04-t05: Type Inference (Local Hindley-Milner)

## Purpose
Implement local type inference so that `let x = 42` infers `x: i32` and `let y = 3.14` infers `y: f64`, without requiring explicit type annotations on local variables. Function signatures must always be explicit (no global inference).

## Context
AXIOM uses bidirectional local type inference based on Hindley-Milner. Bottom-up: leaf nodes get types from literals (integer literals → i32, float literals → f64, string literals → string, true/false → bool). Propagation: binary expressions use `CommonType()`, function calls use declared return type. Top-down: `let x: f64 = 42` pushes expected type f64 onto the integer literal, widening i32→f64.

**Spec references:** `04. Type checker.md` — local HM inference rules

## Inputs
- `compiler/sema/resolver.go` — resolved AST with symbol indices (p04-t04)
- `compiler/types/typetable.go` — TypeTable with CommonType/CanImplicitCast (p04-t02)

## Outputs
- `compiler/sema/inference.go` — InferenceEngine with type propagation
- `compiler/sema/inference_test.go` — ≥18 test cases

## Dependencies
- p04-t04: name-resolver — AST must be resolved before type inference
- p04-t02: type-table-primitives — TypeTable provides type operations

## Detailed Requirements

### InferenceEngine
```go
type InferenceEngine struct {
    ast      *ast.AstTree
    symtable *SymbolTable
    types    *TypeTable
    errors   []diagnostics.Diagnostic
}
func NewInferenceEngine(tree *ast.AstTree, st *SymbolTable, tt *TypeTable) *InferenceEngine
func (ie *InferenceEngine) Infer() []diagnostics.Diagnostic
func (ie *InferenceEngine) TypeOf(nodeIdx uint32) TypeID
```

### Inference Rules

**Literals:**
- Integer literal → `TypeI32` (default), or expected type if context provides one
- Float literal → `TypeF64` (default)
- String literal → `TypeString`
- Bool literal → `TypeBool`
- `nil` → must have expected type from context (error if ambiguous)

**Variables:**
- `let x = expr` → infer type of expr, assign to x's symbol TypeID
- `let x: T = expr` → bidirectional: verify expr type assignable to T
- `mut x := expr` → same as let, with SymFlagMut

**Binary expressions:**
- `a + b` → `CommonType(TypeOf(a), TypeOf(b))`, error if incompatible
- `a == b` → operands must have same type (or implicitly castable), result is `TypeBool`
- `a and b` → both must be `TypeBool`, result `TypeBool`

**Function calls:**
- `foo(args)` → look up foo's FuncType, match arg types to param types, result is return type
- Argument count mismatch → error
- Argument type mismatch → error (with implicit cast attempt)

**If expressions:**
- `if cond: a elif ...: b else: c` → all branches must have same type (or CommonType)

**Return statements:**
- `return expr` → expr type must match enclosing function's declared return type

**Array/index:**
- `arr[i]` → i must be integer, result is element type of arr

### Error Diagnostics
- `"type mismatch: expected T1, found T2"` (code 3001)
- `"cannot infer type of 'nil' without context"` (code 3002)
- `"argument count mismatch: expected N, got M"` (code 3003)
- `"branches of if expression have incompatible types"` (code 3004)
- `"return type mismatch: expected T1, found T2"` (code 3005)

## Implementation Steps
1. Create `compiler/sema/inference.go` with InferenceEngine.
2. Implement `inferExpr(nodeIdx uint32, expected TypeID) TypeID` — bottom-up with optional expected type.
3. Implement literal inference: check NodeKind, return appropriate primitive TypeID.
4. Implement binary expr inference: infer both operands, call CommonType.
5. Implement call inference: resolve function symbol, match args to params.
6. Implement variable declaration inference: infer RHS, update symbol TypeID.
7. Implement bidirectional: when expected type provided, apply implicit cast check.
8. Write tests.

## Test Plan
1. `TestInfer_IntLiteral`: `42` → TypeI32
2. `TestInfer_FloatLiteral`: `3.14` → TypeF64
3. `TestInfer_StringLiteral`: `"hello"` → TypeString
4. `TestInfer_BoolLiteral`: `true` → TypeBool
5. `TestInfer_LetInfer`: `let x = 42` → x is TypeI32
6. `TestInfer_LetExplicit`: `let x: f64 = 42` → x is TypeF64 (widened)
7. `TestInfer_LetMismatch`: `let x: bool = 42` → type mismatch error
8. `TestInfer_BinaryAdd`: `1 + 2` → TypeI32
9. `TestInfer_BinaryMixed`: `1 + 2.0` → TypeF64 (widened)
10. `TestInfer_BinaryCompare`: `1 == 2` → TypeBool
11. `TestInfer_BinaryLogical`: `true and false` → TypeBool
12. `TestInfer_FuncCall`: `fn foo() -> i32; foo()` → TypeI32
13. `TestInfer_FuncCallArgMismatch`: wrong arg type → error
14. `TestInfer_FuncCallArgCount`: wrong arg count → error
15. `TestInfer_IfExpr`: `if c: 1 else: 2` → TypeI32
16. `TestInfer_IfExprMismatch`: `if c: 1 else: "x"` → incompatible branches error
17. `TestInfer_ReturnType`: `fn foo() -> i32: return "x"` → return mismatch error
18. `TestInfer_NilNoContext`: `let x = nil` → cannot infer nil error

## Validation Checklist
- [ ] All primitive literals inferred correctly
- [ ] Bidirectional inference works (explicit type annotation)
- [ ] CommonType used for binary expressions
- [ ] Function call arg types checked
- [ ] Type mismatches produce diagnostics, not panics
- [ ] `go test ./compiler/sema/ -run TestInfer` passes

## Acceptance Criteria
- All 18 tests pass
- `axc check` on compliance suite groups 1–3: no false type errors

## Definition of Done
- [ ] `compiler/sema/inference.go` implemented
- [ ] 18 tests passing
- [ ] No panics on error paths

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Integer literal defaulting to i32 may surprise users expecting i64 | Document default; allow suffix `42i64` in future |
| Bidirectional inference complex for nested expressions | Start with single-level; extend to nested in follow-up |

## Future Follow-up Tasks
- p04-t06: type checker statements uses inference results
- p04-t07: type checker expressions builds on inference
