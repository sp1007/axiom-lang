# p04-t07: Type Checker — Expressions

## Purpose
Implement type checking for all AXIOM expression forms, validating operator compatibility, function call argument types, field access validity, index bounds types, and cast legality. Together with p04-t06, this completes the core type-safety guarantee of the AXIOM compiler.

## Context
Expression type checking reads the TypeID already annotated in each node's Payload (set by type inference in p04-t05) and validates the semantic rules. For cases where inference is insufficient (e.g., overloaded operators, generic calls), the type checker computes the correct type. All expression checks produce diagnostics — they never panic or silently accept invalid programs.

## Inputs
- Type-inferred AST (all expression Payload fields set to TypeIDs)
- TypeTable (p04-t02)
- SymbolTable (p04-t01)

## Outputs
- `[]Diagnostic` — expression type errors
- Corrected TypeIDs in Payload for expressions where inference was deferred to type checker

## Dependencies
- p04-t05: type-inference-hm — initial TypeIDs set
- p04-t06: type-checker-statements — shares `TypeChecker` struct
- p04-t08: overload-resolution — used for operator and call resolution

## Subsystems Affected
- Semantic analysis: expression checking is the final validation before codegen
- Type safety: guarantees no type confusion reaches code generation

## Detailed Requirements

1. `BinaryExpr` check: verify left and right operand types are compatible with the operator:
   - `+`, `-`, `*`, `/`, `%`: both operands must be numeric (same type or coercible). Result type = wider of the two.
   - `**`: both numeric. Result = f64 if either is float.
   - `==`, `!=`, `<`, `>`, `<=`, `>=`: comparable types (numeric, bool, string). Result = bool.
   - `and`, `or`: both must be bool. Result = bool.
   - `&`, `|`, `^`, `<<`, `>>`: both must be integer. Result = left type.
2. `UnaryExpr` check: `-` → numeric operand; `not` → bool operand; `~` → integer operand.
3. `CallExpr` check: resolve callee symbol's TypeInfo (TKFunc), verify arg count matches param count, verify each arg type matches param type (using overload resolution for methods).
4. `IndexExpr` check: collection must be slice/array type; index must be integer type. Result type = element type.
5. `FieldExpr` check: object must be struct type; field name must exist in struct's TypeInfo.Fields. Result type = field type.
6. `CastExpr` (`as` operator): verify cast is legal. Legal casts: numeric↔numeric, pointer↔pointer, integer↔bool. Illegal: string↔int (must use parse functions).
7. `AwaitExpr`: only valid inside `async fn`; operand must be `Future[T]`; result type = T.
8. `SpawnExpr`: operand must be a function call; result type = `ActorRef`.

## Implementation Steps

1. Create `compiler/sema/checker_exprs.go`.
2. Implement `checkExpr(nodeIdx) uint32` — returns the TypeID of the expression.
3. Handle BinaryExpr: call `overloadResolver.Resolve(op, leftType, rightType)` to get result type.
4. Handle CallExpr: look up function TypeInfo, validate arg types one by one.
5. Handle FieldExpr: look up struct TypeInfo, find field by NameID.
6. Handle IndexExpr: verify collection type is TKSlice or TKArray, verify index is integer.
7. Handle CastExpr: implement cast validity table.
8. For each invalid operation: emit diagnostic with specific message.
9. Integration test: compile compliance tests 001-010 (primitives) → 0 errors.

## Test Plan

- `TestBinaryExprTypes`: `1 + 2` → i32, `1.0 + 2.0` → f64, `1 + "a"` → error
- `TestCallArgCount`: `fn foo(a: i32)` called with `foo(1, 2)` → "expected 1 args, got 2"
- `TestCallArgType`: `fn foo(a: i32)` called with `foo("hello")` → type mismatch
- `TestFieldAccess`: struct Foo{x: i32} — `foo.x` → i32; `foo.y` → "no field y on Foo"
- `TestIndexExpr`: `arr[0]` where arr is [i32] → i32; `arr["a"]` → "index must be integer"
- `TestCastLegal`: `42 as f32` → f32, `42 as bool` → bool
- `TestCastIllegal`: `"hello" as i32` → "illegal cast from string to i32"

## Validation Checklist

- [ ] All 8 expression kinds checked
- [ ] Operator type rules enforced
- [ ] Call argument count and types validated
- [ ] Field access validated against struct definition
- [ ] Illegal casts rejected
- [ ] No false positives for valid AXIOM programs

## Acceptance Criteria

- Compliance tests 001-010 compile with 0 errors
- Each error case produces exactly one diagnostic with file:line:col
- No panic on any valid or invalid AST

## Definition of Done

- [ ] `compiler/sema/checker_exprs.go` implemented
- [ ] `go test ./compiler/sema/ -run TestCheckExpr` passes
- [ ] Integrated with checker_stmts.go into unified CheckProgram()

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Numeric promotion rules complex | Define explicit promotion table; test all combinations |
| Generic call type checking requires monomorphization | Defer generic call checking to after monomorphization (p05-t02) |

## Future Follow-up Tasks

- p05-t02: monomorphization validates generic call sites
- p05-t03: sum types add variant pattern type checking
- p06-t02: ownership rules layer on top of type checking
