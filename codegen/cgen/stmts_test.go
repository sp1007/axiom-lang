package cgen_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// stmtTestCtx builds a minimal context for statement/expression generation tests.
type stmtTestCtx struct {
	tree    *ast.AstTree
	intern  *ast.InternPool
	symbols *sema.SymbolTable
	table   *types.TypeTable
}

func newStmtTestCtx() *stmtTestCtx {
	intern := ast.NewInternPool(64)
	table := types.NewTypeTable()
	symbols := sema.NewSymbolTable(intern)

	// Build a source with commonly needed tokens
	source := []byte(
		"x 5 y 10 true false 3.14 hello + - * / == != and or not return if cond body " +
			"while for i list foo bar a b 42 0 defer cleanup",
	)
	tokens := []lexer.Token{
		{Kind: lexer.TokenIdent, Offset: 0, Len: 1},    // 0: "x"
		{Kind: lexer.TokenIntLit, Offset: 2, Len: 1},    // 1: "5"
		{Kind: lexer.TokenIdent, Offset: 4, Len: 1},     // 2: "y"
		{Kind: lexer.TokenIntLit, Offset: 6, Len: 2},    // 3: "10"
		{Kind: lexer.TokenTrue, Offset: 9, Len: 4},      // 4: "true"
		{Kind: lexer.TokenFalse, Offset: 14, Len: 5},    // 5: "false"
		{Kind: lexer.TokenFloatLit, Offset: 20, Len: 4}, // 6: "3.14"
		{Kind: lexer.TokenStringLit, Offset: 25, Len: 5},// 7: "hello" (no quotes in token)
		{Kind: lexer.TokenPlus, Offset: 31, Len: 1},     // 8: "+"
		{Kind: lexer.TokenMinus, Offset: 33, Len: 1},    // 9: "-"
		{Kind: lexer.TokenStar, Offset: 35, Len: 1},     // 10: "*"
		{Kind: lexer.TokenSlash, Offset: 37, Len: 1},    // 11: "/"
		{Kind: lexer.TokenEqEq, Offset: 39, Len: 2},     // 12: "=="
		{Kind: lexer.TokenBangEq, Offset: 42, Len: 2},   // 13: "!="
		{Kind: lexer.TokenAnd, Offset: 45, Len: 3},      // 14: "and"
		{Kind: lexer.TokenOr, Offset: 49, Len: 2},       // 15: "or"
		{Kind: lexer.TokenNot, Offset: 52, Len: 3},      // 16: "not"
		{Kind: lexer.TokenReturn, Offset: 56, Len: 6},   // 17: "return"
		{Kind: lexer.TokenIf, Offset: 63, Len: 2},       // 18: "if"
		{Kind: lexer.TokenIdent, Offset: 66, Len: 4},    // 19: "cond"
		{Kind: lexer.TokenIdent, Offset: 71, Len: 4},    // 20: "body"
		{Kind: lexer.TokenWhile, Offset: 76, Len: 5},    // 21: "while"
		{Kind: lexer.TokenFor, Offset: 82, Len: 3},      // 22: "for"
		{Kind: lexer.TokenIdent, Offset: 86, Len: 1},    // 23: "i"
		{Kind: lexer.TokenIdent, Offset: 88, Len: 4},    // 24: "list"
		{Kind: lexer.TokenIdent, Offset: 93, Len: 3},    // 25: "foo"
		{Kind: lexer.TokenIdent, Offset: 97, Len: 3},    // 26: "bar"
		{Kind: lexer.TokenIdent, Offset: 101, Len: 1},   // 27: "a"
		{Kind: lexer.TokenIdent, Offset: 103, Len: 1},   // 28: "b"
		{Kind: lexer.TokenIntLit, Offset: 105, Len: 2},  // 29: "42"
		{Kind: lexer.TokenIntLit, Offset: 108, Len: 1},  // 30: "0"
		{Kind: lexer.TokenDefer, Offset: 110, Len: 5},   // 31: "defer"
		{Kind: lexer.TokenIdent, Offset: 116, Len: 7},   // 32: "cleanup"
	}

	tree := ast.NewTree(source, tokens)
	return &stmtTestCtx{tree: tree, intern: intern, symbols: symbols, table: table}
}

