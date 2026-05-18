# p05-t03: Sum Types (Tagged Unions)

## Purpose
Implement sum types (tagged unions / discriminated unions) — the `type X = A | B` syntax — including their type representation, pattern matching exhaustiveness checking, and C-backend representation as tagged C unions.

## Context
Sum types are central to AXIOM's error handling and data modeling. `type Result[T, E] = Ok(T) | Err(E)` is the primary error handling mechanism (no exceptions). Pattern matching via `match` must cover all variants exhaustively. In the C backend, sum types map to a struct with a tag field (uint8) and a union of payload types.

## Inputs
- Parsed AST with sum type declarations (`NodeKind.SumTypeDecl`)
- TypeTable for registering new sum type TypeIDs
- Type checker for exhaustiveness analysis

## Outputs
- `TypeInfo{Kind:TKSumType, Variants:[]VariantInfo}` registered in TypeTable
- Exhaustiveness checker for match expressions
- `compiler/types/sumtype.go`

## Dependencies
- p05-t01: generic-type-representation — `Result[T, E]` is a generic sum type
- p04-t06: type-checker-statements — match statement uses exhaustiveness checker
- p04-t02: type-table-primitives — new TypeKind.TKSumType added

## Subsystems Affected
- Type system: TKSumType kind and VariantInfo added
- Pattern matching: exhaustiveness checking
- C-backend: tagged union generation

## Detailed Requirements

1. `VariantInfo` struct:
   ```go
   type VariantInfo struct {
       NameID      uint32  // interned variant name (e.g., "Ok", "Err")
       PayloadType uint32  // TypeID of payload (0 if unit variant)
       Tag         uint8   // numeric tag value (0, 1, 2, ...)
   }
   ```
2. Extend `TypeInfo` with `Variants []VariantInfo` (non-nil only when TKSumType).
3. Syntax: `type Result[T, E] = Ok(T) | Err(E)` — generic sum type.
4. Unit variant: `type Color = Red | Green | Blue` — each variant has PayloadType=0 (TypeNil).
5. Pattern matching: `match x { Ok(v): expr | Err(e): expr }` — `v` is bound to the Ok payload, `e` to Err payload.
6. Exhaustiveness check: for a match on sum type, all variants must have an arm OR a wildcard `_` arm exists.
7. Type checking in match: in each arm, bind the payload variable with correct TypeID.
8. C-backend representation:
   ```c
   enum ax_Result_tag { ax_Result_Ok = 0, ax_Result_Err = 1 };
   typedef struct {
       enum ax_Result_tag tag;
       union {
           ax_i32 ok;      // T=i32 for Result[i32, string]
           ax_string err;  // E=string
       } data;
   } ax_Result_i32_string;
   ```
9. Constructor functions: `Ok(v)` → `(ax_Result_i32_string){.tag=ax_Result_Ok, .data.ok=v}`.

## Implementation Steps

1. Create `compiler/types/sumtype.go` with `VariantInfo`.
2. Add `TKSumType` to `TypeKind` enum.
3. In parser: parse `type X = A(T) | B(U)` → create `NodeKind.SumTypeDecl` with variant children.
4. In type checker: scan SumTypeDecl nodes, register each as `TypeInfo{TKSumType, Variants:[...]}`.
5. Implement exhaustiveness checker: `CheckExhaustive(matchNode, sumTypeID)` — returns missing variants.
6. In match arm type checking: create new scope with bound variable having the payload TypeID.
7. In C-backend: generate tagged union struct and constructor expressions.
8. Write tests: `TestSumTypeBasic`, `TestSumTypeMatch`, `TestSumTypeExhaustive`, `TestSumTypeGeneric`.

## Test Plan

- `TestSumTypeColor`: `type Color = Red | Green | Blue` → TypeInfo with 3 unit variants
- `TestSumTypeResult`: `type Result[T,E] = Ok(T) | Err(E)` → generic sum type
- `TestMatchExhaustive`: match with all arms → no error
- `TestMatchNonExhaustive`: missing Err arm → "non-exhaustive match: missing Err"
- `TestMatchWildcard`: match with `_` arm → exhaustive regardless
- `TestMatchBinding`: `Ok(v)` → v bound as TypeID of payload type
- `TestSumTypeCompliance`: compliance tests 051-060 (error handling group) pass

## Validation Checklist

- [ ] Sum type TypeInfo registered with correct variants
- [ ] Tags assigned sequentially starting at 0
- [ ] Exhaustiveness checker catches missing variants
- [ ] Wildcard arm satisfies exhaustiveness
- [ ] Payload binding gives correct TypeID in match arm
- [ ] C-backend generates valid tagged union

## Acceptance Criteria

- `type Result[T,E] = Ok(T) | Err(E)` compiles without errors
- Non-exhaustive match produces exactly one diagnostic naming missing variants
- Compliance tests 051-060 pass

## Definition of Done

- [ ] `compiler/types/sumtype.go` implemented
- [ ] Exhaustiveness checking integrated into type checker
- [ ] C-backend sum type codegen implemented (p08 prerequisite)
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Nested sum types in match (match inside Ok arm) | Use recursive arm checking |
| Generic sum types require monomorphization | `Result[i32, string]` is a monomorphized instance; integrate with p05-t02 |

## Future Follow-up Tasks

- p08-t02: cgen-declarations generates tagged union C structs
- p16-t01: std.testing.assert uses Result type internally
