package ast

// VisitFn is called for each node during tree traversal.
// Return false to stop traversal.
type VisitFn func(tree *AstTree, nodeIdx uint32) bool

// WalkPreOrder visits every reachable node in pre-order (parent before children).
// Traversal starts from root and visits depth-first.
func WalkPreOrder(tree *AstTree, root uint32, visit VisitFn) {
	walkPreOrder(tree, root, visit)
}

func walkPreOrder(tree *AstTree, idx uint32, visit VisitFn) bool {
	if !visit(tree, idx) {
		return false
	}
	child := tree.Nodes[idx].FirstChild
	for child != NullIdx {
		if !walkPreOrder(tree, child, visit) {
			return false
		}
		child = tree.Nodes[child].NextSibling
	}
	return true
}

// WalkPostOrder visits every reachable node in post-order (children before parent).
func WalkPostOrder(tree *AstTree, root uint32, visit VisitFn) {
	walkPostOrder(tree, root, visit)
}

func walkPostOrder(tree *AstTree, idx uint32, visit VisitFn) bool {
	child := tree.Nodes[idx].FirstChild
	for child != NullIdx {
		if !walkPostOrder(tree, child, visit) {
			return false
		}
		child = tree.Nodes[child].NextSibling
	}
	return visit(tree, idx)
}

// WalkChildren visits only the direct children of the given node.
func WalkChildren(tree *AstTree, parent uint32, visit VisitFn) {
	child := tree.Nodes[parent].FirstChild
	for child != NullIdx {
		if !visit(tree, child) {
			return
		}
		child = tree.Nodes[child].NextSibling
	}
}

// ReachableCount returns the number of reachable nodes from root.
// Useful for validation — should equal NodeCount() if tree is well-formed.
func ReachableCount(tree *AstTree, root uint32) int {
	count := 0
	WalkPreOrder(tree, root, func(_ *AstTree, _ uint32) bool {
		count++
		return true
	})
	return count
}