func (c *stmtTestCtx) makeStmtGen() (*cgen.StmtGen, *bytes.Buffer) {
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	queue := cgen.NewTypeDeclQueue()
	sg := cgen.NewStmtGen(w, c.table, c.intern, c.symbols, c.tree, queue)
	return sg, &buf
}

// === Expression Tests ===

func TestExprGen_IntLit(t *testing.T) {
	ctx := newStmtTestCtx()
	queue := cgen.NewTypeDeclQueue()
	eg := cgen.NewExprGen(ctx.table, ctx.intern, ctx.symbols, ctx.tree, queue)

	// Token 29 = "42"
	node := ctx.tree.AddNode(ast.NodeIntLit, 29)
	ctx.tree.Node(node).Payload = uint32(types.TypeI32)

	got := eg.Emit(node)
	if got != "42" {
		t.Errorf("IntLit = %q, want \"42\"", got)
	}
}

func TestExprGen_BoolLit(t *testing.T) {
	ctx := newStmtTestCtx()
	queue := cgen.NewTypeDeclQueue()
	eg := cgen.NewExprGen(ctx.table, ctx.intern, ctx.symbols, ctx.tree, queue)

	trueNode := ctx.tree.AddNode(ast.NodeBoolLit, 4) // "true"
	falseNode := ctx.tree.AddNode(ast.NodeBoolLit, 5) // "false"

	if got := eg.Emit(trueNode); got != "AX_TRUE" {
		t.Errorf("true = %q, want \"AX_TRUE\"", got)
	}
	if got := eg.Emit(falseNode); got != "AX_FALSE" {
		t.Errorf("false = %q, want \"AX_FALSE\"", got)
	}
}

func TestExprGen_BinaryAdd(t *testing.T) {
	ctx := newStmtTestCtx()
	queue := cgen.NewTypeDeclQueue()
	eg := cgen.NewExprGen(ctx.table, ctx.intern, ctx.symbols, ctx.tree, queue)

	// a + b
	left := ctx.tree.AddNode(ast.NodeIdent, 27) // "a"
	right := ctx.tree.AddNode(ast.NodeIdent, 28) // "b"
	binNode := ctx.tree.AddNode(ast.NodeBinaryExpr, 8) // token 8 = "+"
	ctx.tree.SetFirstChild(binNode, left)
	ctx.tree.SetNextSibling(left, right)

	got := eg.Emit(binNode)
	if got != "(a + b)" {
		t.Errorf("a + b = %q, want \"(a + b)\"", got)
	}
}

func TestExprGen_BinaryAnd(t *testing.T) {
	ctx := newStmtTestCtx()
	queue := cgen.NewTypeDeclQueue()
	eg := cgen.NewExprGen(ctx.table, ctx.intern, ctx.symbols, ctx.tree, queue)

	left := ctx.tree.AddNode(ast.NodeIdent, 27) // "a"
	right := ctx.tree.AddNode(ast.NodeIdent, 28) // "b"
	binNode := ctx.tree.AddNode(ast.NodeBinaryExpr, 14) // token 14 = "and"
	ctx.tree.SetFirstChild(binNode, left)
	ctx.tree.SetNextSibling(left, right)

	got := eg.Emit(binNode)
	if got != "(a && b)" {
		t.Errorf("a and b = %q, want \"(a && b)\"", got)
	}
}

func TestExprGen_UnaryNot(t *testing.T) {
	ctx := newStmtTestCtx()
	queue := cgen.NewTypeDeclQueue()
	eg := cgen.NewExprGen(ctx.table, ctx.intern, ctx.symbols, ctx.tree, queue)

	operand := ctx.tree.AddNode(ast.NodeIdent, 27) // "a"
	unNode := ctx.tree.AddNode(ast.NodeUnaryExpr, 16) // token 16 = "not"
	ctx.tree.SetFirstChild(unNode, operand)

	got := eg.Emit(unNode)
	if got != "(!a)" {
		t.Errorf("not a = %q, want \"(!a)\"", got)
	}
}

