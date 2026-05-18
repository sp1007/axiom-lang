# p05-t02: Monomorphization Engine

## Purpose
Implement monomorphization — the process of cloning and specializing generic functions and structs for each unique combination of concrete type arguments. This produces non-generic, fully typed AST/TypeInfo entries that can be directly compiled without runtime type dispatch.

## Context
When the type checker encounters `sort[i32](my_list)`, monomorphization creates a new specialized version `sort_i32` with all occurrences of `T` replaced by `i32`. The specialized version is type-checked like a normal function. Results are cached — calling `sort[i32]` 100 times produces exactly one monomorphized copy. Name mangling distinguishes specialized versions: `_AX_std_sort_i32`.

## Inputs
- `GenericTemplate` instances from TypeTable (p05-t01)
- Call sites with concrete type arguments (identified by type checker)
- AST tree for cloning

## Outputs
- New `TypeInfo` entries for each instantiation (e.g., `Stack[i32]` gets its own TypeID)
- New function symbols for each instantiation (e.g., `sort_i32`)
- Updated call sites: `sort[T]` → `sort_i32` (SymID updated)

## Dependencies
- p05-t01: generic-type-representation — GenericTemplate data structures
- p04-t07: type-checker-expressions — call sites with type argument lists
- p04-t02: type-table-primitives — new TypeIDs registered here

## Subsystems Affected
- Type table: new TypeIDs added for each instantiation
- Symbol table: new function symbols added
- AST: cloned subtrees for instantiated generic bodies
- Code generation: each monomorphized function compiled independently

## Detailed Requirements

1. `Monomorphizer` struct: `tt *TypeTable, st *SymbolTable, tree *AstTree, cache map[string]uint32`
2. Cache key: `"<templateSymID>:<type1_id>,<type2_id>..."` — string representation for map key.
3. `Instantiate(templateSymID uint32, typeArgs []uint32) uint32` → returns SymID of instantiation:
   - Check cache; if hit, return cached SymID
   - Clone AST subtree of generic function
   - Replace all TypeVar references (T, U) with concrete typeArgs
   - Register new TypeInfo with `IsGeneric=false`
   - Type-check the cloned subtree
   - Add to cache
   - Return new SymID
4. AST cloning: `cloneSubtree(nodeIdx uint32) uint32` — deep copy of all nodes, returns new root index.
5. Type substitution in cloned tree: walk all TypeExpr nodes, replace TypeParam references with concrete TypeIDs.
6. Name mangling: `_AX_<module>_<name>_<type1>_<type2>` — used as the new function's C name.
7. Generic struct instantiation: `Stack[i32]` → clone struct definition, replace T→i32 in all field types, register as new TypeID.

## Implementation Steps

1. Create `compiler/sema/mono.go` with `Monomorphizer`.
2. Implement `cloneSubtree()` in `ast/tree.go` — allocates new nodes, copies all fields.
3. Implement `substituteTypeParams(nodeIdx, paramMap map[uint32]uint32)` — walk clone, replace TypeParam IDs.
4. Implement `Instantiate()`: check cache → clone → substitute → type-check clone → register → cache.
5. Hook into type checker: when `callExpr` has type arguments `[T=i32]`, call `Instantiate()` and update the CallExpr's callee SymID.
6. Implement `mangleName(module, name string, typeArgs []uint32) string`.
7. Write tests: `TestMonoSortI32`, `TestMonoCached`, `TestMonoStack`.

## Test Plan

- `TestMonoBasic`: call `identity[i32](5)` → creates `identity_i32` function
- `TestMonoCaching`: call `identity[i32]` twice → only one monomorphized copy
- `TestMonoTwoTypes`: call `pair[i32, string]` → creates `pair_i32_string`
- `TestMonoStruct`: `Stack[i32]` and `Stack[string]` → two different TypeIDs
- `TestMonoRecursive`: `fn fib[T](n: T) -> T` calling itself → terminates (cached before recursing)

## Validation Checklist

- [ ] Same instantiation called N times → exactly one monomorphized copy
- [ ] Cloned AST does not share nodes with original (deep copy)
- [ ] Type substitution replaces all occurrences of type params
- [ ] Cloned function passes type checking
- [ ] Name mangling produces valid C identifier

## Acceptance Criteria

- `fn sort[T](list: [T]) -> [T]` instantiated as `sort_i32` with all T→i32
- Generic struct `Stack[i32]` gets distinct TypeID from `Stack[string]`
- Compliance tests 041-050 (generics group) pass

## Definition of Done

- [ ] `compiler/sema/mono.go` implemented
- [ ] AST cloning implemented
- [ ] Unit tests pass
- [ ] Integrated with type checker

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| AST clone bloat (many instantiations) | Use cache aggressively; compiler warning if > 1000 instantiations |
| Infinite loop in recursive generic (fib[T] calling fib[T]) | Cache before recursing; detect same-type recursion early |

## Future Follow-up Tasks

- p09-t06: AIR builder handles monomorphized functions the same as regular functions
- p10-t04: inlining works on monomorphized functions
