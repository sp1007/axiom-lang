# p04-t02: Type Table & Primitive Types

## Purpose
Implement the `TypeTable` — the central registry of all types in the compilation unit. Every expression, variable, and function in AXIOM has a `TypeID` indexing into this table. Primitive types occupy indices 1–16 (frozen). `TypeID(0)` = Unknown/Unresolved sentinel.

## Context
AXIOM is statically typed with local type inference. The type system includes: primitives (`i8`–`i64`, `u8`–`u64`, `f32`, `f64`, `bool`, `string`, `char8`, `void`, `isize`, `usize`), structs, functions, generics (p05), sum types (p05), tuples, arrays, and built-in parameterized types. This task implements table infrastructure and primitive types only.

**Spec references:** `04. Type checker.md`, `docs/plan.md` Section 2 Phase 2

## Inputs
- `compiler/sema/symbols.go` — `Symbol.TypeID` field (p04-t01)
- `compiler/diagnostics/diagnostics.go` — error reporting (p01-t01)

## Outputs
- `compiler/types/typeid.go` — TypeID type, primitive constants, category methods
- `compiler/types/typeentry.go` — TypeEntry, TypeKind, FieldEntry, StructType, FuncType
- `compiler/types/typetable.go` — TypeTable with registration, lookup, query methods
- `compiler/types/typetable_test.go` — ≥20 unit tests

## Dependencies
- p04-t01: symbol-table — Symbol.TypeID references this table
- p03-t02: string-intern-pool — type names use interned strings

## Subsystems Affected
- Type checker (p04-t05–t07), monomorphization (p05-t02), C-backend (p08-t01), AIR builder (p09-t06), native backend (p11-t07)

## Detailed Requirements

### TypeID Constants (FROZEN)
```go
type TypeID uint32
const (
    TypeUnknown TypeID = 0   // unresolved sentinel
    TypeI8      TypeID = 1
    TypeI16     TypeID = 2
    TypeI32     TypeID = 3
    TypeI64     TypeID = 4
    TypeU8      TypeID = 5
    TypeU16     TypeID = 6
    TypeU32     TypeID = 7
    TypeU64     TypeID = 8
    TypeF32     TypeID = 9
    TypeF64     TypeID = 10
    TypeBool    TypeID = 11
    TypeString  TypeID = 12
    TypeChar8   TypeID = 13
    TypeVoid    TypeID = 14
    TypeISize   TypeID = 15
    TypeUSize   TypeID = 16
    PrimitiveCount = 16
)
```

Category methods: `IsUnknown()`, `IsPrimitive()`, `IsInteger()`, `IsSigned()`, `IsUnsigned()`, `IsFloat()`, `IsNumeric()`, `IsBool()`, `IsVoid()`, `IsString()`, `SizeOf() uint32`.

### TypeKind Enum
```go
type TypeKind uint8
const (
    KindPrimitive TypeKind = iota
    KindStruct      // struct { fields... }
    KindFunction    // fn(params) -> return
    KindArray       // [T; N] fixed-size
    KindSlice       // Seq[T] dynamic
    KindTuple       // (T1, T2, ...)
    KindSum         // type X = A | B
    KindGeneric     // unresolved [T]
    KindGenericInst // concrete Box[i32]
    KindPointer     // *T raw pointer
    KindRef         // &T / lent reference
    KindOption      // Option[T]
    KindResult      // Result[T, E]
)
```

### TypeEntry, FieldEntry, StructType, FuncType
```go
type TypeEntry struct {
    Kind TypeKind; NameID uint32; Size uint32; Align uint32; Flags uint16; Extra uint32
}
type FieldEntry struct {
    NameID uint32; TypeID TypeID; Offset uint32; Flags uint8
}
type StructType struct {
    Fields []FieldEntry; GenericParams []uint32
}
type FuncType struct {
    Params []TypeID; Return TypeID; Effects []uint32; IsVariadic bool; IsAsync bool
}
```

### TypeTable API
```go
type TypeTable struct {
    entries []TypeEntry; structs []StructType; funcs []FuncType
}
func NewTypeTable() *TypeTable           // pre-populate primitives
func (tt *TypeTable) RegisterStruct(nameID uint32, fields []FieldEntry, generics []uint32) TypeID
func (tt *TypeTable) RegisterFunction(params []TypeID, ret TypeID, effects []uint32) TypeID
func (tt *TypeTable) Entry(id TypeID) *TypeEntry
func (tt *TypeTable) StructInfo(id TypeID) *StructType
func (tt *TypeTable) FuncInfo(id TypeID) *FuncType
func (tt *TypeTable) Count() int
func (tt *TypeTable) IsAssignableTo(from, to TypeID) bool
func (tt *TypeTable) CanImplicitCast(from, to TypeID) bool
func (tt *TypeTable) CommonType(a, b TypeID) (TypeID, bool)
func (tt *TypeTable) FindByName(nameID uint32) (TypeID, bool)
```

