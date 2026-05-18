# p05-t01: Generic Type Representation

## Purpose
Extend the type system to represent generic (parameterized) types and functions. This provides the data structures needed by monomorphization (p05-t02) to clone and specialize generic definitions for each concrete type argument.

## Context
AXIOM uses monomorphized generics — like C++ templates or Rust's generics — rather than type erasure (like Java). Each unique instantiation (e.g., `List[i32]` and `List[string]`) gets its own cloned and specialized copy. Generic templates are stored separately from instantiations; instantiation creates a new TypeID for each unique concrete combination.

## Inputs
- TypeTable from p04-t02
- Parsed AST with generic declarations (`fn sort[T](...)` and `struct Stack[T]`)

## Outputs
- Extended `TypeInfo` with generic fields
- `GenericTemplate` data structure for generic functions/structs
- `compiler/types/generics.go`

## Dependencies
- p04-t02: type-table-primitives — TypeInfo extended here
- p03-t04: parser-statements — generic syntax already parsed into AST

## Subsystems Affected
- Type system: TypeInfo gains IsGeneric, TypeParams fields
- Symbol table: generic functions stored with generic flag
- Monomorphization (p05-t02): consumes GenericTemplate

## Detailed Requirements

1. Extend `TypeInfo` (in `compiler/types/types.go`):
   ```go
   type TypeInfo struct {
       // ... existing fields ...
       IsGeneric      bool
       TypeParams     []GenericParam    // e.g., [T, U]
       TypeConstraints []uint32         // interface TypeID per param (0 = unconstrained)
   }
   type GenericParam struct {
       NameID     uint32  // interned name of the type parameter
       Constraint uint32  // interface TypeID, 0 if none
   }
   ```
2. `GenericTemplate` struct:
   ```go
   type GenericTemplate struct {
       SymID       uint32        // symbol of the generic function/struct
       NodeIdx     uint32        // AST node of the template
       Params      []GenericParam
       Instances   map[string]uint32  // "i32,string" → instantiated TypeID
   }
   ```
3. When parsing `fn sort[T](list: [T]) -> [T]`: register as generic function with TypeParam `T`, no constraint.
4. When parsing `fn max[T: Ord](a: T, b: T) -> T`: TypeParam `T` with constraint = TypeID of `Ord` interface.
5. Type parameter names are scoped to the function/struct — `T` in `sort[T]` is different from `T` in `Stack[T]`.
6. In TypeTable: generic templates stored in `GenericTemplates []GenericTemplate`.
7. Lookup: `TypeTable.GetGenericTemplate(symID uint32) *GenericTemplate`.

## Implementation Steps

1. Add `GenericParam`, `GenericTemplate` types to `compiler/types/generics.go`.
2. Extend `TypeInfo` with `IsGeneric`, `TypeParams`, `TypeConstraints`.
3. In parser: when a FuncDecl or StructDecl has `[T, U]` type parameters, set `IsGeneric=true` and fill `TypeParams`.
4. In name resolver: when encountering `T` inside a generic function body, resolve it to the corresponding `GenericParam` (create a pseudo-symbol for each type parameter).
5. In TypeTable: add `GenericTemplates []GenericTemplate` and `RegisterGenericTemplate()` method.
6. Write unit tests: `TestGenericFuncRegistration`, `TestGenericStructRegistration`, `TestConstrainedGeneric`.

## Test Plan

- `TestRegisterGenericFunc`: `fn sort[T](list: [T])` → registered with 1 type param
- `TestRegisterConstrainedGeneric`: `fn max[T: Ord](a: T, b: T) -> T` → TypeParam with Ord constraint
- `TestGenericStructParams`: `struct Stack[T]{items: [T]}` → 1 type param, field type references T
- `TestTypeParamScope`: T in sort[T] different from T in Stack[T] → separate GenericParams

## Validation Checklist

- [ ] Generic flag set correctly on TypeInfo
- [ ] TypeParams populated with correct names and constraints
- [ ] Type parameters scoped to their declaration
- [ ] GenericTemplate registered in TypeTable
- [ ] Lookup by symID works

## Acceptance Criteria

- `fn sort[T](list: [T]) -> [T]` parsed and registered with no errors
- Type parameter `T` resolves correctly inside the generic function body

## Definition of Done

- [ ] `compiler/types/generics.go` created
- [ ] TypeInfo extended
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Type parameter name conflicts with outer scope | Type params always shadow outer scope; check in resolver |
| Multiple type parameters with same name | Validate in parser: `fn foo[T, T]` → error |

## Future Follow-up Tasks

- p05-t02: monomorphization clones GenericTemplates
- p05-t04: structural duck typing checks constraints
