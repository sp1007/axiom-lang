package ast

import (
	"fmt"

	"github.com/axiom-lang/axiom/compiler/lexer"
)

// NullIdx is the sentinel value for "no node".
// It also happens to be the root node index, which is intentional:
// no node may be the child of another node AND the root simultaneously.
const NullIdx uint32 = 0

// AstTree holds all AST nodes for a single compilation unit.
// Nodes live in a flat slice; tree structure is encoded via index fields.
// The zero value is not valid; use NewTree to create.
type AstTree struct {
	Nodes  []AstNode    // all nodes; index 0 is always NodeProgram
	Extras []uint32     // overflow storage for nodes with many sub-fields
	Source []byte       // original source bytes (zero-copy token text)
	Tokens []lexer.Token // token slice (for token text lookup)
}

// NewTree creates a new AstTree with a root NodeProgram at index 0.
func NewTree(source []byte, tokens []lexer.Token) *AstTree {
	t := &AstTree{
		Nodes:  make([]AstNode, 0, 256),
		Extras: make([]uint32, 0, 64),
		Source: source,
		Tokens: tokens,
	}
	// Root node: NodeProgram at index 0
	t.Nodes = append(t.Nodes, AstNode{Kind: NodeProgram})
	return t
}

// AddNode appends a new node and returns its index.
func (t *AstTree) AddNode(kind NodeKind, tokenIdx uint32) uint32 {
	idx := uint32(len(t.Nodes))
	t.Nodes = append(t.Nodes, AstNode{Kind: kind, TokenIdx: tokenIdx})
	return idx
}

// Node returns a pointer to the node at the given index.
// NOTE: The returned pointer is invalidated if t.Nodes is reallocated
// (e.g., by AddNode). Callers must not cache pointers across AddNode calls.
func (t *AstTree) Node(idx uint32) *AstNode {
	return &t.Nodes[idx]
}

// SetFirstChild sets the first child of parent.
func (t *AstTree) SetFirstChild(parent, child uint32) {
	t.Nodes[parent].FirstChild = child
}

// SetNextSibling sets the next sibling of node.
func (t *AstTree) SetNextSibling(node, sibling uint32) {
	t.Nodes[node].NextSibling = sibling
}

// SetPayload sets the Payload field of the given node.
func (t *AstTree) SetPayload(node uint32, payload uint32) {
	t.Nodes[node].Payload = payload
}

// SetFlags sets (ORs in) the given flags on the node.
func (t *AstTree) SetFlags(node uint32, flags uint16) {
	t.Nodes[node].Flags |= flags
}

// ClearFlags clears (AND-NOTs) the given flags on the node.
func (t *AstTree) ClearFlags(node uint32, flags uint16) {
	t.Nodes[node].Flags &^= flags
}

// AddExtra appends overflow data and returns the start index.
// A node uses ExtraIdx to point here; the first word is typically a count.
func (t *AstTree) AddExtra(values ...uint32) uint32 {
	idx := uint32(len(t.Extras))
	t.Extras = append(t.Extras, values...)
	return idx
}

// TokenText recovers the source text for the given token index.
func (t *AstTree) TokenText(tokenIdx uint32) []byte {
	tok := t.Tokens[tokenIdx]
	return t.Source[tok.Offset : tok.Offset+uint32(tok.Len)]
}

// NodeText returns the source text for the node's primary token.
func (t *AstTree) NodeText(nodeIdx uint32) []byte {
	return t.TokenText(t.Nodes[nodeIdx].TokenIdx)
}

// Children collects all direct child indices of the given node.
func (t *AstTree) Children(nodeIdx uint32) []uint32 {
	var children []uint32
	child := t.Nodes[nodeIdx].FirstChild
	for child != NullIdx {
		children = append(children, child)
		child = t.Nodes[child].NextSibling
	}
	return children
}

// AppendChild appends a child at the end of the parent's child list.
// NOTE: O(n) in the number of existing children. For hot paths, the parser
// should track the last child explicitly using SetNextSibling directly.
func (t *AstTree) AppendChild(parent, child uint32) {
	if t.Nodes[parent].FirstChild == NullIdx {
		t.Nodes[parent].FirstChild = child
		return
	}
	cur := t.Nodes[parent].FirstChild
	for t.Nodes[cur].NextSibling != NullIdx {
		cur = t.Nodes[cur].NextSibling
	}
	t.Nodes[cur].NextSibling = child
}

// NodeCount returns the total number of nodes.
func (t *AstTree) NodeCount() int { return len(t.Nodes) }

// ExtraCount returns the total number of extra uint32 values stored.
func (t *AstTree) ExtraCount() int { return len(t.Extras) }

// Validate checks tree invariants. Returns a list of errors (nil if valid).
// Used in debug builds for sanity checking after parsing.
func (t *AstTree) Validate() []string {
	var errors []string
	if len(t.Nodes) == 0 {
		return []string{"tree has no nodes (missing root)"}
	}
	if t.Nodes[0].Kind != NodeProgram {
		errors = append(errors, "node[0] is not NodeProgram")
	}
	for i, n := range t.Nodes {
		if n.FirstChild != NullIdx && int(n.FirstChild) >= len(t.Nodes) {
			errors = append(errors, fmt.Sprintf("node[%d].FirstChild=%d out of bounds (max=%d)", i, n.FirstChild, len(t.Nodes)-1))
		}
		if n.NextSibling != NullIdx && int(n.NextSibling) >= len(t.Nodes) {
			errors = append(errors, fmt.Sprintf("node[%d].NextSibling=%d out of bounds (max=%d)", i, n.NextSibling, len(t.Nodes)-1))
		}
	}
	return errors
}

// CloneSubtree creates a deep copy of the subtree rooted at nodeIdx.
func (t *AstTree) CloneSubtree(nodeIdx uint32) uint32 {
	if nodeIdx == NullIdx {
		return NullIdx
	}

	orig := t.Nodes[nodeIdx]
	newIdx := t.AddNode(orig.Kind, orig.TokenIdx)
	
	// AstTree might have reallocated, so re-fetch if using pointers,
	// but we use indices so we're safe.
	t.Nodes[newIdx].Payload = orig.Payload
	t.Nodes[newIdx].Flags = orig.Flags
	t.Nodes[newIdx].ExtraIdx = orig.ExtraIdx
	
	// Clone children
	child := orig.FirstChild
	if child != NullIdx {
		firstClone := t.CloneSubtree(child)
		t.Nodes[newIdx].FirstChild = firstClone
		
		prevClone := firstClone
		child = t.Nodes[child].NextSibling
		for child != NullIdx {
			siblingClone := t.CloneSubtree(child)
			t.Nodes[prevClone].NextSibling = siblingClone
			prevClone = siblingClone
			child = t.Nodes[child].NextSibling
		}
	}

	return newIdx
}
