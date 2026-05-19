package sema_test

import (
	"flag"
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

var update = flag.Bool("update", false, "update golden files")

func runSema(src []byte, filename string) []diagnostics.Diagnostic {
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

	// 1. Lazy field analysis / import loading (skipped for MVP since no multi-file)
	
	// 2. Name Resolution
	lazy := sema.NewLazyResolver(st, tt, nil)
	nr := sema.NewNameResolver(tree, pool, st, tt, lazy)
	nrDiags := nr.Resolve()
	allDiags = append(allDiags, nrDiags...)

	// 3. Type Inference
	ie := sema.NewInferenceEngine(tree, st, tt, nil)
	ieDiags := ie.Infer()
	allDiags = append(allDiags, ieDiags...)

	// 4. Type Checker
	tc := sema.NewTypeChecker(tree, pool, st, tt, ie)
	tcDiags := tc.Check()
	allDiags = append(allDiags, tcDiags...)

	// 5. Effects Checker
	ec := sema.NewEffectChecker(tree, pool, st, tt, ie)
	ecDiags := ec.Check()
	allDiags = append(allDiags, ecDiags...)

	return allDiags
}

func TestSemaGolden(t *testing.T) {
	inputs, err := filepath.Glob("../../tests/sema/*.ax")
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) == 0 {
		t.Skip("no .ax test files found in tests/sema/")
	}

	for _, axFile := range inputs {
		name := filepath.Base(axFile)
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(axFile)
			if err != nil {
				t.Fatal(err)
			}

			diags := runSema(src, axFile)

			// Sort diagnostics by line/col (simplified, just by Pos offset for now)
			sort.Slice(diags, func(i, j int) bool {
				return diags[i].Pos.Offset < diags[j].Pos.Offset
			})

			var sb strings.Builder
			if len(diags) > 0 {
				for _, d := range diags {
					// Format as: "file.ax:LINE:COL: error: MSG"
					// We just use a minimal format for golden testing so it is stable
					severityStr := "error"
					if d.Severity == diagnostics.SeverityWarning {
						severityStr = "warning"
					}
					// Ideally we map Pos to Line/Col. For now, since AST has limited pos info in MVP,
					// we just use the diagnostic code and message if Line is 0.
					line := d.Pos.Line
					col := d.Pos.Col
					if line == 0 {
						line = 1 // default
						col = 1
					}
					
					// To make tests stable, just use basename
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

func TestNoFalsePositives(t *testing.T) {
	inputs, err := filepath.Glob("../../tests/sema/valid_*.ax")
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

			diags := runSema(src, axFile)
			if len(diags) > 0 {
				t.Errorf("expected 0 errors for valid file, got %d:\n%v", len(diags), diags)
			}
		})
	}
}
