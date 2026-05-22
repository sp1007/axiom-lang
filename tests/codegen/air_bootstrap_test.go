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
	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/builder"
)

func TestStage1AirCorpus(t *testing.T) {
	// 1. Concatenate all self-hosted compiler frontend files including air and air_builder
	workspaceDir := "../.." // relative to tests/codegen
	tokenPath := filepath.Join(workspaceDir, "bootstrap/stage1/token.ax")
	lexerPath := filepath.Join(workspaceDir, "bootstrap/stage1/lexer.ax")
	astPath := filepath.Join(workspaceDir, "bootstrap/stage1/ast.ax")
	internPath := filepath.Join(workspaceDir, "bootstrap/stage1/intern.ax")
	parserPath := filepath.Join(workspaceDir, "bootstrap/stage1/parser.ax")
	resolverPath := filepath.Join(workspaceDir, "bootstrap/stage1/resolver.ax")
	typecheckPath := filepath.Join(workspaceDir, "bootstrap/stage1/typecheck.ax")
	connectionGraphPath := filepath.Join(workspaceDir, "bootstrap/stage1/connection_graph.ax")
	ownershipPath := filepath.Join(workspaceDir, "bootstrap/stage1/ownership.ax")
	escapePath := filepath.Join(workspaceDir, "bootstrap/stage1/escape.ax")
	ctgcPath := filepath.Join(workspaceDir, "bootstrap/stage1/ctgc.ax")
	aliasReusePath := filepath.Join(workspaceDir, "bootstrap/stage1/alias_reuse.ax")
	airPath := filepath.Join(workspaceDir, "bootstrap/stage1/air.ax")
	airBuilderPath := filepath.Join(workspaceDir, "bootstrap/stage1/air_builder.ax")
	cgenPath := filepath.Join(workspaceDir, "bootstrap/stage1/cgen.ax")
	mainPath := filepath.Join(workspaceDir, "bootstrap/stage1/main_air.ax")

	sourceBytes, err := concatenateAxiomFiles(
		tokenPath, lexerPath, astPath, internPath, parserPath, resolverPath, typecheckPath,
		connectionGraphPath, ownershipPath, escapePath, ctgcPath, aliasReusePath,
		airPath, airBuilderPath, cgenPath, mainPath,
	)
	if err != nil {
		t.Fatalf("failed to concatenate air files: %v", err)
	}

	// 2. Compile using C Backend
	tmpDir, err := os.MkdirTemp("", "axiom-air-bootstrap-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "air_test")
	if filepath.Separator == '\\' {
		binPath += ".exe"
	}

	if err := compileCBackendIgnoringAtDiagnostics(t, sourceBytes, binPath); err != nil {
		t.Fatalf("failed to compile self-hosted air generator: %v", err)
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
			// Only select files that are valid simple programs
			if strings.HasPrefix(base, "00") || base == "valid_assign.ax" || base == "valid_fibonacci.ax" || base == "valid_shadow.ax" || base == "valid_hello.ax" {
				axFiles = append(axFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to scan corpus: %v", err)
	}

	t.Logf("Found %d simplified .ax files to build AIR and compare", len(axFiles))

	// 4. Build AIR and compare output strings exactly
	for _, axPath := range axFiles {
		t.Run(filepath.Base(axPath), func(t *testing.T) {
			srcBytes, err := os.ReadFile(axPath)
			if err != nil {
				t.Fatalf("failed to read source file: %v", err)
			}

			// Get reference Go AIR builder output
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

			checker := sema.NewTypeChecker(goTree, goIntern, goSymbols, goTypes, ie)
			checkDiags := checker.Check()
			if len(checkDiags) > 0 {
				t.Skipf("Go typechecker verify failed on %s (skipping expected semantic failure): %v", axPath, checkDiags)
			}

			// Build AIR using Go reference builder
			mb := builder.NewModuleBuilder(goTree, goSymbols, goTypes, goIntern)
			goMod := mb.Build()

			var goBuf bytes.Buffer
			air.PrintModule(&goBuf, goMod)
			goOutput := strings.TrimSpace(goBuf.String())

			// Get self-hosted Axiom AIR generator output by running the compiled binary
			cmd := exec.Command(binPath, axPath)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("running self-hosted AIR generator failed: %v\nStderr: %s", err, stderr.String())
			}

			axiomOutput := strings.TrimSpace(stdout.String())

			// Normalise newlines (\r\n -> \n) for comparison
			goOutput = strings.ReplaceAll(goOutput, "\r\n", "\n")
			axiomOutput = strings.ReplaceAll(axiomOutput, "\r\n", "\n")

			// Check for exact match
			if goOutput != axiomOutput {
				t.Errorf("AIR output mismatch on %s!\n\n=== EXPECTED (Go Reference) ===\n%s\n\n=== ACTUAL (Self-Hosted) ===\n%s\n",
					axPath, goOutput, axiomOutput)
			}
		})
	}
}
