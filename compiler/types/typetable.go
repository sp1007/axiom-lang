package types

// TypeTable is the central registry of all types in a compilation unit.
// It assigns a unique TypeID to every structural type and stores their metadata.
type TypeTable struct {
	entries      []TypeEntry
	structs      []StructType
	funcs        []FuncType
	sumtypes     []SumType
	templates    []GenericTemplate
	interfaces   []InterfaceType
	genericInsts []GenericInstInfo
}

// NewTypeTable creates a new TypeTable pre-populated with primitive types.
func NewTypeTable() *TypeTable {
	tt := &TypeTable{
		entries:      make([]TypeEntry, 0, 256),
		structs:      make([]StructType, 0, 64),
		funcs:        make([]FuncType, 0, 64),
		sumtypes:     make([]SumType, 0, 32),
		templates:    make([]GenericTemplate, 0, 16),
		interfaces:   make([]InterfaceType, 0, 16),
		genericInsts: make([]GenericInstInfo, 0, 32),
	}

	// 0: Unknown sentinel
	tt.entries = append(tt.entries, TypeEntry{Kind: KindPrimitive, Size: 0, Align: 0})

	// 1-16: Primitives (must match constants in typeid.go)
	// We leave NameID as 0 here since primitive names are interned in SymbolTable.
	// If we want NameIDs here, we'd need to pass the InternPool. For now, 0 is fine.
	for i := TypeI8; i <= TypeUSize; i++ {
		tt.entries = append(tt.entries, TypeEntry{
			Kind:  KindPrimitive,
			Size:  i.SizeOf(),
			Align: i.SizeOf(), // simple alignment assumption for primitives
		})
	}

	// 17-20: Built-in interfaces
	for i := 0; i < 4; i++ {
		ifaceIdx := uint32(len(tt.interfaces))
		tt.interfaces = append(tt.interfaces, InterfaceType{Methods: nil}) // Mock methods for builtins
		tt.entries = append(tt.entries, TypeEntry{
			Kind:   KindInterface,
			NameID: 0, // Should be interned later or hardcoded
			Size:   16,
			Align:  8,
			Flags:  0,
			Extra:  ifaceIdx,
		})
	}

	// 21: ActorRef (built-in opaque struct for now)
	tt.entries = append(tt.entries, TypeEntry{
		Kind:   KindStruct,
		NameID: 0, // interned later
		Size:   8, // pointer size
		Align:  8,
		Flags:  0,
		Extra:  0, // mock struct extra
	})

	return tt
}

// RegisterStruct registers a new struct type and returns its TypeID.
func (tt *TypeTable) RegisterStruct(nameID uint32, fields []FieldEntry, generics []uint32) TypeID {
	structIdx := uint32(len(tt.structs))
	tt.structs = append(tt.structs, StructType{
		Fields:        fields,
		GenericParams: generics,
	})

	id := TypeID(len(tt.entries))
	tt.entries = append(tt.entries, TypeEntry{
		Kind:   KindStruct,
		NameID: nameID,
		Size:   0, // computed later during layout phase
		Align:  0, // computed later
		Extra:  structIdx,
	})

	return id
}

// RegisterFunction registers a new function type and returns its TypeID.
func (tt *TypeTable) RegisterFunction(params []TypeID, ret TypeID, effects []uint32) TypeID {
	funcIdx := uint32(len(tt.funcs))
	tt.funcs = append(tt.funcs, FuncType{
		Params:  params,
		Return:  ret,
		Effects: effects,
	})

	id := TypeID(len(tt.entries))
	tt.entries = append(tt.entries, TypeEntry{
		Kind:  KindFunction,
		Size:  8, // function pointer size
		Align: 8,
		Extra: funcIdx,
	})

	return id
}

// RegisterGenericType registers a new unresolved generic parameter type.
func (tt *TypeTable) RegisterGenericType(nameID uint32) TypeID {
	id := TypeID(len(tt.entries))
	tt.entries = append(tt.entries, TypeEntry{
		Kind:   KindGeneric,
		NameID: nameID,
		Size:   0,
		Align:  0,
	})
	return id
}

// RegisterGenericTemplate adds a new generic template and returns its internal index.
// The index is returned, not a TypeID, because generic templates are not instantiated types.
func (tt *TypeTable) RegisterGenericTemplate(tmpl GenericTemplate) uint32 {
	idx := uint32(len(tt.templates))
	tt.templates = append(tt.templates, tmpl)
	return idx
}

// GenericTemplate returns a pointer to a registered GenericTemplate by index.
func (tt *TypeTable) GenericTemplate(idx uint32) *GenericTemplate {
	if int(idx) >= len(tt.templates) {
		panic("TypeTable: invalid generic template index")
	}
	return &tt.templates[idx]
}

