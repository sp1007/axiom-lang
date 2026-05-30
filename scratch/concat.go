//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	workspaceDir := "."
	
	files := []string{
		"bootstrap/stage1/print_helpers.ax",
		"bootstrap/stage1/token.ax",
		"bootstrap/stage1/lexer.ax",
		"bootstrap/stage1/ast.ax",
		"bootstrap/stage1/intern.ax",
		"bootstrap/stage1/parser.ax",
		"bootstrap/stage1/resolver.ax",
		"bootstrap/stage1/typetable.ax",
		"bootstrap/stage1/mono.ax",
		"bootstrap/stage1/typecheck.ax",
		"bootstrap/stage1/connection_graph.ax",
		"bootstrap/stage1/ownership.ax",
		"bootstrap/stage1/escape.ax",
		"bootstrap/stage1/ctgc.ax",
		"bootstrap/stage1/alias_reuse.ax",
		"bootstrap/stage1/air.ax",
		"bootstrap/stage1/air_builder.ax",
		"bootstrap/stage1/ssa_opt.ax",
		"bootstrap/stage1/cgen.ax",
		"bootstrap/stage1/wasm.ax",
		"bootstrap/stage1/x86_regs.ax",
		"bootstrap/stage1/x86_selector.ax",
		"bootstrap/stage1/x86_regalloc.ax",
		"bootstrap/stage1/x86_asm_emitter.ax",
		"bootstrap/stage1/x86_modrm.ax",
		"bootstrap/stage1/x86_encoding.ax",
		"bootstrap/stage1/x86_emitter.ax",
		"bootstrap/stage1/x86_elf64.ax",
		"bootstrap/stage1/x86_coff.ax",
		"bootstrap/stage1/linker.ax",
		"bootstrap/stage1/fmt.ax",
		"bootstrap/stage1/main_air.ax",
	}

	var imports []string
	var body []string
	for _, f := range files {
		p := filepath.Join(workspaceDir, f)
		content, err := os.ReadFile(p)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", p, err)
			os.Exit(1)
		}
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "import ") {
				if strings.HasPrefix(trimmed, "import bootstrap.stage1.") {
					continue
				}
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
	
	outputPath := filepath.Join(workspaceDir, "bootstrap/stage1/tmp_concatenated_air.ax")
	err := os.WriteFile(outputPath, []byte(result), 0644)
	if err != nil {
		fmt.Printf("Error writing output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully generated tmp_concatenated_air.ax")
}
