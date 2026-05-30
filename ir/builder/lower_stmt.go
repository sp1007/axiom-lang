package builder

import (
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/ir/air"
)

// funcLowering handles AIR generation for a single function.
// It holds per-function state: the block builder, the local variable map,
// and scope tracking for SSA value resolution.
type funcLowering struct {
	mb     *ModuleBuilder
	fb     *air.AirFuncBuilder
	locals map[uint32]uint32 // name ID → SSA register holding current value
	params []uint32          // parameter type IDs

	// Loop tracking for break/continue
	loopConds []uint32
	loopExits []uint32

	// Track whether the current block already has a terminator.
	terminated bool
}

// newFuncLowering creates a new function lowering context.
func newFuncLowering(mb *ModuleBuilder, fb *air.AirFuncBuilder, paramTypes []uint32) *funcLowering {
	return &funcLowering{
		mb:     mb,
		fb:     fb,
		locals: make(map[uint32]uint32, 32),
		params: paramTypes,
	}
}

// registerParams creates SSA registers for function parameters and stores
// them in the locals map.
func (fl *funcLowering) registerParams(funcIdx uint32, funcNode *ast.AstNode) {
	child := funcNode.FirstChild
	paramIdx := 0
	for child != ast.NullIdx {
		cn := fl.mb.tree.Node(child)
		if cn.Kind == ast.NodeParamDecl {
			// Derive name ID from token or payload
			nameID := cn.Payload
			if nameID == 0 {
				text := fl.mb.tree.TokenText(cn.TokenIdx)
				nameID = fl.mb.intern.Intern(text)
			}

			// Create a register for the parameter
			reg := fl.fb.FreshReg()

			// Emit a synthetic copy from the implicit parameter slot
			// In AIR, parameters are implicit starting values in the entry block.
			typeID := uint16(0)
			if paramIdx < len(fl.params) {
				typeID = uint16(fl.params[paramIdx])
			}
			fl.fb.Emit(air.AirInst{
				Opcode: air.OpCopy,
				TypeID: typeID,
				Dest:   reg,
				Src1:   uint32(paramIdx + 1), // 1-based param slot (0 is reserved)
			})

			fl.locals[nameID] = reg
			paramIdx++
		}
		child = cn.NextSibling
	}
}

// ensureReturn adds a void return at the end of the function if the last
// block does not already have a terminator.
func (fl *funcLowering) ensureReturn() {
	if !fl.terminated {
		fl.fb.Emit(air.AirInst{
			Opcode: air.OpReturn,
		})
	}
}

// lowerBlock lowers a NodeBlock (sequence of statements).
func (fl *funcLowering) lowerBlock(blockIdx uint32) {
	node := fl.mb.tree.Node(blockIdx)
	child := node.FirstChild
	for child != ast.NullIdx {
		cn := fl.mb.tree.Node(child)
		fl.lowerStmt(child, cn)
		child = cn.NextSibling
	}
}

// lowerStmt dispatches a statement node to the appropriate lowering method.
func (fl *funcLowering) lowerStmt(idx uint32, node *ast.AstNode) {
	// If we already emitted a terminator, skip dead code
	if fl.terminated {
		return
	}

	switch node.Kind {
	case ast.NodeVarDecl:
		fl.lowerVarDecl(idx, node)
	case ast.NodeAssignStmt:
		fl.lowerAssign(idx, node)
	case ast.NodeReturnStmt:
		fl.lowerReturn(idx, node)
	case ast.NodeIfStmt:
		fl.lowerIf(idx, node)
	case ast.NodeWhileStmt:
		fl.lowerWhile(idx, node)
	case ast.NodeForStmt:
		fl.lowerFor(idx, node)
	case ast.NodeBlock:
		fl.lowerBlock(idx)
	case ast.NodeDestroyStmt:
		fl.lowerDestroy(idx, node)
	case ast.NodeAliasStmt:
		fl.lowerAliasStmt(idx, node)
	case ast.NodeDeferStmt:
		// Defer is handled at scope exit; record for later emission
		fl.lowerDefer(idx, node)
	case ast.NodeBreakStmt:
		fl.lowerBreak(idx, node)
	case ast.NodeContinueStmt:
		fl.lowerContinue(idx, node)
	default:
		// Expression statement (e.g., function call)
		fl.lowerExpr(idx, node)
	}
}

