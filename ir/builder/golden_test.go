package builder_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/builder"
)

// TestAIRGolden runs golden tests: each .ax file in tests/air/ has a matching
// .air file with the expected AIR text output. If the .air file is missing,
// it is created (update mode). If present, the output is compared.
func TestAIRGolden(t *testing.T) {
	testDir := filepath.Join("..", "..", "tests", "air")

	entries, err := os.ReadDir(testDir)
	if err != nil {
		t.Skipf("cannot read golden test dir %s: %v", testDir, err)
	}

	var axFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".ax") {
			axFiles = append(axFiles, filepath.Join(testDir, e.Name()))
		}
	}

	if len(axFiles) == 0 {
		t.Skip("no .ax golden test files found")
	}

	for _, axPath := range axFiles {
		name := strings.TrimSuffix(filepath.Base(axPath), ".ax")
		t.Run(name, func(t *testing.T) {
			// Read source
			source, err := os.ReadFile(axPath)
			if err != nil {
				t.Fatalf("cannot read %s: %v", axPath, err)
			}

			// Compile to AIR
			tokens, _, _ := lexer.Lex(source)
			intern := ast.NewInternPool(256)
			tree, _ := parser.Parse(tokens, source, intern)
			table := types.NewTypeTable()
			symbols := sema.NewSymbolTable(intern)

			mb := builder.NewModuleBuilder(tree, symbols, table, intern)
			mod := mb.Build()

			// Print AIR
			var sb strings.Builder
			air.PrintModule(&sb, mod)
			got := sb.String()

			// Check/update golden file
			goldenPath := strings.TrimSuffix(axPath, ".ax") + ".air"
			updateGolden := os.Getenv("UPDATE_GOLDEN") == "1"

			goldenData, err := os.ReadFile(goldenPath)
			if err != nil || updateGolden {
				// Write golden file
				if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
					t.Fatalf("cannot write golden file %s: %v", goldenPath, err)
				}
				if updateGolden {
					t.Logf("updated golden file: %s", goldenPath)
				} else {
					t.Logf("created golden file: %s", goldenPath)
				}
				return
			}

			want := string(goldenData)
			if got != want {
				t.Errorf("AIR output mismatch for %s\n--- WANT ---\n%s\n--- GOT ---\n%s", name, want, got)
			}
		})
	}
}