// FindGenericTemplate searches for a generic template by its SymID.
func (tt *TypeTable) FindGenericTemplate(symID uint32) (*GenericTemplate, bool) {
	for i := range tt.templates {
		if tt.templates[i].SymID == symID {
			return &tt.templates[i], true
		}
	}
	return nil, false
}

// Entry returns the TypeEntry for the given TypeID.
func (tt *TypeTable) Entry(id TypeID) *TypeEntry {
	if int(id) >= len(tt.entries) {
		panic("TypeTable: invalid TypeID")
	}
	return &tt.entries[id]
}

// RegisterSumType registers a new sum type and returns its TypeID.
func (tt *TypeTable) RegisterSumType(nameID uint32, variants []VariantInfo, generics []uint32) TypeID {
	sumIdx := uint32(len(tt.sumtypes))
	tt.sumtypes = append(tt.sumtypes, SumType{
		Variants:      variants,
		GenericParams: generics,
	})

	idx := len(tt.entries)
	tt.entries = append(tt.entries, TypeEntry{
		Kind:   KindSum,
		NameID: nameID,
		Size:   0, // size depends on largest variant + tag
		Align:  0,
		Flags:  0,
		Extra:  sumIdx,
	})

	return TypeID(idx)
}

// StructInfo returns the StructType definition for a struct TypeID.
// Panics if the type is not a struct.
func (tt *TypeTable) StructInfo(id TypeID) *StructType {
	entry := tt.Entry(id)
	if entry.Kind != KindStruct {
		panic("TypeTable: TypeID is not a struct")
	}
	return &tt.structs[entry.Extra]
}

// FuncInfo returns the FuncType definition for a function TypeID.
// Panics if the type is not a function.
func (tt *TypeTable) FuncInfo(id TypeID) *FuncType {
	entry := tt.Entry(id)
	if entry.Kind != KindFunction {
		panic("TypeTable: TypeID is not a function")
	}
	return &tt.funcs[entry.Extra]
}

// SumInfo returns the full SumType metadata for a given TypeID.
// Panics if the type is not a sum type.
func (tt *TypeTable) SumInfo(id TypeID) *SumType {
	entry := tt.Entry(id)
	if entry.Kind != KindSum {
		panic("TypeTable: TypeID is not a sum type")
	}
	return &tt.sumtypes[entry.Extra]
}

// RegisterInterface registers a new interface type and returns its TypeID.
func (tt *TypeTable) RegisterInterface(nameID uint32, methods []MethodSig) TypeID {
	ifaceIdx := uint32(len(tt.interfaces))
	tt.interfaces = append(tt.interfaces, InterfaceType{
		Methods: methods,
	})

	id := TypeID(len(tt.entries))
	tt.entries = append(tt.entries, TypeEntry{
		Kind:   KindInterface,
		NameID: nameID,
		Size:   16, // size of fat pointer (ptr + vtable)
		Align:  8,
		Flags:  0,
		Extra:  ifaceIdx,
	})

	return id
}

// InterfaceInfo returns the full InterfaceType metadata for a given TypeID.
// Panics if the type is not an interface.
func (tt *TypeTable) InterfaceInfo(id TypeID) *InterfaceType {
	entry := tt.Entry(id)
	if entry.Kind != KindInterface {
		panic("TypeTable: TypeID is not an interface type")
	}
	return &tt.interfaces[entry.Extra]
}

// Entries returns all registered TypeEntries.
func (tt *TypeTable) Entries() []TypeEntry {
	return tt.entries
}

// Count returns the total number of registered types.
func (tt *TypeTable) Count() int {
	return len(tt.entries)
}

// FindByName searches for a type by its interned name ID.
// Returns the TypeID and true if found, (TypeUnknown, false) if not found.
// This performs a linear scan; it is typically used for resolving top-level type names.
func (tt *TypeTable) FindByName(nameID uint32) (TypeID, bool) {
	for i, e := range tt.entries {
		if e.NameID == nameID {
			return TypeID(i), true
		}
	}
	return TypeUnknown, false
}

// IsAssignableTo checks if a value of type 'from' can be assigned to a variable of type 'to'.
func (tt *TypeTable) IsAssignableTo(from, to TypeID) bool {
	if from == to {
		return true
	}
	return tt.CanImplicitCast(from, to)
}

// CanImplicitCast checks if type 'from' can be implicitly converted to type 'to'.
func (tt *TypeTable) CanImplicitCast(from, to TypeID) bool {
	if !from.IsPrimitive() || !to.IsPrimitive() {
		return false // No implicit casts for non-primitives yet
	}

	// Floating point widening: f32 -> f64
	if from == TypeF32 && to == TypeF64 {
		return true
	}

	// Integer widening: must preserve signedness and strictly widen
	if from.IsSigned() && to.IsSigned() {
		return from.SizeOf() < to.SizeOf() && from != TypeISize && to != TypeISize
	}
	if from.IsUnsigned() && to.IsUnsigned() {
		return from.SizeOf() < to.SizeOf() && from != TypeUSize && to != TypeUSize
	}

	return false
}

