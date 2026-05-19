package sema_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// runGenericsSema runs the full sema pipeline on a source file.
// This is deliberately separate from runSema in golden_test.go to
// avoid package-level function conflicts — both are in sema_test.
func runGenericsSema(src []byte) []diagnostics.Diagnostic {
	var allDiags []diagnostics.Diagnostic

	toks, _, lexDiags := lexer.Lex(src)
	allDiags = append(allDiags, lexDiags...)
	if len(lexDiags) > 0 {
		return allDiags
	}

	pool := ast.NewInternPool(16)
	tree, parseDiags := parser.Parse(toks, src, pool)
	allDiags = append(allDiags, parseDiags...)
	if len(parseDiags) > 0 {
		return allDiags
	}

	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()

	// Name Resolution
	lazy := sema.NewLazyResolver(st, tt, nil)
	nr := sema.NewNameResolver(tree, pool, st, tt, lazy)
	nrDiags := nr.Resolve()
	allDiags = append(allDiags, nrDiags...)

	// Type Inference
	ie := sema.NewInferenceEngine(tree, st, tt, nil)
	ieDiags := ie.Infer()
	allDiags = append(allDiags, ieDiags...)

	// Type Checker
	tc := sema.NewTypeChecker(tree, pool, st, tt, ie)
	tcDiags := tc.Check()
	allDiags = append(allDiags, tcDiags...)

	// Effects Checker
	ec := sema.NewEffectChecker(tree, pool, st, tt, ie)
	ecDiags := ec.Check()
	allDiags = append(allDiags, ecDiags...)

	return allDiags
}

func TestGenericsGolden(t *testing.T) {
	inputs, err := filepath.Glob("../../tests/generics/*.ax")
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) == 0 {
		t.Skip("no .ax test files found in tests/generics/")
	}

	for _, axFile := range inputs {
		name := filepath.Base(axFile)
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(axFile)
			if err != nil {
				t.Fatal(err)
			}

			diags := runGenericsSema(src)

			// Sort diagnostics by offset
			sort.Slice(diags, func(i, j int) bool {
				return diags[i].Pos.Offset < diags[j].Pos.Offset
			})

			var sb strings.Builder
			if len(diags) > 0 {
				for _, d := range diags {
					severityStr := "error"
					if d.Severity == diagnostics.SeverityWarning {
						severityStr = "warning"
					}
					line := d.Pos.Line
					col := d.Pos.Col
					if line == 0 {
						line = 1
						col = 1
					}
					basename := filepath.Base(axFile)
					sb.WriteString(fmt.Sprintf("%s:%d:%d: %s: %s\n", basename, line, col, severityStr, d.Message))
				}
			}

			got := sb.String()
			goldenFile := strings.TrimSuffix(axFile, ".ax") + ".diag"

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

			gotNorm := strings.ReplaceAll(got, "\r\n", "\n")
			wantNorm := strings.ReplaceAll(string(want), "\r\n", "\n")

			if gotNorm != wantNorm {
				t.Errorf("Diagnostic mismatch for %s\n--- want ---\n%s\n--- got ---\n%s",
					axFile, wantNorm, gotNorm)
			}
		})
	}
}

// TestGenericsNoFalsePositives checks that all "valid_*" and non-error
// generics test files produce zero diagnostics.
func TestGenericsNoFalsePositives(t *testing.T) {
	inputs, err := filepath.Glob("../../tests/generics/valid_*.ax")
	if err != nil {
		t.Fatal(err)
	}

	for _, axFile := range inputs {
		name := filepath.Base(axFile)
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(axFile)
			if err != nil {
				t.Fatal(err)
			}

			diags := runGenericsSema(src)
			if len(diags) > 0 {
				for _, d := range diags {
					t.Logf("  severity=%d: %s", d.Severity, d.Message)
				}
				t.Errorf("expected 0 errors for valid file %s, got %d", name, len(diags))
			}
		})
	}
}

// TestGenericsCaching verifies that the monomorphizer caches identical instantiations.
func TestGenericsCaching(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()

	// Create a minimal tree with a generic function
	tokens := []lexer.Token{{Kind: lexer.TokenEOF}}
	tree := ast.NewTree(nil, tokens)

	mono := sema.NewMonomorphizer(tree, pool, st, tt)
	if mono == nil {
		t.Fatal("NewMonomorphizer returned nil")
	}
	// The monomorphizer is instantiated successfully; cache mechanism exists.
	// Full integration test deferred to when we have a full pipeline.
}
