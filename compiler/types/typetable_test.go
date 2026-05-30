package types_test

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/types"
)

func TestNewTypeTable_PrimitiveCount(t *testing.T) {
	tt := types.NewTypeTable()
	if tt.Count() != 22 {
		t.Errorf("expected 22 built-in types (unknown + 16 primitives + 4 builtin interfaces + ActorRef), got %d", tt.Count())
	}
}

func TestPrimitiveIDs_Frozen(t *testing.T) {
	if types.TypeUnknown != 0 { t.Error("TypeUnknown should be 0") }
	if types.TypeI32 != 3 { t.Error("TypeI32 should be 3") }
	if types.TypeBool != 11 { t.Error("TypeBool should be 11") }
	if types.TypeVoid != 14 { t.Error("TypeVoid should be 14") }
	if types.PrimitiveCount != 16 { t.Error("PrimitiveCount should be 16") }
}

func TestPrimitiveSizes(t *testing.T) {
	if types.TypeI8.SizeOf() != 1 { t.Error("i8 size != 1") }
	if types.TypeI32.SizeOf() != 4 { t.Error("i32 size != 4") }
	if types.TypeF64.SizeOf() != 8 { t.Error("f64 size != 8") }
	if types.TypeString.SizeOf() != 16 { t.Error("string size != 16") }
	if types.TypeVoid.SizeOf() != 0 { t.Error("void size != 0") }
}

func TestPrimitiveCategories(t *testing.T) {
	if !types.TypeI32.IsInteger() { t.Error("i32 should be integer") }
	if !types.TypeI32.IsSigned() { t.Error("i32 should be signed") }
	if types.TypeI32.IsUnsigned() { t.Error("i32 should not be unsigned") }
	
	if !types.TypeU32.IsInteger() { t.Error("u32 should be integer") }
	if !types.TypeU32.IsUnsigned() { t.Error("u32 should be unsigned") }
	if types.TypeU32.IsSigned() { t.Error("u32 should not be signed") }

	if !types.TypeF64.IsFloat() { t.Error("f64 should be float") }
	if types.TypeF64.IsInteger() { t.Error("f64 should not be integer") }

	if !types.TypeI32.IsNumeric() { t.Error("i32 should be numeric") }
	if !types.TypeF64.IsNumeric() { t.Error("f64 should be numeric") }

	if !types.TypeBool.IsBool() { t.Error("bool should be bool") }
	if types.TypeBool.IsNumeric() { t.Error("bool should not be numeric") }
}

func TestTypeUnknown(t *testing.T) {
	if !types.TypeUnknown.IsUnknown() { t.Error("TypeUnknown.IsUnknown() failed") }
	if types.TypeUnknown.IsPrimitive() { t.Error("TypeUnknown should not be primitive") }
}

func TestRegisterStruct(t *testing.T) {
	tt := types.NewTypeTable()
	fields := []types.FieldEntry{
		{NameID: 100, TypeID: types.TypeI32, Offset: 0, Flags: 0},
	}
	id := tt.RegisterStruct(200, fields, nil)
	
	if id < 17 { t.Errorf("expected struct TypeID >= 17, got %d", id) }
	
	entry := tt.Entry(id)
	if entry.Kind != types.KindStruct { t.Errorf("expected KindStruct") }
	if entry.NameID != 200 { t.Errorf("expected NameID 200") }
}

func TestRegisterFunction(t *testing.T) {
	tt := types.NewTypeTable()
	params := []types.TypeID{types.TypeI32, types.TypeBool}
	id := tt.RegisterFunction(params, types.TypeString, nil)
	
	entry := tt.Entry(id)
	if entry.Kind != types.KindFunction { t.Errorf("expected KindFunction") }
}

func TestStructInfo_Fields(t *testing.T) {
	tt := types.NewTypeTable()
	fields := []types.FieldEntry{
		{NameID: 100, TypeID: types.TypeI32, Offset: 0, Flags: 0},
	}
	id := tt.RegisterStruct(200, fields, nil)
	
	info := tt.StructInfo(id)
	if len(info.Fields) != 1 { t.Fatalf("expected 1 field") }
	if info.Fields[0].TypeID != types.TypeI32 { t.Errorf("expected field type i32") }
}

