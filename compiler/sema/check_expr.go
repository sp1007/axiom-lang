package sema

import (
	"fmt"
	"strings"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

// checkExpr checks an expression node for type safety rules not covered by inference.
func (tc *TypeChecker) checkExpr(nodeIdx uint32) {
	if nodeIdx == 0 {
		return
	}

	node := &tc.ast.Nodes[nodeIdx]

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[PANIC-DEBUG] nodeIdx=%d kind=%v payload=%d\n", nodeIdx, node.Kind, node.Payload)
			panic(r)
		}
	}()

	switch node.Kind {
	case ast.NodeUnaryExpr:
		op := node.Flags
		operand := node.FirstChild
		if operand != 0 {
			tc.checkStmt(operand) // continue traversal
			operandType := tc.infer.TypeOf(operand)
			if operandType != types.TypeUnknown {
				// op flags: 1 = -, 2 = not, 3 = ~
				if op == 1 { // -
					if !operandType.IsInteger() && !operandType.IsFloat() {
						tc.errorf(nodeIdx, 3020, "invalid operator '-' for type %d", operandType)
					}
				} else if op == 2 { // not
					if operandType != types.TypeBool {
						tc.errorf(nodeIdx, 3021, "invalid operator 'not' for type %d", operandType)
					}
				} else if op == 3 { // ~
					if !operandType.IsInteger() {
						tc.errorf(nodeIdx, 3022, "invalid operator '~' for type %d", operandType)
					}
				}
			}
		}

	case ast.NodeIndexExpr:
		collection := node.FirstChild
		if collection != 0 {
			tc.checkStmt(collection)
			arg := tc.ast.Nodes[collection].NextSibling
			for arg != 0 {
				tc.checkStmt(arg)
				arg = tc.ast.Nodes[arg].NextSibling
			}
		}

		index := uint32(0)
		if collection != 0 {
			index = tc.ast.Nodes[collection].NextSibling
		}

		if collection != 0 && index != 0 {

			colType := tc.infer.TypeOf(collection)
			idxType := tc.infer.TypeOf(index)

			if colType != types.TypeUnknown {
				entry := tc.types.Entry(colType)
				if entry.Kind != types.KindArray && entry.Kind != types.KindSlice && entry.Kind != types.KindPointer && colType != types.TypeString {
					// Hack to ignore generic function instantiation
					if entry.Kind == types.KindFunction || colType == types.TypeUnknown {
						// skip
					} else {
						// Is collection an Ident to a generic template?
						isGeneric := false
						if tc.ast.Nodes[collection].Kind == ast.NodeIdent || tc.ast.Nodes[collection].Kind == ast.NodeFieldExpr {
							symIdx := tc.ast.Nodes[collection].Payload
							if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
								if tc.symtable.SymbolAt(symIdx).Flags&SymFlagGeneric != 0 {
									isGeneric = true
								}
							}
						}
						if !isGeneric {
							tc.errorf(nodeIdx, 3023, "cannot index into non-array/slice type %d", colType)
						}
					}
				}
			}

			if idxType != types.TypeUnknown {
				isGenericApp := false
				if colType != types.TypeUnknown {
					entry := tc.types.Entry(colType)
					if entry.Kind == types.KindFunction {
						isGenericApp = true
					}
				}

				if tc.ast.Nodes[collection].Kind == ast.NodeIdent || tc.ast.Nodes[collection].Kind == ast.NodeFieldExpr {
					name := string(tc.ast.NodeText(collection))
					if name == "compiler_intrinsic" || name == "@compiler_intrinsic" {
						isGenericApp = true
					}
					symIdx := tc.ast.Nodes[collection].Payload
					if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
						if tc.symtable.SymbolAt(symIdx).Flags&SymFlagGeneric != 0 {
							isGenericApp = true
						}
					}
				} else if tc.ast.Nodes[collection].Kind == ast.NodeCallExpr {
					// e.g. @compiler_intrinsic("size_of")
					callee := tc.ast.Nodes[collection].FirstChild
					if callee != 0 {
						calleeText := string(tc.ast.NodeText(callee))
						if calleeText == "compiler_intrinsic" || calleeText == "@compiler_intrinsic" {
							isGenericApp = true
						}
					}
				}

				if !isGenericApp && !idxType.IsInteger() {
					tc.errorf(nodeIdx, 3024, "index must be integer, found %d", idxType)
				}
			}
		}

	case ast.NodeFieldExpr:
		symIdx := node.Payload
		if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
			sym := tc.symtable.SymbolAt(symIdx)
			if sym.Kind == SymModule {
				break
			}
		}
		obj := node.FirstChild
		if obj != 0 {
			tc.checkStmt(obj)
			objType := tc.infer.TypeOf(obj)
			lhsIsModule := false
			if tc.ast.Nodes[obj].Kind == ast.NodeFieldExpr || tc.ast.Nodes[obj].Kind == ast.NodeIdent {
				symIdx := tc.ast.Nodes[obj].Payload
				if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
					sym := tc.symtable.SymbolAt(symIdx)
					if sym.Kind == SymModule {
						lhsIsModule = true
					}
				}
			}
			if !lhsIsModule {
				curr := obj
				for tc.ast.Nodes[curr].Kind == ast.NodeFieldExpr {
					curr = tc.ast.Nodes[curr].FirstChild
				}
				if tc.ast.Nodes[curr].Kind == ast.NodeIdent {
					symIdx := tc.ast.Nodes[curr].Payload
					if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
						sym := tc.symtable.SymbolAt(symIdx)
						if sym.Kind == SymModule {
							lhsIsModule = true
						}
					}
				}
			}
			if !lhsIsModule && tc.parents != nil {
				currParent := tc.parents[nodeIdx]
				for currParent != 0 && tc.ast.Nodes[currParent].Kind == ast.NodeFieldExpr {
					symIdx := tc.ast.Nodes[currParent].Payload
					if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
						sym := tc.symtable.SymbolAt(symIdx)
						if sym.Kind == SymModule {
							lhsIsModule = true
							break
						}
					}
					currParent = tc.parents[currParent]
				}
			}

			// Check for built-in .len property access on string, slice, array
			fieldName := ""
			fieldNodeIdx := node.FirstChild
			if fieldNodeIdx != 0 {
				fieldNodeIdx = tc.ast.Nodes[fieldNodeIdx].NextSibling
				if fieldNodeIdx != 0 && tc.ast.Nodes[fieldNodeIdx].Kind == ast.NodeIdent {
					fieldName = string(tc.ast.NodeText(fieldNodeIdx))
				}
			}
			if fieldName == "" && node.Payload != 0 {
				if int(node.Payload) <= tc.intern.Len() {
					fieldName = string(tc.intern.Get(uint32(node.Payload)))
				}
			}
			isLenAccess := false
			if (fieldName == "len" || fieldName == "ptr") && objType != types.TypeUnknown {
				actualObjType := objType
				entry := tc.types.Entry(actualObjType)
				if entry.Kind == types.KindPointer {
					actualObjType = tc.types.PointerElem(actualObjType)
					entry = tc.types.Entry(actualObjType)
				}
				if actualObjType == types.TypeString || entry.Kind == types.KindSlice || entry.Kind == types.KindArray {
					isLenAccess = true
				}
			}

			isStructAccess := false
			var structType types.TypeID = objType
			var entry *types.TypeEntry
			isResolvedSym := false
			symIdx := node.Payload
			if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
				sym := tc.symtable.SymbolAt(symIdx)
				symName := string(tc.symtable.intern.Get(sym.NameID))
				if symName == fieldName {
					if sym.Kind == SymFunc || sym.Kind == SymVar || sym.Kind == SymConst || sym.Kind == SymStruct || sym.Kind == SymInterface {
						isResolvedSym = true
					}
				}
			}

			if !lhsIsModule && objType != types.TypeUnknown {
				entry = tc.types.Entry(objType)
				fmt.Printf("[DEBUG-CHECK-FIELD] objType=%d entry.Kind=%d fieldName='%s' isResolvedSym=%v\n", objType, entry.Kind, fieldName, isResolvedSym)
				if entry.Kind == types.KindPointer {
					objType = tc.types.PointerElem(objType)
					entry = tc.types.Entry(objType)
				}
				structType = objType
				if entry.Kind == types.KindGenericInst {
					for idx := 0; idx < tc.types.Count(); idx++ {
						e := tc.types.Entry(types.TypeID(idx))
						var name1, name2 string
						if e.NameID != 0 {
							name1 = tc.symtable.intern.Get(e.NameID)
						}
						if entry.NameID != 0 {
							name2 = tc.symtable.intern.Get(entry.NameID)
						}
						if (e.Kind == types.KindStruct || e.Kind == types.KindSum || e.Kind == types.KindGenericInst) && name1 != "" && name1 == name2 {
							structType = types.TypeID(idx)
							entry = e
							break
						}
					}
				}
				if entry.Kind == types.KindStruct || entry.Kind == types.KindSum {
					if !isResolvedSym {
						isStructAccess = true
					}
				} else if entry.Kind == types.KindGeneric {
					constraintID := tc.types.GenericConstraint(objType)
					fmt.Printf("[DEBUG-CHECK-FIELD] generic constraintID=%d for objType=%d\n", constraintID, objType)
					if constraintID != 0 {
						ifaceInfo := tc.types.InterfaceInfo(constraintID)
						fieldNameID := tc.symtable.intern.InternString(fieldName)
						fmt.Printf("[DEBUG-CHECK-FIELD] looking for fieldNameID=%d in interface %d methods:\n", fieldNameID, constraintID)
						for _, method := range ifaceInfo.Methods {
							fmt.Printf("[DEBUG-CHECK-FIELD]   method.NameID=%d ('%s')\n", method.NameID, tc.symtable.intern.Get(method.NameID))
							if method.NameID == fieldNameID {
								isStructAccess = true
								break
							}
						}
					}
				}
				fmt.Printf("[DEBUG-CHECK-FIELD] final isStructAccess=%v\n", isStructAccess)
			}

			isBuiltinMethod := false
			if entry != nil && entry.Kind == types.KindGenericInst {
				name := string(tc.symtable.intern.Get(entry.NameID))
				if name == "Vec" && (fieldName == "len" || fieldName == "get" || fieldName == "push" || fieldName == "data" || fieldName == "destroy" || fieldName == "cap") {
					isBuiltinMethod = true
				} else if name == "Option" && (fieldName == "unwrap" || fieldName == "unwrap_or" || fieldName == "expect" || fieldName == "is_ok" || fieldName == "is_some" || fieldName == "is_none" || fieldName == "map" || fieldName == "flat_map" || fieldName == "ok_or" || fieldName == "filter" || fieldName == "or_other") {
					isBuiltinMethod = true
				} else if name == "Result" && (fieldName == "unwrap" || fieldName == "unwrap_err" || fieldName == "unwrap_or" || fieldName == "expect" || fieldName == "is_ok" || fieldName == "is_err" || fieldName == "map" || fieldName == "map_err" || fieldName == "flat_map" || fieldName == "ok" || fieldName == "err") {
					isBuiltinMethod = true
				}
			}
			if tc.ast.Nodes[obj].Kind == ast.NodeIdent || tc.ast.Nodes[obj].Kind == ast.NodeFieldExpr {
				objName := string(tc.ast.NodeText(obj))
				if (objName == "Vec" || strings.HasSuffix(objName, ".Vec") || objName == "Option" || strings.HasSuffix(objName, ".Option") || objName == "Result" || strings.HasSuffix(objName, ".Result")) && fieldName == "new" {
					isBuiltinMethod = true
				}
			} else if tc.ast.Nodes[obj].Kind == ast.NodeIndexExpr {
				col := tc.ast.Nodes[obj].FirstChild
				if col != 0 && (tc.ast.Nodes[col].Kind == ast.NodeIdent || tc.ast.Nodes[col].Kind == ast.NodeFieldExpr) {
					objName := string(tc.ast.NodeText(col))
					if (objName == "Vec" || strings.HasSuffix(objName, ".Vec") || objName == "Option" || strings.HasSuffix(objName, ".Option") || objName == "Result" || strings.HasSuffix(objName, ".Result")) && fieldName == "new" {
						isBuiltinMethod = true
					}
				}
			}

			if isLenAccess || isBuiltinMethod {
				// Valid built-in property access for .len and generic methods, skip error checks.
			} else if !isStructAccess {
				nodeIsModule := false
				if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
					sym := tc.symtable.SymbolAt(symIdx)
					if sym.Kind == SymModule {
						nodeIsModule = true
					}
				}
				if lhsIsModule || nodeIsModule {
					if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
						sym := tc.symtable.SymbolAt(symIdx)
						if sym.Kind == SymFunc || sym.Kind == SymVar || sym.Kind == SymConst || sym.Kind == SymModule {
							isResolvedSym = true
						}
					}
				}
				if !lhsIsModule && !isResolvedSym && objType != types.TypeUnknown {
					tc.errorf(nodeIdx, 3025, "cannot access field on non-struct type %d", objType)
				}
			} else {
				// Verify that the field or method exists on the struct or sum type
				found := false
				
				if entry.Kind == types.KindStruct {
					structInfo := tc.types.StructInfo(structType)
					for _, field := range structInfo.Fields {
						fieldNameInStruct := tc.symtable.intern.Get(field.NameID)
						if fieldNameInStruct == fieldName {
							found = true
							break
						}
					}
				} else if entry.Kind == types.KindGeneric {
					constraintID := tc.types.GenericConstraint(objType)
					if constraintID != 0 {
						ifaceInfo := tc.types.InterfaceInfo(constraintID)
						fieldNameID := tc.symtable.intern.InternString(fieldName)
						for _, method := range ifaceInfo.Methods {
							if method.NameID == fieldNameID {
								found = true
								break
							}
						}
					}
				}
				
				if !found && tc.ifaces != nil {
					methods := tc.ifaces.getMethodsOfStruct(structType)
					for _, method := range methods {
						methodNameInStruct := tc.symtable.intern.Get(method.NameID)
						if methodNameInStruct == fieldName {
							found = true
							break
						}
					}
				}

				// Also fall back to getMethodsOfStruct using the original objType
				if !found && tc.ifaces != nil && objType != structType {
					methods := tc.ifaces.getMethodsOfStruct(objType)
					for _, method := range methods {
						methodNameInStruct := tc.symtable.intern.Get(method.NameID)
						if methodNameInStruct == fieldName {
							found = true
							break
						}
					}
				}

				if !found {
					tc.errorf(nodeIdx, 3026, "type %d has no field or method '%s'", objType, fieldName)
				}
			}
		}

	case ast.NodeCastExpr:
		expr := node.FirstChild
		if expr != 0 {
			tc.checkStmt(expr)
			exprType := tc.infer.TypeOf(expr)
			
			// In AST, the target type is usually stored in Payload or as a TypeRef child.
			// Let's assume Payload has the TypeID (this happens after NameResolver resolves TypeRef).
			// If we mock the TypeRef payload to be the TypeID:
			targetType := types.TypeID(node.Payload)
			
			if exprType != types.TypeUnknown && targetType != 0 && targetType != types.TypeUnknown {
				// Legal casts: numeric<->numeric, bool<->int
				valid := false
				
				if (exprType.IsInteger() || exprType.IsFloat() || exprType == types.TypeChar8) && (targetType.IsInteger() || targetType.IsFloat() || targetType == types.TypeChar8) {
					valid = true
				} else if (exprType.IsInteger() || exprType == types.TypeChar8) && targetType == types.TypeBool {
					valid = true
				} else if exprType == types.TypeBool && (targetType.IsInteger() || targetType == types.TypeChar8) {
					valid = true
				}
				exprEntry := tc.types.Entry(exprType)
				targetEntry := tc.types.Entry(targetType)
				if exprEntry.Kind == types.KindPointer && targetEntry.Kind == types.KindPointer {
					valid = true
				} else if exprEntry.Kind == types.KindPointer && targetType.IsInteger() {
					valid = true
				} else if exprType.IsInteger() && targetEntry.Kind == types.KindPointer {
					valid = true
				} else if exprEntry.Kind == types.KindPointer && targetType == types.TypeString {
					valid = true
				} else if exprType == types.TypeString && targetEntry.Kind == types.KindPointer {
					valid = true
				}
				
				if !valid {
					fmt.Printf("[DEBUG] ILLEGAL CAST: exprType=%d (Kind=%d), targetType=%d (Kind=%d)\n", exprType, exprEntry.Kind, targetType, targetEntry.Kind)
					tc.errorf(nodeIdx, 3026, "illegal cast from %d to %d", exprType, targetType)
				}
			}
		}

	case ast.NodeAwaitExpr:
		if !tc.inAsyncFn {
			tc.errorf(nodeIdx, 3011, "await can only be used inside async functions")
		}
		expr := node.FirstChild
		if expr != 0 {
			tc.checkStmt(expr)
			exprType := tc.infer.TypeOf(expr)
			
			if exprType != types.TypeUnknown {
				isFuture, _ := IsFutureType(tc.types, exprType)
				if !isFuture {
					tc.errorf(nodeIdx, 3010, "await requires Future[T], found %d", exprType)
				}
			}
		}

	case ast.NodeSpawnExpr:
		expr := node.FirstChild
		if expr != 0 {
			tc.checkStmt(expr)
			// we could verify that it's a function call, but inference already assigned ActorRef
		}

	default:
		child := node.FirstChild
		for child != 0 {
			tc.checkStmt(child)
			child = tc.ast.Nodes[child].NextSibling
		}
	}
}
