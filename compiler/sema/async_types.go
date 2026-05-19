package sema

import "github.com/axiom-lang/axiom/compiler/types"

// CreateFutureType returns an instantiated Future[T] type for the given inner type.
func CreateFutureType(tt *types.TypeTable, innerType types.TypeID) types.TypeID {
	// For MVP, we mock Future[T] as a struct with a single generic parameter.
	// Since we don't have a full instantiation pipeline here, we just create a struct
	// type on the fly that acts as Future[T].
	
	// We could find the template and instantiate it properly, but the TypeTable
	// doesn't have an InstantiateGenericStruct method yet. Let's just create one.
	return tt.RegisterStruct(0, []types.FieldEntry{
		{NameID: 0, TypeID: innerType, Offset: 0, Flags: 0},
	}, []uint32{uint32(innerType)})
}

// IsFutureType checks if the given type is a Future[T] and returns its inner type T.
func IsFutureType(tt *types.TypeTable, typeID types.TypeID) (bool, types.TypeID) {
	entry := tt.Entry(typeID)
	if entry.Kind != types.KindStruct {
		return false, types.TypeUnknown
	}
	
	info := tt.StructInfo(typeID)
	// We identify Future[T] by having exactly 1 generic param in MVP mock
	// (or checking NameID if we instantiated it with NameID = "Future")
	if len(info.GenericParams) == 1 {
		return true, types.TypeID(info.GenericParams[0])
	}
	
	return false, types.TypeUnknown
}
