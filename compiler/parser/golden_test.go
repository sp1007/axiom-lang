package parser_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
)

var update = flag.Bool("update", false, "update golden files")

func TestParserGolden(t *testing.T) {
	inputs, err := filepath.Glob("testdata/*.ax")
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) == 0 {
		t.Fatal("no .ax test files found in testdata/")
	}

	for _, axFile := range inputs {
		name := filepath.Base(axFile)
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(axFile)
			if err != nil {
				t.Fatal(err)
			}

			toks, _, lexDiags := lexer.Lex(src)
			if len(lexDiags) > 0 {
				t.Fatalf("lexer errors: %v", lexDiags)
			}

			pool := ast.NewInternPool(16)
			tree, parseDiags := parser.Parse(toks, src, pool)

			var sb strings.Builder
			ast.Print(&sb, tree, pool)

			if len(parseDiags) > 0 {
				sb.WriteString("\nDIAGNOSTICS:\n")
				for _, d := range parseDiags {
					sb.WriteString(diagnostics.FormatDiagnostic(d, src, axFile, diagnostics.DefaultFormatOptions()))
				}
			}

			got := sb.String()
			goldenFile := strings.TrimSuffix(axFile, ".ax") + ".ast"

			if *update {
				if err := os.WriteFile(goldenFile, []byte(got), 0644); err != nil {
					t.Fatal(err)
				}
				t.Logf("updated %s", goldenFile)
				return
			}

			want, err := os.ReadFile(goldenFile)
			if err != nil {
				t.Fatalf("missing golden file %s; run with -update to create", goldenFile)
			}

			// Normalize newlines for Windows
			gotNorm := strings.ReplaceAll(got, "\r\n", "\n")
			wantNorm := strings.ReplaceAll(string(want), "\r\n", "\n")

			if gotNorm != wantNorm {
				t.Errorf("AST mismatch for %s\n--- want ---\n%s\n--- got ---\n%s",
					axFile, wantNorm, gotNorm)
			}
		})
	}
}
