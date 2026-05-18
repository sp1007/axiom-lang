# p05-t04: Structural Duck Typing (Interfaces)

## Purpose
Implement structural interface satisfaction — a struct automatically satisfies an interface if it has all required methods with matching signatures, without requiring an explicit `implements` declaration. This is AXIOM's duck typing system, analogous to Go interfaces.

## Context
AXIOM interfaces are purely structural: `interface Printable { fn print(self) }` is satisfied by any struct that has a `print(self)` method with the same signature. No `implements Printable` needed. This is checked at the point of use (when a struct is passed where an interface is expected), not at declaration. The check is performed via structural comparison of method signatures.

## Inputs
- TypeTable with struct TypeInfos (fields and methods)
- TypeTable with interface TypeInfos (required methods)
- Type checker call sites where interface types are expected

## Outputs
- `compiler/sema/interfaces.go` — `ImplementsInterface()` check
- Updated symbol table: `Symbol.ImplementedInterfaces []uint32` (TypeIDs of satisfied interfaces)
- `[]Diagnostic` — "does not implement interface X: missing method Y"

## Dependencies
- p05-t01: generic-type-representation — interfaces are types with constraints
- p04-t07: type-checker-expressions — call sites where interface satisfaction is checked
- p04-t02: type-table-primitives — TKInterface kind

## Subsystems Affected
- Type checker: interface satisfaction checked at assignment and call sites
- Generic constraints: `fn max[T: Ord]` checks T satisfies Ord interface
- Overload resolution: interface match adds 1 point to overload score

## Detailed Requirements

1. `MethodSig` struct: `{NameID uint32, ParamTypes []uint32, ReturnType uint32}`
2. `TypeInfo` for interfaces: `Kind=TKInterface`, `Methods []MethodSig` (required methods).
3. `ImplementsInterface(structTypeID, ifaceTypeID uint32) (bool, []MethodSig)` — returns true/false + missing methods.
4. Algorithm: for each required method in the interface, check if the struct has a method with:
   - Matching NameID
   - Same number of parameters
   - Same parameter TypeIDs (structural, not nominal)
   - Same return TypeID
5. Method lookup: scan the struct's associated methods in the SymbolTable (functions with `self` first param of that struct type).
6. At call site: `fn accept(p: Printable)` called with `MyStruct{}` → check `ImplementsInterface(MyStruct, Printable)`.
7. For generic constraints: `fn max[T: Ord](a: T, b: T)` → when instantiated with `i32`, check `ImplementsInterface(i32, Ord)`. Built-in types (i32, string, etc.) implement standard interfaces (Ord, Hash, Eq) by default.
8. Error message: `"type Foo does not implement interface Bar: missing method baz(i32) -> string"`.

## Implementation Steps

1. Create `compiler/sema/interfaces.go`.
2. Implement `ImplementsInterface(structTypeID, ifaceTypeID uint32)`.
3. Implement method signature extraction: `getMethodsOfStruct(structTypeID uint32) []MethodSig`.
4. Pre-register built-in interface implementations: `i32` implements `Ord`, `Eq`, `Hash`; `string` implements `Ord`, `Eq`, `Hash`, `Display`.
5. Hook into type checker: when assigning struct to interface variable or passing to interface parameter, call `ImplementsInterface`.
6. Hook into overload resolution: interface satisfaction scores 1 (lowest priority).
7. Write tests: `TestImplicitImplementation`, `TestMissingMethod`, `TestBuiltinInterfaces`.

## Test Plan

- `TestImplicitImplementation`: struct with `print(self)` passed to `fn show(p: Printable)` → OK
- `TestMissingMethod`: struct without `print` passed to `Printable` parameter → error with specific method
- `TestWrongSignature`: method exists but wrong return type → "does not implement: print signature mismatch"
- `TestBuiltinOrd`: `fn max[T: Ord](a: T, b: T)` called with `i32` → OK (i32 pre-implements Ord)
- `TestGenericConstraintFail`: `Stack[MyStruct]` where MyStruct doesn't implement required interface → error
- `TestCompliance041`: compliance tests 041-050 (interface group) pass

## Validation Checklist

- [ ] Structural check works without explicit `implements` declaration
- [ ] Missing methods reported by name in error
- [ ] Built-in types (i32, string) implement standard interfaces
- [ ] Interface satisfaction cached (not recomputed per call site)
- [ ] Wrong method signature (not just missing) caught

## Acceptance Criteria

- Compliance tests 041-050 pass
- Duck typing works for all method signatures (no nominal registration required)

## Definition of Done

- [ ] `compiler/sema/interfaces.go` implemented
- [ ] Built-in interface implementations pre-registered
- [ ] `go test ./compiler/sema/ -run TestInterface` passes

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Interface check O(N²) for many methods | Cache results per (structType, ifaceType) pair |
| Self-referential interfaces (interface requires method returning same interface) | Supported; use TypeID references, no infinite loop |

## Future Follow-up Tasks

- p15-t05: actor spawn uses Isolated[T] which is a structural interface check
- p16-t11: std.concurrency defines Actor, Channel interfaces
