package sema

import (
	"fmt"
	"strings"

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


	for idx, sym := range chk.symtable.Symbols {
		if sym.Kind == SymFunc {
			tID := types.TypeID(sym.TypeID)
			funcName := string(chk.symtable.intern.Get(sym.NameID))
			if strings.Contains(funcName, "contains") || strings.Contains(funcName, "get") {
				entryKind := "unknown"
				if tID != types.TypeUnknown {
					entryKind = fmt.Sprintf("%v", chk.types.Entry(tID).Kind)
				}
				fmt.Printf("[DEBUG-GET-METHODS-SYM] funcName=%s (idx %d) tID=%d kind=%s\n", funcName, idx, tID, entryKind)
			}
			if tID != types.TypeUnknown && chk.types.Entry(tID).Kind == types.KindFunction {
				fInfo := chk.types.FuncInfo(tID)
				if len(fInfo.Params) > 0 {
					firstParamType := fInfo.Params[0]
					// Check if first parameter is exactly the struct type, or a ref to it.
					matched := chk.baseTypeEquals(firstParamType, structType)
					funcName := string(chk.symtable.intern.Get(sym.NameID))
					if strings.Contains(funcName, "contains") || strings.Contains(funcName, "insert") || strings.Contains(funcName, "remove") {
						fmt.Printf("[DEBUG-GET-METHODS] funcName=%s (idx %d) firstParamType=%d structType=%d matched=%v\n", 
							funcName, idx, firstParamType, structType, matched)
					}
					if matched {
						nameID := sym.NameID
						if chk.symtable.InstantiatedToOriginalName != nil {
							if origNameID, ok := chk.symtable.InstantiatedToOriginalName[uint32(idx)]; ok {
								nameID = origNameID
							}
						}

						// Strip reference/pointer modifiers from firstParamType
						selfBaseTypeID := firstParamType
						entrySelf := chk.types.Entry(selfBaseTypeID)
						if entrySelf.Kind == types.KindPointer {
							selfBaseTypeID = chk.types.PointerElem(selfBaseTypeID)
							entrySelf = chk.types.Entry(selfBaseTypeID)
						} else if entrySelf.Kind == types.KindRef {
							selfBaseTypeID = types.TypeID(entrySelf.Extra)
							entrySelf = chk.types.Entry(selfBaseTypeID)
						}

						methodParams := append([]types.TypeID(nil), fInfo.Params[1:]...)
						methodReturn := fInfo.Return

						// If self is a GenericInst and structType has GenericParams, perform mapping/substitution
						structEntry := chk.types.Entry(structType)
						if entrySelf.Kind == types.KindGenericInst && (structEntry.Kind == types.KindStruct || structEntry.Kind == types.KindSum || structEntry.Kind == types.KindGenericInst) {
							var structGPs []uint32
							if structEntry.Kind == types.KindStruct {
								structInfo := chk.types.StructInfo(structType)
								structGPs = structInfo.GenericParams
							} else if structEntry.Kind == types.KindSum {
								sumInfo := chk.types.SumInfo(structType)
								structGPs = sumInfo.GenericParams
							}

							if len(structGPs) == 0 {
								// Find the original template struct to get its GenericParams
								baseStructName := getBaseName(string(chk.symtable.intern.Get(structEntry.NameID)))
								baseNameID := chk.symtable.intern.InternString(baseStructName)
								if tmplSymIdx, found := chk.symtable.ResolveGlobal(baseNameID); found {
									tmplSym := chk.symtable.SymbolAt(tmplSymIdx)
									tmplTypeID := types.TypeID(tmplSym.TypeID)
									if tmplTypeID != types.TypeUnknown {
										tmplEntry := chk.types.Entry(tmplTypeID)
										if tmplEntry.Kind == types.KindStruct {
											structGPs = chk.types.StructInfo(tmplTypeID).GenericParams
										} else if tmplEntry.Kind == types.KindSum {
											structGPs = chk.types.SumInfo(tmplTypeID).GenericParams
										}
									}
								}
							}

							selfTypeArgs := chk.types.GenericInstArgs(selfBaseTypeID)
							structTypeArgs := chk.getTypeArgs(structType, structEntry)
							if len(structTypeArgs) == 0 && len(structGPs) > 0 {
								structTypeArgs = make([]types.TypeID, len(structGPs))
								for i, gp := range structGPs {
									structTypeArgs[i] = types.TypeID(gp)
								}
							}

							fmt.Printf("[DEBUG-GET-METHODS-SUB] selfBaseTypeID=%d, selfTypeArgs=%v, structTypeArgs=%v, structGPs=%v\n",
								selfBaseTypeID, selfTypeArgs, structTypeArgs, structGPs)

							if len(selfTypeArgs) > 0 && len(selfTypeArgs) == len(structTypeArgs) {
								paramsToSub := make([]uint32, len(selfTypeArgs))
								argsToSub := make([]types.TypeID, len(structTypeArgs))
								for i, arg := range selfTypeArgs {
									paramsToSub[i] = uint32(arg)
									argsToSub[i] = structTypeArgs[i]
								}

								// Substitute remaining params and return type
								for i, pVal := range methodParams {
									methodParams[i] = chk.types.SubstituteGenericType(pVal, paramsToSub, argsToSub)
								}
								oldRet := methodReturn
								methodReturn = chk.types.SubstituteGenericType(methodReturn, paramsToSub, argsToSub)
								fmt.Printf("[DEBUG-GET-METHODS-SUB-RES] oldRet=%d, newRet=%d, paramsToSub=%v, argsToSub=%v\n",
									oldRet, methodReturn, paramsToSub, argsToSub)
							}
						}

						// Add method without the first 'self' parameter
						m := types.MethodSig{
							NameID: nameID,
							Params: methodParams,
							Return: methodReturn,
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

func getBaseName(name string) string {
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		name = parts[len(parts)-1]
	}
	if !strings.HasPrefix(name, "_AX_") {
		return name
	}
	parts := strings.SplitN(name, "__", 2)
	if len(parts) < 2 {
		parts = strings.Split(name, "_")
		if len(parts) < 4 {
			return name
		}
		return parts[3]
	}
	firstParts := strings.Split(parts[0], "_")
	return firstParts[len(firstParts)-1]
}


func (chk *Interfaces) getTypeArgs(tID types.TypeID, entry *types.TypeEntry) []types.TypeID {
	if entry.Kind == types.KindGenericInst {
		return chk.types.GenericInstArgs(tID)
	}
	if (entry.Kind == types.KindStruct || entry.Kind == types.KindSum) && entry.NameID != 0 {
		nameStr := string(chk.symtable.intern.Get(entry.NameID))
		if strings.HasPrefix(nameStr, "_AX_") {
			_, args := parseMangledName(chk.types, chk.symtable.intern, nameStr)
			return args
		}
	}
	return nil
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
	if name1 != 0 && name2 != 0 {
		nameStr1 := string(chk.symtable.intern.Get(name1))
		nameStr2 := string(chk.symtable.intern.Get(name2))
		baseName1 := getBaseName(nameStr1)
		baseName2 := getBaseName(nameStr2)
		if baseName1 != "" && baseName1 == baseName2 {
			if (entry1.Kind == types.KindGenericInst || entry1.Kind == types.KindStruct || entry1.Kind == types.KindSum) &&
				(entry2.Kind == types.KindGenericInst || entry2.Kind == types.KindStruct || entry2.Kind == types.KindSum) {
				// The generic status of the two types must match to prevent matching generic templates with concrete types.
				if isGeneric(chk.types, t1) == isGeneric(chk.types, target) || isGeneric(chk.types, t1) || isGeneric(chk.types, target) {
					// Additionally, if they are both concrete, their type arguments must match exactly.
					if !isGeneric(chk.types, t1) && !isGeneric(chk.types, target) {
						args1 := chk.getTypeArgs(t1, entry1)
						args2 := chk.getTypeArgs(target, entry2)
						if len(args1) != len(args2) {
							return false
						}
						for i, arg := range args1 {
							if arg != args2[i] {
								return false
							}
						}
					}
					return true
				}
			}
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