// lowerVarDecl lowers a variable declaration: let x: T = expr
func (fl *funcLowering) lowerVarDecl(idx uint32, node *ast.AstNode) {
	// The name ID is stored in node.Payload (intern pool ID)
	nameID := node.Payload

	// Type ID would come from semantic analysis.
	// For MVP, we infer from the initializer's type or use 0.
	typeID := uint16(0)

	// Lower the initializer expression (if present)
	initReg := uint32(0)
	child := node.FirstChild
	for child != ast.NullIdx {
		cn := fl.mb.tree.Node(child)
		// Skip type annotation nodes; take the last non-type child as init
		if cn.Kind != ast.NodeTypeExpr && cn.Kind != ast.NodePtrType &&
			cn.Kind != ast.NodeSliceType && cn.Kind != ast.NodeArrayType &&
			cn.Kind != ast.NodeFuncType && cn.Kind != ast.NodeGenericType {
			initReg = fl.lowerExpr(child, cn)
		}
		child = cn.NextSibling
	}

	// If we have an initializer, check ownership flags
	if initReg != 0 {
		srcReg := initReg
		// Check if ownership operations are needed
		if node.Flags&(ast.FlagEscapesToHeap|ast.FlagUsesArena|ast.FlagIsMoved) != 0 {
			srcReg = fl.lowerOwnershipAware(idx, node, initReg)
		}

		reg := fl.fb.FreshReg()
		varTypeID := uint16(0)
		if node.Payload != 0 && fl.mb.symbols != nil && int(node.Payload) < len(fl.mb.symbols.Symbols) {
			sym := fl.mb.symbols.SymbolAt(node.Payload)
			if sym.TypeID != 0 {
				varTypeID = uint16(sym.TypeID)
			}
		}

		fl.fb.Emit(air.AirInst{
			Opcode: air.OpCopy,
			TypeID: varTypeID,
			Dest:   reg,
			Src1:   srcReg,
		})
		fl.locals[nameID] = reg
	} else {
		// No initializer: create a zero/undef value
		reg := fl.fb.FreshReg()
		fl.fb.Emit(air.AirInst{
			Opcode: air.OpIConst,
			TypeID: typeID,
			Dest:   reg,
			Src1:   0, // zero value
		})
		fl.locals[nameID] = reg
	}
}

// lowerAssign lowers an assignment statement: x = expr
func (fl *funcLowering) lowerAssign(idx uint32, node *ast.AstNode) {
	lhsIdx := node.FirstChild
	if lhsIdx == ast.NullIdx {
		return
	}
	lhsNode := fl.mb.tree.Node(lhsIdx)

	rhsIdx := lhsNode.NextSibling
	if rhsIdx == ast.NullIdx {
		return
	}
	rhsNode := fl.mb.tree.Node(rhsIdx)
	valReg := fl.lowerExpr(rhsIdx, rhsNode)

	// Determine type
	typeID := fl.nodeType(lhsNode)
	if typeID == 0 {
		typeID = 3 // default i32
	}

	if lhsNode.Kind == ast.NodeIndexExpr {
		// Index assignment: arr[idx] = expr
		arrIdx := lhsNode.FirstChild
		if arrIdx == ast.NullIdx {
			return
		}
		arrNode := fl.mb.tree.Node(arrIdx)
		arrReg := fl.lowerExpr(arrIdx, arrNode)

		idxExprIdx := arrNode.NextSibling
		idxReg := uint32(0)
		if idxExprIdx != ast.NullIdx {
			idxNode := fl.mb.tree.Node(idxExprIdx)
			idxReg = fl.lowerExpr(idxExprIdx, idxNode)
		}

		fl.fb.Emit(air.AirInst{
			Opcode: air.OpStore,
			TypeID: typeID,
			Dest:   arrReg,
			Src1:   valReg,
			Src2:   idxReg,
		})
	} else if lhsNode.Kind == ast.NodeDerefExpr {
		// Dereference assignment: ptr.* = expr
		ptrIdx := lhsNode.FirstChild
		if ptrIdx == ast.NullIdx {
			return
		}
		ptrNode := fl.mb.tree.Node(ptrIdx)
		ptrReg := fl.lowerExpr(ptrIdx, ptrNode)

		fl.fb.Emit(air.AirInst{
			Opcode: air.OpStore,
			TypeID: typeID,
			Dest:   ptrReg,
			Src1:   valReg,
			Src2:   0,
		})
	} else if lhsNode.Kind == ast.NodeFieldExpr {
		// Field assignment: obj.field = expr
		objIdx := lhsNode.FirstChild
		if objIdx == ast.NullIdx {
			return
		}
		objNode := fl.mb.tree.Node(objIdx)
		objReg := fl.lowerExpr(objIdx, objNode)

		fieldIdx := lhsNode.ExtraIdx

		fl.fb.Emit(air.AirInst{
			Opcode: air.OpSetField,
			TypeID: typeID,
			Dest:   valReg,
			Src1:   objReg,
			Src2:   fieldIdx,
		})
	} else {
		// Normal variable assignment: x = expr
		nameID := lhsNode.Payload
		if nameID == 0 {
			text := fl.mb.tree.TokenText(lhsNode.TokenIdx)
			nameID = fl.mb.intern.Intern(text)
		}

		if existingReg, ok := fl.locals[nameID]; ok {
			fl.fb.Emit(air.AirInst{
				Opcode: air.OpCopy,
				TypeID: typeID,
				Dest:   existingReg,
				Src1:   valReg,
			})
		} else {
			fl.locals[nameID] = valReg
		}
	}
}

