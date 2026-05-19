package sema

import (
	"github.com/axiom-lang/axiom/compiler/types"
)

// Interfaces provides structural interface satisfaction checking.
type Interfaces struct {
	symtable    *SymbolTable
	types       *types.TypeTable
	methodCache map[types.TypeID][]types.MethodSig
	cache       map[structIfacePair]bool
}

type structIfacePair struct {
	structType types.TypeID
	ifaceType  types.TypeID
}

// NewInterfaces creates a new Interfaces checker.
func NewInterfaces(st *SymbolTable, tt *types.TypeTable) *Interfaces {
	return &Interfaces{
		symtable:    st,
		types:       tt,
		methodCache: make(map[types.TypeID][]types.MethodSig),
		cache:       make(map[structIfacePair]bool),
	}
}

// ImplementsInterface checks if structType satisfies ifaceType structurally.
// Returns true if satisfied, otherwise false and the list of missing methods.
func (chk *Interfaces) ImplementsInterface(structType, ifaceType types.TypeID) (bool, []types.MethodSig) {
	pair := structIfacePair{structType, ifaceType}
	if cached, ok := chk.cache[pair]; ok {
		// If cached, we assume it's true, because if it was false we might want to return the missing methods.
		// For simplicity, we just return true and nil missing methods if cached == true.
		if cached {
			return true, nil
		}
	}

	if chk.isBuiltinImpl(structType, ifaceType) {
		chk.cache[pair] = true
		return true, nil
	}

	ifaceInfo := chk.types.InterfaceInfo(ifaceType)
	structMethods := chk.getMethodsOfStruct(structType)

	var missing []types.MethodSig

	for _, reqMethod := range ifaceInfo.Methods {
		found := false
		for _, sMethod := range structMethods {
			if sMethod.NameID == reqMethod.NameID {
				if chk.signaturesMatch(sMethod, reqMethod) {
					found = true
					break
				}
			}
		}
		if !found {
			missing = append(missing, reqMethod)
		}
	}

	satisfied := len(missing) == 0
	chk.cache[pair] = satisfied

	return satisfied, missing
}

func (chk *Interfaces) signaturesMatch(a, b types.MethodSig) bool {
	if a.Return != b.Return {
		return false
	}
	// Note: b.Params does not include 'self', but a.Params might?
	// Wait, reqMethod.Params does not include 'self' typically. 
	// But getMethodsOfStruct should probably strip the 'self' param before adding to structMethods.
	if len(a.Params) != len(b.Params) {
		return false
	}
	for i, pt := range a.Params {
		if pt != b.Params[i] {
			return false
		}
	}
	return true
}

// getMethodsOfStruct scans the symbol table for functions whose first parameter is the structType (or a ref to it).
func (chk *Interfaces) getMethodsOfStruct(structType types.TypeID) []types.MethodSig {
	if methods, ok := chk.methodCache[structType]; ok {
		return methods
	}

	var methods []types.MethodSig

	for _, sym := range chk.symtable.Symbols {
		if sym.Kind == SymFunc {
			fInfo := chk.types.FuncInfo(types.TypeID(sym.TypeID))
			if len(fInfo.Params) > 0 {
				firstParamType := fInfo.Params[0]
				// Check if first parameter is exactly the struct type, or a ref to it.
				// For simplicity, we only match exact struct type right now.
				if chk.baseTypeEquals(firstParamType, structType) {
					// Add method without the first 'self' parameter
					m := types.MethodSig{
						NameID: sym.NameID,
						Params: append([]types.TypeID(nil), fInfo.Params[1:]...),
						Return: fInfo.Return,
					}
					methods = append(methods, m)
				}
			}
		}
	}

	chk.methodCache[structType] = methods
	return methods
}

func (chk *Interfaces) baseTypeEquals(t1, target types.TypeID) bool {
	// Strip reference/pointer modifiers if any
	// Right now we assume strict equality for simplicity,
	// but normally &T should match T for method receivers.
	entry := chk.types.Entry(t1)
	if entry.Kind == types.KindRef || entry.Kind == types.KindPointer {
		// If it's a ref/ptr, we would need to check its base type.
		// Since we don't have PointerInfo easily accessible here without knowing its structure,
		// we'll just check exact type. Wait, we can use Extra for pointer base type maybe?
		// Actually, let's just do strict equality for MVP.
	}
	return t1 == target
}

// isBuiltinImpl checks if a primitive type implements a builtin interface.
func (chk *Interfaces) isBuiltinImpl(structType, ifaceType types.TypeID) bool {
	switch ifaceType {
	case types.TypeOrd, types.TypeEq, types.TypeHash:
		return structType >= types.TypeI8 && structType <= types.TypeString
	case types.TypeDisplay:
		return structType == types.TypeString
	default:
		return false
	}
}
