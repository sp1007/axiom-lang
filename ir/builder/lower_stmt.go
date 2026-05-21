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
			// Derive name ID from token
			text := fl.mb.tree.TokenText(cn.TokenIdx)
			nameID := fl.mb.intern.Intern(text)

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
		// Check if ownership operations are needed
		if node.Flags&(ast.FlagEscapesToHeap|ast.FlagUsesArena|ast.FlagIsMoved) != 0 {
			ownerReg := fl.lowerOwnershipAware(idx, node, initReg)
			fl.locals[nameID] = ownerReg
		} else {
			// Default SSA: variable IS the value
			fl.locals[nameID] = initReg
		}
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
	// Assignment in SSA: create a new register for the target
	// The name ID is in node.Payload
	nameID := node.Payload
	if nameID == 0 {
		// Fallback: try from token text
		text := fl.mb.tree.TokenText(node.TokenIdx)
		nameID = fl.mb.intern.Intern(text)
	}

	// Lower the RHS expression
	child := node.FirstChild
	if child != ast.NullIdx {
		cn := fl.mb.tree.Node(child)
		valReg := fl.lowerExpr(child, cn)
		fl.locals[nameID] = valReg
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
	// Create blocks: then, else, merge
	thenBlock := fl.fb.NewBlock()
	elseBlock := fl.fb.NewBlock()
	mergeBlock := fl.fb.NewBlock()

	// Lower condition (first child)
	child := node.FirstChild
	if child == ast.NullIdx {
		return
	}
	cn := fl.mb.tree.Node(child)
	condReg := fl.lowerExpr(child, cn)

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
	child = cn.NextSibling
	if child != ast.NullIdx {
		cn = fl.mb.tree.Node(child)
		if cn.Kind == ast.NodeBlock {
			fl.lowerBlock(child)
		}
		child = cn.NextSibling
	}
	if !fl.terminated {
		fl.fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: mergeBlock})
		fl.fb.AddEdge(thenBlock, mergeBlock)
	}

	// Else/elif block
	fl.fb.SwitchTo(elseBlock)
	fl.terminated = false
	hasElse := false
	for child != ast.NullIdx {
		cn = fl.mb.tree.Node(child)
		if cn.Kind == ast.NodeElseClause || cn.Kind == ast.NodeElifClause {
			hasElse = true
			elseBody := cn.FirstChild
			if elseBody != ast.NullIdx {
				ecn := fl.mb.tree.Node(elseBody)
				if ecn.Kind == ast.NodeBlock {
					fl.lowerBlock(elseBody)
				} else if cn.Kind == ast.NodeElifClause {
					// Elif is a nested if inside the else block
					fl.lowerIf(child, cn)
				}
			}
		}
		child = cn.NextSibling
	}
	if !fl.terminated {
		fl.fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: mergeBlock})
		fl.fb.AddEdge(elseBlock, mergeBlock)
	}
	if !hasElse {
		// Empty else falls through
		fl.fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: mergeBlock})
		fl.fb.AddEdge(elseBlock, mergeBlock)
	}

	// Merge block
	fl.fb.SwitchTo(mergeBlock)
	fl.terminated = false
}

// lowerWhile lowers a while loop.
func (fl *funcLowering) lowerWhile(idx uint32, node *ast.AstNode) {
	condBlock := fl.fb.NewBlock()
	bodyBlock := fl.fb.NewBlock()
	exitBlock := fl.fb.NewBlock()

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
		fl.fb.AddEdge(bodyBlock, condBlock)
	}

	// Exit
	fl.fb.SwitchTo(exitBlock)
	fl.terminated = false
}

// lowerFor lowers a for-in loop (range-based).
func (fl *funcLowering) lowerFor(idx uint32, node *ast.AstNode) {
	// For now, treat like a while loop with iterator variable
	// Full range lowering requires iterator protocol; emit loop structure

	condBlock := fl.fb.NewBlock()
	bodyBlock := fl.fb.NewBlock()
	exitBlock := fl.fb.NewBlock()

	// Get loop variable name
	text := fl.mb.tree.TokenText(node.TokenIdx)
	nameID := fl.mb.intern.Intern(text)

	// Initialize loop variable to 0
	iterReg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpIConst,
		TypeID: uint16(3), // i32
		Dest:   iterReg,
		Src1:   0,
	})
	fl.locals[nameID] = iterReg

	// Jump to condition
	curBlock := fl.fb.CurrentBlock()
	fl.fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: condBlock})
	fl.fb.AddEdge(curBlock, condBlock)

	// Condition: i < limit (simplified — full lowering needs range expr)
	fl.fb.SwitchTo(condBlock)
	fl.terminated = false

	// Lower the range expression (second child)
	limitReg := uint32(0)
	child := node.FirstChild
	if child != ast.NullIdx {
		cn := fl.mb.tree.Node(child)
		child = cn.NextSibling // skip iterator variable decl, get range expr
		if child != ast.NullIdx {
			cn = fl.mb.tree.Node(child)
			limitReg = fl.lowerExpr(child, cn)
		}
	}

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

	// Lower body (last child should be the block)
	if child != ast.NullIdx {
		cn := fl.mb.tree.Node(child)
		child = cn.NextSibling
		if child != ast.NullIdx {
			cn = fl.mb.tree.Node(child)
			if cn.Kind == ast.NodeBlock {
				fl.lowerBlock(child)
			}
		}
	}

	if !fl.terminated {
		// Increment iterator: i = i + 1
		oneReg := fl.fb.FreshReg()
		fl.fb.Emit(air.AirInst{
			Opcode: air.OpIConst,
			TypeID: uint16(3), // i32
			Dest:   oneReg,
			Src1:   1,
		})
		newIterReg := fl.fb.FreshReg()
		fl.fb.Emit(air.AirInst{
			Opcode: air.OpIAdd,
			TypeID: uint16(3),
			Dest:   newIterReg,
			Src1:   fl.locals[nameID],
			Src2:   oneReg,
		})
		fl.locals[nameID] = newIterReg

		fl.fb.Emit(air.AirInst{Opcode: air.OpJump, Src1: condBlock})
		fl.fb.AddEdge(bodyBlock, condBlock)
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
