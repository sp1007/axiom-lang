package cgen_test

import (
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

// helper creates a TypeTable, InternPool, and TypeDeclQueue for tests.
func helper() (*types.TypeTable, *ast.InternPool, *cgen.TypeDeclQueue) {
	return types.NewTypeTable(), ast.NewInternPool(64), cgen.NewTypeDeclQueue()
}

// --- Primitive Type Mappings ---

func TestPrimitiveTypes(t *testing.T) {
	table, intern, queue := helper()

	cases := []struct {
		id   types.TypeID
		want string
	}{
		{types.TypeI8, "ax_i8"},
		{types.TypeI16, "ax_i16"},
		{types.TypeI32, "ax_i32"},
		{types.TypeI64, "ax_i64"},
		{types.TypeU8, "ax_u8"},
		{types.TypeU16, "ax_u16"},
		{types.TypeU32, "ax_u32"},
		{types.TypeU64, "ax_u64"},
		{types.TypeF32, "ax_f32"},
		{types.TypeF64, "ax_f64"},
		{types.TypeBool, "ax_bool"},
		{types.TypeString, "ax_string"},
		{types.TypeChar8, "ax_char"},
		{types.TypeVoid, "void"},
		{types.TypeISize, "ax_isize"},
		{types.TypeUSize, "ax_usize"},
	}

	for _, tc := range cases {
		got := cgen.CTypeName(tc.id, table, intern, queue)
		if got != tc.want {
			t.Errorf("CTypeName(%v) = %q, want %q", tc.id, got, tc.want)
		}
	}
}

func TestPrimitivesDoNotEnqueue(t *testing.T) {
	table, intern, queue := helper()

	// Primitive types should not enqueue any declarations
	cgen.CTypeName(types.TypeI32, table, intern, queue)
	cgen.CTypeName(types.TypeBool, table, intern, queue)
	cgen.CTypeName(types.TypeString, table, intern, queue)

	if queue.Len() != 0 {
		t.Errorf("primitives should not enqueue declarations, got %d", queue.Len())
	}
}

// --- Pointer Types ---

func TestPointerToPrimitive(t *testing.T) {
	table, intern, queue := helper()
	ptrI32 := table.RegisterPointer(types.TypeI32)
	got := cgen.CTypeName(ptrI32, table, intern, queue)
	if got != "ax_i32*" {
		t.Errorf("pointer to i32 = %q, want \"ax_i32*\"", got)
	}
}

func TestPointerToStruct(t *testing.T) {
	table, intern, queue := helper()
	fooName := intern.InternString("Foo")
	fooID := table.RegisterStruct(fooName, nil, nil)
	ptrFoo := table.RegisterPointer(fooID)
	got := cgen.CTypeName(ptrFoo, table, intern, queue)
	if got != "struct ax_Foo*" {
		t.Errorf("pointer to Foo = %q, want \"struct ax_Foo*\"", got)
	}
}

func TestDoublePointer(t *testing.T) {
	table, intern, queue := helper()
	ptrI32 := table.RegisterPointer(types.TypeI32)
	pptrI32 := table.RegisterPointer(ptrI32)
	got := cgen.CTypeName(pptrI32, table, intern, queue)
	if got != "ax_i32**" {
		t.Errorf("double pointer to i32 = %q, want \"ax_i32**\"", got)
	}
}

// --- Slice Types ---

func TestSliceType(t *testing.T) {
	table, intern, queue := helper()
	sliceF32 := table.RegisterSlice(types.TypeF32)
	got := cgen.CTypeName(sliceF32, table, intern, queue)
	if got != "ax_slice_ax_f32" {
		t.Errorf("slice of f32 = %q, want \"ax_slice_ax_f32\"", got)
	}
	drained := queue.Drain()
	if len(drained) == 0 {
		t.Error("slice type should be enqueued for declaration")
	}
}

func TestSliceOfStruct(t *testing.T) {
	table, intern, queue := helper()
	barName := intern.InternString("Bar")
	barID := table.RegisterStruct(barName, nil, nil)
	sliceBar := table.RegisterSlice(barID)
	got := cgen.CTypeName(sliceBar, table, intern, queue)
	if got != "ax_slice_struct_ax_Bar" {
		t.Errorf("slice of Bar = %q, want \"ax_slice_struct_ax_Bar\"", got)
	}
}

// --- Struct Types ---

func TestStructType(t *testing.T) {
	table, intern, queue := helper()
	pointName := intern.InternString("Point")
	xName := intern.InternString("x")
	yName := intern.InternString("y")
	fields := []types.FieldEntry{
		{NameID: xName, TypeID: types.TypeI32},
		{NameID: yName, TypeID: types.TypeF64},
	}
	id := table.RegisterStruct(pointName, fields, nil)
	got := cgen.CTypeName(id, table, intern, queue)
	if got != "struct ax_Point" {
		t.Errorf("struct Point = %q, want \"struct ax_Point\"", got)
	}

	drained := queue.Drain()
	if len(drained) == 0 {
		t.Error("struct type should be enqueued")
	}
}

// --- Sum Types (tagged unions) ---

func TestSumType(t *testing.T) {
	table, intern, queue := helper()
	resultName := intern.InternString("Result")
	okName := intern.InternString("Ok")
	errName := intern.InternString("Err")

	variants := []types.VariantInfo{
		{NameID: okName, PayloadType: types.TypeI32, Tag: 0},
		{NameID: errName, PayloadType: types.TypeString, Tag: 1},
	}
	id := table.RegisterSumType(resultName, variants, nil)
	got := cgen.CTypeName(id, table, intern, queue)
	if got != "struct ax_Result" {
		t.Errorf("sum type Result = %q, want \"struct ax_Result\"", got)
	}
}

// --- Generic Instantiations ---

func TestGenericInst(t *testing.T) {
	table, intern, queue := helper()
	stackName := intern.InternString("Stack")
	id := table.RegisterGenericInst(stackName, []types.TypeID{types.TypeI32})
	got := cgen.CTypeName(id, table, intern, queue)
	if got != "struct ax_Stack_ax_i32" {
		t.Errorf("Stack[i32] = %q, want \"struct ax_Stack_ax_i32\"", got)
	}
}

func TestGenericInstMultipleArgs(t *testing.T) {
	table, intern, queue := helper()
	mapName := intern.InternString("Map")
	id := table.RegisterGenericInst(mapName, []types.TypeID{types.TypeString, types.TypeI64})
	got := cgen.CTypeName(id, table, intern, queue)
	if got != "struct ax_Map_ax_string_ax_i64" {
		t.Errorf("Map[string, i64] = %q, want \"struct ax_Map_ax_string_ax_i64\"", got)
	}
}

func TestNestedGenericInst(t *testing.T) {
	table, intern, queue := helper()
	stackName := intern.InternString("Stack")
	innerStack := table.RegisterGenericInst(stackName, []types.TypeID{types.TypeI32})
	outerStack := table.RegisterGenericInst(stackName, []types.TypeID{innerStack})
	got := cgen.CTypeName(outerStack, table, intern, queue)
	if got != "struct ax_Stack_struct_ax_Stack_ax_i32" {
		t.Errorf("Stack[Stack[i32]] = %q, want \"struct ax_Stack_struct_ax_Stack_ax_i32\"", got)
	}
}

// --- Function Types ---

func TestFuncType(t *testing.T) {
	table, intern, queue := helper()
	params := []types.TypeID{types.TypeI32, types.TypeString}
	id := table.RegisterFunction(params, types.TypeBool, nil)
	got := cgen.CTypeName(id, table, intern, queue)
	if got != "ax_bool (*)(ax_i32, ax_string)" {
		t.Errorf("fn(i32, string) -> bool = %q", got)
	}
}

func TestFuncTypeNoParams(t *testing.T) {
	table, intern, queue := helper()
	id := table.RegisterFunction(nil, types.TypeVoid, nil)
	got := cgen.CTypeName(id, table, intern, queue)
	if got != "void (*)(void)" {
		t.Errorf("fn() -> void = %q, want \"void (*)(void)\"", got)
	}
}

// --- TypeDeclQueue ---

func TestTypeDeclQueue_Dedup(t *testing.T) {
	queue := cgen.NewTypeDeclQueue()
	queue.Enqueue(42)
	queue.Enqueue(42)
	queue.Enqueue(42)
	drained := queue.Drain()
	if len(drained) != 1 {
		t.Errorf("expected 1 entry after dedup, got %d", len(drained))
	}
}

func TestTypeDeclQueue_Order(t *testing.T) {
	queue := cgen.NewTypeDeclQueue()
	queue.Enqueue(10)
	queue.Enqueue(20)
	queue.Enqueue(30)
	drained := queue.Drain()
	if len(drained) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(drained))
	}
	if drained[0] != 10 || drained[1] != 20 || drained[2] != 30 {
		t.Errorf("order incorrect: %v", drained)
	}
}

