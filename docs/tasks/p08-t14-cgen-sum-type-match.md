# p08-t14: C-Backend Sum Type + Match Expression

## Purpose
Implement C code generation for AXIOM sum types (`type X = A | B`) as tagged unions and `match` expressions as switch/if-else chains. This covers `Result[T, E]`, `Option[T]`, and user-defined sum types.

## Context
Plan §Phase 4 maps sum types to `tagged union struct` in C and match to `switch`/if-else. Sum types are the foundation for error handling (`Result[T, E]`) and null safety (`Option[T]`). The C-backend must generate correct discriminated union layouts and pattern matching dispatch.

## Inputs
- Sum type definitions from p05-t03 (TypeTable entries for sum types)
- TypedAST match expressions with discriminant + arms
- C-backend type mapping from p08-t01

## Outputs
- `codegen/cgen/sumtype.go` — tagged union struct generation
- `codegen/cgen/match.go` — match expression → switch/if-else
- Tests and golden output files

## Dependencies
- p05-t03: sum-types — TypeTable has sum type entries with variants
- p08-t01: cgen-type-mapping — C type mapping infrastructure
- p08-t04: cgen-expressions — expression emission framework

## Subsystems Affected
- C-Backend: new struct layout for sum types, new codegen for match
- Runtime: discriminant tags use `uint8_t` for up to 256 variants

## Detailed Requirements

### 1. Tagged Union Layout

```axiom
type Shape = Circle(f64) | Rect(f64, f64) | Empty
```

Generated C:
```c
typedef struct {
    uint8_t tag;
    union {
        struct { double radius; } Circle;
        struct { double w; double h; } Rect;
        // Empty has no data
    } data;
} _AX_Shape;

enum _AX_Shape_Tag { _AX_Shape_Circle = 0, _AX_Shape_Rect = 1, _AX_Shape_Empty = 2 };
```

### 2. Result[T, E] and Option[T]

These are built-in sum types with specialized C layouts:

```c
// Option[i32]
typedef struct { uint8_t tag; int32_t value; } _AX_Option_i32;
enum { _AX_Option_i32_None = 0, _AX_Option_i32_Some = 1 };

// Result[i32, string]
typedef struct {
    uint8_t tag;
    union { int32_t ok; ax_string err; } data;
} _AX_Result_i32_string;
enum { _AX_Result_i32_string_Ok = 0, _AX_Result_i32_string_Err = 1 };
```

### 3. Match Expression Code Generation

```axiom
match shape:
    Circle(r) => 3.14 * r * r
    Rect(w, h) => w * h
    Empty => 0.0
```

Generated C:
```c
double _match_result;
switch (shape.tag) {
    case _AX_Shape_Circle: {
        double r = shape.data.Circle.radius;
        _match_result = 3.14 * r * r;
        break;
    }
    case _AX_Shape_Rect: {
        double w = shape.data.Rect.w;
        double h = shape.data.Rect.h;
        _match_result = w * h;
        break;
    }
    case _AX_Shape_Empty: {
        _match_result = 0.0;
        break;
    }
}
```

### 4. Exhaustiveness

Type checker (p05-t03) verifies exhaustiveness. C-backend can add `default: __builtin_unreachable();` for safety.

### 5. Nested Patterns

Nested patterns like `Some(Ok(x))` desugar to nested switches. MVP: only single-level patterns. Nested patterns deferred to future RFC.

## Implementation Steps

1. Create `codegen/cgen/sumtype.go` — `EmitSumTypeDecl()` generating tagged union struct + enum.
2. Create `codegen/cgen/match.go` — `EmitMatchExpr()` generating switch statement.
3. Register all sum types in type mapping phase (p08-t01 integration).
4. Handle `Result[T,E]` and `Option[T]` as monomorphized sum types.
5. Emit constructor helpers: `_AX_Option_i32_some(int32_t v)`, `_AX_Option_i32_none()`.
6. Write tests for each variant pattern.

## Test Plan

- `TestSumTypeDecl`: sum type → correct tagged union struct
- `TestMatchSimple`: 3-arm match → switch with correct bindings
- `TestResultOkErr`: `Result[i32, string]` layout correct
- `TestOptionSomeNone`: `Option[i32]` constructor and match
- `TestMatchExhaustive`: missing arm → `default: __builtin_unreachable()`
- Golden test: `tests/golden/cgen/sum_type.ax` → expected C output

## Validation Checklist

- [ ] Tagged union struct has correct alignment and size
- [ ] Match arms bind variables correctly
- [ ] Result and Option monomorphized correctly
- [ ] Generated C compiles without warnings
- [ ] Exhaustiveness default case present

## Acceptance Criteria

- Compliance tests using `Result[T,E]` and `Option[T]` compile and run correctly
- Match expressions produce correct runtime behavior

## Definition of Done

- [ ] `codegen/cgen/sumtype.go` implemented
- [ ] `codegen/cgen/match.go` implemented
- [ ] Golden tests pass
- [ ] E2E test: program with sum types runs correctly

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Union alignment varies across platforms | Use `__attribute__((aligned))` or compute max field size |
| Large sum types waste memory | Future: optimize small variants with NaN-boxing (Phase 10+) |

## Future Follow-up Tasks

- p09-t06: AIR builder for match expressions
- p10-t02: constant folding for match on known discriminant
- p16-t12: std.result_option uses these C-backend sum types
