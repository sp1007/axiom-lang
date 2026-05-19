package main

import (
	"fmt"
	"os"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
)

// runDumpAST implements the "dump-ast" subcommand.
func runDumpAST(filename string) int {
	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "axc: error reading %s: %v\n", filename, err)
		return 1
	}

	// 1. Lex
	tokens, _, lexDiags := lexer.Lex(src)
	
	// 2. Parse
	pool := ast.NewInternPool(256)
	tree, parseDiags := parser.Parse(tokens, src, pool)

	// Combine diagnostics
	diags := append(lexDiags, parseDiags...)

	// Print AST to stdout
	ast.Print(os.Stdout, tree, pool)

	// Print diagnostics to stderr
	if len(diags) > 0 {
		opts := diagnostics.DefaultFormatOptions()
		for _, d := range diags {
			fmt.Fprint(os.Stderr, diagnostics.FormatDiagnostic(d, src, filename, opts))
		}
		return 1
	}

	return 0
}
