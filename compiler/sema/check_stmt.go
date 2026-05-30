package sema

import (
	"fmt"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/types"
)

// TypeChecker performs statement-level type checking on the AST.
type TypeChecker struct {
	ast      *ast.AstTree
	intern   *ast.InternPool
	symtable *SymbolTable
	types    *types.TypeTable
	infer    *InferenceEngine
	ifaces   *Interfaces
	errors   []diagnostics.Diagnostic
	
	currentFuncReturnType types.TypeID
	insideLoop            bool
	currentMatchScrutinee types.TypeID
	inAsyncFn             bool
	parents               []uint32
}

// Errors returns the list of diagnostic errors found during type checking.
func (tc *TypeChecker) Errors() []diagnostics.Diagnostic {
	return tc.errors
}

// NewTypeChecker creates a new TypeChecker.
func NewTypeChecker(tree *ast.AstTree, intern *ast.InternPool, st *SymbolTable, tt *types.TypeTable, ie *InferenceEngine) *TypeChecker {
	return &TypeChecker{
		ast:      tree,
		intern:   intern,
		symtable: st,
		types:    tt,
		infer:    ie,
		ifaces:   NewInterfaces(st, tt),
	}
}

// errorf appends a type error diagnostic.
func (tc *TypeChecker) errorf(nodeIdx uint32, code int, format string, args ...any) {
	tc.errors = append(tc.errors, diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     uint32(code),
		Message:  fmt.Sprintf(format, args...),
		Pos:      nodePos(tc.ast, nodeIdx),
	})
}

// warnf appends a type warning diagnostic.
func (tc *TypeChecker) warnf(nodeIdx uint32, code int, format string, args ...any) {
	tc.errors = append(tc.errors, diagnostics.Diagnostic{
		Severity: diagnostics.SeverityWarning,
		Code:     uint32(code),
		Message:  fmt.Sprintf(format, args...),
		Pos:      nodePos(tc.ast, nodeIdx),
	})
}

// Check walks the AST and performs statement-level type checking.
func (tc *TypeChecker) Check() []diagnostics.Diagnostic {
	if tc.ast == nil || tc.ast.NodeCount() == 0 {
		return tc.errors
	}
	tc.parents = make([]uint32, len(tc.ast.Nodes))
	for i := 0; i < len(tc.ast.Nodes); i++ {
		child := tc.ast.Nodes[i].FirstChild
		for child != 0 {
			tc.parents[child] = uint32(i)
			child = tc.ast.Nodes[child].NextSibling
		}
	}
	tc.checkStmt(0)
	return tc.errors
}

// CheckNode performs type checking starting from a specific node.
func (tc *TypeChecker) CheckNode(nodeIdx uint32) {
	if tc.parents == nil {
		tc.parents = make([]uint32, len(tc.ast.Nodes))
		for i := 0; i < len(tc.ast.Nodes); i++ {
			child := tc.ast.Nodes[i].FirstChild
			for child != 0 {
				tc.parents[child] = uint32(i)
				child = tc.ast.Nodes[child].NextSibling
			}
		}
	}
	tc.checkStmt(nodeIdx)
}

