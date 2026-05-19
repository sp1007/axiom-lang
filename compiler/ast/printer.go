package ast

import (
	"fmt"
	"io"
	"strings"
)

// Printer renders an AstTree as human-readable indented text.
type Printer struct {
	w      io.Writer
	tree   *AstTree
	pool   *InternPool // may be nil
	indent int
}

// Print writes the entire AST to w.
// pool may be nil (identifiers shown as raw token text).
func Print(w io.Writer, tree *AstTree, pool *InternPool) {
	p := &Printer{w: w, tree: tree, pool: pool}
	p.printNode(0)
}

// PrintNode prints a single subtree rooted at nodeIdx.
func PrintNode(w io.Writer, tree *AstTree, nodeIdx uint32, pool *InternPool) {
	p := &Printer{w: w, tree: tree, pool: pool}
	p.printNode(nodeIdx)
}

// PrintToString renders the entire AST as a string.
func PrintToString(tree *AstTree, pool *InternPool) string {
	var sb strings.Builder
	Print(&sb, tree, pool)
	return sb.String()
}

// printNode recursively prints a node and all its children.
func (p *Printer) printNode(idx uint32) {
	node := p.tree.Node(idx)

	// Write indentation
	for i := 0; i < p.indent; i++ {
		fmt.Fprint(p.w, "  ")
	}

	// Node kind
	fmt.Fprint(p.w, node.Kind.String())

	// Flags
	p.writeFlags(node)

	// Token-based info
	p.writeTokenInfo(node)

	// Payload
	if node.Payload != 0 {
		fmt.Fprintf(p.w, " @%d", node.Payload)
	}

	fmt.Fprintln(p.w)

	// Recurse into children
	p.indent++
	child := node.FirstChild
	for child != NullIdx {
		p.printNode(child)
		child = p.tree.Node(child).NextSibling
	}
	p.indent--
}

// writeFlags prints active flags as [flag1 flag2 ...].
func (p *Printer) writeFlags(node *AstNode) {
	type flagInfo struct {
		mask uint16
		name string
	}
	flags := []flagInfo{
		{FlagIsPub, "pub"}, {FlagIsMut, "mut"}, {FlagIsAsync, "async"},
		{FlagIsExtern, "extern"}, {FlagIsSink, "sink"}, {FlagIsLent, "lent"},
		{FlagIsPacked, "packed"}, {FlagEscapesToHeap, "heap"},
		{FlagUsesArena, "arena"}, {FlagIsGeneric, "generic"}, {FlagIsMoved, "moved"},
	}
	var active []string
	for _, f := range flags {
		if node.Flags&f.mask != 0 {
			active = append(active, f.name)
		}
	}
	if len(active) > 0 {
		fmt.Fprintf(p.w, " [%s]", strings.Join(active, " "))
	}
}

// writeTokenInfo prints token-based information per node kind.
func (p *Printer) writeTokenInfo(node *AstNode) {
	if p.tree.Tokens == nil {
		return
	}
	if node.TokenIdx == 0 && node.Kind != NodeProgram {
		return
	}
	if int(node.TokenIdx) >= len(p.tree.Tokens) {
		return
	}

	text := string(p.tree.TokenText(node.TokenIdx))
	switch node.Kind {
	case NodeIdent, NodeFuncDecl, NodeStructDecl, NodeInterfaceDecl,
		NodeParamDecl, NodeFieldDecl, NodeVariantDecl, NodeTypeAliasDecl:
		fmt.Fprintf(p.w, " name=%q", text)
	case NodeIntLit, NodeFloatLit, NodeBoolLit, NodeNilLit:
		fmt.Fprintf(p.w, " value=%q", text)
	case NodeStringLit, NodeCharLit:
		if len(text) > 40 {
			text = text[:40] + "..."
		}
		fmt.Fprintf(p.w, " value=%q", text)
	case NodeBinaryExpr, NodeUnaryExpr, NodeAssignStmt:
		fmt.Fprintf(p.w, " op=%q", text)
	}
}
