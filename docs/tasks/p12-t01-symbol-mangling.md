# p12-t01: Symbol Name Mangling

## Purpose
Implement a deterministic symbol mangling scheme for AXIOM functions and types, ensuring unique, demangleable names in object files and enabling linker symbol resolution across modules.

## Context
AXIOM functions need unique linker-visible names. A function `add` in module `math` with type `(i32, i32) -> i32` conflicts with `add` in module `string`. Mangling encodes the module path, function name, and type signature into a unique string. The mangling scheme must also handle generic monomorphizations (e.g., `List[i32]` vs `List[f64]`).

## Inputs
- Function symbol info: module path, function name, parameter types, return type
- Generic instantiation info from p05-t02 (monomorphization)
- Exported vs internal symbols (visibility)

## Outputs
- `codegen/mangle.go` — symbol mangling and demangling
- Mangled symbol strings embedded in object file symbol tables

## Dependencies
- p04-t02: type-table — TypeInfo for encoding type signatures
- p05-t02: monomorphization — generic instantiation names

## Subsystems Affected
- Object file emitters (p11-t12, p12-t02, p12-t03): use mangled names in symbol tables
- Linker (p12-t04): matches symbols by mangled name

## Detailed Requirements

Mangling scheme: `_AX_<module>_<name>_<typesig>`

Examples:
```
math::add(i32, i32) -> i32     → _AX_math_add_ii_i
string::format(str, i32) -> str → _AX_string_format_si_s
List[i32]::push(i32)           → _AX_List_push_Ti32_v
main::main()                   → _AX_main_main_v
```

Type encoding table:
```
i8=b, i16=s, i32=i, i64=l, u8=B, u16=S, u32=I, u64=L
f32=f, f64=d, bool=o, str=t, void=v, ptr=p
Array[T]=AT, Slice[T]=ST, Option[T]=OT
Generic T=T<name>
```

```go
func Mangle(module, name string, params []uint32, ret uint32) string
func Demangle(mangled string) (module, name string, params []uint32, ret uint32, err error)
func MangleGeneric(module, name string, typeArgs []uint32, params []uint32, ret uint32) string
```

Special cases:
- `main()` → `_AX_main_main_v` but also exported as `main` (entry point)
- `extern` functions: not mangled (use as-is)
- Methods: `_AX_<module>_<TypeName>_<method>_<sig>`

## Implementation Steps

1. Create `codegen/mangle.go`.
2. Define type encoding table (TypeID → char).
3. Implement `Mangle()` — encode module + name + types into `_AX_...` string.
4. Implement `Demangle()` — parse mangled string back to components.
5. Handle generic monomorphizations with type argument encoding.
6. Handle extern (unmangled) and entry-point (dual export) cases.
7. Write round-trip tests: mangle → demangle → assert equal.

## Test Plan
- `TestMangleBasic`: `math::add(i32,i32)->i32` → `_AX_math_add_ii_i`
- `TestMangleGeneric`: `List[i32]::push(i32)` → correct mangled form
- `TestDemangleRoundtrip`: mangle → demangle → equal
- `TestMangleExtern`: extern function → unmangled name
- `TestMangleMain`: main → exported as `main` symbol

## Validation Checklist
- [ ] All type primitives have unique single-char encoding
- [ ] Demangling is deterministic inverse of mangling
- [ ] Generics include type argument in name
- [ ] Module separator is `_` (no special chars for linker compatibility)

## Acceptance Criteria
- `axc demangle _AX_math_add_ii_i` prints `math::add(i32, i32) -> i32`

## Definition of Done
- [ ] `codegen/mangle.go` implemented
- [ ] Round-trip tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Name collision between different type encodings | Enforce uniqueness in type encoding table, test all primitive pairs |
| Mangled names exceed platform symbol length limits | Cap at 255 chars; hash remainder for very long generics |

## Future Follow-up Tasks
- p12-t04: dynamic linking uses mangled names for .so exports
- p17: LSP demangler for hover display
