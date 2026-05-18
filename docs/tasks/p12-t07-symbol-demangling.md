# p12-t07: Symbol Demangling (Debug)

## Purpose
Implement a symbol demangling utility that converts AXIOM's mangled symbol names (e.g., `_AX_module_Box_i32_compute`) back to human-readable form (e.g., `module.Box[i32].compute`). Used by the panic handler's stack trace, debugger integration, and `axc` diagnostic tools.

## Context
Plan §Phase 3 defines the mangle scheme: `_AX_<module>_<name>_<type1>_<type2>`. The demangler is the inverse operation, needed for readable stack traces and debug info.

## Inputs
- Mangling scheme from p12-t01 (symbol-mangling)
- Panic handler from p07-t03 (needs demangled names for stack traces)

## Outputs
- `linker/demangle.go` — `Demangle(mangled string) string`
- `linker/demangle_test.go`

## Dependencies
- p12-t01: symbol-mangling — mangling rules defined

## Subsystems Affected
- Panic handler: stack traces show demangled names
- DWARF info (p11-t13): debug symbol names demangled
- `axc` diagnostic messages: function names in errors

## Detailed Requirements

```go
// Demangle converts _AX_module_name_type1_type2 to module.name[type1, type2].
func Demangle(mangled string) string

// Examples:
// "_AX_main_compute"              → "main.compute"
// "_AX_collections_Box_i32"       → "collections.Box[i32]"
// "_AX_math_max_f64_f64"          → "math.max[f64, f64]"
// "_AX_main_main"                 → "main.main"
// "printf"                        → "printf" (not mangled, return as-is)
```

Rules:
1. If string doesn't start with `_AX_`, return as-is (extern symbol).
2. Split by `_` after prefix.
3. First segment = module, second = function/type name.
4. Remaining segments = type parameters (wrapped in `[]`).

## Implementation Steps

1. Create `linker/demangle.go`.
2. Implement prefix detection and splitting.
3. Handle edge cases: names containing underscores (use double-underscore `__` as literal underscore escape).
4. Write tests for all mangle patterns.

## Test Plan

- `TestDemangleSimple`: `_AX_main_foo` → `main.foo`
- `TestDemangleGeneric`: `_AX_std_Box_i32` → `std.Box[i32]`
- `TestDemangleMultiTypeArgs`: `_AX_std_Map_string_i32` → `std.Map[string, i32]`
- `TestDemangleExtern`: `printf` → `printf`
- `TestDemangleUnderscore`: `_AX_my__module_my__func` → `my_module.my_func`

## Acceptance Criteria

- Demangled names appear in panic handler stack traces
- Round-trip: `Demangle(Mangle(name))` produces readable output

## Definition of Done

- [ ] `linker/demangle.go` implemented
- [ ] Tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Ambiguous underscore in names | Use double-underscore escape convention; document in p12-t01 |

## Future Follow-up Tasks

- p07-t03: Panic handler uses `Demangle()` for stack traces
- p17-t07: Profiler uses `Demangle()` for flame graphs
