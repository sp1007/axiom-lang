package ast

import (
	"strings"
	"testing"
)

func TestPrinterEmptyProgram(t *testing.T) {
	tree := NewTree(nil, nil)
	got := PrintToString(tree, nil)
	if !strings.Contains(got, "Program") {
		t.Errorf("expected 'Program' in output, got:\n%s", got)
	}
}

func TestPrinterFuncDecl(t *testing.T) {
	tree := NewTree(nil, nil)
	fnIdx := tree.AddNode(NodeFuncDecl, 0)
	tree.AppendChild(0, fnIdx)
	tree.SetFlags(fnIdx, FlagIsPub)
	got := PrintToString(tree, nil)
	if !strings.Contains(got, "FuncDecl") {
		t.Error("expected FuncDecl")
	}
	if !strings.Contains(got, "[pub]") {
		t.Errorf("expected [pub] flag in output:\n%s", got)
	}
}

func TestPrinterNestedBlocks(t *testing.T) {
	tree := NewTree(nil, nil)
	fn := tree.AddNode(NodeFuncDecl, 0)
	block := tree.AddNode(NodeBlock, 0)
	stmt := tree.AddNode(NodeReturnStmt, 0)
	tree.AppendChild(0, fn)
	tree.AppendChild(fn, block)
	tree.AppendChild(block, stmt)
	got := PrintToString(tree, nil)
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) < 4 {
		t.Fatalf("expected ≥4 lines, got %d:\n%s", len(lines), got)
	}
	if !strings.HasPrefix(lines[1], "  FuncDecl") {
		t.Errorf("line 1: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "    Block") {
		t.Errorf("line 2: %q", lines[2])
	}
	if !strings.HasPrefix(lines[3], "      Ret") {
		t.Errorf("line 3: %q", lines[3])
	}
}

func TestPrinterDeterministic(t *testing.T) {
	tree := NewTree(nil, nil)
	for i := 0; i < 10; i++ {
		idx := tree.AddNode(NodeFuncDecl, 0)
		tree.AppendChild(0, idx)
	}
	out1 := PrintToString(tree, nil)
	out2 := PrintToString(tree, nil)
	if out1 != out2 {
		t.Error("printer is not deterministic")
	}
}

func TestNodeKindStringAll(t *testing.T) {
	for k := NodeKind(0); k < NodeKindCount; k++ {
		s := k.String()
		if s == "" {
			t.Errorf("NodeKind(%d).String() is empty", k)
		}
	}
}

func TestNodeKindStringSpecific(t *testing.T) {
	if NodeProgram.String() != "Program" {
		t.Errorf("NodeProgram.String() = %q", NodeProgram.String())
	}
	if NodeFuncDecl.String() != "FuncDecl" {
		t.Errorf("NodeFuncDecl.String() = %q", NodeFuncDecl.String())
	}
	if NodeError.String() != "Error" {
		t.Errorf("NodeError.String() = %q", NodeError.String())
	}
}

func TestPrinterPayload(t *testing.T) {
	tree := NewTree(nil, nil)
	idx := tree.AddNode(NodeFuncDecl, 0)
	tree.AppendChild(0, idx)
	tree.SetPayload(idx, 42)
	got := PrintToString(tree, nil)
	if !strings.Contains(got, "@42") {
		t.Errorf("expected @42 in output:\n%s", got)
	}
}

func TestPrinterMultipleFlags(t *testing.T) {
	tree := NewTree(nil, nil)
	idx := tree.AddNode(NodeFuncDecl, 0)
	tree.AppendChild(0, idx)
	tree.SetFlags(idx, FlagIsPub|FlagIsAsync|FlagIsExtern)
	got := PrintToString(tree, nil)
	if !strings.Contains(got, "pub") || !strings.Contains(got, "async") || !strings.Contains(got, "extern") {
		t.Errorf("expected all flags in output:\n%s", got)
	}
}
