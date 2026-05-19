package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
)

// tokenJSON is the JSON-serializable representation of a token.
type tokenJSON struct {
	Kind   string `json:"kind"`
	Offset uint32 `json:"offset"`
	Len    uint16 `json:"len"`
	Text   string `json:"text,omitempty"`
}

// runDumpTokens implements the "dump-tokens" subcommand.
// It tokenizes the given file and prints the token stream as JSON.
func runDumpTokens(filename string, compact, noText, stats bool) int {
	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "axc: error reading %s: %v\n", filename, err)
		return 1
	}

	tokens, _, diags := lexer.Lex(src)

	// Print any diagnostics to stderr
	for _, d := range diags {
		opts := diagnostics.DefaultFormatOptions()
		fmt.Fprint(os.Stderr, diagnostics.FormatDiagnostic(d, src, filename, opts))
	}

	if stats {
		printTokenStats(tokens)
		return exitCode(diags)
	}

	// Build JSON array
	jsonTokens := make([]tokenJSON, len(tokens))
	for i, tok := range tokens {
		jt := tokenJSON{
			Kind:   tok.Kind.String(),
			Offset: tok.Offset,
			Len:    tok.Len,
		}
		if !noText && tok.Len > 0 {
			end := tok.Offset + uint32(tok.Len)
			if end <= uint32(len(src)) {
				jt.Text = string(src[tok.Offset:end])
			}
		}
		jsonTokens[i] = jt
	}

	var output []byte
	if compact {
		output, _ = json.Marshal(jsonTokens)
	} else {
		output, _ = json.MarshalIndent(jsonTokens, "", "  ")
	}
	fmt.Println(string(output))

	return exitCode(diags)
}

// printTokenStats prints a summary of token counts by kind.
func printTokenStats(tokens []lexer.Token) {
	counts := make(map[string]int)
	for _, tok := range tokens {
		counts[tok.Kind.String()]++
	}

	// Sort kinds alphabetically
	kinds := make([]string, 0, len(counts))
	for k := range counts {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)

	fmt.Println("Token Statistics:")
	total := 0
	for _, k := range kinds {
		fmt.Printf("  %-20s %d\n", k+":", counts[k])
		total += counts[k]
	}
	fmt.Printf("  %-20s %d\n", "Total:", total)
}

// exitCode returns 0 if no errors, 1 if any error diagnostics exist.
func exitCode(diags []diagnostics.Diagnostic) int {
	for _, d := range diags {
		if d.Severity == diagnostics.SeverityError {
			return 1
		}
	}
	return 0
}