func TestFuncInfo_Params(t *testing.T) {
	tt := types.NewTypeTable()
	params := []types.TypeID{types.TypeI32, types.TypeBool}
	id := tt.RegisterFunction(params, types.TypeString, nil)
	
	info := tt.FuncInfo(id)
	if len(info.Params) != 2 { t.Fatalf("expected 2 params") }
	if info.Return != types.TypeString { t.Errorf("expected string return") }
}

func TestIsAssignableTo_SameType(t *testing.T) {
	tt := types.NewTypeTable()
	if !tt.IsAssignableTo(types.TypeI32, types.TypeI32) {
		t.Error("i32 should be assignable to i32")
	}
}

func TestIsAssignableTo_DifferentType(t *testing.T) {
	tt := types.NewTypeTable()
	if tt.IsAssignableTo(types.TypeI32, types.TypeString) {
		t.Error("i32 should not be assignable to string")
	}
}

func TestCanImplicitCast_Widening(t *testing.T) {
	tt := types.NewTypeTable()
	if !tt.CanImplicitCast(types.TypeI8, types.TypeI32) {
		t.Error("i8 should implicitly cast to i32")
	}
	if !tt.CanImplicitCast(types.TypeU16, types.TypeU64) {
		t.Error("u16 should implicitly cast to u64")
	}
}

func TestCanImplicitCast_Narrowing(t *testing.T) {
	tt := types.NewTypeTable()
	if tt.CanImplicitCast(types.TypeI32, types.TypeI8) {
		t.Error("i32 should not implicitly cast to i8")
	}
}

func TestCanImplicitCast_FloatWidening(t *testing.T) {
	tt := types.NewTypeTable()
	if !tt.CanImplicitCast(types.TypeF32, types.TypeF64) {
		t.Error("f32 should implicitly cast to f64")
	}
	if tt.CanImplicitCast(types.TypeF64, types.TypeF32) {
		t.Error("f64 should not implicitly cast to f32")
	}
}

func TestCanImplicitCast_SignedUnsigned(t *testing.T) {
	tt := types.NewTypeTable()
	if tt.CanImplicitCast(types.TypeI32, types.TypeU32) {
		t.Error("i32 should not cast to u32")
	}
	if tt.CanImplicitCast(types.TypeU32, types.TypeI32) {
		t.Error("u32 should not cast to i32")
	}
}

func TestCommonType_SameType(t *testing.T) {
	tt := types.NewTypeTable()
	if typ, ok := tt.CommonType(types.TypeI32, types.TypeI32); !ok || typ != types.TypeI32 {
		t.Error("common type of i32 and i32 should be i32")
	}
}

func TestCommonType_Widening(t *testing.T) {
	tt := types.NewTypeTable()
	if typ, ok := tt.CommonType(types.TypeI8, types.TypeI32); !ok || typ != types.TypeI32 {
		t.Error("common type of i8 and i32 should be i32")
	}
	// order shouldn't matter
	if typ, ok := tt.CommonType(types.TypeI32, types.TypeI8); !ok || typ != types.TypeI32 {
		t.Error("common type of i32 and i8 should be i32")
	}
}

func TestCommonType_IntFloat(t *testing.T) {
	tt := types.NewTypeTable()
	if typ, ok := tt.CommonType(types.TypeI32, types.TypeF64); !ok || typ != types.TypeF64 {
		t.Error("common type of i32 and f64 should be f64")
	}
}

func TestCommonType_Incompatible(t *testing.T) {
	tt := types.NewTypeTable()
	if _, ok := tt.CommonType(types.TypeBool, types.TypeI32); ok {
		t.Error("bool and i32 should not have a common type")
	}
}

func TestFindByName(t *testing.T) {
	tt := types.NewTypeTable()
	id := tt.RegisterStruct(999, nil, nil)
	
	foundID, ok := tt.FindByName(999)
	if !ok { t.Fatal("should find type by NameID") }
	if foundID != id { t.Errorf("expected %d, got %d", id, foundID) }
}

func TestDeterminism(t *testing.T) {
	tt1 := types.NewTypeTable()
	tt2 := types.NewTypeTable()
	
	id1 := tt1.RegisterStruct(100, nil, nil)
	id2 := tt2.RegisterStruct(100, nil, nil)
	
	if id1 != id2 {
		t.Error("TypeTable is not deterministic")
	}
}
