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

	"github.com/axiom-lang/axiom/compiler/lexer"
)

func TestStage1LexerCorpus(t *testing.T) {
	// 1. Concatenate the self-hosted lexer files
	workspaceDir := "../.." // relative to tests/codegen
	tokenPath := filepath.Join(workspaceDir, "bootstrap/stage1/token.ax")
	lexerPath := filepath.Join(workspaceDir, "bootstrap/stage1/lexer.ax")
	mainPath := filepath.Join(workspaceDir, "bootstrap/stage1/main.ax")

	sourceBytes, err := concatenateAxiomFiles(tokenPath, lexerPath, mainPath)
	if err != nil {
		t.Fatalf("failed to concatenate lexer files: %v", err)
	}
	// Write concatenated file for inspection
	_ = os.WriteFile(filepath.Join(workspaceDir, "bootstrap/stage1/tmp_concatenated_lexer.ax"), sourceBytes, 0644)


	// 2. Compile using C Backend
	tmpDir, err := os.MkdirTemp("", "axiom-lexer-bootstrap-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "lexer_test")
	if filepath.Separator == '\\' {
		binPath += ".exe"
	}

	if err := compileCBackendIgnoringAtDiagnostics(t, sourceBytes, binPath); err != nil {
		t.Fatalf("failed to compile self-hosted lexer: %v", err)
	}

	// 3. Scan all .ax files in the repository corpus
	var axFiles []string
	err = filepath.Walk(workspaceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ax") {
			// Skip temporary scratch/test files in workspace to avoid noise
			if strings.Contains(path, "scratch") || strings.Contains(path, "tmp") {
				return nil
			}
			axFiles = append(axFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to scan corpus: %v", err)
	}

	t.Logf("Found %d .ax files to tokenize and compare", len(axFiles))

	// 4. Tokenize and compare for each file
	for _, axPath := range axFiles {
		t.Run(filepath.Base(axPath), func(t *testing.T) {
			// Read reference file content
			srcBytes, err := os.ReadFile(axPath)
			if err != nil {
				t.Fatalf("failed to read source file: %v", err)
			}

			// Get reference Go lexer output
			goTokens, _, _ := lexer.Lex(srcBytes)

			// Get self-hosted Axiom lexer output by running the binary
			cmd := exec.Command(binPath, axPath)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("running self-hosted lexer failed: %v\nStderr: %s", err, stderr.String())
			}

			// Parse self-hosted lexer CSV tokens
			axiomTokens, err := parseCSVTokens(stdout.String())
			if err != nil {
				t.Fatalf("failed to parse self-hosted CSV output: %v", err)
			}

			// Compare token streams exactly
			if len(goTokens) != len(axiomTokens) {
				t.Errorf("token count mismatch: Go lexer produced %d tokens, Axiom lexer produced %d tokens",
					len(goTokens), len(axiomTokens))
			}

			minLen := len(goTokens)
			if len(axiomTokens) < minLen {
				minLen = len(axiomTokens)
			}

			for i := 0; i < minLen; i++ {
				gt := goTokens[i]
				at := axiomTokens[i]

				if uint8(gt.Kind) != at.Kind || gt.Offset != at.Offset || gt.Len != at.Len {
					t.Errorf("token %d mismatch!\nGo:    kind=%d (%s), offset=%d, len=%d\nAxiom: kind=%d (%s), offset=%d, len=%d",
						i, gt.Kind, gt.Kind.String(), gt.Offset, gt.Len,
						at.Kind, lexer.TokenKind(at.Kind).String(), at.Offset, at.Len)
					
					// Dump context around mismatch
					start := int(gt.Offset) - 20
					if start < 0 {
						start = 0
					}
					end := int(gt.Offset) + int(gt.Len) + 20
					if end > len(srcBytes) {
						end = len(srcBytes)
					}
					t.Logf("Source context: %q", string(srcBytes[start:end]))
					break
				}
			}
		})
	}
}

type ParsedToken struct {
	Kind   uint8
	Offset uint32
	Len    uint16
}

func parseCSVTokens(csv string) ([]ParsedToken, error) {
	var tokens []ParsedToken
	lines := strings.Split(strings.TrimSpace(csv), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return tokens, nil
	}
	for i, line := range lines {
		parts := strings.Split(strings.TrimSpace(line), ",")
		if len(parts) != 3 {
			return nil, fmt.Errorf("line %d: expected 3 parts, got %d", i+1, len(parts))
		}
		kind, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		offset, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		length, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, ParsedToken{
			Kind:   uint8(kind),
			Offset: uint32(offset),
			Len:    uint16(length),
		})
	}
	return tokens, nil
}

func concatenateAxiomFiles(paths ...string) ([]byte, error) {
	var imports []string
	var body []string
	for _, p := range paths {
		content, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "import ") {
				imports = append(imports, line)
			} else {
				body = append(body, line)
			}
		}
	}
	importMap := make(map[string]bool)
	var uniqueImports []string
	for _, imp := range imports {
		trimmed := strings.TrimSpace(imp)
		if !importMap[trimmed] {
			importMap[trimmed] = true
			uniqueImports = append(uniqueImports, imp)
		}
	}
	result := strings.Join(uniqueImports, "\n") + "\n\n" + strings.Join(body, "\n")
	return []byte(result), nil
}
