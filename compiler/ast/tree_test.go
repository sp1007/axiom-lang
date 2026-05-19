package ast

import "testing"

func TestNewTreeRootNode(t *testing.T) {
	tree := NewTree(nil, nil)
	if tree.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", tree.NodeCount())
	}
	if tree.Nodes[0].Kind != NodeProgram {
		t.Fatalf("root must be NodeProgram, got %d", tree.Nodes[0].Kind)
	}
}

func TestAddNode(t *testing.T) {
	tree := NewTree(nil, nil)
	idx := tree.AddNode(NodeFuncDecl, 0)
	if idx != 1 {
		t.Fatalf("expected idx=1, got %d", idx)
	}
	if tree.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes, got %d", tree.NodeCount())
	}
	if tree.Node(idx).Kind != NodeFuncDecl {
		t.Errorf("expected NodeFuncDecl, got %d", tree.Node(idx).Kind)
	}
}

func TestAppendChild(t *testing.T) {
	tree := NewTree(nil, nil)
	child1 := tree.AddNode(NodeFuncDecl, 0)
	child2 := tree.AddNode(NodeStructDecl, 0)
	tree.AppendChild(0, child1)
	tree.AppendChild(0, child2)
	children := tree.Children(0)
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if children[0] != child1 {
		t.Errorf("first child = %d, want %d", children[0], child1)
	}
	if children[1] != child2 {
		t.Errorf("second child = %d, want %d", children[1], child2)
	}
}

func TestChildrenOrder(t *testing.T) {
	tree := NewTree(nil, nil)
	indices := make([]uint32, 10)
	for i := 0; i < 10; i++ {
		indices[i] = tree.AddNode(NodeVarDecl, uint32(i))
		tree.AppendChild(0, indices[i])
	}
	children := tree.Children(0)
	if len(children) != 10 {
		t.Fatalf("expected 10 children, got %d", len(children))
	}
	for i, idx := range children {
		if idx != indices[i] {
			t.Errorf("child[%d] = %d, want %d", i, idx, indices[i])
		}
	}
}

func TestSetFlags(t *testing.T) {
	tree := NewTree(nil, nil)
	idx := tree.AddNode(NodeFuncDecl, 0)
	tree.SetFlags(idx, FlagIsPub|FlagIsAsync)
	n := tree.Node(idx)
	if n.Flags&FlagIsPub == 0 {
		t.Error("expected FlagIsPub set")
	}
	if n.Flags&FlagIsAsync == 0 {
		t.Error("expected FlagIsAsync set")
	}
	tree.ClearFlags(idx, FlagIsPub)
	if tree.Node(idx).Flags&FlagIsPub != 0 {
		t.Error("expected FlagIsPub cleared")
	}
	if tree.Node(idx).Flags&FlagIsAsync == 0 {
		t.Error("FlagIsAsync should still be set")
	}
}

func TestSetPayload(t *testing.T) {
	tree := NewTree(nil, nil)
	idx := tree.AddNode(NodeIntLit, 0)
	tree.SetPayload(idx, 42)
	if tree.Node(idx).Payload != 42 {
		t.Errorf("Payload = %d, want 42", tree.Node(idx).Payload)
	}
}

func TestAddExtra(t *testing.T) {
	tree := NewTree(nil, nil)
	idx := tree.AddExtra(10, 20, 30)
	if idx != 0 {
		t.Fatalf("expected Extra idx=0, got %d", idx)
	}
	if tree.Extras[0] != 10 || tree.Extras[1] != 20 || tree.Extras[2] != 30 {
		t.Error("extra values not stored correctly")
	}
	if tree.ExtraCount() != 3 {
		t.Errorf("ExtraCount = %d, want 3", tree.ExtraCount())
	}

	// Second extra group
	idx2 := tree.AddExtra(40, 50)
	if idx2 != 3 {
		t.Fatalf("expected Extra idx=3, got %d", idx2)
	}
}

func TestValidate(t *testing.T) {
	tree := NewTree(nil, nil)
	errs := tree.Validate()
	if len(errs) != 0 {
		t.Fatalf("fresh tree has validation errors: %v", errs)
	}

	// Corrupt a child index
	child := tree.AddNode(NodeFuncDecl, 0)
	tree.Nodes[child].FirstChild = 9999 // out of bounds
	errs = tree.Validate()
	if len(errs) == 0 {
		t.Error("expected validation error for out-of-bounds child")
	}
}

func TestNullIdxIsSentinel(t *testing.T) {
	if NullIdx != 0 {
		t.Fatal("NullIdx must be 0")
	}
}

func TestChildrenEmpty(t *testing.T) {
	tree := NewTree(nil, nil)
	children := tree.Children(0)
	if len(children) != 0 {
		t.Errorf("expected 0 children, got %d", len(children))
	}
}

func TestLargeTree(t *testing.T) {
	tree := NewTree(nil, nil)
	for i := 0; i < 10000; i++ {
		idx := tree.AddNode(NodeVarDecl, 0)
		tree.AppendChild(0, idx)
	}
	if tree.NodeCount() != 10001 {
		t.Errorf("expected 10001 nodes, got %d", tree.NodeCount())
	}
	errs := tree.Validate()
	if len(errs) != 0 {
		t.Fatalf("validation errors: %v", errs)
	}
}

func TestSetFirstChildAndSibling(t *testing.T) {
	tree := NewTree(nil, nil)
	a := tree.AddNode(NodeFuncDecl, 0)
	b := tree.AddNode(NodeStructDecl, 0)
	c := tree.AddNode(NodeConstDecl, 0)

	tree.SetFirstChild(0, a)
	tree.SetNextSibling(a, b)
	tree.SetNextSibling(b, c)

	children := tree.Children(0)
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}
}
