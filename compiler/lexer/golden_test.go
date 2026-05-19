package lexer_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/compiler/lexer"
)

var update = flag.Bool("update", false, "update golden files")

func TestLexerGolden(t *testing.T) {
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

			toks, _, _ := lexer.Lex(src)
			got := formatTokens(toks, src)

			goldenFile := strings.TrimSuffix(axFile, ".ax") + ".tokens"
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
			if got != string(want) {
				t.Errorf("token mismatch for %s\n--- want ---\n%s\n--- got ---\n%s",
					axFile, want, got)
			}
		})
	}
}

// TestGoldenFilesComplete ensures every .ax file has a corresponding .tokens file.
func TestGoldenFilesComplete(t *testing.T) {
	if *update {
		t.Skip("skipping completeness check during -update")
	}
	axFiles, err := filepath.Glob("testdata/*.ax")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range axFiles {
		golden := strings.TrimSuffix(f, ".ax") + ".tokens"
		if _, err := os.Stat(golden); os.IsNotExist(err) {
			t.Errorf("missing golden file for %s", f)
		}
	}
}

// formatTokens converts a token slice to the golden file format.
// Format: KIND offset:len "text"
func formatTokens(toks []lexer.Token, src []byte) string {
	var sb strings.Builder
	for _, tok := range toks {
		text := ""
		if tok.Len > 0 {
			end := tok.Offset + uint32(tok.Len)
			if end <= uint32(len(src)) {
				text = string(src[tok.Offset:end])
			}
		}
		fmt.Fprintf(&sb, "%s %d:%d %q\n", tok.Kind, tok.Offset, tok.Len, text)
	}
	return sb.String()
}
