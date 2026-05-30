package sema

import (
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

// CTGCPass implements Compile-Time Garbage Collection by injecting
// DestroyStmt nodes at scope exits for heap-allocated values that
// have not been moved or returned.
type CTGCPass struct {
	ast   *ast.AstTree
	st    *SymbolTable
	moved map[uint32]bool // set of moved symbol IDs
	Types *types.TypeTable
}

// NewCTGCPass creates a new CTGC pass.
func NewCTGCPass(tree *ast.AstTree, st *SymbolTable, moved map[uint32]bool) *CTGCPass {
	return &CTGCPass{
		ast:   tree,
		st:    st,
		moved: moved,
	}
}

// InjectDestroys walks the function body and injects DestroyStmt nodes
// at block exits for all heap-allocated values still alive (not moved/returned).
func (ctgc *CTGCPass) InjectDestroys(funcNodeIdx uint32) {
	ctgc.walkNode(funcNodeIdx)
}

func (ctgc *CTGCPass) walkNode(nodeIdx uint32) {
	node := &ctgc.ast.Nodes[nodeIdx]

	// Recurse into children first (depth-first)
	child := node.FirstChild
	for child != 0 {
		ctgc.walkNode(child)
		child = ctgc.ast.Nodes[child].NextSibling
	}

	// After processing children, inject destroys at block exits
	if node.Kind == ast.NodeBlock {
		ctgc.injectBlockDestroys(nodeIdx)
	}
}

// injectBlockDestroys collects heap-allocated VarDecls in a block and
// injects DestroyStmt nodes at the block's end in LIFO order.
func (ctgc *CTGCPass) injectBlockDestroys(blockIdx uint32) {
	// Collect VarDecl nodes in this block that need destruction
	var destroyTargets []uint32 // symIDs in declaration order

	child := ctgc.ast.Nodes[blockIdx].FirstChild
	for child != 0 {
		childNode := &ctgc.ast.Nodes[child]
		if childNode.Kind == ast.NodeVarDecl {
			symID := childNode.Payload
			if ctgc.needsDestroy(symID, child) {
				destroyTargets = append(destroyTargets, symID)
			}
		}
		child = childNode.NextSibling
	}

	if len(destroyTargets) == 0 {
		return
	}

	// Inject DestroyStmt in REVERSE order (LIFO: last declared → first destroyed)
	for i := len(destroyTargets) - 1; i >= 0; i-- {
		symID := destroyTargets[i]
		destroyNode := ctgc.ast.AddNode(ast.NodeDestroyStmt, 0)
		ctgc.ast.SetPayload(destroyNode, symID)
		ctgc.ast.AppendChild(blockIdx, destroyNode)
	}
}

// needsDestroy returns true if the variable at symID/nodeIdx needs a DestroyStmt.
func (ctgc *CTGCPass) needsDestroy(symID uint32, nodeIdx uint32) bool {
	if symID == 0 {
		return false
	}

	// Skip if already moved
	if ctgc.moved[symID] {
		return false
	}

	// Skip if not heap-allocated (no EscapesToHeap flag)
	node := &ctgc.ast.Nodes[nodeIdx]
	if node.Flags&ast.FlagEscapesToHeap == 0 {
		return false
	}

	// Skip primitive types (no cleanup needed)
	if symID < uint32(len(ctgc.st.Symbols)) {
		sym := ctgc.st.SymbolAt(symID)
		if ctgc.isPrimitiveType(sym.TypeID) {
			return false
		}
		if sym.TypeID != 0 && ctgc.Types != nil {
			typeID := types.TypeID(sym.TypeID)
			if int(typeID) < ctgc.Types.Count() {
				entry := ctgc.Types.Entry(typeID)
				switch entry.Kind {
				case types.KindPointer, types.KindRef, types.KindFunction:
					return false
				}
			}
		}
	}

	// Skip arena-allocated values
	if node.Flags&ast.FlagUsesArena != 0 {
		return false
	}

	return true
}

// isPrimitiveType returns true if the type is a built-in primitive (i32, f64, bool, etc.)
// that doesn't need heap deallocation.
func (ctgc *CTGCPass) isPrimitiveType(typeID uint32) bool {
	// Primitive types have low TypeIDs (the first entries in the TypeTable).
	// TypeID 0 = void, 1-17 = primitives (i8, i16, i32, i64, u8, u16, u32, u64,
	// f32, f64, bool, string, isize, usize, char, byte, nil).
	// This is a heuristic; a full implementation would check TypeEntry.Kind.
	return typeID > 0 && typeID <= 17
}

// InjectEarlyReturnDestroys finds all ReturnStmt nodes in the function
// and injects DestroyStmt for still-alive heap values before each return.
func (ctgc *CTGCPass) InjectEarlyReturnDestroys(funcNodeIdx uint32) {
	// Collect all VarDecls in the function with EscapesToHeap
	var heapVars []heapVar
	ctgc.collectHeapVars(funcNodeIdx, &heapVars)

	// Find all return statements and inject destroys before them
	ctgc.injectBeforeReturns(funcNodeIdx, heapVars)
}

type heapVar struct {
	symID   uint32
	nodeIdx uint32
}

func (ctgc *CTGCPass) collectHeapVars(nodeIdx uint32, vars *[]heapVar) {
	node := &ctgc.ast.Nodes[nodeIdx]

	if node.Kind == ast.NodeVarDecl {
		symID := node.Payload
		if ctgc.needsDestroy(symID, nodeIdx) {
			*vars = append(*vars, heapVar{symID: symID, nodeIdx: nodeIdx})
		}
	}

	child := node.FirstChild
	for child != 0 {
		ctgc.collectHeapVars(child, vars)
		child = ctgc.ast.Nodes[child].NextSibling
	}
}

func (ctgc *CTGCPass) injectBeforeReturns(nodeIdx uint32, heapVars []heapVar) {
	node := &ctgc.ast.Nodes[nodeIdx]

	// For return statements: the returned value should NOT be destroyed.
	// We need to check if the return value is one of our heap vars.
	if node.Kind == ast.NodeReturnStmt {
		var returnedSym uint32
		retChild := node.FirstChild
		if retChild != 0 {
			retNode := &ctgc.ast.Nodes[retChild]
			if retNode.Kind == ast.NodeIdent {
				returnedSym = retNode.Payload
			}
		}

		// Inject destroys for all heap vars that are NOT the returned value,
		// in LIFO order. Since we can't insert siblings before a node in the
		// flat-array AST, we instead add DestroyStmt as children of the return.
		// The codegen will emit these before the actual return instruction.
		for i := len(heapVars) - 1; i >= 0; i-- {
			hv := heapVars[i]
			if hv.symID != returnedSym && !ctgc.moved[hv.symID] {
				destroyNode := ctgc.ast.AddNode(ast.NodeDestroyStmt, 0)
				ctgc.ast.SetPayload(destroyNode, hv.symID)
				ctgc.ast.AppendChild(nodeIdx, destroyNode)
			}
		}
		return // don't recurse into return's children
	}

	child := node.FirstChild
	for child != 0 {
		ctgc.injectBeforeReturns(child, heapVars)
		child = ctgc.ast.Nodes[child].NextSibling
	}
}

// DestroyCount returns the number of DestroyStmt nodes in the tree under nodeIdx.
func (ctgc *CTGCPass) DestroyCount(nodeIdx uint32) int {
	count := 0
	ctgc.countDestroys(nodeIdx, &count)
	return count
}

func (ctgc *CTGCPass) countDestroys(nodeIdx uint32, count *int) {
	node := &ctgc.ast.Nodes[nodeIdx]
	if node.Kind == ast.NodeDestroyStmt {
		*count++
	}
	child := node.FirstChild
	for child != 0 {
		ctgc.countDestroys(child, count)
		child = ctgc.ast.Nodes[child].NextSibling
	}
}
