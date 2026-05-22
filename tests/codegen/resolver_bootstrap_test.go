package codegen_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

func TestStage1ResolverCorpus(t *testing.T) {
	// 1. Concatenate all self-hosted compiler frontend files including resolver
	workspaceDir := "../.." // relative to tests/codegen
	tokenPath := filepath.Join(workspaceDir, "bootstrap/stage1/token.ax")
	lexerPath := filepath.Join(workspaceDir, "bootstrap/stage1/lexer.ax")
	astPath := filepath.Join(workspaceDir, "bootstrap/stage1/ast.ax")
	internPath := filepath.Join(workspaceDir, "bootstrap/stage1/intern.ax")
	parserPath := filepath.Join(workspaceDir, "bootstrap/stage1/parser.ax")
	resolverPath := filepath.Join(workspaceDir, "bootstrap/stage1/resolver.ax")
	mainPath := filepath.Join(workspaceDir, "bootstrap/stage1/main_resolver.ax")

	sourceBytes, err := concatenateAxiomFiles(tokenPath, lexerPath, astPath, internPath, parserPath, resolverPath, mainPath)
	if err != nil {
		t.Fatalf("failed to concatenate resolver files: %v", err)
	}

	// 2. Compile using C Backend
	tmpDir, err := os.MkdirTemp("", "axiom-resolver-bootstrap-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	// defer os.RemoveAll(tmpDir)
	t.Logf("Preserving temp dir: %s", tmpDir)

	binPath := filepath.Join(tmpDir, "resolver_test")
	if filepath.Separator == '\\' {
		binPath += ".exe"
	}

	if err := compileCBackendIgnoringAtDiagnostics(t, sourceBytes, binPath); err != nil {
		t.Fatalf("failed to compile self-hosted resolver: %v", err)
	}

	// 3. Scan all .ax files in the repository corpus that are valid simple programs
	var axFiles []string
	err = filepath.Walk(filepath.Join(workspaceDir, "tests"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ax") {
			if strings.Contains(path, "scratch") || strings.Contains(path, "tmp") || strings.Contains(path, "err_") {
				return nil
			}
			base := filepath.Base(path)
			if strings.HasPrefix(base, "00") || base == "valid_assign.ax" || base == "valid_fibonacci.ax" || base == "valid_shadow.ax" || base == "valid_hello.ax" {
				axFiles = append(axFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to scan corpus: %v", err)
	}

	t.Logf("Found %d simplified .ax files to resolve and compare AST", len(axFiles))

	// 4. Resolve and compare AST structures exactly
	for _, axPath := range axFiles {
		t.Run(filepath.Base(axPath), func(t *testing.T) {
			srcBytes, err := os.ReadFile(axPath)
			if err != nil {
				t.Fatalf("failed to read source file: %v", err)
			}

			// Get reference Go resolved AST
			goTokens, _, _ := lexer.Lex(srcBytes)
			goIntern := ast.NewInternPool(256)
			goTree, goDiags := parser.Parse(goTokens, srcBytes, goIntern)
			if len(goDiags) > 0 {
				t.Fatalf("Go parser failed on %s: %v", axPath, goDiags)
			}

			goTypes := types.NewTypeTable()
			goSymbols := sema.NewSymbolTable(goIntern)
			goResolver := sema.NewNameResolver(goTree, goIntern, goSymbols, goTypes, nil)
			goDiags = goResolver.Resolve()
			if len(goDiags) > 0 {
				t.Skipf("Go resolver failed on %s (skipping expected semantic failure): %v", axPath, goDiags)
			}

			// Get self-hosted Axiom resolver output by running the binary
			cmd := exec.Command(binPath, axPath)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("running self-hosted resolver failed: %v\nStderr: %s", err, stderr.String())
			}

			// Parse self-hosted resolver output AST
			axiomNodes, err := parseASTOutput(stdout.String())
			if err != nil {
				t.Fatalf("failed to parse self-hosted AST output: %v\nStdout:\n%s", err, stdout.String())
			}

			// Compare AST nodes count
			if len(goTree.Nodes) != len(axiomNodes) {
				t.Errorf("AST nodes count mismatch: Go resolver produced %d nodes, Axiom resolver produced %d nodes\nStdout:\n%s",
					len(goTree.Nodes), len(axiomNodes), stdout.String())
			}

			minLen := len(goTree.Nodes)
			if len(axiomNodes) < minLen {
				minLen = len(axiomNodes)
			}

			for i := 0; i < minLen; i++ {
				gn := goTree.Nodes[i]
				an := axiomNodes[i]

				if uint8(gn.Kind) != an.Kind || gn.Flags != an.Flags || gn.TokenIdx != an.TokenIdx ||
					gn.FirstChild != an.FirstChild || gn.NextSibling != an.NextSibling || gn.Payload != an.Payload {
					t.Errorf("Resolved AST node %d mismatch!\nGo:    kind=%d (%s), flags=%d, token_idx=%d, first_child=%d, next_sibling=%d, payload=%d\nAxiom: kind=%d (%s), flags=%d, token_idx=%d, first_child=%d, next_sibling=%d, payload=%d\nStdout:\n%s",
						i, gn.Kind, gn.Kind.String(), gn.Flags, gn.TokenIdx, gn.FirstChild, gn.NextSibling, gn.Payload,
						an.Kind, ast.NodeKind(an.Kind).String(), an.Flags, an.TokenIdx, an.FirstChild, an.NextSibling, an.Payload,
						stdout.String())
					break
				}
			}
		})
	}
}
