package cgen_test

import (
	"bytes"
	"runtime"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// pipelineTestCtx provides a shared context builder for pipeline tests.
type pipelineTestCtx struct {
	tree    *ast.AstTree
	intern  *ast.InternPool
	symbols *sema.SymbolTable
	table   *types.TypeTable
}

func newPipelineTestCtx() *pipelineTestCtx {
	intern := ast.NewInternPool(64)
	table := types.NewTypeTable()
	symbols := sema.NewSymbolTable(intern)

	// Source with tokens needed for pipeline tests
	// "fn main x i32 5 return add a b result let"
	source := []byte(
		"fn main x i32 5 return add a b result let",
	)
	tokens := []lexer.Token{
		{Kind: lexer.TokenFn, Offset: 0, Len: 2},       // 0: "fn"
		{Kind: lexer.TokenIdent, Offset: 3, Len: 4},     // 1: "main"
		{Kind: lexer.TokenIdent, Offset: 8, Len: 1},     // 2: "x"
		{Kind: lexer.TokenIdent, Offset: 10, Len: 3},    // 3: "i32"
		{Kind: lexer.TokenIntLit, Offset: 14, Len: 1},   // 4: "5"
		{Kind: lexer.TokenReturn, Offset: 16, Len: 6},   // 5: "return"
		{Kind: lexer.TokenIdent, Offset: 23, Len: 3},    // 6: "add"
		{Kind: lexer.TokenIdent, Offset: 27, Len: 1},    // 7: "a"
		{Kind: lexer.TokenIdent, Offset: 29, Len: 1},    // 8: "b"
		{Kind: lexer.TokenIdent, Offset: 31, Len: 6},    // 9: "result"
		{Kind: lexer.TokenLet, Offset: 38, Len: 3},      // 10: "let"
	}

	tree := ast.NewTree(source, tokens)
	return &pipelineTestCtx{tree: tree, intern: intern, symbols: symbols, table: table}
}

// TestPipeline_EmptyModule verifies that an empty module produces valid output structure.
func TestPipeline_EmptyModule(t *testing.T) {
	ctx := newPipelineTestCtx()
	p := cgen.NewPipeline(ctx.table, ctx.intern, ctx.symbols, ctx.tree)

	var buf bytes.Buffer
	err := p.GenerateC(&buf)
	if err != nil {
		t.Fatalf("GenerateC failed: %v", err)
	}

	output := buf.String()

	// Must start with #include
	if !strings.Contains(output, `#include "ax_runtime.h"`) {
		t.Errorf("output missing #include ax_runtime.h:\n%s", output)
	}
}

// TestPipeline_SingleFuncNoBody verifies a function with no body (extern-like).
func TestPipeline_SingleFuncNoBody(t *testing.T) {
	ctx := newPipelineTestCtx()

	// Register function type: fn main() -> void
	funcTypeID := ctx.table.RegisterFunction(nil, types.TypeVoid, nil)

	mainName := ctx.intern.InternString("main")
	symIdx, _ := ctx.symbols.Define(mainName, sema.SymFunc, sema.SymFlagPub, 0)
	ctx.symbols.SymbolAt(symIdx).TypeID = uint32(funcTypeID)

	// Create function node with FlagIsPub
	funcIdx := ctx.tree.AddNode(ast.NodeFuncDecl, 1) // token 1 = "main"
	ctx.tree.Node(funcIdx).Payload = symIdx
	ctx.tree.Node(funcIdx).Flags = ast.FlagIsPub
	ctx.tree.AppendChild(0, funcIdx)

	p := cgen.NewPipeline(ctx.table, ctx.intern, ctx.symbols, ctx.tree)

	var buf bytes.Buffer
	err := p.GenerateC(&buf)
	if err != nil {
		t.Fatalf("GenerateC failed: %v", err)
	}

	output := buf.String()

	// Should have the include
	if !strings.Contains(output, `#include "ax_runtime.h"`) {
		t.Errorf("missing #include:\n%s", output)
	}

	// Should have the prototype in the declarations section
	if !strings.Contains(output, "void ax_main(void);") {
		t.Errorf("missing function prototype:\n%s", output)
	}

	// Should have the function definition
	if !strings.Contains(output, "void ax_main(void) {") {
		t.Errorf("missing function definition:\n%s", output)
	}
}

// TestPipeline_FuncWithBody verifies a function with a body generates properly.
func TestPipeline_FuncWithBody(t *testing.T) {
	ctx := newPipelineTestCtx()

	// Register: fn add(a: i32, b: i32) -> i32
	funcTypeID := ctx.table.RegisterFunction(
		[]types.TypeID{types.TypeI32, types.TypeI32},
		types.TypeI32,
		nil,
	)

	addName := ctx.intern.InternString("add")
	symIdx, _ := ctx.symbols.Define(addName, sema.SymFunc, sema.SymFlagPub, 0)
	ctx.symbols.SymbolAt(symIdx).TypeID = uint32(funcTypeID)

	// Build AST: fn add(a, b) { return 5 }
	funcIdx := ctx.tree.AddNode(ast.NodeFuncDecl, 6) // token 6 = "add"
	ctx.tree.Node(funcIdx).Payload = symIdx
	ctx.tree.Node(funcIdx).Flags = ast.FlagIsPub

	paramA := ctx.tree.AddNode(ast.NodeParamDecl, 7) // token 7 = "a"
	paramB := ctx.tree.AddNode(ast.NodeParamDecl, 8) // token 8 = "b"

	bodyBlock := ctx.tree.AddNode(ast.NodeBlock, 0)

	// return 5
	retNode := ctx.tree.AddNode(ast.NodeReturnStmt, 5) // token 5 = "return"
	litNode := ctx.tree.AddNode(ast.NodeIntLit, 4)      // token 4 = "5"
	ctx.tree.Node(litNode).Payload = uint32(types.TypeI32)
	ctx.tree.SetFirstChild(retNode, litNode)
	ctx.tree.SetFirstChild(bodyBlock, retNode)

	// Wire up func children: paramA -> paramB -> bodyBlock
	ctx.tree.SetFirstChild(funcIdx, paramA)
	ctx.tree.SetNextSibling(paramA, paramB)
	ctx.tree.SetNextSibling(paramB, bodyBlock)

	ctx.tree.AppendChild(0, funcIdx)

	p := cgen.NewPipeline(ctx.table, ctx.intern, ctx.symbols, ctx.tree)

	var buf bytes.Buffer
	err := p.GenerateC(&buf)
	if err != nil {
		t.Fatalf("GenerateC failed: %v", err)
	}

	output := buf.String()

	// Check include
	if !strings.Contains(output, `#include "ax_runtime.h"`) {
		t.Errorf("missing #include:\n%s", output)
	}

	// Check prototype
	if !strings.Contains(output, "ax_i32 ax_add(ax_i32 a, ax_i32 b);") {
		t.Errorf("missing or incorrect function prototype:\n%s", output)
	}

	// Check function definition header
	if !strings.Contains(output, "ax_i32 ax_add(ax_i32 a, ax_i32 b) {") {
		t.Errorf("missing or incorrect function definition:\n%s", output)
	}

	// Check return statement in the body
	if !strings.Contains(output, "return 5;") {
		t.Errorf("missing return statement in body:\n%s", output)
	}

	// Check closing brace
	if !strings.Contains(output, "}") {
		t.Errorf("missing closing brace:\n%s", output)
	}
}

// TestPipeline_FuncWithVarDecl verifies a function with a var declaration.
func TestPipeline_FuncWithVarDecl(t *testing.T) {
	ctx := newPipelineTestCtx()

	// fn main() -> void
	funcTypeID := ctx.table.RegisterFunction(nil, types.TypeVoid, nil)

	mainName := ctx.intern.InternString("main")
	symIdx, _ := ctx.symbols.Define(mainName, sema.SymFunc, sema.SymFlagPub, 0)
	ctx.symbols.SymbolAt(symIdx).TypeID = uint32(funcTypeID)

	funcIdx := ctx.tree.AddNode(ast.NodeFuncDecl, 1) // "main"
	ctx.tree.Node(funcIdx).Payload = symIdx
	ctx.tree.Node(funcIdx).Flags = ast.FlagIsPub

	bodyBlock := ctx.tree.AddNode(ast.NodeBlock, 0)

	// let result: i32 = 5
	varNode := ctx.tree.AddNode(ast.NodeVarDecl, 9) // token 9 = "result"
	ctx.tree.Node(varNode).Payload = uint32(types.TypeI32)
	initExpr := ctx.tree.AddNode(ast.NodeIntLit, 4) // token 4 = "5"
	ctx.tree.Node(initExpr).Payload = uint32(types.TypeI32)
	ctx.tree.SetFirstChild(varNode, initExpr)

	// return
	retNode := ctx.tree.AddNode(ast.NodeReturnStmt, 5)
	ctx.tree.SetFirstChild(bodyBlock, varNode)
	ctx.tree.SetNextSibling(varNode, retNode)

	ctx.tree.SetFirstChild(funcIdx, bodyBlock)
	ctx.tree.AppendChild(0, funcIdx)

	p := cgen.NewPipeline(ctx.table, ctx.intern, ctx.symbols, ctx.tree)

	var buf bytes.Buffer
	err := p.GenerateC(&buf)
	if err != nil {
		t.Fatalf("GenerateC failed: %v", err)
	}

	output := buf.String()

	// Should have the var decl
	if !strings.Contains(output, "ax_i32 result = 5;") {
		t.Errorf("missing var declaration in body:\n%s", output)
	}

	// Should have return
	if !strings.Contains(output, "return;") {
		t.Errorf("missing void return:\n%s", output)
	}
}

// TestPipeline_PrivateFunc verifies static prefix for non-pub functions.
func TestPipeline_PrivateFunc(t *testing.T) {
	ctx := newPipelineTestCtx()

	funcTypeID := ctx.table.RegisterFunction(nil, types.TypeVoid, nil)

	helperName := ctx.intern.InternString("add")
	symIdx, _ := ctx.symbols.Define(helperName, sema.SymFunc, 0, 0) // no Pub flag
	ctx.symbols.SymbolAt(symIdx).TypeID = uint32(funcTypeID)

	funcIdx := ctx.tree.AddNode(ast.NodeFuncDecl, 6) // "add"
	ctx.tree.Node(funcIdx).Payload = symIdx
	// No FlagIsPub
	ctx.tree.AppendChild(0, funcIdx)

	p := cgen.NewPipeline(ctx.table, ctx.intern, ctx.symbols, ctx.tree)

	var buf bytes.Buffer
	err := p.GenerateC(&buf)
	if err != nil {
		t.Fatalf("GenerateC failed: %v", err)
	}

	output := buf.String()

	// Definition should have static prefix
	if !strings.Contains(output, "static void ax_add(void) {") {
		t.Errorf("private function should have static prefix in definition:\n%s", output)
	}
}

// TestPipeline_OutputStructure verifies the overall output ordering.
func TestPipeline_OutputStructure(t *testing.T) {
	ctx := newPipelineTestCtx()

	// Register a simple function
	funcTypeID := ctx.table.RegisterFunction(nil, types.TypeI32, nil)
	mainName := ctx.intern.InternString("main")
	symIdx, _ := ctx.symbols.Define(mainName, sema.SymFunc, sema.SymFlagPub, 0)
	ctx.symbols.SymbolAt(symIdx).TypeID = uint32(funcTypeID)

	funcIdx := ctx.tree.AddNode(ast.NodeFuncDecl, 1) // "main"
	ctx.tree.Node(funcIdx).Payload = symIdx
	ctx.tree.Node(funcIdx).Flags = ast.FlagIsPub

	bodyBlock := ctx.tree.AddNode(ast.NodeBlock, 0)
	retNode := ctx.tree.AddNode(ast.NodeReturnStmt, 5)
	litNode := ctx.tree.AddNode(ast.NodeIntLit, 4)
	ctx.tree.Node(litNode).Payload = uint32(types.TypeI32)
	ctx.tree.SetFirstChild(retNode, litNode)
	ctx.tree.SetFirstChild(bodyBlock, retNode)
	ctx.tree.SetFirstChild(funcIdx, bodyBlock)
	ctx.tree.AppendChild(0, funcIdx)

	p := cgen.NewPipeline(ctx.table, ctx.intern, ctx.symbols, ctx.tree)

	var buf bytes.Buffer
	err := p.GenerateC(&buf)
	if err != nil {
		t.Fatalf("GenerateC failed: %v", err)
	}

	output := buf.String()

	// Verify ordering: #include comes before prototype, prototype comes before definition
	includePos := strings.Index(output, `#include "ax_runtime.h"`)
	protoPos := strings.Index(output, "ax_i32 ax_main(void);")
	defPos := strings.Index(output, "ax_i32 ax_main(void) {")

	if includePos < 0 {
		t.Fatal("missing #include")
	}
	if protoPos < 0 {
		t.Fatal("missing prototype")
	}
	if defPos < 0 {
		t.Fatal("missing definition")
	}

	if includePos >= protoPos {
		t.Errorf("#include (%d) should come before prototype (%d)", includePos, protoPos)
	}
	if protoPos >= defPos {
		t.Errorf("prototype (%d) should come before definition (%d)", protoPos, defPos)
	}
}

// TestPipeline_DetectCCompiler verifies the compiler detection logic.
func TestPipeline_DetectCCompiler(t *testing.T) {
	compiler, err := cgen.DetectCCompiler()

	// On CI or systems without a C compiler, this is expected to fail.
	// We just verify the function doesn't panic.
	if err != nil {
		t.Skipf("No C compiler found (expected in some environments): %v", err)
	}

	// If found, verify it's one of the expected compilers
	valid := map[string]bool{
		"gcc":    true,
		"clang":  true,
		"cl.exe": true,
	}
	if !valid[compiler] {
		t.Errorf("unexpected compiler detected: %q", compiler)
	}
	t.Logf("detected C compiler: %s", compiler)
}

// TestPipeline_GenerateCSourcePath verifies source path generation.
func TestPipeline_GenerateCSourcePath(t *testing.T) {
	cases := []struct {
		module string
		want   string
	}{
		{"main", "main.c"},
		{"math/vector", "math_vector.c"},
		{"std/io", "std_io.c"},
		{"hello", "hello.c"},
	}

	for _, tc := range cases {
		got := cgen.GenerateCSourcePath(tc.module)
		if got != tc.want {
			t.Errorf("GenerateCSourcePath(%q) = %q, want %q", tc.module, got, tc.want)
		}
	}
}

// TestPipeline_OutputBinaryName verifies binary name derivation.
func TestPipeline_OutputBinaryName(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"main.ax", "main"},
		{"hello.ax", "hello"},
		{"app", "app"},
	}

	for _, tc := range cases {
		got := cgen.OutputBinaryName(tc.input)
		expected := tc.want
		if runtime.GOOS == "windows" {
			if !strings.HasSuffix(expected, ".exe") {
				expected += ".exe"
			}
		}
		if got != expected {
			t.Errorf("OutputBinaryName(%q) = %q, want %q", tc.input, got, expected)
		}
	}
}