func TestExprGen_UnaryNegate(t *testing.T) {
	ctx := newStmtTestCtx()
	queue := cgen.NewTypeDeclQueue()
	eg := cgen.NewExprGen(ctx.table, ctx.intern, ctx.symbols, ctx.tree, queue)

	operand := ctx.tree.AddNode(ast.NodeIdent, 27) // "a"
	unNode := ctx.tree.AddNode(ast.NodeUnaryExpr, 9) // token 9 = "-"
	ctx.tree.SetFirstChild(unNode, operand)

	got := eg.Emit(unNode)
	if got != "(-a)" {
		t.Errorf("-a = %q, want \"(-a)\"", got)
	}
}

func TestExprGen_FuncCall(t *testing.T) {
	ctx := newStmtTestCtx()
	queue := cgen.NewTypeDeclQueue()
	eg := cgen.NewExprGen(ctx.table, ctx.intern, ctx.symbols, ctx.tree, queue)

	// foo(a, b)
	funcIdent := ctx.tree.AddNode(ast.NodeIdent, 25) // "foo"
	argA := ctx.tree.AddNode(ast.NodeIdent, 27)      // "a"
	argB := ctx.tree.AddNode(ast.NodeIdent, 28)      // "b"
	callNode := ctx.tree.AddNode(ast.NodeCallExpr, 25)
	ctx.tree.SetFirstChild(callNode, funcIdent)
	ctx.tree.SetNextSibling(funcIdent, argA)
	ctx.tree.SetNextSibling(argA, argB)

	got := eg.Emit(callNode)
	if got != "ax_foo(a, b)" {
		t.Errorf("foo(a, b) = %q, want \"ax_foo(a, b)\"", got)
	}
}

func TestExprGen_IndexSafe(t *testing.T) {
	ctx := newStmtTestCtx()
	queue := cgen.NewTypeDeclQueue()
	eg := cgen.NewExprGen(ctx.table, ctx.intern, ctx.symbols, ctx.tree, queue)

	// arr[i]
	arr := ctx.tree.AddNode(ast.NodeIdent, 24) // "list"
	idx := ctx.tree.AddNode(ast.NodeIdent, 23) // "i"
	indexNode := ctx.tree.AddNode(ast.NodeIndexExpr, 0)
	ctx.tree.SetFirstChild(indexNode, arr)
	ctx.tree.SetNextSibling(arr, idx)

	got := eg.Emit(indexNode)
	if !strings.Contains(got, "ax_bounds_check") {
		t.Errorf("safe index should contain ax_bounds_check: %q", got)
	}
	if !strings.Contains(got, "list") {
		t.Errorf("safe index should reference list: %q", got)
	}
}

func TestExprGen_IndexUnsafe(t *testing.T) {
	ctx := newStmtTestCtx()
	queue := cgen.NewTypeDeclQueue()
	eg := cgen.NewExprGen(ctx.table, ctx.intern, ctx.symbols, ctx.tree, queue)
	unsafeGen := eg.WithUnsafe()

	arr := ctx.tree.AddNode(ast.NodeIdent, 24) // "list"
	idx := ctx.tree.AddNode(ast.NodeIdent, 23) // "i"
	indexNode := ctx.tree.AddNode(ast.NodeIndexExpr, 0)
	ctx.tree.SetFirstChild(indexNode, arr)
	ctx.tree.SetNextSibling(arr, idx)

	got := unsafeGen.Emit(indexNode)
	if strings.Contains(got, "ax_bounds_check") {
		t.Errorf("unsafe index should NOT contain ax_bounds_check: %q", got)
	}
	if !strings.Contains(got, "(list).ptr[i]") {
		t.Errorf("unsafe index should use direct ptr access: %q", got)
	}
}

