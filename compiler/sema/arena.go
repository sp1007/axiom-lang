package sema

import (
	"fmt"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
)

// ArenaPass handles arena block semantics:
// - Marks all VarDecl nodes inside ArenaBlock with FlagUsesArena
// - Injects arena destroy at block exit
// - Verifies arena expressions (type checking deferred to type checker)
type ArenaPass struct {
	ast    *ast.AstTree
	intern *ast.InternPool
	st     *SymbolTable
	errors []diagnostics.Diagnostic
}

// NewArenaPass creates a new ArenaPass.
func NewArenaPass(tree *ast.AstTree, intern *ast.InternPool, st *SymbolTable) *ArenaPass {
	return &ArenaPass{
		ast:    tree,
		intern: intern,
		st:     st,
	}
}

// Process walks the AST and processes all ArenaBlock nodes.
// Returns diagnostics for any arena-related errors.
func (ap *ArenaPass) Process() []diagnostics.Diagnostic {
	if ap.ast == nil || ap.ast.NodeCount() == 0 {
		return ap.errors
	}
	ap.walkNode(0)
	return ap.errors
}

func (ap *ArenaPass) errorf(nodeIdx uint32, code int, format string, args ...any) {
	ap.errors = append(ap.errors, diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     uint32(code),
		Message:  fmt.Sprintf(format, args...),
		Pos:      diagnostics.Pos{},
	})
}

func (ap *ArenaPass) walkNode(nodeIdx uint32) {
	node := &ap.ast.Nodes[nodeIdx]

	if node.Kind == ast.NodeArenaBlock {
		ap.processArenaBlock(nodeIdx)
	}

	child := node.FirstChild
	for child != 0 {
		ap.walkNode(child)
		child = ap.ast.Nodes[child].NextSibling
	}
}

// processArenaBlock handles a single ArenaBlock:
// 1. Marks all VarDecl children with FlagUsesArena
// 2. Injects arena destroy at block exit
func (ap *ArenaPass) processArenaBlock(blockIdx uint32) {
	node := &ap.ast.Nodes[blockIdx]

	// The first child is the arena expression (or list of arena expressions)
	// The remaining children form the body.
	// For MVP: first child = arena ident, rest = body statements.

	// Mark all VarDecl nodes in the body with FlagUsesArena
	ap.markArenaVarDecls(blockIdx)

	// Inject arena destroy at block exit
	// The arena expression is the first child (ident or call)
	arenaExpr := node.FirstChild
	if arenaExpr != 0 {
		arenaNode := &ap.ast.Nodes[arenaExpr]
		if arenaNode.Kind == ast.NodeIdent {
			symID := arenaNode.Payload
			if symID != 0 {
				// Inject DestroyStmt for the arena itself at block exit
				destroyNode := ap.ast.AddNode(ast.NodeDestroyStmt, 0)
				ap.ast.SetPayload(destroyNode, symID)
				ap.ast.AppendChild(blockIdx, destroyNode)
			}
		}
	}
}

// markArenaVarDecls recursively marks all VarDecl nodes under nodeIdx
// with FlagUsesArena.
func (ap *ArenaPass) markArenaVarDecls(nodeIdx uint32) {
	node := &ap.ast.Nodes[nodeIdx]

	if node.Kind == ast.NodeVarDecl {
		ap.ast.SetFlags(nodeIdx, ast.FlagUsesArena)
	}

	child := node.FirstChild
	for child != 0 {
		ap.markArenaVarDecls(child)
		child = ap.ast.Nodes[child].NextSibling
	}
}

// ArenaVarDeclCount returns the number of VarDecl nodes with FlagUsesArena
// under the given node.
func (ap *ArenaPass) ArenaVarDeclCount(nodeIdx uint32) int {
	count := 0
	ap.countArenaVars(nodeIdx, &count)
	return count
}

func (ap *ArenaPass) countArenaVars(nodeIdx uint32, count *int) {
	node := &ap.ast.Nodes[nodeIdx]

	if node.Kind == ast.NodeVarDecl && node.Flags&ast.FlagUsesArena != 0 {
		*count++
	}

	child := node.FirstChild
	for child != 0 {
		ap.countArenaVars(child, count)
		child = ap.ast.Nodes[child].NextSibling
	}
}
