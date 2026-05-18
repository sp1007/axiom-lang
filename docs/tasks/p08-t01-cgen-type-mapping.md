# p08-t01: C-Backend Type Mapping

## Purpose
Implement the AXIOM-to-C type mapping layer in `codegen/cgen/types.go`. This module translates every AXIOM type (primitives, composites, generics, sum types, slices, pointers) into its corresponding C11 representation using the `ax_` type aliases defined in `ax_runtime.h`. This is the foundation that all other C-Backend code generation tasks depend on.

## Context
The C-Backend generates C11 source code from the typed AST. Every generated variable declaration, function parameter, return type, and struct field must carry a valid C type name. The type mapping must be deterministic (same AXIOM type always produces the same C name), complete (covers every type that can appear in valid AXIOM code), and collision-free (two distinct AXIOM types never produce the same C name, including generic instantiations).

## Inputs
- `codegen/cgen/` directory (create if not present)
- `compiler/typecheck/types.go` — TypeTable and type ID definitions (from p04-t02)
- `runtime/ax_runtime.h` — the C type aliases that output names must match
- AXIOM Language Specification: type system section

## Outputs
- `codegen/cgen/types.go` — type mapping implementation
- `codegen/cgen/types_test.go` — unit tests for all type mappings

## Dependencies
- p04-t02 (TypeTable with type IDs for all AXIOM types)
- p07-t04 (ax_runtime.h defines the C names used here)

## Subsystems Affected
- C-Backend (all code generation passes consume this module)
- Symbol mangling (p12-t01 uses type names in mangled identifiers)

## Detailed Requirements

### Primitive Type Mappings
```
i8   → "ax_i8"
i16  → "ax_i16"
i32  → "ax_i32"
i64  → "ax_i64"
u8   → "ax_u8"
u16  → "ax_u16"
u32  → "ax_u32"
u64  → "ax_u64"
f32  → "ax_f32"
f64  → "ax_f64"
bool → "ax_bool"
string → "ax_string"
void → "void"
never → "void"  // never-returning functions; C return type is void
```

### Pointer Types
```
*T (raw pointer) → "ax_<T>*"
*i32             → "ax_i32*"
*Foo             → "struct ax_Foo*"
```

### Slice Types
```
[T]   (slice of T) → struct with ptr/len/cap fields, named "ax_slice_<T_mangled>"
[i32] → "ax_slice_i32"
[Foo] → "ax_slice_Foo"  // struct ax_slice_Foo { struct ax_Foo* ptr; ax_u64 len; ax_u64 cap; };
```

Slice struct definition must be emitted in the declaration section before any use.

### Struct Types
```
struct Foo → "struct ax_Foo"
```

The struct definition is emitted as:
```c
struct ax_Foo {
    ax_i32 x;
    ax_string name;
};
```

### Sum Types (tagged unions)
```
Result[T, E] (sum type with variants Ok(T) and Err(E)) →
    enum ax_Result_tag { ax_Result_Ok = 0, ax_Result_Err = 1 };
    struct ax_Result_i32_string {
        enum ax_Result_tag tag;
        union {
            ax_i32   ok;
            ax_string err;
        } data;
    };
```

The C name for a monomorphized sum type includes its type arguments to ensure uniqueness.

### Generic Instantiations
Monomorphize by appending type arguments to the name:
```
Stack[i32]     → "struct ax_Stack_i32"
Map[string,i64] → "struct ax_Map_string_i64"
Result[i32,string] → "struct ax_Result_i32_string"
```

Nested generics:
```
Stack[Stack[i32]] → "struct ax_Stack_Stack_i32"
```

### Function Types (for function pointers)
```
fn(i32, string) -> bool →
    "ax_bool (*)(ax_i32, ax_string)"
```

### AxRef (heap references in AXIOM)
AXIOM heap-allocated values are represented as `AxRef` in generated C:
```
heap T → "AxRef"  // all heap refs use the same AxRef type; cast on deref
```

### API
```go
// CTypeName returns the C type name string for the given type ID.
// For compound types (structs, slices), it may add entries to the TypeDeclQueue
// so that forward declarations are emitted before first use.
func CTypeName(typeID uint32, table *TypeTable, queue *TypeDeclQueue) string

// CTypeDecl returns the full C struct/enum declaration for a named type.
// Returns "" for primitive types that need no declaration.
func CTypeDecl(typeID uint32, table *TypeTable, queue *TypeDeclQueue) string

// TypeDeclQueue tracks which types need declarations emitted, in dependency order.
type TypeDeclQueue struct {
    seen    map[uint32]bool
    ordered []uint32
}
func (q *TypeDeclQueue) Enqueue(typeID uint32)
func (q *TypeDeclQueue) Drain() []uint32
```

