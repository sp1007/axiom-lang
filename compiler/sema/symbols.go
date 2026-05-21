package sema

// SymKind classifies the kind of entity a symbol represents.
type SymKind uint8

const (
	SymVar          SymKind = iota // let/mut variable binding
	SymFunc                        // fn declaration
	SymStruct                      // struct declaration
	SymInterface                   // interface declaration
	SymTypeAlias                   // type X = Y declaration
	SymVariant                     // sum type variant
	SymParam                       // function parameter
	SymField                       // struct field
	SymGenericParam                // generic type parameter [T]
	SymModule                      // import module name
	SymBuiltinType                 // built-in primitive type (i32, bool, etc.)
	SymEnumVariant                 // sum type variant
	SymConst                       // const declaration
)

// String returns a human-readable name for the symbol kind.
func (k SymKind) String() string {
	switch k {
	case SymVar:
		return "Var"
	case SymFunc:
		return "Func"
	case SymStruct:
		return "Struct"
	case SymInterface:
		return "Interface"
	case SymTypeAlias:
		return "TypeAlias"
	case SymVariant:
		return "Variant"
	case SymParam:
		return "Param"
	case SymField:
		return "Field"
	case SymGenericParam:
		return "GenericParam"
	case SymModule:
		return "Module"
	case SymBuiltinType:
		return "BuiltinType"
	case SymEnumVariant:
		return "EnumVariant"
	case SymConst:
		return "Const"
	default:
		return "Unknown"
	}
}

// SymFlags encodes boolean properties of a symbol.
type SymFlags uint16

const (
	SymFlagPub      SymFlags = 1 << iota // pub visibility
	SymFlagMut                           // mutable (mut keyword)
	SymFlagExtern                        // extern "C" declaration
	SymFlagSink                          // sink parameter (!T)
	SymFlagLent                          // lent (borrowed) parameter
	SymFlagAsync                         // async fn
	SymFlagPure                          // verified pure function
	SymFlagMoved                         // symbol has been moved (invalidated)
	SymFlagUsed                          // symbol has been referenced at least once
	SymFlagComptime                      // compile-time constant (#run result)
	SymFlagGeneric                       // generic template
)

// Symbol represents a named entity in the program.
// Stored in a flat array for cache-friendly access.
// Memory layout: 20 bytes (32-bit fields x 5 = 20, plus uint8 + uint16 = 3 bytes, padded to 24 on some arches, but fields themselves are tightly packed).
type Symbol struct {
	NameID       uint32   // interned name (index into InternPool)
	Kind         SymKind  // what kind of entity
	Flags        SymFlags // boolean properties
	TypeID       uint32   // index into TypeTable (0 = unresolved)
	DeclNode     uint32   // index into AstTree.Nodes (source location, 0 = builtin)
	ScopeID      uint32   // which scope this symbol belongs to
	NextOverload uint32   // index of next overloaded function in same scope (0 = none)
}
