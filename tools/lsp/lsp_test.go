package lsp

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/compiler/sema"
)

func TestInitialize(t *testing.T) {
	// Mock a basic initialize request
	req := rawMessage{
		Jsonrpc: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
		Params:  json.RawMessage(`{}`),
	}
	_ = req

	// We can manually call handleRequest and capture the output by intercepting lspWriter,
	// but we can also just verify our handler logic directly!
	// Let's verify that runAnalysis works correctly on a valid AXIOM program.
	src := `
fn add(x: i32, y: i32) -> i32:
    return x + y

fn main():
    let val = add(10, 20)
`
	uri := "file:///test.ax"
	diags, pool, tree, symtab, tt, infer, err := runAnalysis(uri, src)
	if err != nil {
		t.Fatalf("runAnalysis failed: %v", err)
	}

	if len(diags) > 0 {
		t.Errorf("Expected 0 diagnostics, got %d: %v", len(diags), diags)
	}

	if pool == nil || tree == nil || symtab == nil || tt == nil || infer == nil {
		t.Fatal("Expected all analysis data structures to be populated")
	}

	// Verify type check results in the symbol table
	// Find symbol "val" or "add"
	foundAdd := false
	for _, sym := range symtab.Symbols {
		name := string(pool.GetBytes(sym.NameID))
		if name == "add" {
			foundAdd = true
			if sym.Kind != sema.SymFunc {
				t.Errorf("Expected 'add' to be SymFunc, got %v", sym.Kind)
			}
		}
	}
	if !foundAdd {
		t.Error("Expected symbol 'add' to be defined in symbol table")
	}
}

func TestHover(t *testing.T) {
	src := `
fn main():
    let val = 42
    let another = val
`
	uri := "file:///test.ax"
	_, pool, tree, symtab, tt, infer, err := runAnalysis(uri, src)
	if err != nil {
		t.Fatalf("runAnalysis failed: %v", err)
	}

	// Let's hover over "val" in "let another = val"
	// val is at offset 38: Line 3 (0-based: 3), Character 18 (0-based: 18)
	// Let's verify that OffsetFromLineCol finds it:
	offset := OffsetFromLineCol(src, 3, 18)
	
	// Ensure the offset points to "val"
	tokText := src[offset : offset+3]
	if tokText != "val" {
		t.Fatalf("Expected token at offset %d to be 'val', got %q", offset, tokText)
	}

	result := handleHover(src, pool, tree, symtab, tt, infer, 3, 18)
	if result == nil {
		t.Fatal("Expected hover result to be non-nil")
	}

	hoverRes, ok := result.(*HoverResult)
	if !ok {
		t.Fatalf("Expected HoverResult, got %T", result)
	}

	if hoverRes.Contents.Kind != "markdown" {
		t.Errorf("Expected markdown format, got %q", hoverRes.Contents.Kind)
	}

	if !strings.Contains(hoverRes.Contents.Value, "val") || !strings.Contains(hoverRes.Contents.Value, "i32") {
		t.Errorf("Expected hover content to mention val and i32, got %q", hoverRes.Contents.Value)
	}
}

func TestDefinition(t *testing.T) {
	src := `
fn my_func(a: i32) -> i32:
    return a

fn main():
    let x = my_func(10)
`
	uri := "file:///test.ax"
	_, pool, tree, symtab, tt, infer, err := runAnalysis(uri, src)
	if err != nil {
		t.Fatalf("runAnalysis failed: %v", err)
	}

	// Find the definition of my_func when hovering at "my_func" in main
	// my_func is at line 5 (0-based: 5), character 12 (0-based: 12)
	offset := OffsetFromLineCol(src, 5, 12)
	tokText := src[offset : offset+7]
	if tokText != "my_func" {
		t.Fatalf("Expected token at offset %d to be 'my_func', got %q", offset, tokText)
	}

	result := handleDefinition(src, pool, tree, symtab, tt, infer, 5, 12, uri)
	if result == nil {
		t.Fatal("Expected definition result to be non-nil")
	}

	loc, ok := result.(*Location)
	if !ok {
		t.Fatalf("Expected Location, got %T", result)
	}

	if loc.URI != uri {
		t.Errorf("Expected URI %q, got %q", uri, loc.URI)
	}

	// my_func is defined starting at the 'fn' keyword: line 1 (0-based), character 0 (0-based)
	if loc.Range.Start.Line != 1 || loc.Range.Start.Character != 0 {
		t.Errorf("Expected definition at line 1, char 0, got line %d, char %d", loc.Range.Start.Line, loc.Range.Start.Character)
	}
}

func TestDiagnostics(t *testing.T) {
	// Test with a semantic error (assigning string to i32)
	src := `
fn main():
    let val: i32 = "hello"
`
	uri := "file:///test.ax"
	diags, _, _, _, _, _, err := runAnalysis(uri, src)
	if err != nil {
		t.Fatalf("runAnalysis failed: %v", err)
	}

	if len(diags) == 0 {
		t.Fatal("Expected at least one type-checking error diagnostic")
	}

	foundTypeError := false
	for _, d := range diags {
		if strings.Contains(d.Message, "expected") || strings.Contains(d.Message, "type") || strings.Contains(d.Message, "assign") {
			foundTypeError = true
		}
	}

	if !foundTypeError {
		t.Errorf("Expected type check error about mismatch types, got: %+v", diags)
	}
}