func TestExprGen_CastExpr(t *testing.T) {
	ctx := newStmtTestCtx()
	queue := cgen.NewTypeDeclQueue()
	eg := cgen.NewExprGen(ctx.table, ctx.intern, ctx.symbols, ctx.tree, queue)

	inner := ctx.tree.AddNode(ast.NodeIdent, 0) // "x"
	castNode := ctx.tree.AddNode(ast.NodeCastExpr, 0)
	ctx.tree.Node(castNode).Payload = uint32(types.TypeI32)
	ctx.tree.SetFirstChild(castNode, inner)

	got := eg.Emit(castNode)
	if got != "((ax_i32)(x))" {
		t.Errorf("x as i32 = %q, want \"((ax_i32)(x))\"", got)
	}
}

func TestExprGen_WithUnsafe_DoesNotMutateOriginal(t *testing.T) {
	ctx := newStmtTestCtx()
	queue := cgen.NewTypeDeclQueue()
	eg := cgen.NewExprGen(ctx.table, ctx.intern, ctx.symbols, ctx.tree, queue)

	unsafeGen := eg.WithUnsafe()
	if eg.Unsafe {
		t.Error("WithUnsafe should not mutate original")
	}
	if !unsafeGen.Unsafe {
		t.Error("WithUnsafe result should be unsafe")
	}
}

// === Statement Tests ===

func TestStmtGen_VarDecl(t *testing.T) {
	ctx := newStmtTestCtx()
	sg, buf := ctx.makeStmtGen()

	// let x: i32 = 5
	varNode := ctx.tree.AddNode(ast.NodeVarDecl, 0) // token 0 = "x"
	ctx.tree.Node(varNode).Payload = uint32(types.TypeI32)
	initExpr := ctx.tree.AddNode(ast.NodeIntLit, 1) // token 1 = "5"
	ctx.tree.Node(initExpr).Payload = uint32(types.TypeI32)
	ctx.tree.SetFirstChild(varNode, initExpr)

	sg.EmitStmt(varNode)
	output := buf.String()

	if !strings.Contains(output, "ax_i32 x = 5;") {
		t.Errorf("var decl output = %q, want containing \"ax_i32 x = 5;\"", output)
	}
}

func TestStmtGen_VarDeclNoInit(t *testing.T) {
	ctx := newStmtTestCtx()
	sg, buf := ctx.makeStmtGen()

	// let x: i32 (no init)
	varNode := ctx.tree.AddNode(ast.NodeVarDecl, 0) // "x"
	ctx.tree.Node(varNode).Payload = uint32(types.TypeI32)

	sg.EmitStmt(varNode)
	output := buf.String()

	if !strings.Contains(output, "ax_i32 x = {0};") {
		t.Errorf("var decl no init = %q, want containing \"ax_i32 x = {0};\"", output)
	}
}

func TestStmtGen_Return(t *testing.T) {
	ctx := newStmtTestCtx()
	sg, buf := ctx.makeStmtGen()

	// return x
	retNode := ctx.tree.AddNode(ast.NodeReturnStmt, 17) // "return"
	xNode := ctx.tree.AddNode(ast.NodeIdent, 0) // "x"
	ctx.tree.SetFirstChild(retNode, xNode)

	sg.EmitStmt(retNode)
	output := buf.String()

	if !strings.Contains(output, "return x;") {
		t.Errorf("return = %q, want containing \"return x;\"", output)
	}
}

func TestStmtGen_ReturnVoid(t *testing.T) {
	ctx := newStmtTestCtx()
	sg, buf := ctx.makeStmtGen()

	retNode := ctx.tree.AddNode(ast.NodeReturnStmt, 17) // "return"

	sg.EmitStmt(retNode)
	output := buf.String()

	if !strings.Contains(output, "return;") {
		t.Errorf("void return = %q, want containing \"return;\"", output)
	}
}

