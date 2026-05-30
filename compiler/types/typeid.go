package types

// TypeID uniquely identifies a type in the TypeTable.
type TypeID uint32

// Primitive TypeID constants (FROZEN).
// These MUST match the initialization order in TypeTable.
const (
	TypeUnknown TypeID = 0 // unresolved sentinel
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

	TypeOrd     TypeID = 17
	TypeEq      TypeID = 18
	TypeHash    TypeID = 19
	TypeDisplay TypeID = 20
	TypeActorRef TypeID = 21

	PrimitiveCount = 16
	BuiltinInterfaceCount = 4
)

// IsUnknown returns true if the type is unresolved.
func (t TypeID) IsUnknown() bool { return t == TypeUnknown }

// IsPrimitive returns true if the type is a primitive type.
func (t TypeID) IsPrimitive() bool { return t >= 1 && t <= PrimitiveCount }

// IsInteger returns true if the type is a built-in integer (signed or unsigned).
func (t TypeID) IsInteger() bool { return (t >= TypeI8 && t <= TypeU64) || t == TypeISize || t == TypeUSize }

// IsSigned returns true if the type is a signed integer.
func (t TypeID) IsSigned() bool { return (t >= TypeI8 && t <= TypeI64) || t == TypeISize }

// IsUnsigned returns true if the type is an unsigned integer.
func (t TypeID) IsUnsigned() bool { return (t >= TypeU8 && t <= TypeU64) || t == TypeUSize }

// IsFloat returns true if the type is a built-in float.
func (t TypeID) IsFloat() bool { return t == TypeF32 || t == TypeF64 }

// IsNumeric returns true if the type is an integer or float.
func (t TypeID) IsNumeric() bool { return t.IsInteger() || t.IsFloat() }

// IsBool returns true if the type is boolean.
func (t TypeID) IsBool() bool { return t == TypeBool }

// IsVoid returns true if the type is void.
func (t TypeID) IsVoid() bool { return t == TypeVoid }

// IsString returns true if the type is string.
func (t TypeID) IsString() bool { return t == TypeString }

// SizeOf returns the size in bytes of a primitive type.
// For non-primitive types, it returns 0 (requires full type table query).
func (t TypeID) SizeOf() uint32 {
	switch t {
	case TypeI8, TypeU8, TypeChar8, TypeBool:
		return 1
	case TypeI16, TypeU16:
		return 2
	case TypeI32, TypeU32, TypeF32:
		return 4
	case TypeI64, TypeU64, TypeF64:
		return 8
	case TypeString:
		return 16 // string is a fat pointer (ptr + len)
	case TypeISize, TypeUSize:
		return 8 // assume 64-bit platform for now
	case TypeVoid, TypeUnknown:
		return 0
	}
	return 0
}

// String returns the string representation of a primitive type.
func (t TypeID) String() string {
	switch t {
	case TypeUnknown: return "unknown"
	case TypeI8: return "i8"
	case TypeI16: return "i16"
	case TypeI32: return "i32"
	case TypeI64: return "i64"
	case TypeU8: return "u8"
	case TypeU16: return "u16"
	case TypeU32: return "u32"
	case TypeU64: return "u64"
	case TypeF32: return "f32"
	case TypeF64: return "f64"
	case TypeBool: return "bool"
	case TypeString: return "string"
	case TypeChar8: return "char8"
	case TypeVoid: return "void"
	case TypeISize: return "isize"
	case TypeUSize: return "usize"
	}
	return "type" // Should not happen for primitives
}