### Determinism Requirement
Given the same `TypeTable`, `CTypeName` must always return the identical string for the same `typeID`. This is required for reproducible builds.

## Implementation Steps

### Step 1: Create `codegen/cgen/` directory with `types.go`
```go
package cgen

import (
    "fmt"
    "strings"
    "axiom/compiler/typecheck"
)

// CTypeName returns the C11 type name for a given AXIOM type ID.
func CTypeName(id uint32, table *typecheck.TypeTable, queue *TypeDeclQueue) string {
    ty := table.Get(id)
    switch ty.Kind {
    case typecheck.TyI8:     return "ax_i8"
    case typecheck.TyI16:    return "ax_i16"
    case typecheck.TyI32:    return "ax_i32"
    case typecheck.TyI64:    return "ax_i64"
    case typecheck.TyU8:     return "ax_u8"
    case typecheck.TyU16:    return "ax_u16"
    case typecheck.TyU32:    return "ax_u32"
    case typecheck.TyU64:    return "ax_u64"
    case typecheck.TyF32:    return "ax_f32"
    case typecheck.TyF64:    return "ax_f64"
    case typecheck.TyBool:   return "ax_bool"
    case typecheck.TyString: return "ax_string"
    case typecheck.TyVoid:   return "void"
    case typecheck.TyNever:  return "void"

    case typecheck.TyPointer:
        inner := CTypeName(ty.Args[0], table, queue)
        return inner + "*"

    case typecheck.TySlice:
        inner := CTypeName(ty.Args[0], table, queue)
        name := "ax_slice_" + sanitizeName(inner)
        queue.Enqueue(id) // needs struct decl
        return "struct " + name

    case typecheck.TyStruct:
        queue.Enqueue(id)
        return "struct ax_" + ty.Name

    case typecheck.TySumType:
        args := make([]string, len(ty.Args))
        for i, arg := range ty.Args {
            args[i] = sanitizeName(CTypeName(arg, table, queue))
        }
        name := "ax_" + ty.Name + "_" + strings.Join(args, "_")
        queue.Enqueue(id)
        return "struct " + name

    case typecheck.TyGenericInst:
        args := make([]string, len(ty.Args))
        for i, arg := range ty.Args {
            args[i] = sanitizeName(CTypeName(arg, table, queue))
        }
        name := "ax_" + ty.Name + "_" + strings.Join(args, "_")
        queue.Enqueue(id)
        return "struct " + name

    case typecheck.TyHeapRef:
        return "AxRef"

    case typecheck.TyFuncPtr:
        ret := CTypeName(ty.Ret, table, queue)
        params := make([]string, len(ty.Params))
        for i, p := range ty.Params {
            params[i] = CTypeName(p, table, queue)
        }
        return fmt.Sprintf("%s (*)(%s)", ret, strings.Join(params, ", "))

    default:
        panic(fmt.Sprintf("CTypeName: unknown type kind %d", ty.Kind))
    }
}

// sanitizeName replaces characters invalid in C identifiers with underscores.
func sanitizeName(name string) string {
    var b strings.Builder
    for _, r := range name {
        if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
           (r >= '0' && r <= '9') || r == '_' {
            b.WriteRune(r)
        } else {
            b.WriteRune('_')
        }
    }
    return b.String()
}
```

### Step 2: Implement `TypeDeclQueue`
```go
type TypeDeclQueue struct {
    seen    map[uint32]bool
    ordered []uint32
}

func NewTypeDeclQueue() *TypeDeclQueue {
    return &TypeDeclQueue{seen: make(map[uint32]bool)}
}

func (q *TypeDeclQueue) Enqueue(id uint32) {
    if !q.seen[id] {
        q.seen[id] = true
        q.ordered = append(q.ordered, id)
    }
}

func (q *TypeDeclQueue) Drain() []uint32 {
    out := q.ordered
    q.ordered = nil
    return out
}
```