// TestPipeline_MultipleFunctions verifies correct output with multiple functions.
func TestPipeline_MultipleFunctions(t *testing.T) {
	ctx := newPipelineTestCtx()

	// fn main() -> void
	mainTypeID := ctx.table.RegisterFunction(nil, types.TypeVoid, nil)
	mainName := ctx.intern.InternString("main")
	mainSymIdx, _ := ctx.symbols.Define(mainName, sema.SymFunc, sema.SymFlagPub, 0)
	ctx.symbols.SymbolAt(mainSymIdx).TypeID = uint32(mainTypeID)

	mainFuncIdx := ctx.tree.AddNode(ast.NodeFuncDecl, 1) // "main"
	ctx.tree.Node(mainFuncIdx).Payload = mainSymIdx
	ctx.tree.Node(mainFuncIdx).Flags = ast.FlagIsPub

	mainBody := ctx.tree.AddNode(ast.NodeBlock, 0)
	ctx.tree.SetFirstChild(mainFuncIdx, mainBody)
	ctx.tree.AppendChild(0, mainFuncIdx)

	// fn add(a: i32, b: i32) -> i32
	addTypeID := ctx.table.RegisterFunction(
		[]types.TypeID{types.TypeI32, types.TypeI32},
		types.TypeI32,
		nil,
	)
	addName := ctx.intern.InternString("add")
	addSymIdx, _ := ctx.symbols.Define(addName, sema.SymFunc, sema.SymFlagPub, 0)
	ctx.symbols.SymbolAt(addSymIdx).TypeID = uint32(addTypeID)

	addFuncIdx := ctx.tree.AddNode(ast.NodeFuncDecl, 6) // "add"
	ctx.tree.Node(addFuncIdx).Payload = addSymIdx
	ctx.tree.Node(addFuncIdx).Flags = ast.FlagIsPub

	paramA := ctx.tree.AddNode(ast.NodeParamDecl, 7) // "a"
	paramB := ctx.tree.AddNode(ast.NodeParamDecl, 8) // "b"
	addBody := ctx.tree.AddNode(ast.NodeBlock, 0)

	retNode := ctx.tree.AddNode(ast.NodeReturnStmt, 5)
	litNode := ctx.tree.AddNode(ast.NodeIntLit, 4) // "5"
	ctx.tree.Node(litNode).Payload = uint32(types.TypeI32)
	ctx.tree.SetFirstChild(retNode, litNode)
	ctx.tree.SetFirstChild(addBody, retNode)

	ctx.tree.SetFirstChild(addFuncIdx, paramA)
	ctx.tree.SetNextSibling(paramA, paramB)
	ctx.tree.SetNextSibling(paramB, addBody)
	ctx.tree.AppendChild(0, addFuncIdx)

	p := cgen.NewPipeline(ctx.table, ctx.intern, ctx.symbols, ctx.tree)

	var buf bytes.Buffer
	err := p.GenerateC(&buf)
	if err != nil {
		t.Fatalf("GenerateC failed: %v", err)
	}

	output := buf.String()

	// Both functions should appear
	if !strings.Contains(output, "ax_main") {
		t.Errorf("missing main function:\n%s", output)
	}
	if !strings.Contains(output, "ax_add") {
		t.Errorf("missing add function:\n%s", output)
	}

	// Both should have definitions
	if !strings.Contains(output, "void ax_main(void) {") {
		t.Errorf("missing main definition:\n%s", output)
	}
	if !strings.Contains(output, "ax_i32 ax_add(ax_i32 a, ax_i32 b) {") {
		t.Errorf("missing add definition:\n%s", output)
	}
}
