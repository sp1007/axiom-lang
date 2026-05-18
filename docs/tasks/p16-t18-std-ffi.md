# p16-t18: std.ffi — Foreign Function Interface

## Purpose
Implement the AXIOM foreign function interface (FFI) library providing safe wrappers around C pointer types, memory layout utilities, and calling conventions for integrating with C libraries.

## Context
AXIOM programs often need to call C libraries (OpenSSL, SQLite, etc.). `std.ffi` provides `CString` (null-terminated string), `CPtr[T]` (raw C pointer with bounds), layout query functions (`size_of`, `align_of`), and utilities for converting between AXIOM and C types safely.

## Inputs
- AXIOM `extern` keyword from p04 (semantic layer)
- Runtime allocator from p14 for CString allocation
- Type system TypeInfo for `size_of`/`align_of`

## Outputs
- `stdlib/ffi/ffi.ax` — CString, CPtr, CArray, layout utilities
- `stdlib/ffi/marshal.ax` — AXIOM↔C type conversion

## Dependencies
- p04-t02: type-table — TypeInfo for size/align queries
- p14-t01: axalloc — CString heap allocation
- p08-t01: cgen-type-mapping — C type sizes (already computed)

## Detailed Requirements

```axiom
# stdlib/ffi/ffi.ax

# Null-terminated C string
type CString:
    var ptr: *u8    # null-terminated
    var len: u32    # byte length (not including \0)

    fn from_str(s: str) -> CString   # copies + null-terminates
    fn to_str(self) -> str           # wraps (no copy)
    fn as_ptr(self) -> *u8           # raw pointer for extern calls
    fn free(mut self)                # explicit free (CTGC auto-frees)

# Raw C pointer (unsafe)
type CPtr[T]:
    var ptr: *T
    var valid: bool  # simple validity flag (not generational ref)

    fn null[T]() -> CPtr[T]
    fn from_raw(p: *T) -> CPtr[T]  # unsafe
    fn as_ref(self) -> Option[*T]  # None if null
    fn offset(self, n: i64) -> CPtr[T]

# C array (pointer + length)
type CArray[T]:
    var ptr: *T
    var len: u32

    fn as_slice(self) -> Slice[T]
    fn get(self, idx: u32) -> Option[T]

# Layout queries (compile-time for known types, runtime via TypeInfo)
fn size_of[T]() -> u32       # sizeof(T)
fn align_of[T]() -> u32      # alignof(T)
fn offset_of[T](field: str) -> u32  # offsetof(T, field) — compile-time only

# Type conversion utilities
# stdlib/ffi/marshal.ax
fn i32_to_c(v: i32) -> i32         # identity
fn str_to_cstring(s: str) -> CString
fn cstring_to_str(cs: CString) -> str
fn bool_to_c(b: bool) -> i32       # true=1, false=0
fn c_to_bool(n: i32) -> bool       # n != 0
```

Safety model:
- `CPtr[T]` is `unsafe` to dereference — requires `unsafe {}` block.
- `CString` is auto-freed by CTGC unless `.as_ptr()` passes it to extern fn.
- `size_of`, `align_of` produce compile-time constants when T is known.

Example FFI usage:
```axiom
extern fn strlen(s: *u8) -> u64

fn count_chars(s: str) -> u64:
    let cs = CString.from_str(s)
    strlen(cs.as_ptr())
```

## Implementation Steps

1. Create `stdlib/ffi/ffi.ax` — CString, CPtr, CArray.
2. Implement `CString.from_str()` — alloc + memcpy + null terminate.
3. Implement `size_of[T]()` — emit `sizeof(T)` in C backend or computed value.
4. Implement `offset_of[T](field)` — compile-time struct field offset.
5. Create `stdlib/ffi/marshal.ax` — type converters.
6. Write tests calling real C functions (strlen, memcpy).
7. Document unsafe usage patterns.

## Test Plan
- `TestCStringRoundtrip`: str → CString → to_str() == original
- `TestCStrlenCall`: CString of "hello" → strlen returns 5
- `TestSizeOf`: size_of[i32]() = 4
- `TestAlignOf`: align_of[f64]() = 8
- `TestCPtrNull`: CPtr.null().as_ref() = None

## Validation Checklist
- [ ] CString always null-terminated (fuzz with random strings)
- [ ] CPtr.as_ref() returns None for null pointer
- [ ] size_of values match C sizeof() for all primitive types
- [ ] Unsafe blocks required for CPtr dereference

## Acceptance Criteria
- AXIOM program calling `printf` via CString and extern fn works correctly

## Definition of Done
- [ ] `stdlib/ffi/ffi.ax` implemented
- [ ] All tests pass including real C function calls

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| CString freed while extern fn still using ptr | Document: keep CString alive for duration of extern call |
| size_of wrong for struct types with padding | Compute from TypeInfo struct layout, not sum of field sizes |

## Future Follow-up Tasks
- Variadic C function support (AL register for XMM args in SysV)
- C struct binding generator (parse C headers → AXIOM extern types)
