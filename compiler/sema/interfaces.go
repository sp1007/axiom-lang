package sema

import (
	"fmt"
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

	fmt.Printf("[DEBUG] getMethodsOfStruct called for struct %d\n", structType)
	for idx, sym := range chk.symtable.Symbols {
		if sym.Kind == SymFunc {
			fmt.Printf("  [DEBUG] SymFunc idx=%d NameID=%d TypeID=%d\n", idx, sym.NameID, sym.TypeID)
		}
	}

	for idx, sym := range chk.symtable.Symbols {
		if sym.Kind == SymFunc {
			tID := types.TypeID(sym.TypeID)
			if tID != types.TypeUnknown && chk.types.Entry(tID).Kind == types.KindFunction {
				fInfo := chk.types.FuncInfo(tID)
				if len(fInfo.Params) > 0 {
					firstParamType := fInfo.Params[0]
					// Check if first parameter is exactly the struct type, or a ref to it.
					// For simplicity, we only match exact struct type right now.
					if chk.baseTypeEquals(firstParamType, structType) {
						nameID := sym.NameID
						if chk.symtable.InstantiatedToOriginalName != nil {
							if origNameID, ok := chk.symtable.InstantiatedToOriginalName[uint32(idx)]; ok {
								nameID = origNameID
							}
						}
						// Add method without the first 'self' parameter
						m := types.MethodSig{
							NameID: nameID,
							Params: append([]types.TypeID(nil), fInfo.Params[1:]...),
							Return: fInfo.Return,
						}
						methods = append(methods, m)
					}
				}
			}
		}
	}

	chk.methodCache[structType] = methods
	return methods
}

func (chk *Interfaces) baseTypeEquals(t1, target types.TypeID) bool {
	// Strip reference/pointer modifiers if any
	entry1 := chk.types.Entry(t1)
	if entry1.Kind == types.KindPointer {
		t1 = chk.types.PointerElem(t1)
		entry1 = chk.types.Entry(t1)
	} else if entry1.Kind == types.KindRef {
		t1 = types.TypeID(entry1.Extra)
		entry1 = chk.types.Entry(t1)
	}

	entry2 := chk.types.Entry(target)
	if entry2.Kind == types.KindPointer {
		target = chk.types.PointerElem(target)
		entry2 = chk.types.Entry(target)
	} else if entry2.Kind == types.KindRef {
		target = types.TypeID(entry2.Extra)
		entry2 = chk.types.Entry(target)
	}

	if t1 == target {
		return true
	}

	// If either is a GenericInst or Sum/Struct, check if their base NameID matches
	name1 := entry1.NameID
	name2 := entry2.NameID
	if name1 != 0 && name2 != 0 && name1 == name2 {
		if (entry1.Kind == types.KindGenericInst || entry1.Kind == types.KindStruct || entry1.Kind == types.KindSum) &&
			(entry2.Kind == types.KindGenericInst || entry2.Kind == types.KindStruct || entry2.Kind == types.KindSum) {
			return true
		}
	}

	return false
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
