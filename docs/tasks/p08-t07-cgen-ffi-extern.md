# p08-t07: C-Backend FFI and Extern Declaration Generation

## Purpose
Implement C interop code generation for `extern "C"` declarations in `codegen/cgen/ffi.go`. AXIOM code can call C functions directly using `extern "C"` blocks; the C-Backend must emit correct C function prototypes, handle variadic functions, and respect struct layout attributes (`@packed`, `@align(N)`).

## Context
AXIOM is designed to interoperate with C libraries, the OS API, and other native code. The `extern "C"` syntax allows AXIOM programs to declare and call C functions without AXIOM name mangling. The C-Backend simply emits the C prototype as declared, with no wrapper or thunk.

Struct layout attributes are critical for correct FFI with C libraries that use packed or aligned structs (e.g., network packet headers, SIMD-aligned buffers).

## Inputs
- `extern "C"` declaration nodes in the typed AST
- Struct attribute annotations (`@packed`, `@align(N)`)
- `codegen/cgen/types.go` (p08-t01) for parameter type mapping

## Outputs
- `codegen/cgen/ffi.go` — FFI code generation
- `codegen/cgen/ffi_test.go` — unit tests

## Dependencies
- p08-t01 (type mapping — maps AXIOM types to C types for FFI prototypes)
- p08-t02 (declaration emitter — FFI prototypes are added to the declaration section)

## Subsystems Affected
- C-Backend (FFI declarations affect the declaration section of generated code)
- Linker (extern symbols are resolved by the platform linker, not axc)

## Detailed Requirements

### `extern "C"` Function Declarations
AXIOM syntax:
```
extern "C" fn printf(fmt: string, ...) -> i32
extern "C" fn malloc(size: u64) -> *u8
extern "C" fn free(ptr: *u8) -> void
```

Generated C output (placed in declaration section before function bodies):
```c
// These are standard C declarations; no ax_ mangling
int printf(const char* fmt, ...);
void* malloc(unsigned long long size);
void free(void* ptr);
```

**Type Translation for FFI**
FFI parameters use different type mapping than AXIOM-to-AXIOM code:
- `string` → `const char*` (not `ax_string` — C APIs expect null-terminated strings)
- `*u8` → `void*` (generic pointer in C interop)
- `u64` → `unsigned long long` or the platform's `size_t` equivalent
- `i32` → `int`
- Variadic `...` → `...` (C varargs)

The FFI type translator is separate from the main `CTypeName` function; it uses raw C types rather than `ax_` aliases to avoid ABI mismatches.

### FFI Type Mapping Table
```
i8   → signed char
i16  → short
i32  → int
i64  → long long
u8   → unsigned char
u16  → unsigned short
u32  → unsigned int
u64  → unsigned long long
f32  → float
f64  → double
bool → int  (C99 _Bool promoted to int at FFI boundary)
string → const char*  (null-terminated; AXIOM will add \0 when crossing FFI boundary)
*T   → void*  (or <C type>* if T is a primitive)
void → void
```

### `@packed` Attribute
AXIOM:
```
@packed
struct PacketHeader:
    magic: u32
    length: u16
    flags: u8
    _pad: u8
```

Generated C:
```c
struct __attribute__((packed)) ax_PacketHeader {
    ax_u32 magic;
    ax_u16 length;
    ax_u8  flags;
    ax_u8  _pad;
};
```

On MSVC (Windows), use `#pragma pack(1)` / `#pragma pack()` instead:
```c
#ifdef _MSC_VER
#pragma pack(push, 1)
struct ax_PacketHeader { ... };
#pragma pack(pop)
#else
struct __attribute__((packed)) ax_PacketHeader { ... };
#endif
```

### `@align(N)` Attribute
AXIOM:
```
@align(32)
struct SimdVec:
    data: [f32; 8]
```

Generated C:
```c
struct __attribute__((aligned(32))) ax_SimdVec {
    ax_f32 data[8];
};
```

On MSVC: `__declspec(align(32))`.

### Combined `@packed` + `@align(N)`
If both are present, emit both attributes:
```c
struct __attribute__((packed, aligned(N))) ax_Foo { ... };
```

### Variadic Functions
The `...` AXIOM syntax for variadic extern functions maps directly to C `...`:
```c
int printf(const char* fmt, ...);
```

In the AST, variadic functions have `IsVariadic = true` on the `ExternFuncDecl` node. The last parameter before `...` must be a named parameter.

### Calling Convention Attributes (Future)
Reserve the `@callconv("win64")`, `@callconv("sysv")` attributes for future use. In MVP, emit no calling convention attribute (use platform default).

### `extern "C"` Block Grouping
Multiple `extern "C"` declarations can be grouped:
```
extern "C":
    fn sin(x: f64) -> f64
    fn cos(x: f64) -> f64
    fn tan(x: f64) -> f64
```

All generate individual C prototypes; no `extern "C" { }` C++ block is needed since the output is C, not C++.

### API
```go
// FFIDecl generates a C declaration for an extern "C" function.
func FFIDecl(fn *ast.ExternFuncDecl) string

// FFITypeName returns the C type name for a type used in FFI context.
// Uses raw C types (int, char*, etc.) rather than ax_ aliases.
func FFITypeName(typeID uint32, table *typecheck.TypeTable) string

// StructAttributeAnnotation returns the GCC/Clang attribute string for the struct.
func StructAttributeAnnotation(attrs []ast.Attribute) string
```

