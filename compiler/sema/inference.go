package sema

import (
	"fmt"
	"strings"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/types"
)

// InferenceEngine performs local Hindley-Milner type inference on the AST.
type InferenceEngine struct {
	ast             *ast.AstTree
	symtable        *SymbolTable
	types           *types.TypeTable
	mono            *Monomorphizer
	nodeTypes       map[uint32]types.TypeID
	errors          []diagnostics.Diagnostic
	ifaces          *Interfaces
	structsInferred map[types.TypeID]bool
	
	// Track current function's return type for 'return' statement checking
	currentReturn types.TypeID
}

// NewInferenceEngine creates a new InferenceEngine.
func NewInferenceEngine(tree *ast.AstTree, st *SymbolTable, tt *types.TypeTable, mono *Monomorphizer) *InferenceEngine {
	return &InferenceEngine{
		ast:             tree,
		symtable:        st,
		types:           tt,
		mono:            mono,
		nodeTypes:       make(map[uint32]types.TypeID),
		ifaces:          NewInterfaces(st, tt),
		structsInferred: make(map[types.TypeID]bool),
	}
}

// errorf appends a type error diagnostic.
func (ie *InferenceEngine) errorf(nodeIdx uint32, code int, format string, args ...any) {
	ie.errors = append(ie.errors, diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     uint32(code),
		Message:  fmt.Sprintf(format, args...),
		Pos:      nodePos(ie.ast, nodeIdx),
	})
}

// TypeOf returns the inferred type for a given node.
func (ie *InferenceEngine) TypeOf(nodeIdx uint32) types.TypeID {
	if typ, ok := ie.nodeTypes[nodeIdx]; ok {
		return typ
	}
	return ie.inferNode(nodeIdx, types.TypeUnknown)
}

// preInferFuncSignature infers only the parameter types and return type of a function.
func (ie *InferenceEngine) preInferFuncSignature(nodeIdx uint32) {
	node := &ie.ast.Nodes[nodeIdx]
	var paramTypes []types.TypeID
	var retType types.TypeID = types.TypeUnknown

	symIdx := node.Payload
	hasExistingType := false
	if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
		sym := ie.symtable.SymbolAt(symIdx)
		if sym.TypeID != 0 && sym.TypeID != uint32(types.TypeUnknown) {
			existingTypeID := types.TypeID(sym.TypeID)
			entry := ie.types.Entry(existingTypeID)
			if entry.Kind == types.KindFunction {
				hasExistingType = true
			}
		}
	}

	if !hasExistingType {
		// Walk children once to infer parameter and return types
		hasReturnTypeExpr := false
		child := node.FirstChild
		for child != 0 {
			childNode := &ie.ast.Nodes[child]
			if childNode.Kind == ast.NodeParamDecl {
				pType := ie.inferNode(child, types.TypeUnknown)
				paramTypes = append(paramTypes, pType)
			} else if childNode.Kind == ast.NodeTypeExpr || childNode.Kind == ast.NodeGenericType {
				retType = ie.inferNode(child, types.TypeUnknown)
				hasReturnTypeExpr = true
			}
			child = childNode.NextSibling
		}

		if !hasReturnTypeExpr {
			isMain := false
			if node.TokenIdx+1 < uint32(len(ie.ast.Tokens)) {
				if string(ie.ast.TokenText(node.TokenIdx+1)) == "main" {
					isMain = true
				}
			}
			if isMain {
				retType = types.TypeI32
			} else {
				retType = types.TypeVoid
			}
		}

		// Register the function type in the TypeTable
		funcTypeID := ie.types.RegisterFunction(paramTypes, retType, nil)
		if node.Flags&uint16(ast.FlagIsAsync) != 0 {
			ie.types.FuncInfo(funcTypeID).IsAsync = true
		}

		// Store type on the function's symbol
		if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
			sym := ie.symtable.SymbolAt(symIdx)
			sym.TypeID = uint32(funcTypeID)
		}
	}
}

// Infer walks the AST and infers types for all nodes.
func (ie *InferenceEngine) Infer() []diagnostics.Diagnostic {
	if ie.ast == nil || ie.ast.NodeCount() == 0 {
		return ie.errors
	}

	// Phase 0.5: Pre-register all empty struct symbols so that self-referential/cyclic references can resolve their pointer types
	for i := 0; i < len(ie.ast.Nodes); i++ {
		node := &ie.ast.Nodes[i]
		if node.Kind == ast.NodeStructDecl {
			symIdx := node.Payload
			if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
				sym := ie.symtable.SymbolAt(symIdx)
				if sym.TypeID == 0 || sym.TypeID == uint32(types.TypeUnknown) {
					nameID := sym.NameID
					gps := ie.getGenericParamTypeIDs(uint32(i))
					var gpUints []uint32
					if len(gps) > 0 {
						gpUints = make([]uint32, len(gps))
						for idx, gp := range gps {
							gpUints[idx] = uint32(gp)
						}
					}
					typeID := ie.types.RegisterStruct(nameID, nil, gpUints)
					sym.TypeID = uint32(typeID)
				}
			}
		}
	}

	// Phase 1: Pre-infer all struct and type alias definitions
	for i := 0; i < len(ie.ast.Nodes); i++ {
		node := &ie.ast.Nodes[i]
		if node.Kind == ast.NodeStructDecl || node.Kind == ast.NodeTypeAliasDecl {
			ie.inferNode(uint32(i), types.TypeUnknown)
		}
	}

	// Phase 2: Pre-infer all function signatures
	for i := 0; i < len(ie.ast.Nodes); i++ {
		node := &ie.ast.Nodes[i]
		if node.Kind == ast.NodeFuncDecl {
			ie.preInferFuncSignature(uint32(i))
		}
	}

	// Phase 3: Perform normal type inference
	ie.inferNode(0, types.TypeUnknown)
	return ie.errors
}