// lowerReturn lowers a return statement.
func (fl *funcLowering) lowerReturn(idx uint32, node *ast.AstNode) {
	retVal := uint32(0)
	if node.FirstChild != ast.NullIdx {
		cn := fl.mb.tree.Node(node.FirstChild)
		retVal = fl.lowerExpr(node.FirstChild, cn)
	}

	fl.fb.Emit(air.AirInst{
		Opcode: air.OpReturn,
		Src1:   retVal,
	})
	fl.terminated = true
}

// lowerIf lowers an if/elif/else chain.
func (fl *funcLowering) lowerIf(idx uint32, node *ast.AstNode) {
	mergeBlock := fl.fb.NewBlock()

	condNodeIdx := node.FirstChild
	if condNodeIdx == ast.NullIdx {
		return
	}
	condNode := fl.mb.tree.Node(condNodeIdx)

	bodyNodeIdx := condNode.NextSibling
	if bodyNodeIdx == ast.NullIdx {
		return
	}
	bodyNode := fl.mb.tree.Node(bodyNodeIdx)

	nextClauseIdx := bodyNode.NextSibling

	fl.lowerIfChain(condNodeIdx, condNode, bodyNodeIdx, bodyNode, nextClauseIdx, mergeBlock)

	// Merge block
	fl.fb.SwitchTo(mergeBlock)
	fl.terminated = false
}

func (fl *funcLowering) lowerIfChain(condIdx uint32, condNode *ast.AstNode, bodyIdx uint32, bodyNode *ast.AstNode, nextClauseIdx uint32, mergeBlock uint32) {
	thenBlock := fl.fb.NewBlock()
	elseBlock := fl.fb.NewBlock()

	// Lower condition
	condReg := fl.lowerExpr(condIdx, condNode)

	// Branch
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpBranch,
		Src1:   condReg,
		Src2:   thenBlock,
		Dest:   elseBlock,
	})
	curBlock := fl.fb.CurrentBlock()
	fl.fb.AddEdge(curBlock, thenBlock)
	fl.fb.AddEdge(curBlock, elseBlock)

	// Then block
	fl.fb.SwitchTo(thenBlock)
	fl.terminated = false
	if bodyNode.Kind == ast.NodeBlock {
		fl.lowerBlock(bodyIdx)
	}
	if !fl.terminated {
		fl.fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: mergeBlock})
		fl.fb.AddEdge(fl.fb.CurrentBlock(), mergeBlock)
	}

	// Else block
	fl.fb.SwitchTo(elseBlock)
	fl.terminated = false
	delegated := false
	if nextClauseIdx != ast.NullIdx {
		cn := fl.mb.tree.Node(nextClauseIdx)
		if cn.Kind == ast.NodeElseClause {
			elseBodyIdx := cn.FirstChild
			if elseBodyIdx != ast.NullIdx {
				elseBodyNode := fl.mb.tree.Node(elseBodyIdx)
				if elseBodyNode.Kind == ast.NodeBlock {
					fl.lowerBlock(elseBodyIdx)
				}
			}
		} else if cn.Kind == ast.NodeElifClause {
			elifCondIdx := cn.FirstChild
			if elifCondIdx != ast.NullIdx {
				elifCondNode := fl.mb.tree.Node(elifCondIdx)
				elifBodyIdx := elifCondNode.NextSibling
				if elifBodyIdx != ast.NullIdx {
					elifBodyNode := fl.mb.tree.Node(elifBodyIdx)
					fl.lowerIfChain(elifCondIdx, elifCondNode, elifBodyIdx, elifBodyNode, cn.NextSibling, mergeBlock)
					delegated = true
				}
			}
		}
	}
	if !fl.terminated && !delegated {
		fl.fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: mergeBlock})
		fl.fb.AddEdge(fl.fb.CurrentBlock(), mergeBlock)
	}
}

