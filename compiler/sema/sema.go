package sema

import (
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
)

// nodePos computes the diagnostics.Pos for a given AST node index.
func nodePos(tree *ast.AstTree, nodeIdx uint32) diagnostics.Pos {
	if tree == nil || int(nodeIdx) >= len(tree.Nodes) {
		return diagnostics.Pos{}
	}
	node := tree.Nodes[nodeIdx]
	if int(node.TokenIdx) >= len(tree.Tokens) {
		return diagnostics.Pos{}
	}
	tok := tree.Tokens[node.TokenIdx]
	offset := tok.Offset
	
	// Compute line and col from offset in tree.Source
	var line uint32 = 1
	var col uint32 = 1
	for i := uint32(0); i < offset && i < uint32(len(tree.Source)); i++ {
		if tree.Source[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return diagnostics.Pos{
		Offset: offset,
		Line:   line,
		Col:    col,
	}
}
