package codegen_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

func TestStage1TypecheckCorpus(t *testing.T) {
	// 1. Concatenate all self-hosted compiler frontend files including typecheck
	workspaceDir := "../.." // relative to tests/codegen
	tokenPath := filepath.Join(workspaceDir, "bootstrap/stage1/token.ax")
	lexerPath := filepath.Join(workspaceDir, "bootstrap/stage1/lexer.ax")
	astPath := filepath.Join(workspaceDir, "bootstrap/stage1/ast.ax")
	internPath := filepath.Join(workspaceDir, "bootstrap/stage1/intern.ax")
	parserPath := filepath.Join(workspaceDir, "bootstrap/stage1/parser.ax")
	resolverPath := filepath.Join(workspaceDir, "bootstrap/stage1/resolver.ax")
	typecheckPath := filepath.Join(workspaceDir, "bootstrap/stage1/typecheck.ax")
	mainPath := filepath.Join(workspaceDir, "bootstrap/stage1/main_typecheck.ax")

	sourceBytes, err := concatenateAxiomFiles(tokenPath, lexerPath, astPath, internPath, parserPath, resolverPath, typecheckPath, mainPath)
	if err != nil {
		t.Fatalf("failed to concatenate typechecker files: %v", err)
	}

	// 2. Compile using C Backend
	tmpDir, err := os.MkdirTemp("", "axiom-typecheck-bootstrap-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "typecheck_test")
	if filepath.Separator == '\\' {
		binPath += ".exe"
	}

	if err := compileCBackendIgnoringAtDiagnostics(t, sourceBytes, binPath); err != nil {
		t.Fatalf("failed to compile self-hosted typechecker: %v", err)
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

	t.Logf("Found %d simplified .ax files to typecheck and compare", len(axFiles))

	// 4. Typecheck and compare AST node type assignments exactly
	for _, axPath := range axFiles {
		t.Run(filepath.Base(axPath), func(t *testing.T) {
			srcBytes, err := os.ReadFile(axPath)
			if err != nil {
				t.Fatalf("failed to read source file: %v", err)
			}

			// Get reference Go typechecker output
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

			ie := sema.NewInferenceEngine(goTree, goSymbols, goTypes, nil)
			typeDiags := ie.Infer()
			if len(typeDiags) > 0 {
				t.Skipf("Go typechecker failed on %s (skipping expected semantic failure): %v", axPath, typeDiags)
			}

			// Get self-hosted Axiom typechecker output by running the binary
			cmd := exec.Command(binPath, axPath)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("running self-hosted typechecker failed: %v\nStderr: %s", err, stderr.String())
			}

			// Parse self-hosted typechecker output
			axiomTypes, err := parseTypecheckOutput(stdout.String())
			if err != nil {
				t.Fatalf("failed to parse self-hosted typechecker output: %v\nStdout:\n%s", err, stdout.String())
			}

			// Compare TypeIDs for all nodes in the AST
			for i := 0; i < len(goTree.Nodes); i++ {
				goType := ie.TypeOf(uint32(i))
				axType, exists := axiomTypes[i]
				if !exists {
					t.Errorf("Node %d (kind %s): missing type assignment in self-hosted output", i, goTree.Nodes[i].Kind.String())
					continue
				}

				if uint32(goType) != axType {
					// Check if they are compatible function/struct TypeIDs
					// Go reference might register types slightly differently, so check TypeEntry kind parity
					goEntry := goTypes.Entry(goType)
					// Look up the self-hosted type details or directly report error
					t.Errorf("Node %d (kind %s): TypeID mismatch! Go expected %d (%s kind %v), Axiom produced %d\nStdout:\n%s",
						i, goTree.Nodes[i].Kind.String(), goType, goType.String(), goEntry.Kind, axType, stdout.String())
				}
			}
		})
	}
}

func parseTypecheckOutput(output string) (map[int]uint32, error) {
	nodeTypes := make(map[int]uint32)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "[") {
			continue
		}
		// Expecting format: "%d: type=%d"
		parts := strings.Split(trimmed, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: expected ':' separation, got: %s", i+1, trimmed)
		}
		idx, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid index: %v", i+1, err)
		}
		kv := strings.Split(strings.TrimSpace(parts[1]), "=")
		if len(kv) != 2 || strings.TrimSpace(kv[0]) != "type" {
			return nil, fmt.Errorf("line %d: invalid field: %s", i+1, parts[1])
		}
		val, err := strconv.ParseUint(strings.TrimSpace(kv[1]), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid type: %v", i+1, err)
		}
		nodeTypes[idx] = uint32(val)
	}
	return nodeTypes, nil
}
