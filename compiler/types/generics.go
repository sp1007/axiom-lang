package types

import "github.com/axiom-lang/axiom/compiler/ast"

// GenericParam represents a single type parameter in a generic template.
type GenericParam struct {
	NameID     uint32 // interned name of the type parameter (e.g., 'T')
	Constraint uint32 // interface TypeID, 0 if unconstrained
}

// GenericTemplate represents the declaration of a generic function, struct, or interface.
// It is used by the monomorphization pass to instantiate concrete versions.
type GenericTemplate struct {
	SymID        uint32            // symbol ID of the generic function/struct
	NodeIdx      uint32            // AST node index of the template declaration
	Params       []GenericParam    // type parameters [T, U, ...]
	ParamTypeIDs []TypeID          // TypeIDs of the generic parameters
	Instances    map[string]uint32 // "T_TypeID,U_TypeID" -> instantiated TypeID (or SymID for functions)
	SrcTree      *ast.AstTree      // AST tree where this template is defined
}

// NewGenericTemplate creates a new generic template.
func NewGenericTemplate(symID, nodeIdx uint32, params []GenericParam, paramTypeIDs []TypeID, srcTree *ast.AstTree) GenericTemplate {
	return GenericTemplate{
		SymID:        symID,
		NodeIdx:      nodeIdx,
		Params:       params,
		ParamTypeIDs: paramTypeIDs,
		Instances:    make(map[string]uint32),
		SrcTree:      srcTree,
	}
}
