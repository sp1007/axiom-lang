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

// buildTestContext creates a minimal context for testing the DeclEmitter.
func buildTestContext() (*ast.AstTree, *ast.InternPool, *sema.SymbolTable, *types.TypeTable) {
	intern := ast.NewInternPool(64)
	table := types.NewTypeTable()
	symbols := sema.NewSymbolTable(intern)

	// Minimal source and tokens — we'll add more tokens as needed per test.
	source := []byte("struct Point x i32 y i32 fn distance a Point b Point f64 helper")
	tokens := []lexer.Token{
		{Kind: lexer.TokenStruct, Offset: 0, Len: 6},   // 0: "struct"
		{Kind: lexer.TokenIdent, Offset: 7, Len: 5},     // 1: "Point"
		{Kind: lexer.TokenIdent, Offset: 13, Len: 1},    // 2: "x"
		{Kind: lexer.TokenIdent, Offset: 15, Len: 3},    // 3: "i32"
		{Kind: lexer.TokenIdent, Offset: 19, Len: 1},    // 4: "y"
		{Kind: lexer.TokenIdent, Offset: 21, Len: 3},    // 5: "i32"
		{Kind: lexer.TokenFn, Offset: 25, Len: 2},       // 6: "fn"
		{Kind: lexer.TokenIdent, Offset: 28, Len: 8},    // 7: "distance"
		{Kind: lexer.TokenIdent, Offset: 37, Len: 1},    // 8: "a"
		{Kind: lexer.TokenIdent, Offset: 39, Len: 5},    // 9: "Point"
		{Kind: lexer.TokenIdent, Offset: 45, Len: 1},    // 10: "b"
		{Kind: lexer.TokenIdent, Offset: 47, Len: 5},    // 11: "Point"
		{Kind: lexer.TokenIdent, Offset: 53, Len: 3},    // 12: "f64"
		{Kind: lexer.TokenIdent, Offset: 57, Len: 6},    // 13: "helper"
	}

	tree := ast.NewTree(source, tokens)
	return tree, intern, symbols, table
}

// TestDeclEmitter_StructForwardDecl verifies that struct forward declarations are emitted.
func TestDeclEmitter_StructForwardDecl(t *testing.T) {
	tree, intern, symbols, table := buildTestContext()

	// Build a struct declaration: struct Point { x: i32, y: i32 }
	structIdx := tree.AddNode(ast.NodeStructDecl, 1) // token 1 = "Point"

	fieldX := tree.AddNode(ast.NodeFieldDecl, 2) // token 2 = "x"
	tree.Node(fieldX).Payload = uint32(types.TypeI32)

	fieldY := tree.AddNode(ast.NodeFieldDecl, 4) // token 4 = "y"
	tree.Node(fieldY).Payload = uint32(types.TypeI32)

	tree.SetFirstChild(structIdx, fieldX)
	tree.SetNextSibling(fieldX, fieldY)
	tree.AppendChild(0, structIdx)

	emitter := cgen.NewDeclEmitter(table, intern, symbols, tree)
	emitter.ProcessModule()

	var buf bytes.Buffer
	emitter.EmitTo(&buf)
	output := buf.String()

	if !strings.Contains(output, "struct ax_Point;") {
		t.Errorf("missing forward declaration for struct Point in:\n%s", output)
	}
	if !strings.Contains(output, "struct ax_Point {") {
		t.Errorf("missing struct definition for Point in:\n%s", output)
	}
	if !strings.Contains(output, "ax_i32 x;") {
		t.Errorf("missing field x in struct Point:\n%s", output)
	}
	if !strings.Contains(output, "ax_i32 y;") {
		t.Errorf("missing field y in struct Point:\n%s", output)
	}
}

