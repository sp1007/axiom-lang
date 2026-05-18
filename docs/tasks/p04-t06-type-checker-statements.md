# p04-t06: Type Checker — Statements

## Purpose
Implement type checking for all statement forms in AXIOM: variable declarations, assignments, if/elif/else, for/in loops, match, return, break/continue, defer, spawn, lock, and function declarations. Each statement must be verified for type correctness and annotated with resolved TypeIDs.

## Context
The statement type checker operates on a resolved AST (after name resolution and inference). It validates semantic rules that go beyond simple type matching: assignments must respect mutability, for-loop iterators must implement iteration protocol, match arms must be exhaustive for sum types, return types must match function signature.

**Spec references:** `04. Type checker.md` — statement typing rules

## Inputs
- `compiler/sema/inference.go` — InferenceEngine (p04-t05)
- `compiler/sema/resolver.go` — resolved AST (p04-t04)
- `compiler/types/typetable.go` — TypeTable (p04-t02)
- `compiler/sema/symtable.go` — SymbolTable (p04-t01)

## Outputs
- `compiler/sema/check_stmt.go` — statement type checking functions
- `compiler/sema/check_stmt_test.go` — ≥20 test cases

## Dependencies
- p04-t05: type-inference-hm
- p04-t04: name-resolver
- p04-t02: type-table-primitives
- p04-t01: symbol-table

## Detailed Requirements

### TypeChecker Core
```go
type TypeChecker struct {
    ast      *ast.AstTree
    intern   *ast.InternPool
    symtable *SymbolTable
    types    *TypeTable
    infer    *InferenceEngine
    errors   []diagnostics.Diagnostic
    currentFuncReturnType TypeID  // tracks expected return type in current function
}
func NewTypeChecker(tree *ast.AstTree, intern *ast.InternPool,
    st *SymbolTable, tt *TypeTable) *TypeChecker
func (tc *TypeChecker) Check() []diagnostics.Diagnostic
```

### Statement Rules

**LetDecl/MutDecl:**
- If type annotation present: verify RHS type assignable to declared type
- If no annotation: use inferred type from RHS
- Set symbol's TypeID in SymbolTable

**Assignment (`a = expr`):**
- `a` must have `SymFlagMut` — error if assigning to immutable: `"cannot assign to immutable variable 'a'"`
- RHS type must be assignable to LHS type
- If `a` is heap-owning and RHS is a new value: previous value needs `=destroy` (handled in p06)

**IfStmt:**
- Condition must be `TypeBool` — error: `"if condition must be bool, found T"`
- Body statements checked recursively
- elif/else branches checked recursively

**ForStmt (for x in iter):**
- `iter` must be iterable (Array, Seq, range `0..N`)
- Loop variable `x` type inferred from iterator element type
- For range `0..N`: x is TypeI32; N must be integer
- Body checked with loop variable in scope

**MatchStmt:**
- Scrutinee expression type checked
- Each arm pattern checked for compatibility with scrutinee type
- Sum type match: warn if not exhaustive (all variants covered)
- Binding patterns (`i32(v)`) introduce variable `v` in arm scope

**ReturnStmt:**
- Expression type must match `currentFuncReturnType`
- `return` without expression: function must return `TypeVoid`
- Error: `"return type mismatch: expected T1, found T2"`

**Break/Continue:**
- Must be inside a loop — error: `"break outside of loop"`
- No type requirements

**DeferStmt:**
- Deferred expression must be a function call
- No return value captured

**SpawnStmt:**
- Spawned expression must be a function call
- Function must accept `Isolated[T]` arguments (validated in p06)

**LockStmt (`lock x as y:`):**
- `x` must be `Locker[T]` type
- `y` is bound as `mut T` within block

### Error Diagnostics
- `"cannot assign to immutable variable 'X'"` (code 3010)
- `"if condition must be bool, found T"` (code 3011)
- `"for iterator must be iterable, found T"` (code 3012)
- `"return type mismatch: expected T1, found T2"` (code 3005)
- `"break outside of loop"` (code 3013)
- `"continue outside of loop"` (code 3014)
- `"defer must be a function call"` (code 3015)
- `"non-exhaustive match: missing variant V"` (code 3016, warning)

## Implementation Steps
1. Create `compiler/sema/check_stmt.go`.
2. Implement `checkStmt(nodeIdx uint32)` — dispatch on NodeKind.
3. Implement each statement rule as a separate method.
4. Track `currentFuncReturnType` when entering/exiting function bodies.
5. Track `insideLoop bool` for break/continue validation.
6. Write tests.

## Test Plan
1. `TestCheck_LetWithAnnotation`: `let x: i32 = 42` → OK
2. `TestCheck_LetTypeMismatch`: `let x: bool = 42` → error
3. `TestCheck_MutAssign`: `mut x := 1; x = 2` → OK
4. `TestCheck_ImmutableAssign`: `let x = 1; x = 2` → error "cannot assign to immutable"
5. `TestCheck_IfCondBool`: `if true: ...` → OK
6. `TestCheck_IfCondNotBool`: `if 42: ...` → error "must be bool"
7. `TestCheck_ForRange`: `for i in 0..10: i` → i is i32
8. `TestCheck_ForArray`: `for x in arr: x` → x is element type
9. `TestCheck_ReturnMatch`: `fn foo() -> i32: return 42` → OK
10. `TestCheck_ReturnMismatch`: `fn foo() -> i32: return "x"` → error
11. `TestCheck_ReturnVoid`: `fn foo(): return` → OK (void function)
12. `TestCheck_BreakInLoop`: `for i in 0..10: break` → OK
13. `TestCheck_BreakOutsideLoop`: `break` at top level → error
14. `TestCheck_DeferCall`: `defer close(f)` → OK
15. `TestCheck_DeferNotCall`: `defer 42` → error
16. `TestCheck_MatchExhaustive`: all variants covered → OK
17. `TestCheck_MatchNonExhaustive`: missing variant → warning
18. `TestCheck_MatchBinding`: `match x: i32(v) => v` → v has i32 type
19. `TestCheck_NestedScopes`: variables in nested if blocks → correctly scoped
20. `TestCheck_FunctionBody`: full function with params, locals, return → all types valid

## Validation Checklist
- [ ] All statement forms type-checked
- [ ] Mutability enforced on assignment
- [ ] If conditions verified as bool
- [ ] Return types match function signature
- [ ] Break/continue only inside loops
- [ ] `go test ./compiler/sema/ -run TestCheck` passes

## Acceptance Criteria
- All 20 tests pass
- `axc check` on compliance suite groups 1–4: zero false errors

## Definition of Done
- [ ] `compiler/sema/check_stmt.go` implemented
- [ ] 20 tests passing

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Match exhaustiveness check complex for nested sum types | Start with flat sum types only; nested in follow-up |
| For-loop iterable protocol not yet defined | Accept Array and range `0..N` initially; generalize with Iterator interface |

## Future Follow-up Tasks
- p04-t07: type checker expressions (completes full type checking)
- p04-t08: overload resolution for function calls
- p06-t02: ownership rules extend assignment checking
