# p08-t02: C-Backend Declaration Generation

## Purpose
Implement the declaration generation phase of the C-Backend in `codegen/cgen/decls.go`. This phase processes all top-level declarations from the typed AST to produce C forward declarations, struct definitions, enum definitions, and function prototypes — all emitted before any function bodies, so that generated C files compile without order-dependency issues.

## Context
C requires that types and function signatures be declared before their first use. AXIOM does not have this restriction — functions can call each other in any order. To bridge this gap, the C-Backend performs two passes over the module: first, emit all declarations (this task); second, emit function bodies. This two-pass approach means the C-Backend never needs to compute call graph ordering.

The output of this task is the "header section" of the generated `.c` file, placed immediately after `#include "ax_runtime.h"`.

## Inputs
- `codegen/cgen/types.go` (p08-t01) — `CTypeName` and `TypeDeclQueue`
- Typed AST module (from p04-t08) — all top-level declarations
- `TypeTable` — for resolving type IDs to type structures

## Outputs
- `codegen/cgen/decls.go` — declaration emitter
- `codegen/cgen/decls_test.go` — unit tests

## Dependencies
- p08-t01 (type mapping)
- p04-t08 (typed AST representation)

## Subsystems Affected
- C-Backend (declarations must precede all function bodies)
- Linker (function prototypes determine symbol visibility)

## Detailed Requirements