// TestDeclEmitter_FuncPrototype verifies function prototype generation.
func TestDeclEmitter_FuncPrototype(t *testing.T) {
	tree, intern, symbols, table := buildTestContext()

	funcTypeID := table.RegisterFunction(
		[]types.TypeID{types.TypeI32, types.TypeI32},
		types.TypeF64,
		nil,
	)

	distName := intern.InternString("distance")
	symIdx, _ := symbols.Define(distName, sema.SymFunc, sema.SymFlagPub, 0)
	symbols.SymbolAt(symIdx).TypeID = uint32(funcTypeID)

	funcIdx := tree.AddNode(ast.NodeFuncDecl, 7) // token 7 = "distance"
	tree.Node(funcIdx).Payload = symIdx
	tree.Node(funcIdx).Flags = ast.FlagIsPub

	paramA := tree.AddNode(ast.NodeParamDecl, 8)  // "a"
	paramB := tree.AddNode(ast.NodeParamDecl, 10)  // "b"
	tree.SetFirstChild(funcIdx, paramA)
	tree.SetNextSibling(paramA, paramB)

	tree.AppendChild(0, funcIdx)

	emitter := cgen.NewDeclEmitter(table, intern, symbols, tree)
	emitter.ProcessModule()

	var buf bytes.Buffer
	emitter.EmitTo(&buf)
	output := buf.String()

	// Public function should NOT have "static"
	if strings.Contains(output, "static ax_f64 ax_distance") {
		t.Errorf("public function should not have static prefix:\n%s", output)
	}
	if !strings.Contains(output, "ax_f64 ax_distance(ax_i32 a, ax_i32 b);") {
		t.Errorf("missing or incorrect function prototype:\n%s", output)
	}
}

// TestDeclEmitter_PrivateFunc verifies private functions get static prefix.
func TestDeclEmitter_PrivateFunc(t *testing.T) {
	tree, intern, symbols, table := buildTestContext()

	funcTypeID := table.RegisterFunction(nil, types.TypeVoid, nil)

	helperName := intern.InternString("helper")
	symIdx, _ := symbols.Define(helperName, sema.SymFunc, 0, 0) // no FlagPub
	symbols.SymbolAt(symIdx).TypeID = uint32(funcTypeID)

	funcIdx := tree.AddNode(ast.NodeFuncDecl, 13) // token 13 = "helper"
	tree.Node(funcIdx).Payload = symIdx
	// No FlagIsPub
	tree.AppendChild(0, funcIdx)

	emitter := cgen.NewDeclEmitter(table, intern, symbols, tree)
	emitter.ProcessModule()

	var buf bytes.Buffer
	emitter.EmitTo(&buf)
	output := buf.String()

	if !strings.Contains(output, "static void ax_helper(void);") {
		t.Errorf("private function should have static prefix:\n%s", output)
	}
}

// TestDeclEmitter_IncludeHeader verifies that #include "ax_runtime.h" is always first.
func TestDeclEmitter_IncludeHeader(t *testing.T) {
	tree, intern, symbols, table := buildTestContext()

	emitter := cgen.NewDeclEmitter(table, intern, symbols, tree)
	emitter.ProcessModule()

	var buf bytes.Buffer
	emitter.EmitTo(&buf)
	output := buf.String()

	if !strings.HasPrefix(output, `#include "ax_runtime.h"`) {
		t.Errorf("output should start with #include, got:\n%.100s", output)
	}
}

// TestDeclEmitter_EmptyModule verifies that an empty module still outputs the include.
func TestDeclEmitter_EmptyModule(t *testing.T) {
	tree, intern, symbols, table := buildTestContext()

	emitter := cgen.NewDeclEmitter(table, intern, symbols, tree)
	emitter.ProcessModule()

	var buf bytes.Buffer
	emitter.EmitTo(&buf)
	output := buf.String()

	if !strings.Contains(output, `#include "ax_runtime.h"`) {
		t.Errorf("even empty module should have include:\n%s", output)
	}
}

// TestMangleFuncName verifies function name mangling.
func TestMangleFuncName(t *testing.T) {
	cases := []struct {
		module string
		name   string
		want   string
	}{
		{"", "fibonacci", "ax_fibonacci"},
		{"math", "sqrt", "ax_math_sqrt"},
		{"", "main", "ax_main"},
	}
	for _, tc := range cases {
		got := cgen.MangleFuncName(tc.module, tc.name)
		if got != tc.want {
			t.Errorf("MangleFuncName(%q, %q) = %q, want %q", tc.module, tc.name, got, tc.want)
		}
	}
}

// TestMangleGlobalName verifies global variable name mangling.
func TestMangleGlobalName(t *testing.T) {
	cases := []struct {
		module string
		name   string
		want   string
	}{
		{"", "MAX", "ax_MAX"},
		{"config", "VERSION", "ax_config_VERSION"},
	}
	for _, tc := range cases {
		got := cgen.MangleGlobalName(tc.module, tc.name)
		if got != tc.want {
			t.Errorf("MangleGlobalName(%q, %q) = %q, want %q", tc.module, tc.name, got, tc.want)
		}
	}
}
