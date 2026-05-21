package sema

import (
	"fmt"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/types"
)

// InferenceEngine performs local Hindley-Milner type inference on the AST.
type InferenceEngine struct {
	ast       *ast.AstTree
	symtable  *SymbolTable
	types     *types.TypeTable
	mono      *Monomorphizer
	nodeTypes map[uint32]types.TypeID
	errors    []diagnostics.Diagnostic
	ifaces    *Interfaces
	
	// Track current function's return type for 'return' statement checking
	currentReturn types.TypeID
}

// NewInferenceEngine creates a new InferenceEngine.
func NewInferenceEngine(tree *ast.AstTree, st *SymbolTable, tt *types.TypeTable, mono *Monomorphizer) *InferenceEngine {
	return &InferenceEngine{
		ast:       tree,
		symtable:  st,
		types:     tt,
		mono:      mono,
		nodeTypes: make(map[uint32]types.TypeID),
		ifaces:    NewInterfaces(st, tt),
	}
}

// errorf appends a type error diagnostic.
func (ie *InferenceEngine) errorf(nodeIdx uint32, code int, format string, args ...any) {
	ie.errors = append(ie.errors, diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     uint32(code),
		Message:  fmt.Sprintf(format, args...),
		Pos:      diagnostics.Pos{}, // Mock pos for now
	})
}

// TypeOf returns the inferred type for a given node.
func (ie *InferenceEngine) TypeOf(nodeIdx uint32) types.TypeID {
	if typ, ok := ie.nodeTypes[nodeIdx]; ok {
		return typ
	}
	return ie.inferNode(nodeIdx, types.TypeUnknown)
}

// Infer walks the AST and infers types for all nodes.
func (ie *InferenceEngine) Infer() []diagnostics.Diagnostic {
	if ie.ast == nil || ie.ast.NodeCount() == 0 {
		return ie.errors
	}
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
				retType = types.TypeVoid
			}
			
			// Register the function type in the TypeTable
			funcTypeID := ie.types.RegisterFunction(paramTypes, retType, nil)
			
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
			if sym.TypeID == 0 || sym.TypeID == uint32(types.TypeUnknown) {
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
						}
					}
					child = childNode.NextSibling
				}

				nameID := sym.NameID
				typeID := ie.types.RegisterStruct(nameID, fields, nil)
				sym.TypeID = uint32(typeID)
				resultType = typeID
			} else {
				resultType = types.TypeID(sym.TypeID)
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
		// Very simplified if-stmt inference
		cond := node.FirstChild
		if cond != 0 {
			ie.inferNode(cond, types.TypeBool)
			
			thenBranch := ie.ast.Nodes[cond].NextSibling
			if thenBranch != 0 {
				thenType := ie.inferNode(thenBranch, expected)
				resultType = thenType
				
				elseBranch := ie.ast.Nodes[thenBranch].NextSibling
				if elseBranch != 0 {
					elseType := ie.inferNode(elseBranch, expected)
					common, ok := ie.types.CommonType(thenType, elseType)
					if !ok {
						ie.errorf(nodeIdx, 3004, "branches of if expression have incompatible types")
						resultType = types.TypeUnknown
					} else {
						resultType = common
					}
				}
			}
		}

	case ast.NodeIndexExpr:
		collection := node.FirstChild
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
							instSymIdx, diags := ie.mono.InstantiateFunction(symIdx, typeArgs)
							ie.errors = append(ie.errors, diags...)
							
							node.Payload = instSymIdx // save instantiated symIdx
							instSym := ie.symtable.SymbolAt(instSymIdx)
							resultType = types.TypeID(instSym.TypeID)
							break
						}
					}
				}
			}
			
			// Normal indexing (fallback)
			ie.inferNode(collection, types.TypeUnknown)
			idx := ie.ast.Nodes[collection].NextSibling
			if idx != 0 {
				ie.inferNode(idx, types.TypeUnknown)
			}
		}

	case ast.NodeIdent:
		symIdx := node.Payload
		if symIdx != 0 && int(symIdx) < len(ie.symtable.Symbols) {
			sym := ie.symtable.SymbolAt(symIdx)
			resultType = types.TypeID(sym.TypeID)
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

	case ast.NodeFieldExpr:
		obj := node.FirstChild
		if obj != 0 {
			objType := ie.inferNode(obj, types.TypeUnknown)
			if objType != types.TypeUnknown {
				entry := ie.types.Entry(objType)
				if entry.Kind == types.KindStruct {
					structInfo := ie.types.StructInfo(objType)
					fieldNameID := uint32(node.Payload)
					for _, field := range structInfo.Fields {
						if field.NameID == fieldNameID {
							resultType = field.TypeID
							break
						}
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
				if t1 != t2 && !ie.types.CanImplicitCast(t1, t2) && !ie.types.CanImplicitCast(t2, t1) {
					// allowed to have mismatch if they are implicit castable, else error
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
				} else if entry.Kind == types.KindStruct {
					resultType = calleeTypeID
					arg := ie.ast.Nodes[callee].NextSibling
					for arg != 0 {
						ie.inferNode(arg, types.TypeUnknown)
						arg = ie.ast.Nodes[arg].NextSibling
					}
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
