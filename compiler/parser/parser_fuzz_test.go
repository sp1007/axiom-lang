package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
)

// FuzzParser feeds random byte sequences into the full lex+parse pipeline
// to ensure it never panics, hangs, or crashes on invalid input.
func FuzzParser(f *testing.F) {
	// Add seed corpus from testdata/*.ax
	inputs, err := filepath.Glob("testdata/*.ax")
	if err == nil {
		for _, axFile := range inputs {
			data, err := os.ReadFile(axFile)
			if err == nil {
				f.Add(data)
			}
		}
	}
	
	// Add some additional interesting seed inputs
	seeds := []string{
		"",
		"fn main():\n    return 0\n",
		"let x: i32 = 42",
		"if x > 0:\n    y()\nelse:\n    z()\n",
		"struct Foo:\n    pub mut x: i32\n",
		"match x:\n    Ok(v): v\n    Err(e): panic(e)\n",
		"pub async fn fetch(url: string) -> Future[string]:",
		"\t\t\t\t", // tabs
		"   \n   \n   \n",
		"@#$%^&*()!~`",
		"\x00\x01\x02\xff\xfe\xfd",
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Must not panic
		tokens, _, _ := lexer.Lex(data)
		pool := ast.NewInternPool(16)
		tree, _ := parser.Parse(tokens, data, pool)

		// Basic invariants
		if tree == nil {
			t.Fatal("Parse() returned nil AstTree")
		}
		if tree.NodeCount() == 0 {
			t.Fatal("tree has no nodes (missing root)")
		}
		if tree.Nodes[0].Kind != ast.NodeProgram {
			t.Fatal("root must be NodeProgram")
		}
	})
}