func TestStmtGen_IfStmt(t *testing.T) {
	ctx := newStmtTestCtx()
	sg, buf := ctx.makeStmtGen()

	// if cond: body
	condNode := ctx.tree.AddNode(ast.NodeIdent, 19) // "cond"
	bodyBlock := ctx.tree.AddNode(ast.NodeBlock, 20) // block

	ifNode := ctx.tree.AddNode(ast.NodeIfStmt, 18) // "if"
	ctx.tree.SetFirstChild(ifNode, condNode)
	ctx.tree.SetNextSibling(condNode, bodyBlock)

	sg.EmitStmt(ifNode)
	output := buf.String()

	if !strings.Contains(output, "if (cond) {") {
		t.Errorf("if stmt = %q, want containing \"if (cond) {\"", output)
	}
	if !strings.Contains(output, "}") {
		t.Errorf("if stmt missing closing brace: %q", output)
	}
}

func TestStmtGen_WhileStmt(t *testing.T) {
	ctx := newStmtTestCtx()
	sg, buf := ctx.makeStmtGen()

	condNode := ctx.tree.AddNode(ast.NodeIdent, 19) // "cond"
	bodyBlock := ctx.tree.AddNode(ast.NodeBlock, 20)

	whileNode := ctx.tree.AddNode(ast.NodeWhileStmt, 21) // "while"
	ctx.tree.SetFirstChild(whileNode, condNode)
	ctx.tree.SetNextSibling(condNode, bodyBlock)

	sg.EmitStmt(whileNode)
	output := buf.String()

	if !strings.Contains(output, "while (cond) {") {
		t.Errorf("while stmt = %q, want containing \"while (cond) {\"", output)
	}
}

func TestStmtGen_Destroy(t *testing.T) {
	ctx := newStmtTestCtx()
	sg, buf := ctx.makeStmtGen()

	xNode := ctx.tree.AddNode(ast.NodeIdent, 0) // "x"
	destroyNode := ctx.tree.AddNode(ast.NodeDestroyStmt, 0)
	ctx.tree.SetFirstChild(destroyNode, xNode)

	sg.EmitStmt(destroyNode)
	output := buf.String()

	if !strings.Contains(output, "ax_free(x);") {
		t.Errorf("destroy = %q, want containing \"ax_free(x);\"", output)
	}
}

// === DeferStack Tests ===

func TestDeferStack_LIFO(t *testing.T) {
	ds := cgen.NewDeferStack()
	ds.PushScope()
	ds.Push(10)
	ds.Push(20)
	ds.Push(30)

	result := ds.PopScope()
	if len(result) != 3 {
		t.Fatalf("expected 3 defers, got %d", len(result))
	}
	// LIFO: 30, 20, 10
	if result[0] != 30 || result[1] != 20 || result[2] != 10 {
		t.Errorf("LIFO order incorrect: %v", result)
	}
}

func TestDeferStack_NestedScopes(t *testing.T) {
	ds := cgen.NewDeferStack()
	ds.PushScope()
	ds.Push(1)
	ds.PushScope()
	ds.Push(2)
	ds.Push(3)

	inner := ds.PopScope()
	if len(inner) != 2 {
		t.Fatalf("inner scope: expected 2 defers, got %d", len(inner))
	}

	outer := ds.PopScope()
	if len(outer) != 1 {
		t.Fatalf("outer scope: expected 1 defer, got %d", len(outer))
	}
	if outer[0] != 1 {
		t.Errorf("outer defer = %d, want 1", outer[0])
	}
}

// === IndentWriter Tests ===

func TestIndentWriter_Indentation(t *testing.T) {
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)

	w.Line("level0")
	w.Indent()
	w.Line("level1")
	w.Indent()
	w.Line("level2")
	w.Dedent()
	w.Line("level1again")
	w.Dedent()
	w.Line("level0again")

	output := buf.String()
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	expected := []string{
		"level0",
		"    level1",
		"        level2",
		"    level1again",
		"level0again",
	}

	if len(lines) != len(expected) {
		t.Fatalf("expected %d lines, got %d:\n%s", len(expected), len(lines), output)
	}

	for i, want := range expected {
		if lines[i] != want {
			t.Errorf("line %d = %q, want %q", i, lines[i], want)
		}
	}
}