## Implementation Steps

### Step 1: Implement `FFITypeName`
```go
func FFITypeName(id uint32, table *typecheck.TypeTable) string {
    ty := table.Get(id)
    switch ty.Kind {
    case typecheck.TyI8:     return "signed char"
    case typecheck.TyI16:    return "short"
    case typecheck.TyI32:    return "int"
    case typecheck.TyI64:    return "long long"
    case typecheck.TyU8:     return "unsigned char"
    case typecheck.TyU16:    return "unsigned short"
    case typecheck.TyU32:    return "unsigned int"
    case typecheck.TyU64:    return "unsigned long long"
    case typecheck.TyF32:    return "float"
    case typecheck.TyF64:    return "double"
    case typecheck.TyBool:   return "int"
    case typecheck.TyString: return "const char*"
    case typecheck.TyVoid:   return "void"
    case typecheck.TyPointer:
        inner := FFITypeName(ty.Args[0], table)
        if inner == "void" { return "void*" }
        return inner + "*"
    default:
        // Struct types in FFI: use the ax_ struct name (shared layout)
        return CTypeName(id, table, nil)
    }
}
```

### Step 2: Implement `FFIDecl`
```go
func FFIDecl(fn *ast.ExternFuncDecl, table *typecheck.TypeTable) string {
    ret := FFITypeName(fn.RetTypeID, table)
    params := make([]string, 0, len(fn.Params)+1)
    for _, p := range fn.Params {
        params = append(params, FFITypeName(p.TypeID, table))
    }
    if fn.IsVariadic {
        params = append(params, "...")
    }
    if len(params) == 0 {
        params = []string{"void"}
    }
    return fmt.Sprintf("%s %s(%s);", ret, fn.Name, strings.Join(params, ", "))
}
```

### Step 3: Implement `StructAttributeAnnotation`
```go
func StructAttributeAnnotation(attrs []ast.Attribute) string {
    var gcc, msvc strings.Builder
    for _, a := range attrs {
        switch a.Name {
        case "packed":
            gcc.WriteString("packed, ")
        case "align":
            n := a.Args[0].(*ast.IntLit).Value
            gcc.WriteString(fmt.Sprintf("aligned(%d), ", n))
            msvc.WriteString(fmt.Sprintf("__declspec(align(%d)) ", n))
        }
    }
    gccStr := strings.TrimSuffix(gcc.String(), ", ")
    if gccStr == "" { return "" }
    return fmt.Sprintf("#ifdef _MSC_VER\n%s\n#else\n__attribute__((%s))\n#endif",
        msvc.String(), gccStr)
}
```

### Step 4: Integrate into `DeclEmitter`
In `DeclEmitter.ProcessModule`, after processing struct declarations, iterate over `ExternFuncDecl` nodes and call `FFIDecl` for each.

### Step 5: Write `ffi_test.go`
Test: extern function prototype, variadic function, `@packed` struct, `@align(32)` struct, combined attributes.

## Test Plan
1. `extern "C" fn malloc(size: u64) -> *u8` → `void* malloc(unsigned long long size);`
2. `extern "C" fn printf(fmt: string, ...) -> i32` → `int printf(const char* fmt, ...);`
3. `extern "C" fn free(ptr: *u8) -> void` → `void free(void*);`
4. `@packed struct` → correct `__attribute__((packed))` wrapper with MSVC guard
5. `@align(32) struct` → `__attribute__((aligned(32)))`
6. `@packed @align(16) struct` → `__attribute__((packed, aligned(16)))`
7. FFI function with no params: emits `void` in param list
8. FFI function with AXIOM struct type: uses `ax_` type name

## Validation Checklist
- [ ] FFI prototypes use raw C types, not `ax_` aliases
- [ ] Variadic `...` is the last element in the parameter list
- [ ] `@packed` and `@align` emit correct GCC and MSVC variants
- [ ] Zero-parameter FFI function emits `(void)` not `()`
- [ ] All tests pass

## Acceptance Criteria
- Generated FFI declarations compile in a C file that also includes `ax_runtime.h`
- `@packed` struct can be round-tripped through FFI with correct byte layout
- All tests pass

## Definition of Done
- `codegen/cgen/ffi.go` exists with `FFIDecl`, `FFITypeName`, `StructAttributeAnnotation`
- `codegen/cgen/ffi_test.go` exists and passes
- `go test ./codegen/cgen/` passes

## Risks & Mitigations
- **Risk**: `string` FFI type mapping to `const char*` requires that AXIOM strings passed to C are null-terminated. **Mitigation**: Document that `extern "C"` string parameters receive the `.ptr` field of `ax_string` which must be null-terminated; AXIOM string literals are null-terminated in C output.
- **Risk**: MSVC `#pragma pack` / `__declspec(align)` interaction with GCC `__attribute__`. **Mitigation**: Use `#ifdef _MSC_VER` guards consistently; test on both compilers in CI.

## Future Follow-up Tasks
- p12-t04: Dynamic linking requires correct extern prototypes for imported symbols
- Future: `@callconv` attribute for non-default calling conventions
- Future: Automatic null-termination wrapper for string FFI parameters
