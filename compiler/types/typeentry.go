package types

// TypeKind specifies the structural kind of a type.
type TypeKind uint8

const (
	KindPrimitive   TypeKind = iota // built-in primitive (i32, bool, etc.)
	KindStruct                      // struct { fields... }
	KindFunction                    // fn(params) -> return
	KindArray                       // [T; N] fixed-size array
	KindSlice                       // [T] dynamic slice
	KindTuple                       // (T1, T2, ...)
	KindSum                         // type X = A | B
	KindGeneric                     // unresolved [T] generic parameter
	KindGenericInst                 // concrete instantiation, e.g., Box[i32]
	KindPointer                     // *T raw pointer
	KindRef                         // &T / lent reference
	KindOption                      // Option[T]
	KindResult                      // Result[T, E]
	KindInterface                   // interface { methods... }
)

const (
	TypeFlagIsGeneric uint16 = 1 << iota
)

// TypeEntry is the common metadata header for all types in the TypeTable.
type TypeEntry struct {
	Kind   TypeKind
	NameID uint32 // interned name (0 if anonymous type)
	Size   uint32 // size in bytes
	Align  uint32 // alignment in bytes
	Flags  uint16 // boolean properties (e.g., is generic)
	Extra  uint32 // index into secondary tables (e.g., structs, funcs) depending on Kind
}

// FieldEntry describes a single field within a StructType.
type FieldEntry struct {
	NameID uint32
	TypeID TypeID
	Offset uint32
	Flags  uint8 // e.g., pub, mut
}

// StructType holds the full structural definition of a struct type.
type StructType struct {
	Fields        []FieldEntry
	GenericParams []uint32 // TypeIDs of generic parameters if this is a generic definition
}

// FuncType holds the signature of a function type.
type FuncType struct {
	Params     []TypeID
	Return     TypeID
	Effects    []uint32 // e.g., raises effects
	IsVariadic bool
	IsAsync    bool
}

// MethodSig represents a method signature required by an interface.
type MethodSig struct {
	NameID uint32
	Params []TypeID
	Return TypeID
}

// InterfaceType holds the structural requirements for an interface.
type InterfaceType struct {
	Methods []MethodSig
}