### Output Ordering
The generated C file's declaration section must follow this order:
1. `#include "ax_runtime.h"`
2. Type alias declarations (`typedef`)
3. Forward declarations for all structs (`struct ax_Foo;`)
4. Struct definitions (in topological dependency order — fields' types must be defined first)
5. Sum type enum definitions (`enum ax_Result_tag { ... };`)
6. Sum type struct definitions (`struct ax_Result_... { ... };`)
7. Slice struct definitions (`struct ax_slice_... { ... };`)
8. Global variable declarations (`extern ax_i32 ax_module_global_name;`)
9. Function prototypes (all exported and internal functions)

### Struct Declaration Generation
For each struct type in the module:
```c
// Forward declaration (emitted first, allows pointer types to work)
struct ax_Foo;

// Full definition
struct ax_Foo {
    ax_i32   x;
    ax_string name;
    struct ax_Bar* next;  // pointer to another struct (forward decl suffices)
};
```

Topological sort algorithm:
- Build a dependency graph where struct A depends on struct B if A contains B by value (not by pointer)
- Pointer fields only require a forward declaration, not a full definition
- Emit in topological order; if a cycle is detected (impossible in AXIOM's value semantics but guarded against), emit a diagnostic

### Sum Type Declaration Generation
For a sum type `Result[T, E]` with variants `Ok(T)` and `Err(E)`:
```c
// Tag enum
enum ax_Result_T_E_tag {
    ax_Result_Ok  = 0,
    ax_Result_Err = 1,
};

// Sum struct
struct ax_Result_T_E {
    enum ax_Result_T_E_tag tag;
    union {
        <C type for T> ok;
        <C type for E> err;
    } data;
};
```

Where `T` and `E` in the names are mangled using the same rules as generic instantiation in p08-t01.

### Function Prototype Generation
For each function declared in the module:
```c
// Regular function
ax_i32 ax_module_fibonacci(ax_i32 n);

// Variadic
ax_i32 ax_module_printf(ax_string fmt, ...);

// Method (AXIOM methods are lowered to free functions with receiver as first param)
void ax_module_Stack_push(struct ax_Stack_ax_i32* self, ax_i32 value);
```

The name mangling uses the scheme from p12-t01 (prefixed with module name).

Visibility:
- Public AXIOM functions: no `static` (visible to linker)
- Private AXIOM functions (module-private): `static` prefix

### Global Variable Declarations
AXIOM global variables:
- Immutable globals: `const ax_i32 ax_module_MAX = 100;` (initialized at declaration)
- Mutable globals: `ax_i32 ax_module_counter;` (initialized in `ax_module_init()`)
- `extern` declarations for globals defined in other modules

### `DeclEmitter` API
```go
// DeclEmitter accumulates C declarations and can write them to an io.Writer.
type DeclEmitter struct {
    table   *typecheck.TypeTable
    queue   *TypeDeclQueue
    structs []string   // struct forward decls and definitions
    enums   []string   // enum definitions
    protos  []string   // function prototypes
    globals []string   // global variable declarations
}

func NewDeclEmitter(table *typecheck.TypeTable) *DeclEmitter

// ProcessModule processes all top-level decls in the module.
func (e *DeclEmitter) ProcessModule(mod *ast.TypedModule)

// EmitTo writes the full declaration section to w.
func (e *DeclEmitter) EmitTo(w io.Writer)
```

## Implementation Steps

### Step 1: Implement topological sort for struct types
```go
func topoSortStructs(structs []*typecheck.StructType, table *typecheck.TypeTable) []*typecheck.StructType {
    // Build adjacency list: struct A → structs it depends on by value
    // Use DFS with cycle detection (cycle = panic; impossible in valid AXIOM)
    visited := make(map[uint32]bool)
    var sorted []*typecheck.StructType
    var visit func(s *typecheck.StructType)
    visit = func(s *typecheck.StructType) {
        if visited[s.ID] { return }
        visited[s.ID] = true
        for _, field := range s.Fields {
            ft := table.Get(field.TypeID)
            if ft.Kind == typecheck.TyStruct && !isPointerField(field) {
                visit(table.GetStruct(ft.ID))
            }
        }
        sorted = append(sorted, s)
    }
    for _, s := range structs { visit(s) }
    return sorted
}
```

### Step 2: Implement struct declaration emitter
```go
func (e *DeclEmitter) emitStruct(s *typecheck.StructType) {
    // Forward declaration
    e.structs = append(e.structs, fmt.Sprintf("struct ax_%s;", s.Name))

    // Full definition
    var fields strings.Builder
    for _, f := range s.Fields {
        ctype := CTypeName(f.TypeID, e.table, e.queue)
        fields.WriteString(fmt.Sprintf("    %s %s;\n", ctype, f.Name))
    }
    e.structs = append(e.structs, fmt.Sprintf(
        "struct ax_%s {\n%s};", s.Name, fields.String()))
}
```

### Step 3: Implement function prototype emitter
```go
func (e *DeclEmitter) emitFuncProto(fn *ast.TypedFuncDecl) {
    ret := CTypeName(fn.RetType, e.table, e.queue)
    params := make([]string, len(fn.Params))
    for i, p := range fn.Params {
        params[i] = CTypeName(p.TypeID, e.table, e.queue) + " " + p.Name
    }
    visibility := ""
    if fn.IsPrivate { visibility = "static " }
    proto := fmt.Sprintf("%s%s %s(%s);",
        visibility, ret, mangleFuncName(fn), strings.Join(params, ", "))
    e.protos = append(e.protos, proto)
}
```

### Step 4: Implement `EmitTo`
```go
func (e *DeclEmitter) EmitTo(w io.Writer) {
    fmt.Fprintln(w, `#include "ax_runtime.h"`)
    fmt.Fprintln(w)
    for _, s := range e.structs  { fmt.Fprintln(w, s) }
    for _, en := range e.enums   { fmt.Fprintln(w, en) }
    for _, g := range e.globals  { fmt.Fprintln(w, g) }
    for _, p := range e.protos   { fmt.Fprintln(w, p) }
    fmt.Fprintln(w)
}
```

### Step 5: Write `decls_test.go`
Test with a synthetic module containing:
- A struct `Point{x:i32, y:i32}`
- A struct `Line{start:Point, end:Point}` (depends on Point by value)
- A function `fn distance(a: Point, b: Point) -> f64`
- Verify `Line` is declared after `Point` in the output
- Verify the function prototype contains correct types

## Test Plan
1. Simple struct → correct `struct ax_Foo { ... };` emitted
2. Dependency order: `Line` contains `Point` by value → `Point` emitted before `Line`
3. Pointer fields: `*Point` in a struct requires only forward decl of `Point`
4. Function prototype: return type and parameter types are correct C names
5. Private function: `static` prefix emitted
6. Variadic function: `...` in prototype
7. Sum type: enum + struct emitted with correct tag values
8. Global variable: `const ax_i32 ax_module_X = 5;` emitted
9. `#include "ax_runtime.h"` is always the first line
10. Struct forward declarations appear before full definitions

## Validation Checklist
- [ ] Generated C compiles without warnings (`gcc -c -Wall -Wextra generated.c`)
- [ ] Topological sort handles diamond dependencies correctly
- [ ] Private functions get `static` prefix
- [ ] Public functions do NOT get `static`
- [ ] Sum types emit correct tag enum values (0, 1, 2, ...)
- [ ] All tests pass

## Acceptance Criteria
- A generated C file with declarations compiles cleanly with GCC
- Struct dependency order is always topologically correct
- Function prototypes exactly match the function definitions emitted by p08-t03

## Definition of Done
- `codegen/cgen/decls.go` exists with `DeclEmitter` implementation
- `codegen/cgen/decls_test.go` exists and passes
- `go test ./codegen/cgen/` passes

## Risks & Mitigations
- **Risk**: Circular struct definitions (A contains B, B contains A by value). **Mitigation**: AXIOM type checker prevents this; add an assertion in the topo sort to panic with an internal compiler error if a cycle is detected.
- **Risk**: Name collision between module-level functions and method-lowered functions. **Mitigation**: Mangled names include both module name and receiver type for methods.

## Future Follow-up Tasks
- p08-t03: Statement codegen uses the function body structure set up by declarations
- p08-t09: Build pipeline calls `DeclEmitter.ProcessModule` as the first codegen step
- p12-t01: Symbol mangling (provides `mangleFuncName` used here)