// lowerWhile lowers a while loop.
func (fl *funcLowering) lowerWhile(idx uint32, node *ast.AstNode) {
	condBlock := fl.fb.NewBlock()
	bodyBlock := fl.fb.NewBlock()
	exitBlock := fl.fb.NewBlock()

	fl.loopConds = append(fl.loopConds, condBlock)
	fl.loopExits = append(fl.loopExits, exitBlock)
	defer func() {
		fl.loopConds = fl.loopConds[:len(fl.loopConds)-1]
		fl.loopExits = fl.loopExits[:len(fl.loopExits)-1]
	}()

	// Jump from current to condition
	curBlock := fl.fb.CurrentBlock()
	fl.fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: condBlock})
	fl.fb.AddEdge(curBlock, condBlock)

	// Condition block
	fl.fb.SwitchTo(condBlock)
	fl.terminated = false
	child := node.FirstChild
	if child == ast.NullIdx {
		return
	}
	cn := fl.mb.tree.Node(child)
	condReg := fl.lowerExpr(child, cn)

	fl.fb.Emit(air.AirInst{
		Opcode: air.OpBranch,
		Src1:   condReg,
		Src2:   bodyBlock,
		Dest:   exitBlock,
	})
	fl.fb.AddEdge(condBlock, bodyBlock)
	fl.fb.AddEdge(condBlock, exitBlock)

	// Body block
	fl.fb.SwitchTo(bodyBlock)
	fl.terminated = false
	child = cn.NextSibling
	if child != ast.NullIdx {
		cn = fl.mb.tree.Node(child)
		if cn.Kind == ast.NodeBlock {
			fl.lowerBlock(child)
		}
	}
	if !fl.terminated {
		fl.fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: condBlock})
		fl.fb.AddEdge(fl.fb.CurrentBlock(), condBlock)
	}

	// Exit
	fl.fb.SwitchTo(exitBlock)
	fl.terminated = false
}

// lowerFor lowers a for-in loop (range-based).
func (fl *funcLowering) lowerFor(idx uint32, node *ast.AstNode) {
	condBlock := fl.fb.NewBlock()
	bodyBlock := fl.fb.NewBlock()
	exitBlock := fl.fb.NewBlock()

	fl.loopConds = append(fl.loopConds, condBlock)
	fl.loopExits = append(fl.loopExits, exitBlock)
	defer func() {
		fl.loopConds = fl.loopConds[:len(fl.loopConds)-1]
		fl.loopExits = fl.loopExits[:len(fl.loopExits)-1]
	}()

	// Get loop variable name/symbol from payload
	symIdx := node.Payload
	var nameID uint32
	var typeID uint16 = 3 // default i32

	if symIdx != 0 && fl.mb.symbols != nil && int(symIdx) < len(fl.mb.symbols.Symbols) {
		sym := fl.mb.symbols.SymbolAt(symIdx)
		nameID = sym.NameID
		if sym.TypeID != 0 {
			typeID = uint16(sym.TypeID)
		}
	} else {
		nameID = symIdx
	}

	rangeExprIdx := node.FirstChild
	var startReg uint32 = 0
	var limitReg uint32 = 0
	isRange := false

	if rangeExprIdx != ast.NullIdx {
		rangeNode := fl.mb.tree.Node(rangeExprIdx)
		if rangeNode.Kind == ast.NodeBinaryExpr {
			opText := string(fl.mb.tree.TokenText(rangeNode.TokenIdx))
			if opText == ".." {
				isRange = true
				startNodeIdx := rangeNode.FirstChild
				if startNodeIdx != ast.NullIdx {
					startNode := fl.mb.tree.Node(startNodeIdx)
					startReg = fl.lowerExpr(startNodeIdx, startNode)
					endNodeIdx := startNode.NextSibling
					if endNodeIdx != ast.NullIdx {
						endNode := fl.mb.tree.Node(endNodeIdx)
						limitReg = fl.lowerExpr(endNodeIdx, endNode)
					}
				}
			}
		}
	}

	iterReg := fl.fb.FreshReg()
	if isRange {
		fl.fb.Emit(air.AirInst{
			Opcode: air.OpCopy,
			TypeID: typeID,
			Dest:   iterReg,
			Src1:   startReg,
		})
	} else {
		fl.fb.Emit(air.AirInst{
			Opcode: air.OpIConst,
			TypeID: typeID,
			Dest:   iterReg,
			Src1:   0,
		})
		if rangeExprIdx != ast.NullIdx {
			rangeNode := fl.mb.tree.Node(rangeExprIdx)
			limitReg = fl.lowerExpr(rangeExprIdx, rangeNode)
		}
	}

	fl.locals[nameID] = iterReg

	// Jump to condition
	curBlock := fl.fb.CurrentBlock()
	fl.fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: condBlock})
	fl.fb.AddEdge(curBlock, condBlock)

	// Condition
	fl.fb.SwitchTo(condBlock)
	fl.terminated = false

	// Comparison: currentIter < limit
	cmpReg := fl.fb.FreshReg()
	currentIter := fl.locals[nameID]
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpLt,
		TypeID: uint16(11), // bool
		Dest:   cmpReg,
		Src1:   currentIter,
		Src2:   limitReg,
	})

	fl.fb.Emit(air.AirInst{
		Opcode: air.OpBranch,
		Src1:   cmpReg,
		Src2:   bodyBlock,
		Dest:   exitBlock,
	})
	fl.fb.AddEdge(condBlock, bodyBlock)
	fl.fb.AddEdge(condBlock, exitBlock)

	// Body block
	fl.fb.SwitchTo(bodyBlock)
	fl.terminated = false

	// Lower body
	if rangeExprIdx != ast.NullIdx {
		rangeNode := fl.mb.tree.Node(rangeExprIdx)
		bodyNodeIdx := rangeNode.NextSibling
		if bodyNodeIdx != ast.NullIdx {
			bodyNode := fl.mb.tree.Node(bodyNodeIdx)
			if bodyNode.Kind == ast.NodeBlock {
				fl.lowerBlock(bodyNodeIdx)
			}
		}
	}

	if !fl.terminated {
		// Increment iterator: i = i + 1
		oneReg := fl.fb.FreshReg()
		fl.fb.Emit(air.AirInst{
			Opcode: air.OpIConst,
			TypeID: typeID,
			Dest:   oneReg,
			Src1:   1,
		})
		iterReg := fl.locals[nameID]
		newValReg := fl.fb.FreshReg()
		fl.fb.Emit(air.AirInst{
			Opcode: air.OpIAdd,
			TypeID: typeID,
			Dest:   newValReg,
			Src1:   iterReg,
			Src2:   oneReg,
		})
		fl.fb.Emit(air.AirInst{
			Opcode: air.OpCopy,
			TypeID: typeID,
			Dest:   iterReg,
			Src1:   newValReg,
		})

		fl.fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: condBlock})
		fl.fb.AddEdge(fl.fb.CurrentBlock(), condBlock)
	}

	// Exit
	fl.fb.SwitchTo(exitBlock)
	fl.terminated = false
}

