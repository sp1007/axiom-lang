package parser_test

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
)

func requireNoErrors(t *testing.T, diags []diagnostics.Diagnostic) {
	t.Helper()
	for _, d := range diags {
		if d.Severity == diagnostics.SeverityError {
			t.Fatalf("unexpected error: %s (offset %d)", d.Message, d.Pos.Offset)
		}
	}
}

func lex(src string) []lexer.Token {
	toks, _, _ := lexer.Lex([]byte(src))
	return toks
}

func TestParseFuncDecl(t *testing.T) {
	src := "fn main():\n    return\n"
	toks := lex(src)
	pool := ast.NewInternPool(16)
	tree, diags := parser.Parse(toks, []byte(src), pool)
	requireNoErrors(t, diags)

	children := tree.Children(0)
	if len(children) != 1 {
		t.Fatalf("expected 1 top-level decl, got %d", len(children))
	}
	node := tree.Node(children[0])
	if node.Kind != ast.NodeFuncDecl {
		t.Errorf("expected NodeFuncDecl, got %s", node.Kind)
	}
	if node.Flags&ast.FlagIsPub != 0 {
		t.Error("expected no FlagIsPub on non-pub fn")
	}
	// Verify intern pool has "main"
	name := pool.Get(node.Payload)
	if name != "main" {
		t.Errorf("expected function name 'main', got %q", name)
	}
}

func TestParsePubFuncDecl(t *testing.T) {
	src := "pub fn foo():\n    return\n"
	toks := lex(src)
	pool := ast.NewInternPool(16)
	tree, diags := parser.Parse(toks, []byte(src), pool)
	requireNoErrors(t, diags)

	children := tree.Children(0)
	if len(children) != 1 {
		t.Fatalf("expected 1 top-level decl, got %d", len(children))
	}
	node := tree.Node(children[0])
	if node.Kind != ast.NodeFuncDecl {
		t.Errorf("expected NodeFuncDecl, got %s", node.Kind)
	}
	if node.Flags&ast.FlagIsPub == 0 {
		t.Error("expected FlagIsPub to be set")
	}
}

func TestParseImportDecl(t *testing.T) {
	src := "import std.io\n"
	toks := lex(src)
	pool := ast.NewInternPool(16)
	tree, diags := parser.Parse(toks, []byte(src), pool)
	requireNoErrors(t, diags)

	children := tree.Children(0)
	if len(children) != 1 {
		t.Fatalf("expected 1 top-level decl, got %d", len(children))
	}
	node := tree.Node(children[0])
	if node.Kind != ast.NodeImportDecl {
		t.Errorf("expected NodeImportDecl, got %s", node.Kind)
	}
	// Verify path was interned as "std.io"
	path := pool.Get(node.Payload)
	if path != "std.io" {
		t.Errorf("expected import path 'std.io', got %q", path)
	}
}

func TestParseVarDecl(t *testing.T) {
	src := "fn main():\n    let x: i32 = 42\n"
	toks := lex(src)
	pool := ast.NewInternPool(16)
	tree, diags := parser.Parse(toks, []byte(src), pool)
	requireNoErrors(t, diags)

	topChildren := tree.Children(0)
	if len(topChildren) != 1 {
		t.Fatalf("expected 1 top-level decl, got %d", len(topChildren))
	}
	fn := topChildren[0]
	if tree.Node(fn).Kind != ast.NodeFuncDecl {
		t.Fatalf("expected NodeFuncDecl, got %s", tree.Node(fn).Kind)
	}

	fnChildren := tree.Children(fn)
	if len(fnChildren) == 0 {
		t.Fatal("expected function body block as child, got none")
	}
	block := fnChildren[0]
	if tree.Node(block).Kind != ast.NodeBlock {
		t.Fatalf("expected NodeBlock, got %s", tree.Node(block).Kind)
	}

	blockChildren := tree.Children(block)
	if len(blockChildren) != 1 {
		t.Fatalf("expected 1 statement in block, got %d", len(blockChildren))
	}
	varDecl := blockChildren[0]
	if tree.Node(varDecl).Kind != ast.NodeVarDecl {
		t.Errorf("expected NodeVarDecl, got %s", tree.Node(varDecl).Kind)
	}
	// Immutable binding: no FlagIsMut
	if tree.Node(varDecl).Flags&ast.FlagIsMut != 0 {
		t.Error("expected no FlagIsMut on 'let' declaration")
	}
	// Name should be "x"
	name := pool.Get(tree.Node(varDecl).Payload)
	if name != "x" {
		t.Errorf("expected var name 'x', got %q", name)
	}
}

