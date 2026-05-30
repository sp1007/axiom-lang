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
)

func TestStage1ParserCorpus(t *testing.T) {
	// 1. Concatenate all self-hosted compiler frontend files
	workspaceDir := "../.." // relative to tests/codegen
	printHelpersPath := filepath.Join(workspaceDir, "bootstrap/stage1/print_helpers.ax")
	tokenPath := filepath.Join(workspaceDir, "bootstrap/stage1/token.ax")
	lexerPath := filepath.Join(workspaceDir, "bootstrap/stage1/lexer.ax")
	astPath := filepath.Join(workspaceDir, "bootstrap/stage1/ast.ax")
	internPath := filepath.Join(workspaceDir, "bootstrap/stage1/intern.ax")
	parserPath := filepath.Join(workspaceDir, "bootstrap/stage1/parser.ax")
	mainPath := filepath.Join(workspaceDir, "bootstrap/stage1/main_parser.ax")

	sourceBytes, err := concatenateAxiomFiles(printHelpersPath, tokenPath, lexerPath, astPath, internPath, parserPath, mainPath)
	if err != nil {
		t.Fatalf("failed to concatenate parser files: %v", err)
	}
	// Write concatenated file for inspection
	_ = os.WriteFile(filepath.Join(workspaceDir, "bootstrap/stage1/tmp_concatenated_parser.ax"), sourceBytes, 0644)

	// 2. Compile using C Backend
	tmpDir, err := os.MkdirTemp("", "axiom-parser-bootstrap-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "parser_test")
	if filepath.Separator == '\\' {
		binPath += ".exe"
	}

	if err := compileCBackendIgnoringAtDiagnostics(t, sourceBytes, binPath, ""); err != nil {
		t.Fatalf("failed to compile self-hosted parser: %v", err)
	}

	// 3. Scan all .ax files in the repository corpus that are valid simple programs
	// Wait, since parser.ax currently implements a subset of features (like pub/fn declarations, simple assignments/variables, blocks, control flow, arithmetic/Pratt expressions),
	// we will run comparison tests on the valid basic test programs in tests/codegen/ or simple programs in tests/sema/.
	var axFiles []string
	err = filepath.Walk(filepath.Join(workspaceDir, "tests"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ax") {
			// Skip temporary scratch/test files in workspace to avoid noise
			if strings.Contains(path, "scratch") || strings.Contains(path, "tmp") || strings.Contains(path, "err_") {
				return nil
			}
			// Only parse simpler semantic test files or basic tests, as complex nested ADT / multi-module is in Go
			// Let's include simple valid ones:
			base := filepath.Base(path)
			if strings.HasPrefix(base, "00") || base == "valid_hello.ax" || base == "valid_assign.ax" || base == "valid_fibonacci.ax" || base == "valid_shadow.ax" {
				axFiles = append(axFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to scan corpus: %v", err)
	}

	t.Logf("Found %d simplified .ax files to parse and compare AST", len(axFiles))

	// 4. Parse and compare AST structures exactly
	for _, axPath := range axFiles {
		t.Run(filepath.Base(axPath), func(t *testing.T) {
			// Read reference file content
			srcBytes, err := os.ReadFile(axPath)
			if err != nil {
				t.Fatalf("failed to read source file: %v", err)
			}

			// Get reference Go parser AST
			goTokens, _, _ := lexer.Lex(srcBytes)
			goIntern := ast.NewInternPool(256)
			goTree, goDiags := parser.Parse(goTokens, srcBytes, goIntern)
			if len(goDiags) > 0 {
				t.Fatalf("Go parser failed on %s: %v", axPath, goDiags)
			}

			// Get self-hosted Axiom parser output by running the binary
			cmd := exec.Command(binPath, axPath)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("running self-hosted parser failed: %v\nStderr: %s", err, stderr.String())
			}

			// Parse self-hosted parser output AST
			axiomNodes, err := parseASTOutput(stdout.String())
			if err != nil {
				t.Fatalf("failed to parse self-hosted AST output: %v\nStdout:\n%s", err, stdout.String())
			}

			// Compare AST nodes count
			if len(goTree.Nodes) != len(axiomNodes) {
				t.Errorf("AST nodes count mismatch: Go parser produced %d nodes, Axiom parser produced %d nodes\nStdout:\n%s",
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
					t.Errorf("AST node %d mismatch!\nGo:    kind=%d (%s), flags=%d, token_idx=%d, first_child=%d, next_sibling=%d, payload=%d\nAxiom: kind=%d (%s), flags=%d, token_idx=%d, first_child=%d, next_sibling=%d, payload=%d",
						i, gn.Kind, gn.Kind.String(), gn.Flags, gn.TokenIdx, gn.FirstChild, gn.NextSibling, gn.Payload,
						an.Kind, ast.NodeKind(an.Kind).String(), an.Flags, an.TokenIdx, an.FirstChild, an.NextSibling, an.Payload)
					break
				}
			}
		})
	}
}

type ParsedNode struct {
	Kind        uint8
	Flags       uint16
	TokenIdx    uint32
	FirstChild  uint32
	NextSibling uint32
	Payload     uint32
	ExtraIdx    uint32
}

func parseASTOutput(output string) ([]ParsedNode, error) {
	var nodes []ParsedNode
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nodes, nil
	}
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "[") {
			continue
		}
		// Expecting format: "0: kind=1, flags=0, token_idx=0, first_child=0, next_sibling=0, payload=0, extra_idx=0"
		parts := strings.Split(trimmed, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: expected ':' separation, got: %s", i+1, trimmed)
		}
		idx, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid index: %v", i+1, err)
		}
		if idx != len(nodes) {
			return nil, fmt.Errorf("line %d: node index mismatch, expected %d, got %d", i+1, len(nodes), idx)
		}

		fields := strings.Split(strings.TrimSpace(parts[1]), ",")
		var node ParsedNode
		for _, field := range fields {
			kv := strings.Split(strings.TrimSpace(field), "=")
			if len(kv) != 2 {
				return nil, fmt.Errorf("line %d: invalid field: %s", i+1, field)
			}
			key := strings.TrimSpace(kv[0])
			val, err := strconv.Atoi(strings.TrimSpace(kv[1]))
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid field value: %v", i+1, err)
			}

			switch key {
			case "kind":
				node.Kind = uint8(val)
			case "flags":
				node.Flags = uint16(val)
			case "token_idx":
				node.TokenIdx = uint32(val)
			case "first_child":
				node.FirstChild = uint32(val)
			case "next_sibling":
				node.NextSibling = uint32(val)
			case "payload":
				node.Payload = uint32(val)
			case "extra_idx":
				node.ExtraIdx = uint32(val)
			default:
				return nil, fmt.Errorf("line %d: unknown field key: %s", i+1, key)
			}
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