// lowerDestroy lowers a compiler-injected destroy statement (CTGC).
func (fl *funcLowering) lowerDestroy(idx uint32, node *ast.AstNode) {
	if node.FirstChild != ast.NullIdx {
		cn := fl.mb.tree.Node(node.FirstChild)
		valReg := fl.lowerExpr(node.FirstChild, cn)
		fl.fb.Emit(air.AirInst{
			Opcode: air.OpDestroy,
			Src1:   valReg,
		})
	} else {
		symID := node.Payload
		if symID != 0 && int(symID) < len(fl.mb.symbols.Symbols) {
			sym := fl.mb.symbols.SymbolAt(symID)
			if reg, ok := fl.locals[sym.NameID]; ok {
				fl.fb.Emit(air.AirInst{
					Opcode: air.OpDestroy,
					Src1:   reg,
				})
			}
		}
	}
}

// lowerDefer records a deferred expression for emission at scope exit.
// For MVP, defer is lowered as an inline call at the current position.
func (fl *funcLowering) lowerDefer(idx uint32, node *ast.AstNode) {
	// MVP: lower the deferred expression inline
	// Full implementation requires scope tracking + reverse-order emission
	if node.FirstChild != ast.NullIdx {
		cn := fl.mb.tree.Node(node.FirstChild)
		fl.lowerExpr(node.FirstChild, cn)
	}
}

func (fl *funcLowering) lowerBreak(idx uint32, node *ast.AstNode) {
	if len(fl.loopExits) == 0 {
		return
	}
	target := fl.loopExits[len(fl.loopExits)-1]
	fl.fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: target})
	fl.fb.AddEdge(fl.fb.CurrentBlock(), target)
	fl.terminated = true
}

func (fl *funcLowering) lowerContinue(idx uint32, node *ast.AstNode) {
	if len(fl.loopConds) == 0 {
		return
	}
	target := fl.loopConds[len(fl.loopConds)-1]
	fl.fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: target})
	fl.fb.AddEdge(fl.fb.CurrentBlock(), target)
	fl.terminated = true
}