func TestTypeDeclQueue_DrainResets(t *testing.T) {
	queue := cgen.NewTypeDeclQueue()
	queue.Enqueue(1)
	queue.Drain()
	drained := queue.Drain()
	if len(drained) != 0 {
		t.Errorf("drain should return empty after second call, got %d", len(drained))
	}
}

// --- Determinism ---

func TestDeterminism(t *testing.T) {
	// CTypeName must return the same string for the same TypeID across calls
	table, intern, queue := helper()
	stackName := intern.InternString("Stack")
	id := table.RegisterGenericInst(stackName, []types.TypeID{types.TypeI32})

	r1 := cgen.CTypeName(id, table, intern, queue)
	r2 := cgen.CTypeName(id, table, intern, queue)
	if r1 != r2 {
		t.Errorf("non-deterministic: first=%q, second=%q", r1, r2)
	}
}

func TestDeterminism_AcrossInstances(t *testing.T) {
	// Two independent TypeTable+InternPool with same registrations should produce same names
	t1, i1, q1 := helper()
	t2, i2, q2 := helper()

	name1 := i1.InternString("Vec")
	name2 := i2.InternString("Vec")

	id1 := t1.RegisterGenericInst(name1, []types.TypeID{types.TypeI32})
	id2 := t2.RegisterGenericInst(name2, []types.TypeID{types.TypeI32})

	r1 := cgen.CTypeName(id1, t1, i1, q1)
	r2 := cgen.CTypeName(id2, t2, i2, q2)
	if r1 != r2 {
		t.Errorf("cross-instance non-determinism: %q vs %q", r1, r2)
	}
}

