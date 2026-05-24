package sema

import (
	"fmt"

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
		index := uint32(0)
		if collection != 0 {
			index = tc.ast.Nodes[collection].NextSibling
		}

		if collection != 0 && index != 0 {
			tc.checkStmt(collection)
			tc.checkStmt(index)

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
						if tc.ast.Nodes[collection].Kind == ast.NodeIdent {
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

				if tc.ast.Nodes[collection].Kind == ast.NodeIdent {
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
		obj := node.FirstChild
		if obj != 0 {
			tc.checkStmt(obj)
			objType := tc.infer.TypeOf(obj)
			fmt.Printf("[DEBUG-FIELD] nodeIdx=%d LHS=%d LHS-Type=%d\n", nodeIdx, obj, objType)
			
			lhsIsModule := false
			curr := obj
			for tc.ast.Nodes[curr].Kind == ast.NodeFieldExpr {
				curr = tc.ast.Nodes[curr].FirstChild
			}
			if tc.ast.Nodes[curr].Kind == ast.NodeIdent {
				name := string(tc.ast.NodeText(curr))
				nameID := tc.intern.InternString(name)
				symIdx, found := tc.symtable.Resolve(nameID)
				if found && int(symIdx) < len(tc.symtable.Symbols) {
					sym := tc.symtable.SymbolAt(symIdx)
					if sym.Kind == SymModule {
						lhsIsModule = true
					}
				}
			}

			isStructAccess := false
			var structType types.TypeID = objType
			var entry *types.TypeEntry
			if !lhsIsModule && objType != types.TypeUnknown {
				entry = tc.types.Entry(objType)
				if entry.Kind == types.KindPointer {
					objType = tc.types.PointerElem(objType)
					entry = tc.types.Entry(objType)
				}
				structType = objType
				if entry.Kind == types.KindGenericInst {
					for idx := 0; idx < tc.types.Count(); idx++ {
						e := tc.types.Entry(types.TypeID(idx))
						if (e.Kind == types.KindStruct || e.Kind == types.KindSum) && e.NameID == entry.NameID {
							structType = types.TypeID(idx)
							entry = e
							break
						}
					}
				}
				if entry.Kind == types.KindStruct || entry.Kind == types.KindSum {
					isStructAccess = true
				}
			}

			if !isStructAccess {
				isResolvedSym := false
				symIdx := node.Payload
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
				if !isResolvedSym && objType != types.TypeUnknown {
					tc.errorf(nodeIdx, 3025, "cannot access field on non-struct type %d", objType)
				}
			} else {
				// Verify that the field or method exists on the struct or sum type
				found := false
				fieldNameID := uint32(node.Payload)
				
				if entry.Kind == types.KindStruct {
					structInfo := tc.types.StructInfo(structType)
					for _, field := range structInfo.Fields {
						if field.NameID == fieldNameID {
							found = true
							break
						}
					}
				}
				
				if !found && tc.ifaces != nil {
					methods := tc.ifaces.getMethodsOfStruct(structType)
					for _, method := range methods {
						if method.NameID == fieldNameID {
							found = true
							break
						}
					}
				}

				// Also fall back to getMethodsOfStruct using the original objType
				if !found && tc.ifaces != nil && objType != structType {
					methods := tc.ifaces.getMethodsOfStruct(objType)
					for _, method := range methods {
						if method.NameID == fieldNameID {
							found = true
							break
						}
					}
				}

				if !found {
					fieldName := ""
					if fieldNameID != 0 {
						fieldName = tc.intern.Get(fieldNameID)
					}
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