// checkStmt dispatches to specific statement checking logic based on NodeKind.
func (tc *TypeChecker) checkStmt(nodeIdx uint32) {
	if nodeIdx == 0 && tc.ast.Nodes[0].Kind != ast.NodeProgram {
		return
	}

	node := &tc.ast.Nodes[nodeIdx]

	switch node.Kind {
	case ast.NodeProgram, ast.NodeBlock:
		child := node.FirstChild
		for child != 0 {
			tc.checkStmt(child)
			child = tc.ast.Nodes[child].NextSibling
		}

	case ast.NodeStructDecl:
		symIdx := node.Payload
		if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
			sym := tc.symtable.SymbolAt(symIdx)
			if sym.TypeID == 0 { // Not yet registered
				// In a full implementation, we'd iterate over fields and gather types.
				// For now, we just register an empty struct type and assign it.
				nameID := sym.NameID
				typeID := tc.types.RegisterStruct(nameID, nil, nil)
				sym.TypeID = uint32(typeID)
			}
		}

		child := node.FirstChild
		for child != 0 {
			tc.checkStmt(child)
			child = tc.ast.Nodes[child].NextSibling
		}

	case ast.NodeTypeAliasDecl:
		symIdx := node.Payload
		var sumNode uint32
		child := node.FirstChild
		for child != 0 {
			if tc.ast.Nodes[child].Kind == ast.NodeSumType {
				sumNode = child
				break
			}
			child = tc.ast.Nodes[child].NextSibling
		}

		if sumNode != 0 {
			if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
				sym := tc.symtable.SymbolAt(symIdx)
				if sym.TypeID == 0 {
					var variants []types.VariantInfo
					var tag uint8 = 0

					vNode := tc.ast.Nodes[sumNode].FirstChild
					for vNode != 0 {
						if tc.ast.Nodes[vNode].Kind == ast.NodeVariantDecl {
							vSymIdx := tc.ast.Nodes[vNode].Payload
							vSym := tc.symtable.SymbolAt(vSymIdx)
							
							var payloadType types.TypeID = 0
							typeExprNode := tc.ast.Nodes[vNode].FirstChild
							if typeExprNode != 0 {
								payloadType = tc.infer.TypeOf(typeExprNode)
								// Wait, TypeOf(typeExprNode) evaluates the type expression. 
								// InferenceEngine should know how to evaluate NodeTypeExpr!
							}
							
							variants = append(variants, types.VariantInfo{
								NameID:      vSym.NameID,
								PayloadType: payloadType,
								Tag:         tag,
							})
							tag++
						}
						vNode = tc.ast.Nodes[vNode].NextSibling
					}

					typeID := tc.types.RegisterSumType(sym.NameID, variants, nil)
					sym.TypeID = uint32(typeID)

					// Also assign this typeID to all the variant symbols so we can typecheck constructors
					vNode = tc.ast.Nodes[sumNode].FirstChild
					for vNode != 0 {
						if tc.ast.Nodes[vNode].Kind == ast.NodeVariantDecl {
							vSymIdx := tc.ast.Nodes[vNode].Payload
							vSym := tc.symtable.SymbolAt(vSymIdx)
							vSym.TypeID = uint32(typeID)
						}
						vNode = tc.ast.Nodes[vNode].NextSibling
					}
				}
			}
		}

	case ast.NodeFuncDecl:
		prevReturn := tc.currentFuncReturnType
		prevAsync := tc.inAsyncFn
		prevInsideLoop := tc.insideLoop

		tc.insideLoop = false

		if node.Flags&uint16(ast.FlagIsAsync) != 0 {
			tc.inAsyncFn = true
		}
		
		// In a real compiler, we resolve the return type from the function signature.
		// We'll mock it by looking at the symbol's type if available, else Void.
		symIdx := node.Payload
		if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
			sym := tc.symtable.SymbolAt(symIdx)
			if sym.Kind == SymFunc {
				funcTypeID := types.TypeID(sym.TypeID)
				if funcTypeID != 0 && funcTypeID != types.TypeUnknown {
					entry := tc.types.Entry(funcTypeID)
					if entry.Kind == types.KindFunction {
						fInfo := tc.types.FuncInfo(funcTypeID)
						tc.currentFuncReturnType = fInfo.Return
					}
				}
			}
		}

		// Only typecheck body if not generic
		if node.Flags&uint16(ast.FlagIsGeneric) == 0 {
			child := node.FirstChild
			for child != 0 {
				tc.checkStmt(child)
				child = tc.ast.Nodes[child].NextSibling
			}
		}
		
		tc.currentFuncReturnType = prevReturn
		tc.inAsyncFn = prevAsync
		tc.insideLoop = prevInsideLoop

	case ast.NodeVarDecl:
		// Trigger inference engine to evaluate type/init
		tc.infer.TypeOf(nodeIdx)
		
		child := node.FirstChild
		for child != 0 {
			tc.checkStmt(child)
			child = tc.ast.Nodes[child].NextSibling
		}

	case ast.NodeAssignStmt:
		lhs := node.FirstChild
		rhs := uint32(0)
		if lhs != 0 {
			rhs = tc.ast.Nodes[lhs].NextSibling
		}

		if lhs != 0 && rhs != 0 {
			lhsNode := &tc.ast.Nodes[lhs]
			if lhsNode.Kind == ast.NodeIdent {
				symIdx := lhsNode.Payload
				if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
					sym := tc.symtable.SymbolAt(symIdx)
					// Check mutability
					if sym.Flags&SymFlagMut == 0 {
						name := tc.intern.Get(sym.NameID)
						tc.errorf(nodeIdx, 3010, "cannot assign to immutable variable '%s'", name)
					}
				}
			}
			
			// Check assignability
			lhsType := tc.infer.TypeOf(lhs)
			rhsType := tc.infer.inferNode(rhs, lhsType)
			if lhsType != types.TypeUnknown && rhsType != types.TypeUnknown {
				if !tc.isAssignableTo(rhsType, lhsType) {
					tc.errorf(nodeIdx, 3001, "cannot assign type %d to %d", rhsType, lhsType)
				}
			}
			
			tc.checkStmt(lhs)
			tc.checkStmt(rhs)
		}

	case ast.NodeIfStmt:
		cond := node.FirstChild
		if cond != 0 {
			condType := tc.infer.TypeOf(cond)
			if condType != types.TypeUnknown && condType != types.TypeBool {
				tc.errorf(nodeIdx, 3011, "if condition must be bool, found %d", condType)
			}
		}
		
		child := node.FirstChild
		for child != 0 {
			tc.checkStmt(child)
			child = tc.ast.Nodes[child].NextSibling
		}

	case ast.NodeWhileStmt:
		prevInsideLoop := tc.insideLoop
		tc.insideLoop = true
		child := node.FirstChild
		for child != 0 {
			tc.checkStmt(child)
			child = tc.ast.Nodes[child].NextSibling
		}
		tc.insideLoop = prevInsideLoop

	case ast.NodeForStmt:
		prevInsideLoop := tc.insideLoop
		tc.insideLoop = true
		
		iterExpr := node.FirstChild
		var elemType types.TypeID = types.TypeI32
		
		if iterExpr != 0 {
			tc.checkStmt(iterExpr)
			iterType := tc.infer.TypeOf(iterExpr)
			if iterType != types.TypeUnknown {
				entry := tc.types.Entry(iterType)
				if entry.Kind == types.KindSlice {
					elemType = tc.types.SliceElem(iterType)
				} else if entry.Kind == types.KindPointer {
					ptrElem := tc.types.PointerElem(iterType)
					ptrEntry := tc.types.Entry(ptrElem)
					if ptrEntry.Kind == types.KindArray {
						elemType = tc.types.ArrayElem(ptrElem)
					}
				}
			}
		}
		
		symIdx := node.Payload
		if symIdx != 0 && tc.symtable != nil && int(symIdx) < len(tc.symtable.Symbols) {
			sym := tc.symtable.SymbolAt(symIdx)
			sym.TypeID = uint32(elemType)
		}
		
		child := node.FirstChild
		for child != 0 {
			tc.checkStmt(child)
			child = tc.ast.Nodes[child].NextSibling
		}
		
		tc.insideLoop = prevInsideLoop

	case ast.NodeBreakStmt, ast.NodeContinueStmt:
		if !tc.insideLoop {
			tc.errorf(nodeIdx, 3013, "%s outside loop", node.Kind.String())
		}

	case ast.NodeReturnStmt:
		expr := node.FirstChild
		if expr != 0 {
			exprType := tc.infer.TypeOf(expr)
			if tc.currentFuncReturnType != 0 && tc.currentFuncReturnType != types.TypeUnknown {
				if exprType != types.TypeUnknown && !tc.isAssignableTo(exprType, tc.currentFuncReturnType) {
					tc.errorf(nodeIdx, 3005, "return type mismatch: expected %d, found %d", tc.currentFuncReturnType, exprType)
				}
			}
			tc.checkStmt(expr)
		} else {
			if tc.currentFuncReturnType != 0 && tc.currentFuncReturnType != types.TypeUnknown && tc.currentFuncReturnType != types.TypeVoid {
				tc.errorf(nodeIdx, 3005, "return type mismatch: expected %d, found void", tc.currentFuncReturnType)
			}
		}

	case ast.NodeMatchStmt:
		// Scrutinee
		scrutinee := node.FirstChild
		if scrutinee != 0 {
			tc.checkStmt(scrutinee)
			scrutineeType := tc.infer.TypeOf(scrutinee)

			prevScrutinee := tc.currentMatchScrutinee
			tc.currentMatchScrutinee = scrutineeType

			// Exhaustiveness check variables
			var sumInfo *types.SumType
			if scrutineeType != 0 && tc.types.Entry(scrutineeType).Kind == types.KindSum {
				sumInfo = tc.types.SumInfo(scrutineeType)
			}
			seenVariants := make(map[uint32]bool)
			hasWildcard := false

			// check arms...
			arm := tc.ast.Nodes[scrutinee].NextSibling
			for arm != 0 {
				tc.checkStmt(arm)
				
				// Extract arm pattern info for exhaustiveness
				if sumInfo != nil {
					pattern := tc.ast.Nodes[arm].FirstChild
					patNode := tc.ast.Nodes[pattern]
					if patNode.Kind == ast.NodeIdent {
						symIdx := patNode.Payload
						if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
							sym := tc.symtable.SymbolAt(symIdx)
							if sym.NameID == tc.intern.Intern([]byte("_")) {
								hasWildcard = true
							} else {
								seenVariants[sym.NameID] = true
							}
						}
					} else if patNode.Kind == ast.NodeCallExpr {
						callee := patNode.FirstChild
						if tc.ast.Nodes[callee].Kind == ast.NodeIdent {
							symIdx := tc.ast.Nodes[callee].Payload
							if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
								sym := tc.symtable.SymbolAt(symIdx)
								seenVariants[sym.NameID] = true
							}
						}
					} else if patNode.Kind == ast.NodeVariantPat {
						symIdx := patNode.Payload
						if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
							sym := tc.symtable.SymbolAt(symIdx)
							seenVariants[sym.NameID] = true
						}
					} else if patNode.Kind == ast.NodeBindingPat {
						symIdx := patNode.Payload
						if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
							sym := tc.symtable.SymbolAt(symIdx)
							if sym.Kind == SymVariant {
								seenVariants[sym.NameID] = true
							} else {
								hasWildcard = true
							}
						}
					} else if patNode.Kind == ast.NodeWildcardPat {
						hasWildcard = true
					}
				}

				arm = tc.ast.Nodes[arm].NextSibling
			}

			// Report non-exhaustive match
			if sumInfo != nil && !hasWildcard {
				var missing []string
				for _, v := range sumInfo.Variants {
					if !seenVariants[v.NameID] {
						missing = append(missing, string(tc.intern.Get(v.NameID)))
					}
				}
				if len(missing) > 0 {
					tc.errorf(nodeIdx, 3030, "non-exhaustive match: missing %v", missing)
				}
			}

			tc.currentMatchScrutinee = prevScrutinee
		}
		
	case ast.NodeMatchArm:
		pattern := node.FirstChild
		body := uint32(0)
		if pattern != 0 {
			body = tc.ast.Nodes[pattern].NextSibling
		}

		if tc.currentMatchScrutinee != 0 && tc.types.Entry(tc.currentMatchScrutinee).Kind == types.KindSum {
			sumInfo := tc.types.SumInfo(tc.currentMatchScrutinee)

			if tc.ast.Nodes[pattern].Kind == ast.NodeCallExpr {
				callee := tc.ast.Nodes[pattern].FirstChild
				arg := tc.ast.Nodes[callee].NextSibling
				
				if tc.ast.Nodes[callee].Kind == ast.NodeIdent {
					symIdx := tc.ast.Nodes[callee].Payload
					if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
						sym := tc.symtable.SymbolAt(symIdx)
						
						// Find variant payload type
						var payloadType types.TypeID = 0
						for _, v := range sumInfo.Variants {
							if v.NameID == sym.NameID {
								payloadType = v.PayloadType
								break
							}
						}

						// If arg is Ident, bind it
						if arg != 0 && tc.ast.Nodes[arg].Kind == ast.NodeIdent {
							argSymIdx := tc.ast.Nodes[arg].Payload
							if argSymIdx != 0 && int(argSymIdx) < len(tc.symtable.Symbols) {
								argSym := tc.symtable.SymbolAt(argSymIdx)
								argSym.TypeID = uint32(payloadType)
							}
						}
					}
				}
			} else if tc.ast.Nodes[pattern].Kind == ast.NodeVariantPat {
				symIdx := tc.ast.Nodes[pattern].Payload
				arg := tc.ast.Nodes[pattern].FirstChild
				
				if symIdx != 0 && int(symIdx) < len(tc.symtable.Symbols) {
					sym := tc.symtable.SymbolAt(symIdx)
					
					// Find variant payload type
					var payloadType types.TypeID = 0
					for _, v := range sumInfo.Variants {
						if v.NameID == sym.NameID {
							payloadType = v.PayloadType
							break
						}
					}

					// If arg is BindingPat or Ident, bind it
					if arg != 0 && (tc.ast.Nodes[arg].Kind == ast.NodeBindingPat || tc.ast.Nodes[arg].Kind == ast.NodeIdent) {
						argSymIdx := tc.ast.Nodes[arg].Payload
						if argSymIdx != 0 && int(argSymIdx) < len(tc.symtable.Symbols) {
							argSym := tc.symtable.SymbolAt(argSymIdx)
							argSym.TypeID = uint32(payloadType)
						}
					}
				}
			}
		}

		if body != 0 {
			tc.checkStmt(body)
		}

	case ast.NodeDeferStmt:
		expr := node.FirstChild
		if expr != 0 {
			exprNode := &tc.ast.Nodes[expr]
			if exprNode.Kind != ast.NodeCallExpr {
				tc.errorf(nodeIdx, 3015, "defer must be a function call")
			}
			tc.checkStmt(expr)
		}
		
	case ast.NodeSpawnExpr: // Treated as stmt sometimes, or just expr
		expr := node.FirstChild
		if expr != 0 {
			exprNode := &tc.ast.Nodes[expr]
			if exprNode.Kind != ast.NodeCallExpr {
				// error for spawn
				tc.errorf(nodeIdx, 3015, "spawn must be a function call")
			}
			tc.checkStmt(expr)
		}

	case ast.NodeUnaryExpr, ast.NodeIndexExpr, ast.NodeFieldExpr, ast.NodeCastExpr, ast.NodeAwaitExpr:
		tc.checkExpr(nodeIdx)

	case ast.NodeIdent:
		var name string
		if len(tc.ast.Tokens) > 0 && tc.ast.Nodes[nodeIdx].TokenIdx < uint32(len(tc.ast.Tokens)) {
			name = string(tc.ast.NodeText(nodeIdx))
		} else {
			payload := tc.ast.Nodes[nodeIdx].Payload
			if payload != 0 && int(payload) <= tc.intern.Len() {
				name = string(tc.intern.Get(payload))
			}
			if name != "break" && name != "continue" {
				if payload != 0 && int(payload) < len(tc.symtable.Symbols) {
					nameID := tc.symtable.Symbols[payload].NameID
					if nameID != 0 && int(nameID) <= tc.intern.Len() {
						name = string(tc.intern.Get(nameID))
					}
				}
			}
		}
		if name == "break" || name == "continue" {
			if !tc.insideLoop {
				tc.errorf(nodeIdx, 3013, "break/continue outside loop")
			}
		}

	default:
		child := node.FirstChild
		for child != 0 {
			tc.checkStmt(child)
			child = tc.ast.Nodes[child].NextSibling
		}
	}
}

func (tc *TypeChecker) isAssignableTo(from, to types.TypeID) bool {
	if from == to {
		return true
	}
	if tc.types.Entry(to).Kind == types.KindInterface {
		ok, _ := tc.ifaces.ImplementsInterface(from, to)
		return ok
	}
	return tc.types.IsAssignableTo(from, to)
}