// --- CTypeDecl ---

func TestCTypeDecl_Struct(t *testing.T) {
	table, intern, queue := helper()
	xName := intern.InternString("x")
	yName := intern.InternString("y")
	pointName := intern.InternString("Point")

	fields := []types.FieldEntry{
		{NameID: xName, TypeID: types.TypeI32},
		{NameID: yName, TypeID: types.TypeF64},
	}
	id := table.RegisterStruct(pointName, fields, nil)

	decl := cgen.CTypeDecl(id, table, intern, queue)
	if !strings.Contains(decl, "struct ax_Point") {
		t.Errorf("struct decl missing name: %s", decl)
	}
	if !strings.Contains(decl, "ax_i32 x;") {
		t.Errorf("struct decl missing field x: %s", decl)
	}
	if !strings.Contains(decl, "ax_f64 y;") {
		t.Errorf("struct decl missing field y: %s", decl)
	}
}

func TestCTypeDecl_Slice(t *testing.T) {
	table, intern, queue := helper()
	id := table.RegisterSlice(types.TypeI32)

	decl := cgen.CTypeDecl(id, table, intern, queue)
	if !strings.Contains(decl, "ax_i32* ptr;") {
		t.Errorf("slice decl missing ptr field: %s", decl)
	}
	if !strings.Contains(decl, "ax_u64 len;") {
		t.Errorf("slice decl missing len field: %s", decl)
	}
	if !strings.Contains(decl, "ax_u64 cap;") {
		t.Errorf("slice decl missing cap field: %s", decl)
	}
}

func TestCTypeDecl_SumType(t *testing.T) {
	table, intern, queue := helper()
	resultName := intern.InternString("Result")
	okName := intern.InternString("Ok")
	errName := intern.InternString("Err")

	variants := []types.VariantInfo{
		{NameID: okName, PayloadType: types.TypeI32, Tag: 0},
		{NameID: errName, PayloadType: types.TypeString, Tag: 1},
	}
	id := table.RegisterSumType(resultName, variants, nil)

	decl := cgen.CTypeDecl(id, table, intern, queue)
	if !strings.Contains(decl, "enum ax_Result_tag") {
		t.Errorf("sum type decl missing tag enum: %s", decl)
	}
	if !strings.Contains(decl, "ax_Result_Ok = 0") {
		t.Errorf("sum type decl missing Ok variant: %s", decl)
	}
	if !strings.Contains(decl, "ax_Result_Err = 1") {
		t.Errorf("sum type decl missing Err variant: %s", decl)
	}
	if !strings.Contains(decl, "ax_i32 Ok;") {
		t.Errorf("sum type decl missing Ok payload: %s", decl)
	}
	if !strings.Contains(decl, "ax_string Err;") {
		t.Errorf("sum type decl missing Err payload: %s", decl)
	}
}

func TestCTypeDecl_Primitive_Empty(t *testing.T) {
	table, intern, queue := helper()
	decl := cgen.CTypeDecl(types.TypeI32, table, intern, queue)
	if decl != "" {
		t.Errorf("primitive should have empty declaration, got %q", decl)
	}
}

// --- sanitizeName ---

func TestSanitizeName(t *testing.T) {
	table, intern, queue := helper()
	// Test with pointer type to verify sanitization
	ptrI32 := table.RegisterPointer(types.TypeI32)
	slicePtr := table.RegisterSlice(ptrI32)

	got := cgen.CTypeName(slicePtr, table, intern, queue)
	// "ax_i32*" gets sanitized to "ax_i32_"
	if got != "ax_slice_ax_i32_" {
		t.Errorf("slice of pointer = %q, want \"ax_slice_ax_i32_\"", got)
	}
}

// --- Edge Cases ---

func TestUnknownType(t *testing.T) {
	table, intern, queue := helper()
	got := cgen.CTypeName(types.TypeUnknown, table, intern, queue)
	if got != "void" {
		t.Errorf("unknown type = %q, want \"void\"", got)
	}
}

func TestCollisionFreeGenericVsStruct(t *testing.T) {
	// Verify that struct "Stack_i32" and generic Stack[i32] produce different C names
	table, intern, queue := helper()

	// A struct literally named "Stack_i32"
	litName := intern.InternString("Stack_i32")
	litID := table.RegisterStruct(litName, nil, nil)
	litC := cgen.CTypeName(litID, table, intern, queue)

	// A generic Stack[i32]
	stackName := intern.InternString("Stack")
	genID := table.RegisterGenericInst(stackName, []types.TypeID{types.TypeI32})
	genC := cgen.CTypeName(genID, table, intern, queue)

	if litC == genC {
		t.Errorf("name collision: struct Stack_i32 = %q, generic Stack[i32] = %q", litC, genC)
	}
}
