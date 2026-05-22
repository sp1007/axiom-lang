package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
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

	// Walk all AST nodes
	fmt.Println("=== Walk all nodes for idx or U32Vec ===")
	for i := 0; i < len(tree.Nodes); i++ {
		node := tree.Nodes[i]
		
		// Get payload as name if relevant
		var name string
		if node.Kind == ast.NodeFuncDecl || node.Kind == ast.NodeStructDecl || node.Kind == ast.NodeVarDecl || node.Kind == ast.NodeParamDecl || node.Kind == ast.NodeFieldDecl || node.Kind == ast.NodeConstDecl {
			if node.Payload != 0 {
				name = string(intern.Get(node.Payload))
			}
		}

		if name == "idx" || name == "U32Vec" {
			tokOffset := int(tokens[node.TokenIdx].Offset)
			origFile := "unknown"
			lineNum := 1
			for _, fo := range fileOffsets {
				if tokOffset >= fo.Start && tokOffset < fo.End {
					origFile = fo.Name
					fileData, _ := os.ReadFile(filepath.Join("bootstrap", "stage1", fo.Name))
					relOffset := tokOffset - fo.Start
					for j := 0; j < relOffset && j < len(fileData); j++ {
						if fileData[j] == '\n' {
							lineNum++
						}
					}
					break
				}
			}
			fmt.Printf("NodeIdx: %d, Kind: %v, Name: %s, File: %s, Line: %d\n", i, node.Kind, name, origFile, lineNum)
		}
	}
}
