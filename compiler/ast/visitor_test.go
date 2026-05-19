package ast

import "testing"

func TestWalkPreOrder(t *testing.T) {
	tree := NewTree(nil, nil)
	a := tree.AddNode(NodeFuncDecl, 0)
	b := tree.AddNode(NodeVarDecl, 0)
	c := tree.AddNode(NodeReturnStmt, 0)
	tree.SetFirstChild(0, a)
	tree.SetFirstChild(a, b)
	tree.SetNextSibling(b, c)

	// Pre-order: root(0), a, b, c
	var visited []uint32
	WalkPreOrder(tree, 0, func(_ *AstTree, idx uint32) bool {
		visited = append(visited, idx)
		return true
	})

	expected := []uint32{0, a, b, c}
	if len(visited) != len(expected) {
		t.Fatalf("expected %d visits, got %d", len(expected), len(visited))
	}
	for i, v := range visited {
		if v != expected[i] {
			t.Errorf("visited[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

func TestWalkPostOrder(t *testing.T) {
	tree := NewTree(nil, nil)
	a := tree.AddNode(NodeFuncDecl, 0)
	b := tree.AddNode(NodeVarDecl, 0)
	c := tree.AddNode(NodeReturnStmt, 0)
	tree.SetFirstChild(0, a)
	tree.SetFirstChild(a, b)
	tree.SetNextSibling(b, c)

	// Post-order: b, c, a, root(0)
	var visited []uint32
	WalkPostOrder(tree, 0, func(_ *AstTree, idx uint32) bool {
		visited = append(visited, idx)
		return true
	})

	expected := []uint32{b, c, a, 0}
	if len(visited) != len(expected) {
		t.Fatalf("expected %d visits, got %d", len(expected), len(visited))
	}
	for i, v := range visited {
		if v != expected[i] {
			t.Errorf("visited[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

func TestWalkPreOrderEarlyTermination(t *testing.T) {
	tree := NewTree(nil, nil)
	a := tree.AddNode(NodeFuncDecl, 0)
	b := tree.AddNode(NodeVarDecl, 0)
	tree.SetFirstChild(0, a)
	tree.SetNextSibling(a, b)

	count := 0
	WalkPreOrder(tree, 0, func(_ *AstTree, _ uint32) bool {
		count++
		return count < 2 // stop after 2 visits
	})

	if count != 2 {
		t.Errorf("expected 2 visits with early termination, got %d", count)
	}
}

func TestWalkChildren(t *testing.T) {
	tree := NewTree(nil, nil)
	a := tree.AddNode(NodeFuncDecl, 0)
	b := tree.AddNode(NodeStructDecl, 0)
	c := tree.AddNode(NodeConstDecl, 0)
	tree.AppendChild(0, a)
	tree.AppendChild(0, b)
	tree.AppendChild(0, c)

	// Add grandchild to a — should NOT be visited
	d := tree.AddNode(NodeVarDecl, 0)
	tree.AppendChild(a, d)

	var visited []uint32
	WalkChildren(tree, 0, func(_ *AstTree, idx uint32) bool {
		visited = append(visited, idx)
		return true
	})

	if len(visited) != 3 {
		t.Fatalf("expected 3 children visited, got %d", len(visited))
	}
	// d should not be in visited
	for _, v := range visited {
		if v == d {
			t.Error("grandchild should not be visited by WalkChildren")
		}
	}
}

func TestReachableCount(t *testing.T) {
	tree := NewTree(nil, nil)
	a := tree.AddNode(NodeFuncDecl, 0)
	b := tree.AddNode(NodeVarDecl, 0)
	tree.SetFirstChild(0, a)
	tree.SetFirstChild(a, b)

	count := ReachableCount(tree, 0)
	if count != 3 {
		t.Errorf("expected 3 reachable nodes, got %d", count)
	}
}

func TestWalkChildrenEarlyStop(t *testing.T) {
	tree := NewTree(nil, nil)
	for i := 0; i < 5; i++ {
		tree.AppendChild(0, tree.AddNode(NodeVarDecl, 0))
	}

	count := 0
	WalkChildren(tree, 0, func(_ *AstTree, _ uint32) bool {
		count++
		return count < 3
	})
	if count != 3 {
		t.Errorf("expected 3 visits with early stop, got %d", count)
	}
}