### Implicit Casting Rules
- Widening allowed: `i8→i16→i32→i64`, `u8→u16→u32→u64`, `f32→f64`
- NOT allowed: signed↔unsigned, float→int, larger→smaller, string→numeric, bool→numeric

### CommonType Resolution
- Same type → that type
- Both integer same signedness → wider type
- Both float → wider float
- Integer + float → float (with widening)
- Otherwise → type error

## Implementation Steps

1. Create `compiler/types/typeid.go` — TypeID, primitive constants, category methods, SizeOf.
2. Create `compiler/types/typeentry.go` — TypeKind, TypeEntry, FieldEntry, StructType, FuncType.
3. Create `compiler/types/typetable.go`:
   - `NewTypeTable()`: allocate entries cap 256, populate entries[0]=Unknown, entries[1..16]=primitives with correct Size/Align.
   - `RegisterStruct()`: append to entries+structs, return new TypeID.
   - `RegisterFunction()`: append to entries+funcs, return new TypeID.
   - `IsAssignableTo()`: exact match or CanImplicitCast.
   - `CanImplicitCast()`: implement widening rules.
   - `CommonType()`: binary expression type resolution.
4. Create `compiler/types/typetable_test.go`.

## Test Plan

1. `TestNewTypeTable_PrimitiveCount`: Count == 17 (Unknown + 16 primitives)
2. `TestPrimitiveIDs_Frozen`: TypeI32==3, TypeBool==11, etc.
3. `TestPrimitiveSizes`: i8=1, i32=4, f64=8, string=8, void=0
4. `TestPrimitiveCategories`: i32.IsInteger, f64.IsFloat, bool.IsBool
5. `TestTypeUnknown`: IsUnknown==true, IsPrimitive==false
6. `TestRegisterStruct`: returns TypeID ≥ 17
7. `TestRegisterFunction`: returns TypeID, FuncInfo has correct params/return
8. `TestStructInfo_Fields`: fields match registered data
9. `TestFuncInfo_Params`: params and return match
10. `TestIsAssignableTo_SameType`: i32→i32 = true
11. `TestIsAssignableTo_DifferentType`: i32→string = false
12. `TestCanImplicitCast_Widening`: i8→i32 = true
13. `TestCanImplicitCast_Narrowing`: i32→i8 = false
14. `TestCanImplicitCast_FloatWidening`: f32→f64 = true
15. `TestCanImplicitCast_SignedUnsigned`: i32→u32 = false
16. `TestCommonType_SameType`: (i32,i32) → i32
17. `TestCommonType_Widening`: (i8,i32) → i32
18. `TestCommonType_IntFloat`: (i32,f64) → f64
19. `TestCommonType_Incompatible`: (bool,i32) → error
20. `TestFindByName`: register "Foo", find → correct TypeID
21. `TestDeterminism`: same registrations → identical TypeIDs

## Validation Checklist
- [ ] TypeID(0) is Unknown, never assigned to valid types
- [ ] Primitives occupy indices 1–16 exactly
- [ ] User types start at index 17
- [ ] All primitive sizes correct
- [ ] Widening rules match spec
- [ ] `CanImplicitCast(signed, unsigned) == false`
- [ ] `StructInfo()` panics for non-struct TypeID
- [ ] `go test ./compiler/types/` passes

## Acceptance Criteria
- All 21 tests pass
- Primitive TypeID constants are frozen and documented
- Entry lookup is O(1) array index
- Implicit cast rules enforced correctly

## Definition of Done
- [ ] All 4 source files implemented
- [ ] 21 tests passing
- [ ] `go vet ./compiler/types/` zero warnings
- [ ] No circular imports with sema package

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| TypeID constants must match SymbolTable built-in IDs | Define in types/typeid.go; sema imports types (one-way) |
| Struct field offset alignment calculation complex | Defer to ComputeLayout(); store offset=0 for now |
| FindByName is O(N) | Acceptable for name resolution; optimize later if needed |

## Future Follow-up Tasks
- p04-t04: name resolver uses FindByName
- p04-t05: type inference uses CommonType/CanImplicitCast
- p05-t01: generic type representation extends TypeTable
- p08-t01: C-backend maps TypeIDs to C strings
