package sema

import (
	"github.com/axiom-lang/axiom/compiler/ast"
)

// AliasReuse implements the CTGC alias reuse optimization.
// When a heap allocation immediately follows a destroy of the same type,
// the existing memory is reused instead of free+malloc.
// This eliminates allocator round-trips in loop-body allocation patterns.
type AliasReuse struct {
	ast *ast.AstTree
	st  *SymbolTable
	cg  *ConnectionGraph
}

// NewAliasReuse creates a new alias reuse optimization pass.
func NewAliasReuse(tree *ast.AstTree, st *SymbolTable, cg *ConnectionGraph) *AliasReuse {
	return &AliasReuse{
		ast: tree,
		st:  st,
		cg:  cg,
	}
}

// Optimize scans the AST for destroy+alloc patterns and replaces them with alias reuse.
// Returns the number of aliases applied.
func (ar *AliasReuse) Optimize(funcNodeIdx uint32) int {
	count := 0
	ar.walkBlocks(funcNodeIdx, &count)
	return count
}

func (ar *AliasReuse) walkBlocks(nodeIdx uint32, count *int) {
	node := &ar.ast.Nodes[nodeIdx]

	if node.Kind == ast.NodeBlock {
		*count += ar.optimizeBlock(nodeIdx)
	}

	child := node.FirstChild
	for child != 0 {
		ar.walkBlocks(child, count)
		child = ar.ast.Nodes[child].NextSibling
	}
}

// optimizeBlock scans a block's children for DestroyStmt followed by VarDecl
// of the same type, and replaces the pair with an AliasStmt.
func (ar *AliasReuse) optimizeBlock(blockIdx uint32) int {
	count := 0

	child := ar.ast.Nodes[blockIdx].FirstChild
	for child != 0 {
		childNode := &ar.ast.Nodes[child]
		nextSibling := childNode.NextSibling

		if childNode.Kind == ast.NodeDestroyStmt && nextSibling != 0 {
			nextNode := &ar.ast.Nodes[nextSibling]
			if nextNode.Kind == ast.NodeVarDecl {
				destroySym := childNode.Payload
				allocSym := nextNode.Payload

				if ar.canReuse(destroySym, allocSym, nextSibling) {
					// Replace DestroyStmt with AliasStmt
					childNode.Kind = ast.NodeAliasStmt
					// Payload stores the "from" sym (destroyed value)
					// ExtraIdx stores the "to" sym (new allocation)
					childNode.Payload = destroySym
					childNode.ExtraIdx = allocSym

					// Clear EscapesToHeap on the VarDecl since it's now reused
					nextNode.Flags &^= ast.FlagEscapesToHeap

					// Add ReusedBy edge to ConnectionGraph
					if ar.cg != nil {
						if fromNode, ok := ar.cg.NodeOfSym(destroySym); ok {
							if toNode, ok2 := ar.cg.NodeOfSym(allocSym); ok2 {
								ar.cg.AddEdge(fromNode, toNode, EdgeReusedBy)
							}
						}
					}

					count++
				}
			}
		}

		child = nextSibling
	}

	return count
}

// canReuse checks whether a destroy+alloc pair can be replaced with alias reuse.
func (ar *AliasReuse) canReuse(destroySym, allocSym uint32, allocNodeIdx uint32) bool {
	if destroySym == 0 || allocSym == 0 {
		return false
	}

	// Both symbols must exist
	if int(destroySym) >= len(ar.st.Symbols) || int(allocSym) >= len(ar.st.Symbols) {
		return false
	}

	destroyInfo := ar.st.SymbolAt(destroySym)
	allocInfo := ar.st.SymbolAt(allocSym)

	// Same type required
	if destroyInfo.TypeID != allocInfo.TypeID {
		return false
	}

	// Alloc must be heap-allocated
	allocNode := &ar.ast.Nodes[allocNodeIdx]
	if allocNode.Flags&ast.FlagEscapesToHeap == 0 {
		return false
	}

	// No outstanding borrows of the destroyed value
	if ar.cg != nil {
		if destroyNode, ok := ar.cg.NodeOfSym(destroySym); ok {
			borrows := ar.cg.InEdges(destroyNode, EdgeBorrows)
			if len(borrows) > 0 {
				return false // active borrows → unsafe to reuse
			}
		}
	}

	return true
}