func TestParseIfStmt(t *testing.T) {
	src := "fn f():\n    if x:\n        y()\n    else:\n        z()\n"
	toks := lex(src)
	pool := ast.NewInternPool(16)
	tree, diags := parser.Parse(toks, []byte(src), pool)
	requireNoErrors(t, diags)

	fn := tree.Children(0)[0]
	if tree.Node(fn).Kind != ast.NodeFuncDecl {
		t.Fatalf("expected NodeFuncDecl, got %s", tree.Node(fn).Kind)
	}

	block := tree.Children(fn)[0]
	if tree.Node(block).Kind != ast.NodeBlock {
		t.Fatalf("expected NodeBlock, got %s", tree.Node(block).Kind)
	}

	stmts := tree.Children(block)
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	ifStmt := stmts[0]
	if tree.Node(ifStmt).Kind != ast.NodeIfStmt {
		t.Fatalf("expected NodeIfStmt, got %s", tree.Node(ifStmt).Kind)
	}

	// IfStmt children: cond, body, elseClause
	ifChildren := tree.Children(ifStmt)
	hasElse := false
	for _, c := range ifChildren {
		if tree.Node(c).Kind == ast.NodeElseClause {
			hasElse = true
		}
	}
	if !hasElse {
		t.Error("expected ElseClause as child of IfStmt")
	}
}

func TestParseStructDecl(t *testing.T) {
	src := "struct Point:\n    x: f64\n    y: f64\n"
	toks := lex(src)
	pool := ast.NewInternPool(16)
	tree, diags := parser.Parse(toks, []byte(src), pool)
	requireNoErrors(t, diags)

	children := tree.Children(0)
	if len(children) != 1 {
		t.Fatalf("expected 1 top-level decl, got %d", len(children))
	}
	node := tree.Node(children[0])
	if node.Kind != ast.NodeStructDecl {
		t.Errorf("expected NodeStructDecl, got %s", node.Kind)
	}

	name := pool.Get(node.Payload)
	if name != "Point" {
		t.Errorf("expected struct name 'Point', got %q", name)
	}

	fields := tree.Children(children[0])
	if len(fields) != 2 {
		t.Fatalf("expected 2 field declarations, got %d", len(fields))
	}
	for i, f := range fields {
		if tree.Node(f).Kind != ast.NodeFieldDecl {
			t.Errorf("field[%d]: expected NodeFieldDecl, got %s", i, tree.Node(f).Kind)
		}
	}

	fieldNames := []string{"x", "y"}
	for i, f := range fields {
		got := pool.Get(tree.Node(f).Payload)
		if got != fieldNames[i] {
			t.Errorf("field[%d]: expected name %q, got %q", i, fieldNames[i], got)
		}
	}
}

func TestParseTreeAlwaysNonNil(t *testing.T) {
	pool := ast.NewInternPool(16)
	tree, _ := parser.Parse(nil, nil, pool)
	if tree == nil {
		t.Fatal("Parse must return non-nil tree even on nil input")
	}
	if tree.Node(0).Kind != ast.NodeProgram {
		t.Errorf("root node must be NodeProgram, got %s", tree.Node(0).Kind)
	}
}

func TestParseAsyncFuncDecl(t *testing.T) {
	src := "async fn fetch():\n    return\n"
	toks := lex(src)
	pool := ast.NewInternPool(16)
	tree, diags := parser.Parse(toks, []byte(src), pool)
	requireNoErrors(t, diags)

	children := tree.Children(0)
	if len(children) != 1 {
		t.Fatalf("expected 1 top-level decl, got %d", len(children))
	}
	node := tree.Node(children[0])
	if node.Kind != ast.NodeFuncDecl {
		t.Errorf("expected NodeFuncDecl, got %s", node.Kind)
	}
	if node.Flags&ast.FlagIsAsync == 0 {
		t.Error("expected FlagIsAsync to be set")
	}
}

func TestParseMutVarDecl(t *testing.T) {
	src := "fn f():\n    mut count: i32 = 0\n"
	toks := lex(src)
	pool := ast.NewInternPool(16)
	tree, diags := parser.Parse(toks, []byte(src), pool)
	requireNoErrors(t, diags)

	fn := tree.Children(0)[0]
	block := tree.Children(fn)[0]
	stmts := tree.Children(block)
	if len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(stmts))
	}
	varDecl := stmts[0]
	if tree.Node(varDecl).Kind != ast.NodeVarDecl {
		t.Fatalf("expected NodeVarDecl, got %s", tree.Node(varDecl).Kind)
	}
	if tree.Node(varDecl).Flags&ast.FlagIsMut == 0 {
		t.Error("expected FlagIsMut on 'mut' declaration")
	}
}

func TestParseImportWithSelectors(t *testing.T) {
	src := "import std.fs { read, write }\n"
	toks := lex(src)
	pool := ast.NewInternPool(16)
	tree, diags := parser.Parse(toks, []byte(src), pool)
	requireNoErrors(t, diags)

	node := tree.Node(tree.Children(0)[0])
	if node.Kind != ast.NodeImportDecl {
		t.Fatalf("expected NodeImportDecl, got %s", node.Kind)
	}
	path := pool.Get(node.Payload)
	if path != "std.fs" {
		t.Errorf("expected path 'std.fs', got %q", path)
	}
	// Two Ident children for the selectors
	selectors := tree.Children(tree.Children(0)[0])
	if len(selectors) != 2 {
		t.Errorf("expected 2 import selectors, got %d", len(selectors))
	}
}

func TestParseTreeValidate(t *testing.T) {
	src := "fn main():\n    let x: i32 = 42\n    return x\n"
	toks := lex(src)
	pool := ast.NewInternPool(16)
	tree, diags := parser.Parse(toks, []byte(src), pool)
	requireNoErrors(t, diags)

	errs := tree.Validate()
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("tree validation error: %s", e)
		}
	}
}