// CommonType computes the resulting type of a binary operation between two types.
// Returns (TypeID, true) if compatible, or (TypeUnknown, false) if incompatible.
func (tt *TypeTable) CommonType(a, b TypeID) (TypeID, bool) {
	if a == b {
		return a, true
	}

	if !a.IsPrimitive() || !b.IsPrimitive() {
		return TypeUnknown, false
	}

	// Float + Float -> wider float
	if a.IsFloat() && b.IsFloat() {
		if a == TypeF64 || b == TypeF64 {
			return TypeF64, true
		}
		return TypeF32, true
	}

	// Int + Int -> wider integer (if same signedness)
	if a.IsSigned() && b.IsSigned() {
		if a.SizeOf() >= b.SizeOf() {
			return a, true
		}
		return b, true
	}
	if a.IsUnsigned() && b.IsUnsigned() {
		if a.SizeOf() >= b.SizeOf() {
			return a, true
		}
		return b, true
	}

	// Int + Float -> Float (widening integer to float)
	if a.IsInteger() && b.IsFloat() {
		return b, true
	}
	if a.IsFloat() && b.IsInteger() {
		return a, true
	}

	// Otherwise, incompatible (e.g. signed + unsigned, numeric + string)
	return TypeUnknown, false
}

// RegisterPointer registers a pointer-to-T type and returns its TypeID.
// Extra stores the inner TypeID of the pointee.
func (tt *TypeTable) RegisterPointer(innerTypeID TypeID) TypeID {
	id := TypeID(len(tt.entries))
	tt.entries = append(tt.entries, TypeEntry{
		Kind:  KindPointer,
		Size:  8, // pointer size (64-bit)
		Align: 8,
		Extra: uint32(innerTypeID),
	})
	return id
}

// RegisterSlice registers a slice-of-T type and returns its TypeID.
// Extra stores the inner TypeID of the element.
func (tt *TypeTable) RegisterSlice(elemTypeID TypeID) TypeID {
	id := TypeID(len(tt.entries))
	tt.entries = append(tt.entries, TypeEntry{
		Kind:  KindSlice,
		Size:  24, // ptr + len + cap (3 x 8)
		Align: 8,
		Extra: uint32(elemTypeID),
	})
	return id
}

// RegisterGenericInst registers a monomorphized generic instantiation (e.g., Box[i32]).
// nameID is the interned name of the template (e.g., "Box"), typeArgs are the concrete type arguments.
// The GenericInstInfo is stored in a separate slice; Extra indexes into it.
func (tt *TypeTable) RegisterGenericInst(nameID uint32, typeArgs []TypeID) TypeID {
	instIdx := uint32(len(tt.genericInsts))
	tt.genericInsts = append(tt.genericInsts, GenericInstInfo{
		TypeArgs: typeArgs,
	})

	id := TypeID(len(tt.entries))
	tt.entries = append(tt.entries, TypeEntry{
		Kind:   KindGenericInst,
		NameID: nameID,
		Size:   0, // depends on instantiation
		Align:  0,
		Extra:  instIdx,
	})

	return id
}

// GenericInstInfo holds the type arguments for a concrete generic instantiation.
type GenericInstInfo struct {
	TypeArgs []TypeID
}

// GenericInstArgs returns the type arguments for a generic instantiation TypeID.
// Panics if the type is not a KindGenericInst.
func (tt *TypeTable) GenericInstArgs(id TypeID) []TypeID {
	entry := tt.Entry(id)
	if entry.Kind != KindGenericInst {
		panic("TypeTable: TypeID is not a generic instantiation")
	}
	return tt.genericInsts[entry.Extra].TypeArgs
}

// PointerElem returns the element TypeID of a pointer type.
// Panics if the type is not KindPointer.
func (tt *TypeTable) PointerElem(id TypeID) TypeID {
	entry := tt.Entry(id)
	if entry.Kind != KindPointer {
		panic("TypeTable: TypeID is not a pointer")
	}
	return TypeID(entry.Extra)
}

// SliceElem returns the element TypeID of a slice type.
// Panics if the type is not KindSlice.
func (tt *TypeTable) SliceElem(id TypeID) TypeID {
	entry := tt.Entry(id)
	if entry.Kind != KindSlice {
		panic("TypeTable: TypeID is not a slice")
	}
	return TypeID(entry.Extra)
}
