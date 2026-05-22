//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

type FileOffset struct {
	Name  string
	Start int
	End   int
}

func main() {
	files := []string{
		"token.ax", "lexer.ax", "ast.ax", "intern.ax", "parser.ax", "resolver.ax",
		"typecheck.ax", "air.ax", "air_builder.ax", "main_air.ax",
	}

	var sourceBuilder strings.Builder
	var fileOffsets []FileOffset

	for _, f := range files {
		path := filepath.Join("bootstrap", "stage1", f)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", f, err)
			continue
		}
		start := sourceBuilder.Len()
		sourceBuilder.Write(data)
		end := sourceBuilder.Len()
		fileOffsets = append(fileOffsets, FileOffset{Name: f, Start: start, End: end})
	}

	sourceBytes := []byte(sourceBuilder.String())

	// Lex
	tokens, _, _ := lexer.Lex(sourceBytes)

	// Parse
	intern := ast.NewInternPool(256)
	tree, parseDiags := parser.Parse(tokens, sourceBytes, intern)
	if len(parseDiags) > 0 {
		fmt.Printf("Parse errors: %v\n", parseDiags)
		return
	}

	symbols := sema.NewSymbolTable(intern)
	table := types.NewTypeTable()
	resolver := sema.NewNameResolver(tree, intern, symbols, table, nil)
	errs := resolver.Resolve()
	if len(errs) > 0 {
		fmt.Println("=== Name Resolution Errors ===")
		for _, e := range errs {
			// Find which file the error offset belongs to
			origFile := "unknown"
			lineNum := 1
			charOffset := 0
			tokOffset := int(e.Pos.Offset)
			snippet := ""
			if tokOffset >= 0 && tokOffset < len(sourceBytes) {
				end := tokOffset + 60
				if end > len(sourceBytes) {
					end = len(sourceBytes)
				}
				snippet = string(sourceBytes[tokOffset:end])
				snippet = strings.ReplaceAll(snippet, "\n", " <NL> ")
			}
			
			for _, fo := range fileOffsets {
				if tokOffset >= fo.Start && tokOffset < fo.End {
					origFile = fo.Name
					// Calculate line number in that file
					fileData, _ := os.ReadFile(filepath.Join("bootstrap", "stage1", fo.Name))
					relOffset := tokOffset - fo.Start
					for j := 0; j < relOffset && j < len(fileData); j++ {
						if fileData[j] == '\n' {
							lineNum++
						}
					}
					charOffset = relOffset
					break
				}
			}
			fmt.Printf("Error: %s, File: %s, Line: %d, Offset: %d, Snippet: %q\n", 
				e.Message, origFile, lineNum, charOffset, snippet)
		}
	}

}