// inferNode infers the type of a node, optionally using an expected type for bidirectional inference.
func (ie *InferenceEngine) inferNode(nodeIdx uint32, expected types.TypeID) types.TypeID {
	node := &ie.ast.Nodes[nodeIdx]
	var resultType types.TypeID = types.TypeUnknown

	switch node.Kind {
	case ast.NodeProgram, ast.NodeBlock:
		child := node.FirstChild
		for child != 0 {
			ie.inferNode(child, types.TypeUnknown)
			child = ie.ast.Nodes[child].NextSibling
		}

	case ast.NodeFuncDecl:
		var paramTypes []types.TypeID
		var retType types.TypeID = types.TypeUnknown
		
		symIdx := node.Payload
		hasExistingType := false
		if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
			sym := ie.symtable.SymbolAt(symIdx)
			if sym.TypeID != 0 && sym.TypeID != uint32(types.TypeUnknown) {
				existingTypeID := types.TypeID(sym.TypeID)
				entry := ie.types.Entry(existingTypeID)
				if entry.Kind == types.KindFunction {
					funcInfo := ie.types.FuncInfo(existingTypeID)
					paramTypes = funcInfo.Params
					retType = funcInfo.Return
					hasExistingType = true
				}
			}
		}

		if !hasExistingType {
			// Walk children once to infer parameter and return types
			hasReturnTypeExpr := false
			child := node.FirstChild
			for child != 0 {
				childNode := &ie.ast.Nodes[child]
				if childNode.Kind == ast.NodeParamDecl {
					pType := ie.inferNode(child, types.TypeUnknown)
					paramTypes = append(paramTypes, pType)
				} else if childNode.Kind == ast.NodeTypeExpr || childNode.Kind == ast.NodeGenericType {
					retType = ie.inferNode(child, types.TypeUnknown)
					hasReturnTypeExpr = true
				}
				child = childNode.NextSibling
			}
			
			if !hasReturnTypeExpr {
				isMain := false
				if node.TokenIdx+1 < uint32(len(ie.ast.Tokens)) {
					if string(ie.ast.TokenText(node.TokenIdx+1)) == "main" {
						isMain = true
					}
				}
				if isMain {
					retType = types.TypeI32
				} else {
					retType = types.TypeVoid
				}
			}
			
			// Register the function type in the TypeTable
			funcTypeID := ie.types.RegisterFunction(paramTypes, retType, nil)
			if node.Flags&uint16(ast.FlagIsAsync) != 0 {
				ie.types.FuncInfo(funcTypeID).IsAsync = true
			}
			
			// Store type on the function's symbol
			if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
				sym := ie.symtable.SymbolAt(symIdx)
				sym.TypeID = uint32(funcTypeID)
			}
			resultType = funcTypeID
		} else {
			resultType = types.TypeID(ie.symtable.SymbolAt(symIdx).TypeID)
		}

		prevReturn := ie.currentReturn
		ie.currentReturn = retType

		// Infer function body and other children
		child := node.FirstChild
		for child != 0 {
			childNode := &ie.ast.Nodes[child]
			if childNode.Kind != ast.NodeParamDecl && childNode.Kind != ast.NodeTypeExpr {
				ie.inferNode(child, types.TypeUnknown)
			}
			child = childNode.NextSibling
		}

		ie.currentReturn = prevReturn

	case ast.NodeStructDecl:
		symIdx := node.Payload
		if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
			sym := ie.symtable.SymbolAt(symIdx)
			typeID := types.TypeID(sym.TypeID)
			if typeID != 0 && typeID != types.TypeUnknown {
				if !ie.structsInferred[typeID] {
					ie.structsInferred[typeID] = true
					var fields []types.FieldEntry
					child := node.FirstChild
					for child != 0 {
						childNode := &ie.ast.Nodes[child]
						if childNode.Kind == ast.NodeFieldDecl {
							fSymIdx := childNode.Payload
							if fSymIdx != 0 && int(fSymIdx) < len(ie.symtable.Symbols) {
								fSym := ie.symtable.SymbolAt(fSymIdx)
								var fieldType types.TypeID = types.TypeUnknown
								fTypeNode := childNode.FirstChild
								if fTypeNode != 0 {
									fieldType = ie.inferNode(fTypeNode, types.TypeUnknown)
								}
								fSym.TypeID = uint32(fieldType)
								fields = append(fields, types.FieldEntry{
									NameID: fSym.NameID,
									TypeID: fieldType,
								})
								fmt.Printf("[DEBUG] StructDecl typeID=%d append field NameID=%d TypeID=%d\n", typeID, fSym.NameID, fieldType)
							}
						}
						child = childNode.NextSibling
					}
					// Update the existing registered struct's fields directly!
					sInfo := ie.types.StructInfo(typeID)
					sInfo.Fields = fields
					gps := ie.getGenericParamTypeIDs(nodeIdx)
					if len(gps) > 0 {
						gpUints := make([]uint32, len(gps))
						for idx, gp := range gps {
							gpUints[idx] = uint32(gp)
						}
						sInfo.GenericParams = gpUints
					}
				}
				resultType = typeID
			}
		}

		child := node.FirstChild
		for child != 0 {
			childNode := &ie.ast.Nodes[child]
			if childNode.Kind != ast.NodeFieldDecl {
				ie.inferNode(child, types.TypeUnknown)
			}
			child = childNode.NextSibling
		}

	case ast.NodeTypeAliasDecl:
		symIdx := node.Payload
		var sumNode uint32
		child := node.FirstChild
		for child != 0 {
			if ie.ast.Nodes[child].Kind == ast.NodeSumType {
				sumNode = child
				break
			}
			child = ie.ast.Nodes[child].NextSibling
		}
		if sumNode != 0 {
			if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
				sym := ie.symtable.SymbolAt(symIdx)
				if sym.TypeID == 0 || sym.TypeID == uint32(types.TypeUnknown) {
					var variants []types.VariantInfo
					var tag uint8 = 0

					vNode := ie.ast.Nodes[sumNode].FirstChild
					for vNode != 0 {
						if ie.ast.Nodes[vNode].Kind == ast.NodeVariantDecl {
							vSymIdx := ie.ast.Nodes[vNode].Payload
							vSym := ie.symtable.SymbolAt(vSymIdx)
							
							var payloadType types.TypeID = 0
							typeExprNode := ie.ast.Nodes[vNode].FirstChild
							if typeExprNode != 0 {
								payloadType = ie.inferNode(typeExprNode, types.TypeUnknown)
							}
							
							variants = append(variants, types.VariantInfo{
								NameID:      vSym.NameID,
								PayloadType: payloadType,
								Tag:         tag,
							})
							tag++
						}
						vNode = ie.ast.Nodes[vNode].NextSibling
					}

					var gpUints []uint32
					gps := ie.getGenericParamTypeIDs(nodeIdx)
					if len(gps) > 0 {
						gpUints = make([]uint32, len(gps))
						for idx, gp := range gps {
							gpUints[idx] = uint32(gp)
						}
					}
					typeID := ie.types.RegisterSumType(sym.NameID, variants, gpUints)
					sym.TypeID = uint32(typeID)

					// Assign typeID to all variant symbols
					vNode = ie.ast.Nodes[sumNode].FirstChild
					for vNode != 0 {
						if ie.ast.Nodes[vNode].Kind == ast.NodeVariantDecl {
							vSymIdx := ie.ast.Nodes[vNode].Payload
							vSym := ie.symtable.SymbolAt(vSymIdx)
							vSym.TypeID = uint32(typeID)
						}
						vNode = ie.ast.Nodes[vNode].NextSibling
					}
					resultType = typeID
				} else {
					resultType = types.TypeID(sym.TypeID)
				}
			}
		}

	case ast.NodeVariantDecl:
		symIdx := node.Payload
		if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
			sym := ie.symtable.SymbolAt(symIdx)
			resultType = types.TypeID(sym.TypeID)
		}

	case ast.NodeParamDecl:
		symIdx := node.Payload
		var paramType types.TypeID = types.TypeUnknown
		typeNode := node.FirstChild
		if typeNode != 0 {
			paramType = ie.inferNode(typeNode, types.TypeUnknown)
		}
		if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
			sym := ie.symtable.SymbolAt(symIdx)
			sym.TypeID = uint32(paramType)
		}
		resultType = paramType

	case ast.NodeConstDecl:
		// ConstDecl children: optional type annotation, then init expr.
		var expectedType types.TypeID = types.TypeUnknown
		
		symIdx := node.Payload
		var sym *Symbol
		if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
			sym = ie.symtable.SymbolAt(symIdx)
			expectedType = types.TypeID(sym.TypeID)
		}

		var typeNode, initExpr uint32
		child := node.FirstChild
		for child != 0 {
			if ie.ast.Nodes[child].Kind == ast.NodeTypeExpr || ie.ast.Nodes[child].Kind == ast.NodeGenericType {
				typeNode = child
			} else {
				initExpr = child
			}
			child = ie.ast.Nodes[child].NextSibling
		}

		if typeNode != 0 {
			expectedType = ie.inferNode(typeNode, types.TypeUnknown)
			if sym != nil {
				sym.TypeID = uint32(expectedType)
			}
		}

		if initExpr != 0 {
			inferred := ie.inferNode(initExpr, expectedType)
			
			if expectedType != types.TypeUnknown && expectedType != 0 {
				if !ie.isAssignableTo(inferred, expectedType) {
					ie.errorf(nodeIdx, 3001, "type mismatch: expected %d, found %d", expectedType, inferred)
				}
				resultType = expectedType
			} else {
				resultType = inferred
				expectedType = inferred
				if sym != nil {
					sym.TypeID = uint32(inferred)
				}
			}
		} else {
			resultType = expectedType
		}

	case ast.NodeVarDecl:
		// VarDecl children: optional type annotation, then init expr.
		// For simplification in AST: FirstChild is init expr if no type ref, etc.
		// Let's assume FirstChild is init expr.
		// node.Payload is the variable's NameID or SymbolIdx (if NameResolver ran).
		
		var expectedType types.TypeID = types.TypeUnknown
		
		symIdx := node.Payload
		var sym *Symbol
		if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
			sym = ie.symtable.SymbolAt(symIdx)
			expectedType = types.TypeID(sym.TypeID)
		}

		var typeNode, initExpr uint32
		child := node.FirstChild
		for child != 0 {
			if ie.ast.Nodes[child].Kind == ast.NodeTypeExpr || ie.ast.Nodes[child].Kind == ast.NodeGenericType {
				typeNode = child
			} else {
				initExpr = child
			}
			child = ie.ast.Nodes[child].NextSibling
		}

		if typeNode != 0 {
			expectedType = ie.inferNode(typeNode, types.TypeUnknown)
			if sym != nil {
				sym.TypeID = uint32(expectedType)
			}
		}

		if initExpr != 0 {
			inferred := ie.inferNode(initExpr, expectedType)
			
			if expectedType != types.TypeUnknown && expectedType != 0 {
				if !ie.isAssignableTo(inferred, expectedType) {
					ie.errorf(nodeIdx, 3001, "type mismatch: expected %d, found %d", expectedType, inferred)
				}
				resultType = expectedType
			} else {
				resultType = inferred
				expectedType = inferred
				if sym != nil {
					sym.TypeID = uint32(inferred)
				}
			}
		} else {
			resultType = expectedType
		}

	case ast.NodeReturnStmt:
		expr := node.FirstChild
		if expr != 0 {
			inferred := ie.inferNode(expr, ie.currentReturn)
			if ie.currentReturn != types.TypeUnknown && !ie.isAssignableTo(inferred, ie.currentReturn) {
				ie.errorf(nodeIdx, 3005, "return type mismatch: expected %d, found %d", ie.currentReturn, inferred)
			}
		} else if ie.currentReturn != types.TypeUnknown && ie.currentReturn != types.TypeVoid {
			ie.errorf(nodeIdx, 3005, "return type mismatch: expected %d, found void", ie.currentReturn)
		}

	case ast.NodeIfStmt:
		// Traverses cond, thenBranch, and all subsequent elif/else clauses.
		cond := node.FirstChild
		if cond != 0 {
			ie.inferNode(cond, types.TypeBool)
			
			thenBranch := ie.ast.Nodes[cond].NextSibling
			if thenBranch != 0 {
				thenType := ie.inferNode(thenBranch, expected)
				resultType = thenType
				
				sibling := ie.ast.Nodes[thenBranch].NextSibling
				for sibling != 0 {
					siblingType := ie.inferNode(sibling, expected)
					if resultType != types.TypeUnknown && siblingType != types.TypeUnknown {
						common, ok := ie.types.CommonType(resultType, siblingType)
						if !ok {
							ie.errorf(nodeIdx, 3004, "branches of if expression have incompatible types")
							resultType = types.TypeUnknown
						} else {
							resultType = common
						}
					}
					sibling = ie.ast.Nodes[sibling].NextSibling
				}
			}
		}

	case ast.NodeElifClause:
		cond := node.FirstChild
		if cond != 0 {
			ie.inferNode(cond, types.TypeBool)
			body := ie.ast.Nodes[cond].NextSibling
			if body != 0 {
				resultType = ie.inferNode(body, expected)
			}
		}

	case ast.NodeElseClause:
		body := node.FirstChild
		if body != 0 {
			resultType = ie.inferNode(body, expected)
		}

	case ast.NodeIndexExpr:
		collection := ie.ast.Nodes[nodeIdx].FirstChild
		if collection != 0 {
			if ie.ast.Nodes[collection].Kind == ast.NodeIdent {
				symIdx := ie.ast.Nodes[collection].Payload
				if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
					sym := ie.symtable.SymbolAt(symIdx)
					if sym.Flags & SymFlagGeneric != 0 {
						// Generic instantiation!
						var typeArgs []types.TypeID
						idx := ie.ast.Nodes[collection].NextSibling
						for idx != 0 {
							t := ie.inferNode(idx, types.TypeUnknown)
							typeArgs = append(typeArgs, t)
							idx = ie.ast.Nodes[idx].NextSibling
						}
						
						if ie.mono != nil {
							fmt.Printf("[DEBUG-INDEX] Instantiating generic collection: collectionName=%s typeArgs=%v (hasGenericParam=%t)\n", string(ie.ast.NodeText(collection)), typeArgs, hasGenericParam(ie.types, typeArgs))
							if hasGenericParam(ie.types, typeArgs) {
								if sym.Kind == SymStruct || sym.Kind == SymTypeAlias || sym.Kind == SymVariant {
									resultType = ie.types.RegisterGenericInst(sym.NameID, typeArgs)
								} else {
									resultType = types.TypeID(sym.TypeID)
								}
								break
							}
							instSymIdx, diags := ie.mono.InstantiateFunction(symIdx, typeArgs)
							ie.errors = append(ie.errors, diags...)
							
							ie.ast.SetPayload(nodeIdx, instSymIdx) // save instantiated symIdx
							instSym := ie.symtable.SymbolAt(instSymIdx)
							resultType = types.TypeID(instSym.TypeID)
							break
						}
					}
				}
			}
			
			// Normal indexing (fallback)
			collectionType := ie.inferNode(collection, types.TypeUnknown)
			idx := ie.ast.Nodes[collection].NextSibling
			if idx != 0 {
				ie.inferNode(idx, types.TypeUnknown)
			}
			if collectionType != types.TypeUnknown {
				cEntry := ie.types.Entry(collectionType)
				if cEntry.Kind == types.KindPointer {
					resultType = ie.types.PointerElem(collectionType)
				} else if cEntry.Kind == types.KindSlice {
					resultType = ie.types.SliceElem(collectionType)
				} else if cEntry.Kind == types.KindArray {
					resultType = ie.types.ArrayElem(collectionType)
				}
			}
		}

	case ast.NodeIdent:
		symIdx := node.Payload
		if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
			sym := ie.symtable.SymbolAt(symIdx)
			resultType = types.TypeID(sym.TypeID)
			if resultType != types.TypeUnknown {
				entry := ie.types.Entry(resultType)
				if entry.Kind == types.KindSum && expected != types.TypeUnknown {
					expEntry := ie.types.Entry(expected)
					if expEntry.Kind == types.KindSum || expEntry.Kind == types.KindGenericInst {
						resultType = expected
					}
				}
			}
		}

	case ast.NodeTypeExpr:
		symIdx := node.Payload
		if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
			sym := ie.symtable.SymbolAt(symIdx)
			resultType = types.TypeID(sym.TypeID)
		}
		
		child := node.FirstChild
		if child != 0 {
			innerType := ie.inferNode(child, types.TypeUnknown)
			if resultType == types.TypeUnknown {
				resultType = innerType
			}
		}

	case ast.NodePtrType:
		innerNodeIdx := node.FirstChild
		if innerNodeIdx != 0 {
			innerTypeID := ie.inferNode(innerNodeIdx, types.TypeUnknown)
			resultType = ie.types.RegisterPointer(innerTypeID)
		}

	case ast.NodeSliceType:
		innerNodeIdx := node.FirstChild
		if innerNodeIdx != 0 {
			innerTypeID := ie.inferNode(innerNodeIdx, types.TypeUnknown)
			resultType = ie.types.RegisterSlice(innerTypeID)
		}

	case ast.NodeArrayType:
		innerNodeIdx := node.FirstChild
		if innerNodeIdx != 0 {
			innerTypeID := ie.inferNode(innerNodeIdx, types.TypeUnknown)
			sizeNodeIdx := ie.ast.Nodes[innerNodeIdx].NextSibling
			var length uint32 = 1
			if sizeNodeIdx != 0 {
				sizeText := string(ie.ast.TokenText(ie.ast.Nodes[sizeNodeIdx].TokenIdx))
				var parsedLen uint64
				fmt.Sscanf(sizeText, "%d", &parsedLen)
				length = uint32(parsedLen)
			}
			resultType = ie.types.RegisterArray(innerTypeID, length)
		}

	case ast.NodeFieldExpr:
		obj := node.FirstChild
		isStructAccess := false
		if obj != 0 {
			objType := ie.inferNode(obj, types.TypeUnknown)
			fmt.Printf("[DEBUG-FIELD] FieldExpr nodeIdx=%d objType=%d fieldNameID=%d\n", nodeIdx, objType, uint32(node.Payload))
			if objType != types.TypeUnknown {
				entry := ie.types.Entry(objType)
				if entry.Kind == types.KindPointer {
					objType = ie.types.PointerElem(objType)
					entry = ie.types.Entry(objType)
				}
				var structType types.TypeID = objType
				var genericParams []uint32
				var genericArgs []types.TypeID
				if entry.Kind == types.KindGenericInst {
					genericArgs = ie.types.GenericInstArgs(objType)
					for idx := 0; idx < ie.types.Count(); idx++ {
						e := ie.types.Entry(types.TypeID(idx))
						if (e.Kind == types.KindStruct || e.Kind == types.KindSum) && e.NameID == entry.NameID {
							structType = types.TypeID(idx)
							entry = e
							break
						}
					}
				}
				fmt.Printf("[DEBUG-FIELD] objType resolved to entry.Kind=%v structType=%d NameID=%d genericArgs=%v\n", entry.Kind, structType, entry.NameID, genericArgs)
				if entry.Kind == types.KindStruct || entry.Kind == types.KindSum {
					isStructAccess = true
					var structInfo *types.StructType
					var sumInfo *types.SumType
					if entry.Kind == types.KindStruct {
						structInfo = ie.types.StructInfo(structType)
						genericParams = structInfo.GenericParams
					} else {
						sumInfo = ie.types.SumInfo(structType)
						genericParams = sumInfo.GenericParams
					}
					fieldNameID := uint32(node.Payload)
					found := false
					if entry.Kind == types.KindStruct {
						fmt.Printf("[DEBUG-FIELD] struct fields count=%d genericParams=%v\n", len(structInfo.Fields), genericParams)
						for _, field := range structInfo.Fields {
							fmt.Printf("[DEBUG-FIELD]   checking field NameID=%d field.TypeID=%d\n", field.NameID, field.TypeID)
							if field.NameID == fieldNameID {
								fieldType := field.TypeID
								if len(genericParams) > 0 && len(genericArgs) > 0 {
									fieldType = ie.types.SubstituteGenericType(fieldType, genericParams, genericArgs)
								}
								resultType = fieldType
								found = true
								fmt.Printf("[DEBUG-FIELD]   found field fieldType=%d resultType=%d\n", field.TypeID, resultType)
								break
							}
						}
					}
					if !found && ie.ifaces != nil {
						methods := ie.ifaces.getMethodsOfStruct(structType)
						for _, method := range methods {
							if method.NameID == fieldNameID {
								substitutedParams := make([]types.TypeID, len(method.Params))
								for i, pVal := range method.Params {
									substitutedParams[i] = ie.types.SubstituteGenericType(pVal, genericParams, genericArgs)
								}
								substitutedRet := ie.types.SubstituteGenericType(method.Return, genericParams, genericArgs)
								resultType = ie.types.RegisterFunction(substitutedParams, substitutedRet, nil)
								found = true
								break
							}
						}
						if !found && ie.mono != nil {
							structEntry := ie.types.Entry(objType)
							if (structEntry.Kind == types.KindStruct || structEntry.Kind == types.KindSum) && structEntry.NameID != 0 {
								structNameStr := string(ie.mono.intern.Get(structEntry.NameID))
								if strings.HasPrefix(structNameStr, "_AX_") {
									baseName, typeArgs := parseMangledName(ie.types, ie.mono.intern, structNameStr)
									if baseName != "" {
										if len(typeArgs) > 0 {
											methodName := string(ie.mono.intern.Get(fieldNameID))
											var foundTemplateSymIdx uint32 = 0
											for idx, sym := range ie.symtable.Symbols {
												if sym.Kind == SymFunc && sym.Flags&SymFlagGeneric != 0 {
													symName := string(ie.mono.intern.Get(sym.NameID))
													if symName == methodName {
														tID := types.TypeID(sym.TypeID)
														if tID != types.TypeUnknown && ie.types.Entry(tID).Kind == types.KindFunction {
															fInfo := ie.types.FuncInfo(tID)
															if len(fInfo.Params) > 0 {
																recParam := fInfo.Params[0]
																entry = ie.types.Entry(recParam)
																if entry.Kind == types.KindPointer {
																	recParam = ie.types.PointerElem(recParam)
																	entry = ie.types.Entry(recParam)
																} else if entry.Kind == types.KindRef {
																	recParam = types.TypeID(entry.Extra)
																	entry = ie.types.Entry(recParam)
																}
																recStructName := ""
																if entry.NameID != 0 {
																	recStructName = string(ie.mono.intern.Get(entry.NameID))
																}
																fmt.Printf("[DEBUG-OVERLOAD] methodName=%s baseName=%s recParam=%d kind=%v nameID=%d recStructName=%s\n", methodName, baseName, recParam, entry.Kind, entry.NameID, recStructName)
																if (entry.Kind == types.KindStruct || entry.Kind == types.KindSum || entry.Kind == types.KindGenericInst) && entry.NameID != 0 {
																	if recStructName == baseName {
																		foundTemplateSymIdx = uint32(idx)
																		break
																	}
																}
															}
														}
													}
												}
											}
											if foundTemplateSymIdx != 0 {
												if hasGenericParam(ie.types, typeArgs) {
													methSym := ie.symtable.SymbolAt(foundTemplateSymIdx)
													resultType = types.TypeID(methSym.TypeID)
													found = true
													break
												}
												fmt.Printf("[DEBUG] Instantiating generic method template: %s for typeArgs %v. structName=%s\n", methodName, typeArgs, structNameStr)
												instSymIdx, diags := ie.mono.InstantiateFunction(foundTemplateSymIdx, typeArgs)
												fmt.Printf("[DEBUG] Instantiated method %s: instSymIdx=%d, errs=%v\n", methodName, instSymIdx, diags)
												ie.errors = append(ie.errors, diags...)
												ie.ifaces.methodCache = make(map[types.TypeID][]types.MethodSig)
												methods = ie.ifaces.getMethodsOfStruct(objType)
												for _, method := range methods {
													if method.NameID == fieldNameID {
														resultType = ie.types.RegisterFunction(method.Params, method.Return, nil)
														found = true
														break
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}

		if !isStructAccess {
			// If NameResolver resolved this field expression to a module/global symbol:
			lhsIsModule := false
			lhsNode := &ie.ast.Nodes[obj]
			if lhsNode.Kind == ast.NodeIdent || lhsNode.Kind == ast.NodeFieldExpr {
				lhsSymIdx := lhsNode.Payload
				if lhsSymIdx != 0 && int(lhsSymIdx) < len(ie.symtable.Symbols) {
					lhsSym := ie.symtable.SymbolAt(lhsSymIdx)
					if lhsSym.Kind == SymModule {
						lhsIsModule = true
					}
				}
			}
			nodeIsModule := false
			symIdx := node.Payload
			if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
				sym := ie.symtable.SymbolAt(symIdx)
				if sym.Kind == SymModule {
					nodeIsModule = true
				}
			}
			if lhsIsModule || nodeIsModule {
				if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
					sym := ie.symtable.SymbolAt(symIdx)
					if sym.Kind == SymFunc || sym.Kind == SymVar || sym.Kind == SymConst {
						resultType = types.TypeID(sym.TypeID)
					}
				}
			}
		}

	case ast.NodeBinaryExpr:
		lhs := node.FirstChild
		rhs := uint32(0)
		if lhs != 0 {
			rhs = ie.ast.Nodes[lhs].NextSibling
		}

		if lhs != 0 && rhs != 0 {
			// Check operator from node (mocked via Flags for this simplification)
			// Flags: 1 = ==, 2 = and, 0 = add/numeric
			op := node.Flags
			
			if op == 1 { // ==
				t1 := ie.inferNode(lhs, types.TypeUnknown)
				t2 := ie.inferNode(rhs, types.TypeUnknown)
				if t1 != t2 && !ie.types.CanImplicitCast(t1, t2) && !ie.types.CanImplicitCast(t2, t1) && !(t1.IsNumeric() && t2.IsNumeric()) {
					// allowed to have mismatch if they are implicit castable, or both are numeric, else error
					ie.errorf(nodeIdx, 3001, "type mismatch: cannot compare %d and %d", t1, t2)
				}
				resultType = types.TypeBool
			} else if op == 2 { // and
				ie.inferNode(lhs, types.TypeBool)
				ie.inferNode(rhs, types.TypeBool)
				resultType = types.TypeBool
			} else { // arithmetic
				t1 := ie.inferNode(lhs, types.TypeUnknown)
				t2 := ie.inferNode(rhs, types.TypeUnknown)
				
				common, ok := ie.types.CommonType(t1, t2)
				if !ok {
					ie.errorf(nodeIdx, 3001, "type mismatch: expected compatible types, found %d and %d", t1, t2)
				}
				resultType = common
			}
		}

	case ast.NodeCallExpr:
		callee := node.FirstChild
		if callee != 0 {
			if ie.isCompilerIntrinsicSizeOf(callee) {
				resultType = types.TypeU64
				arg := ie.ast.Nodes[callee].NextSibling
				for arg != 0 {
					ie.inferNode(arg, types.TypeUnknown)
					arg = ie.ast.Nodes[arg].NextSibling
				}
				ie.nodeTypes[nodeIdx] = resultType
				return resultType
			}
			if resType, isIntrinsic := ie.inferCompilerIntrinsicCall(callee, expected); isIntrinsic {
				resultType = resType
				ie.nodeTypes[nodeIdx] = resultType
				return resultType
			}
			calleeTypeID := ie.inferNode(callee, types.TypeUnknown)
			if calleeTypeID != types.TypeUnknown {
				entry := ie.types.Entry(calleeTypeID)
				if entry.Kind == types.KindFunction {
					funcInfo := ie.types.FuncInfo(calleeTypeID)
					resultType = funcInfo.Return
					
					if funcInfo.IsAsync {
						resultType = CreateFutureType(ie.types, funcInfo.Return)
					}
					
					// Check args
					argCount := 0
					arg := ie.ast.Nodes[callee].NextSibling
					for arg != 0 {
						if argCount < len(funcInfo.Params) {
							paramType := funcInfo.Params[argCount]
							argType := ie.inferNode(arg, paramType)
							if !ie.isAssignableTo(argType, paramType) {
								ie.errorf(arg, 3001, "type mismatch: expected %d, found %d", paramType, argType)
							}
						} else {
							ie.inferNode(arg, types.TypeUnknown) // extra arg
						}
						argCount++
						arg = ie.ast.Nodes[arg].NextSibling
					}
					
					if argCount != len(funcInfo.Params) {
						ie.errorf(nodeIdx, 3003, "argument count mismatch: expected %d, got %d", len(funcInfo.Params), argCount)
					}
				} else if entry.Kind == types.KindStruct || entry.Kind == types.KindGenericInst || entry.Kind == types.KindSum {
					resultType = calleeTypeID
					if entry.Kind == types.KindSum && expected != types.TypeUnknown {
						expEntry := ie.types.Entry(expected)
						if expEntry.Kind == types.KindSum || expEntry.Kind == types.KindGenericInst {
							resultType = expected
						}
					}
					arg := ie.ast.Nodes[callee].NextSibling
					for arg != 0 {
						ie.inferNode(arg, types.TypeUnknown)
						arg = ie.ast.Nodes[arg].NextSibling
					}
				} else {
					// Fallback for non-callable known types (e.g. primitive call error)
					arg := ie.ast.Nodes[callee].NextSibling
					for arg != 0 {
						ie.inferNode(arg, types.TypeUnknown)
						arg = ie.ast.Nodes[arg].NextSibling
					}
				}
			} else {
				// Fallback for unknown callee types (e.g. alloc, free, memcpy calls)
				arg := ie.ast.Nodes[callee].NextSibling
				for arg != 0 {
					ie.inferNode(arg, types.TypeUnknown)
					arg = ie.ast.Nodes[arg].NextSibling
				}
			}
		}

	case ast.NodeAwaitExpr:
		expr := node.FirstChild
		if expr != 0 {
			exprType := ie.inferNode(expr, types.TypeUnknown)
			isFuture, innerType := IsFutureType(ie.types, exprType)
			if isFuture {
				resultType = innerType
			} else {
				ie.errorf(nodeIdx, 3010, "await requires Future[T], found %d", exprType)
			}
		}

	case ast.NodeComptime:
		child := node.FirstChild
		if child != 0 {
			// First, infer the child node's type
			childType := ie.inferNode(child, expected)
			if childType == types.TypeUnknown {
				resultType = types.TypeUnknown
				break
			}
			
			// Evaluate the child node
			ce := NewComptimeEvaluator(ie.ast, ie.symtable.intern, ie.symtable, ie.types)
			val, diag := ce.Eval(child)
			if diag != nil {
				ie.errors = append(ie.errors, *diag)
				resultType = types.TypeUnknown
				break
			}
			
			// Determine literal kind and format value
			var litKind ast.NodeKind
			var valStr string
			switch val.Kind {
			case types.TypeBool:
				litKind = ast.NodeBoolLit
				if val.BoolVal {
					valStr = "true"
				} else {
					valStr = "false"
				}
			case types.TypeString:
				litKind = ast.NodeStringLit
				valStr = `"` + val.StrVal + `"`
			case types.TypeF32, types.TypeF64:
				litKind = ast.NodeFloatLit
				valStr = fmt.Sprintf("%g", val.FloatVal)
			default:
				litKind = ast.NodeIntLit
				valStr = fmt.Sprintf("%d", val.IntVal)
			}
			
			// Append the new source bytes and register the new token
			newOffset := uint32(len(ie.ast.Source))
			ie.ast.Source = append(ie.ast.Source, []byte(valStr)...)
			
			newTokenIdx := uint32(len(ie.ast.Tokens))
			var tokKind lexer.TokenKind
			switch val.Kind {
			case types.TypeBool:
				if val.BoolVal {
					tokKind = lexer.TokenTrue
				} else {
					tokKind = lexer.TokenFalse
				}
			case types.TypeString:
				tokKind = lexer.TokenStringLit
			case types.TypeF32, types.TypeF64:
				tokKind = lexer.TokenFloatLit
			default:
				tokKind = lexer.TokenIntLit
			}
			
			ie.ast.Tokens = append(ie.ast.Tokens, lexer.Token{
				Kind:   tokKind,
				Offset: newOffset,
				Len:    uint16(len(valStr)),
			})
			
			// Mutate NodeComptime into the corresponding literal node
			node.Kind = litKind
			node.TokenIdx = newTokenIdx
			node.FirstChild = ast.NullIdx
			node.Payload = uint32(val.Kind) // Store TypeID in Payload
			
			resultType = val.Kind
		} else {
			resultType = types.TypeVoid
		}

	case ast.NodeSpawnExpr:
		expr := node.FirstChild
		if expr != 0 {
			ie.inferNode(expr, types.TypeUnknown)
			resultType = types.TypeActorRef
		}

	// Literals
	case ast.NodeIntLit:
		if expected.IsInteger() || expected.IsFloat() {
			resultType = expected
		} else {
			resultType = types.TypeI32
		}
	case ast.NodeFloatLit:
		resultType = types.TypeF64
	case ast.NodeStringLit:
		resultType = types.TypeString
	case ast.NodeBoolLit:
		resultType = types.TypeBool
	case ast.NodeNilLit:
		if expected != types.TypeUnknown {
			resultType = expected
		} else {
			ie.errorf(nodeIdx, 3002, "cannot infer type of 'nil' without context")
			resultType = types.TypeUnknown
		}

	case ast.NodeUnaryExpr:
		op := node.Flags
		operand := node.FirstChild
		if operand != 0 {
			operandType := ie.inferNode(operand, expected)
			fmt.Printf("[DEBUG] Inference NodeUnaryExpr: op=%d operandType=%d\n", op, operandType)
			if op == 2 { // not
				resultType = types.TypeBool
			} else if op == 4 { // & (address-of)
				resultType = ie.types.RegisterPointer(operandType)
			} else { // - or ~
				resultType = operandType
			}
			fmt.Printf("[DEBUG] Inference NodeUnaryExpr resultType=%d\n", resultType)
		}

	case ast.NodeCastExpr:
		expr := ie.ast.Nodes[nodeIdx].FirstChild
		var targetType types.TypeID = types.TypeUnknown
		if expr != 0 {
			exprKind := ie.ast.Nodes[expr].Kind
			exprType := ie.inferNode(expr, types.TypeUnknown)
			fmt.Printf("[DEBUG] CastExpr child expr nodeIdx=%d Kind=%s Type=%d\n", expr, exprKind, exprType)
			targetNodeIdx := ie.ast.Nodes[expr].NextSibling
			if targetNodeIdx != 0 {
				targetType = ie.inferNode(targetNodeIdx, types.TypeUnknown)
				targetNode := &ie.ast.Nodes[targetNodeIdx]
				fmt.Printf("[DEBUG] NodeCastExpr nodeIdx=%d targetNodeIdx=%d Kind=%s Payload=%d targetType=%d\n", nodeIdx, targetNodeIdx, targetNode.Kind, targetNode.Payload, targetType)
			}
		}
		if targetType != types.TypeUnknown {
			ie.ast.SetPayload(nodeIdx, uint32(targetType))
			resultType = targetType
		} else {
			resultType = types.TypeID(ie.ast.Nodes[nodeIdx].Payload)
		}

	case ast.NodeGenericType:
		baseNodeIdx := node.FirstChild
		if baseNodeIdx != 0 {
			baseNode := &ie.ast.Nodes[baseNodeIdx]
			name := string(ie.ast.TokenText(baseNode.TokenIdx))

			if name == "ptr" {
				argNodeIdx := baseNode.NextSibling
				if argNodeIdx != 0 {
					innerTypeID := ie.inferNode(argNodeIdx, types.TypeUnknown)
					resultType = ie.types.RegisterPointer(innerTypeID)
					fmt.Printf("[DEBUG] Inference NodeGenericType: base='%s' argNodeIdx=%d innerTypeID=%d resultType=%d\n", name, argNodeIdx, innerTypeID, resultType)
				}
			} else {
				baseSymIdx := baseNode.Payload
				if baseSymIdx != 0 && int(baseSymIdx) < len(ie.symtable.Symbols) {
					var typeArgs []types.TypeID
					argNodeIdx := baseNode.NextSibling
					for argNodeIdx != 0 {
						argType := ie.inferNode(argNodeIdx, types.TypeUnknown)
						typeArgs = append(typeArgs, argType)
						argNodeIdx = ie.ast.Nodes[argNodeIdx].NextSibling
					}
					if ie.mono != nil && len(typeArgs) > 0 {
						if hasGenericParam(ie.types, typeArgs) {
							sym := ie.symtable.SymbolAt(baseSymIdx)
							resultType = ie.types.RegisterGenericInst(sym.NameID, typeArgs)
							fmt.Printf("[DEBUG] Inference NodeGenericType (User-Defined Generic Template Ref): base='%s' typeArgs=%v instTypeID=%d\n", name, typeArgs, resultType)
						} else {
							instSymIdx, diags := ie.mono.InstantiateFunction(baseSymIdx, typeArgs)
							ie.errors = append(ie.errors, diags...)
							instSym := ie.symtable.SymbolAt(instSymIdx)
							resultType = types.TypeID(instSym.TypeID)
							fmt.Printf("[DEBUG] Inference NodeGenericType (User-Defined): base='%s' typeArgs=%v instTypeID=%d\n", name, typeArgs, resultType)
						}
					}
				}
			}
		}

	case ast.NodeDerefExpr:
		child := node.FirstChild
		if child != 0 {
			childType := ie.inferNode(child, types.TypeUnknown)
			if childType != types.TypeUnknown {
				entry := ie.types.Entry(childType)
				if entry.Kind == types.KindPointer {
					resultType = ie.types.PointerElem(childType)
				} else {
					resultType = childType
				}
			}
		}
		ie.ast.SetPayload(nodeIdx, uint32(resultType))

	default:
		// Recurse for anything else
		child := node.FirstChild
		for child != 0 {
			ie.inferNode(child, types.TypeUnknown)
			child = ie.ast.Nodes[child].NextSibling
		}
	}

	ie.nodeTypes[nodeIdx] = resultType
	return resultType
}

func (ie *InferenceEngine) isAssignableTo(from, to types.TypeID) bool {
	if from == to {
		return true
	}
	if ie.types.Entry(to).Kind == types.KindInterface {
		ok, _ := ie.ifaces.ImplementsInterface(from, to)
		return ok
	}
	return ie.types.IsAssignableTo(from, to)
}

func (ie *InferenceEngine) getGenericParamTypeIDs(templateNodeIdx uint32) []types.TypeID {
	var gpNodeIdx uint32
	child := ie.ast.Nodes[templateNodeIdx].FirstChild
	for child != 0 {
		if ie.ast.Nodes[child].Kind == ast.NodeGenericParams {
			gpNodeIdx = child
			break
		}
		child = ie.ast.Nodes[child].NextSibling
	}
	if gpNodeIdx == 0 {
		return nil
	}

	var paramTypeIDs []types.TypeID
	gpChild := ie.ast.Nodes[gpNodeIdx].FirstChild
	for gpChild != 0 {
		if ie.ast.Nodes[gpChild].Kind == ast.NodeGenericParam {
			found := false
			for i := 0; i < len(ie.symtable.Symbols); i++ {
				sym := ie.symtable.SymbolAt(uint32(i))
				if sym.Kind == SymGenericParam && sym.DeclNode == gpChild {
					paramTypeIDs = append(paramTypeIDs, types.TypeID(sym.TypeID))
					found = true
					break
				}
			}
			if !found {
				gpNameID := ie.ast.Nodes[gpChild].Payload
				typeID := ie.types.RegisterGenericType(gpNameID)
				paramTypeIDs = append(paramTypeIDs, typeID)
			}
		}
		gpChild = ie.ast.Nodes[gpChild].NextSibling
	}
	return paramTypeIDs
}

func parseMangledName(tt *types.TypeTable, intern *ast.InternPool, name string) (string, []types.TypeID) {
	if !strings.HasPrefix(name, "_AX_") {
		return "", nil
	}
	parts := strings.Split(name, "_")
	if len(parts) < 4 {
		return "", nil
	}
	baseName := parts[3]
	prefix := "_AX_" + parts[2] + "_" + parts[3] + "_"
	if len(name) <= len(prefix) {
		return baseName, nil
	}
	remainder := name[len(prefix):]

	var typeArgs []types.TypeID
	primitives := []string{"i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64", "f32", "f64", "bool", "string", "char8", "void", "isize", "usize"}

	for len(remainder) > 0 {
		if remainder[0] == '_' && !strings.HasPrefix(remainder, "_AX_") {
			remainder = remainder[1:]
			continue
		}

		// Find longest matching registered type name
		var bestMatch types.TypeID = types.TypeUnknown
		bestMatchLen := 0

		// 1. Check primitives first
		for _, prim := range primitives {
			if strings.HasPrefix(remainder, prim) {
				l := len(prim)
				if l == len(remainder) || remainder[l] == '_' {
					if l > bestMatchLen {
						bestMatch = typeFromName(tt, intern, prim)
						bestMatchLen = l
					}
				}
			}
		}

		// 2. Check registered types in TypeTable
		for idx := 0; idx < tt.Count(); idx++ {
			entry := tt.Entry(types.TypeID(idx))
			if entry.NameID != 0 {
				tName := string(intern.Get(entry.NameID))
				if tName != "" && strings.HasPrefix(remainder, tName) {
					l := len(tName)
					if l == len(remainder) || remainder[l] == '_' {
						if l > bestMatchLen {
							bestMatch = types.TypeID(idx)
							bestMatchLen = l
						}
					}
				}
			}
		}

		if bestMatch != types.TypeUnknown {
			typeArgs = append(typeArgs, bestMatch)
			remainder = remainder[bestMatchLen:]
		} else {
			// Fallback
			remainder = remainder[1:]
		}
	}

	return baseName, typeArgs
}

func typeFromName(tt *types.TypeTable, intern *ast.InternPool, name string) types.TypeID {
	switch name {
	case "i8": return types.TypeI8
	case "i16": return types.TypeI16
	case "i32": return types.TypeI32
	case "i64": return types.TypeI64
	case "u8": return types.TypeU8
	case "u16": return types.TypeU16
	case "u32": return types.TypeU32
	case "u64": return types.TypeU64
	case "f32": return types.TypeF32
	case "f64": return types.TypeF64
	case "bool": return types.TypeBool
	case "string": return types.TypeString
	case "char8": return types.TypeChar8
	case "void": return types.TypeVoid
	case "isize": return types.TypeISize
	case "usize": return types.TypeUSize
	}
	for idx := 0; idx < tt.Count(); idx++ {
		entry := tt.Entry(types.TypeID(idx))
		if entry.Kind == types.KindStruct && entry.NameID != 0 {
			structName := string(intern.Get(entry.NameID))
			if structName == name {
				return types.TypeID(idx)
			}
		}
	}
	return types.TypeUnknown
}

func (ie *InferenceEngine) isCompilerIntrinsicSizeOf(callee uint32) bool {
	node := ie.ast.Nodes[callee]
	if node.Kind != ast.NodeIndexExpr {
		return false
	}
	children := ie.ast.Children(callee)
	if len(children) < 2 {
		return false
	}
	innerCall := children[0]
	innerNode := ie.ast.Nodes[innerCall]
	if innerNode.Kind != ast.NodeCallExpr {
		return false
	}
	innerChildren := ie.ast.Children(innerCall)
	if len(innerChildren) < 2 {
		return false
	}
	identIdx := innerChildren[0]
	identNode := ie.ast.Nodes[identIdx]
	if identNode.Kind != ast.NodeIdent {
		return false
	}
	name := string(ie.ast.NodeText(identIdx))
	if name != "compiler_intrinsic" && name != "@compiler_intrinsic" {
		return false
	}
	argIdx := innerChildren[1]
	argNode := ie.ast.Nodes[argIdx]
	if argNode.Kind != ast.NodeStringLit {
		return false
	}
	argStr := string(ie.ast.NodeText(argIdx))
	argStr = strings.Trim(argStr, `"` + `'`)
	return argStr == "size_of"
}

func isGeneric(tt *types.TypeTable, tID types.TypeID) bool {
	if tID == types.TypeUnknown {
		return false
	}
	entry := tt.Entry(tID)
	if entry.Kind == types.KindGeneric {
		return true
	}
	if entry.Kind == types.KindPointer {
		return isGeneric(tt, tt.PointerElem(tID))
	}
	if entry.Kind == types.KindRef {
		return isGeneric(tt, types.TypeID(entry.Extra))
	}
	if entry.Kind == types.KindSlice {
		return isGeneric(tt, tt.SliceElem(tID))
	}
	if entry.Kind == types.KindGenericInst {
		for _, arg := range tt.GenericInstArgs(tID) {
			if isGeneric(tt, arg) {
				return true
			}
		}
	}
	return false
}

func hasGenericParam(tt *types.TypeTable, typeArgs []types.TypeID) bool {
	for _, arg := range typeArgs {
		if isGeneric(tt, arg) {
			return true
		}
	}
	return false
}

func (ie *InferenceEngine) inferCompilerIntrinsicCall(callee uint32, expected types.TypeID) (types.TypeID, bool) {
	node := ie.ast.Nodes[callee]
	if node.Kind != ast.NodeIdent {
		return types.TypeUnknown, false
	}
	name := string(ie.ast.NodeText(callee))
	if name != "compiler_intrinsic" && name != "@compiler_intrinsic" {
		return types.TypeUnknown, false
	}

	arg := node.NextSibling
	if arg == 0 {
		return expected, true
	}

	argNode := ie.ast.Nodes[arg]
	if argNode.Kind != ast.NodeStringLit {
		sib := arg
		for sib != 0 {
			ie.inferNode(sib, types.TypeUnknown)
			sib = ie.ast.Nodes[sib].NextSibling
		}
		return expected, true
	}

	argStr := string(ie.ast.NodeText(arg))
	argStr = strings.Trim(argStr, `"`+`'`)

	var resultType types.TypeID = types.TypeUnknown

	switch argStr {
	case "atomic_load":
		ptrArg := ie.ast.Nodes[arg].NextSibling
		if ptrArg != 0 {
			ptrType := ie.inferNode(ptrArg, types.TypeUnknown)
			if ptrType != types.TypeUnknown {
				entry := ie.types.Entry(ptrType)
				if entry.Kind == types.KindPointer {
					resultType = ie.types.PointerElem(ptrType)
				}
			}
		}
		if resultType == types.TypeUnknown {
			resultType = expected
		}

	case "atomic_store":
		ptrArg := ie.ast.Nodes[arg].NextSibling
		if ptrArg != 0 {
			ptrType := ie.inferNode(ptrArg, types.TypeUnknown)
			valArg := ie.ast.Nodes[ptrArg].NextSibling
			if valArg != 0 {
				expectedValType := types.TypeUnknown
				if ptrType != types.TypeUnknown {
					entry := ie.types.Entry(ptrType)
					if entry.Kind == types.KindPointer {
						expectedValType = ie.types.PointerElem(ptrType)
					}
				}
				ie.inferNode(valArg, expectedValType)
			}
		}
		resultType = types.TypeVoid

	case "atomic_swap":
		ptrArg := ie.ast.Nodes[arg].NextSibling
		if ptrArg != 0 {
			ptrType := ie.inferNode(ptrArg, types.TypeUnknown)
			if ptrType != types.TypeUnknown {
				entry := ie.types.Entry(ptrType)
				if entry.Kind == types.KindPointer {
					resultType = ie.types.PointerElem(ptrType)
				}
			}
			valArg := ie.ast.Nodes[ptrArg].NextSibling
			if valArg != 0 {
				ie.inferNode(valArg, resultType)
			}
		}
		if resultType == types.TypeUnknown {
			resultType = expected
		}

	case "atomic_cas":
		ptrArg := ie.ast.Nodes[arg].NextSibling
		if ptrArg != 0 {
			ptrType := ie.inferNode(ptrArg, types.TypeUnknown)
			expectedValType := types.TypeUnknown
			if ptrType != types.TypeUnknown {
				entry := ie.types.Entry(ptrType)
				if entry.Kind == types.KindPointer {
					expectedValType = ie.types.PointerElem(ptrType)
				}
			}
			expArg := ie.ast.Nodes[ptrArg].NextSibling
			if expArg != 0 {
				ie.inferNode(expArg, expectedValType)
				desArg := ie.ast.Nodes[expArg].NextSibling
				if desArg != 0 {
					ie.inferNode(desArg, expectedValType)
				}
			}
		}
		resultType = types.TypeBool

	default:
		sib := ie.ast.Nodes[arg].NextSibling
		for sib != 0 {
			ie.inferNode(sib, types.TypeUnknown)
			sib = ie.ast.Nodes[sib].NextSibling
		}
		resultType = expected
	}

	return resultType, true
}