### Step 3: Write `codegen/cgen/types_test.go`
```go
package cgen_test

import (
    "testing"
    "axiom/codegen/cgen"
    "axiom/compiler/typecheck"
)

func TestPrimitiveTypes(t *testing.T) {
    table := typecheck.NewTypeTable()
    queue := cgen.NewTypeDeclQueue()

    cases := []struct{ id uint32; want string }{
        {table.BuiltinID(typecheck.TyI32), "ax_i32"},
        {table.BuiltinID(typecheck.TyF64), "ax_f64"},
        {table.BuiltinID(typecheck.TyBool), "ax_bool"},
        {table.BuiltinID(typecheck.TyString), "ax_string"},
        {table.BuiltinID(typecheck.TyVoid), "void"},
    }
    for _, tc := range cases {
        got := cgen.CTypeName(tc.id, table, queue)
        if got != tc.want {
            t.Errorf("CTypeName(%d) = %q, want %q", tc.id, got, tc.want)
        }
    }
}

func TestPointerType(t *testing.T) {
    table := typecheck.NewTypeTable()
    queue := cgen.NewTypeDeclQueue()
    ptrI32 := table.MakePointer(table.BuiltinID(typecheck.TyI32))
    got := cgen.CTypeName(ptrI32, table, queue)
    if got != "ax_i32*" {
        t.Errorf("got %q, want \"ax_i32*\"", got)
    }
}

func TestSliceType(t *testing.T) {
    table := typecheck.NewTypeTable()
    queue := cgen.NewTypeDeclQueue()
    sliceF32 := table.MakeSlice(table.BuiltinID(typecheck.TyF32))
    got := cgen.CTypeName(sliceF32, table, queue)
    if got != "struct ax_slice_ax_f32" {
        t.Errorf("got %q", got)
    }
    if len(queue.Drain()) == 0 {
        t.Error("slice type should be enqueued for declaration")
    }
}

func TestGenericInstantiation(t *testing.T) {
    table := typecheck.NewTypeTable()
    queue := cgen.NewTypeDeclQueue()
    stackI32 := table.MakeGenericInst("Stack", []uint32{table.BuiltinID(typecheck.TyI32)})
    got := cgen.CTypeName(stackI32, table, queue)
    if got != "struct ax_Stack_ax_i32" {
        t.Errorf("got %q", got)
    }
}
```

## Test Plan
1. All 12 primitive types map to correct `ax_` names
2. Pointer-to-primitive: `*i32` → `"ax_i32*"`
3. Pointer-to-struct: `*Foo` → `"struct ax_Foo*"`
4. Slice: `[i32]` → `"struct ax_slice_ax_i32"`, enqueued in TypeDeclQueue
5. Struct: `Foo` → `"struct ax_Foo"`, enqueued
6. Generic instantiation: `Stack[i32]` → `"struct ax_Stack_ax_i32"`
7. Nested generic: `Stack[Stack[i32]]` → `"struct ax_Stack_struct_ax_Stack_ax_i32"`
8. Sum type with two variants
9. Function pointer type
10. `never` type maps to `"void"`
11. Determinism: calling `CTypeName` twice for the same ID returns the same string
12. TypeDeclQueue deduplication: enqueuing the same type ID twice results in one entry

## Validation Checklist
- [ ] All 12 primitive mappings correct
- [ ] Pointer composition works for arbitrary depth
- [ ] Generic name mangling is collision-free (no two distinct AXIOM types produce identical C names)
- [ ] `TypeDeclQueue` deduplicates correctly
- [ ] `sanitizeName` handles all characters that can appear in C type names
- [ ] All unit tests pass (`go test ./codegen/cgen/`)
- [ ] No panics on valid type IDs

## Acceptance Criteria
- `CTypeName` covers all type kinds in the TypeTable
- Output names match the definitions in `ax_runtime.h`
- No two distinct AXIOM types produce the same C name
- All tests pass

## Definition of Done
- `codegen/cgen/types.go` exists with complete implementation
- `codegen/cgen/types_test.go` exists with tests for all type classes
- `go test ./codegen/cgen/` passes

## Risks & Mitigations
- **Risk**: Name collision between a struct named `Stack_i32` and a generic `Stack[i32]`. **Mitigation**: User-defined names cannot contain underscores followed by type names (enforced by the parser); the mangling scheme uses `ax_` prefix which user code cannot start with.
- **Risk**: Recursive types (e.g., `struct Node { next: *Node }`) cause infinite recursion in `CTypeName`. **Mitigation**: The `TypeDeclQueue` tracks seen IDs; `CTypeName` for a struct just returns the name string without recursing into fields (the full struct declaration is a separate step in p08-t02).

## Future Follow-up Tasks
- p08-t02: Uses `CTypeName` to emit struct declarations
- p08-t03: Uses `CTypeName` for variable declarations
- p12-t01: Symbol mangling uses `CTypeName` output as part of the mangled identifier
