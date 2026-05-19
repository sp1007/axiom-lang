package types

// GenericParam represents a single type parameter in a generic template.
type GenericParam struct {
	NameID     uint32 // interned name of the type parameter (e.g., 'T')
	Constraint uint32 // interface TypeID, 0 if unconstrained
}

// GenericTemplate represents the declaration of a generic function, struct, or interface.
// It is used by the monomorphization pass to instantiate concrete versions.
type GenericTemplate struct {
	SymID     uint32            // symbol ID of the generic function/struct
	NodeIdx   uint32            // AST node index of the template declaration
	Params    []GenericParam    // type parameters [T, U, ...]
	Instances map[string]uint32 // "T_TypeID,U_TypeID" -> instantiated TypeID (or SymID for functions)
}

// NewGenericTemplate creates a new generic template.
func NewGenericTemplate(symID, nodeIdx uint32, params []GenericParam) GenericTemplate {
	return GenericTemplate{
		SymID:     symID,
		NodeIdx:   nodeIdx,
		Params:    params,
		Instances: make(map[string]uint32),
	}
}
