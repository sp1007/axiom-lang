package sema

import (
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

// checkExpr checks an expression node for type safety rules not covered by inference.
func (tc *TypeChecker) checkExpr(nodeIdx uint32) {
	if nodeIdx == 0 {
		return
	}

	node := &tc.ast.Nodes[nodeIdx]

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
				if entry.Kind != types.KindArray && entry.Kind != types.KindSlice && colType != types.TypeString {
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
				if !idxType.IsInteger() {
					tc.errorf(nodeIdx, 3024, "index must be integer, found %d", idxType)
				}
			}
		}

	case ast.NodeFieldExpr:
		obj := node.FirstChild
		if obj != 0 {
			tc.checkStmt(obj)
			objType := tc.infer.TypeOf(obj)
			
			// Mock struct checking. In full AXIOM, we lookup the struct's fields.
			// The node.Payload has the field NameID.
			if objType != types.TypeUnknown {
				entry := tc.types.Entry(objType)
				if entry.Kind != types.KindStruct {
					tc.errorf(nodeIdx, 3025, "cannot access field on non-struct type %d", objType)
				}
				// Assume the field exists for now if it's a struct or module
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
				
				if (exprType.IsInteger() || exprType.IsFloat()) && (targetType.IsInteger() || targetType.IsFloat()) {
					valid = true
				} else if exprType.IsInteger() && targetType == types.TypeBool {
					valid = true
				} else if exprType == types.TypeBool && targetType.IsInteger() {
					valid = true
				}
				// More cast logic like pointer<->pointer would go here
				
				if !valid {
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
