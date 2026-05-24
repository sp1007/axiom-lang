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
	tokenPath := filepath.Join(workspaceDir, "bootstrap/stage1/token.ax")
	lexerPath := filepath.Join(workspaceDir, "bootstrap/stage1/lexer.ax")
	astPath := filepath.Join(workspaceDir, "bootstrap/stage1/ast.ax")
	internPath := filepath.Join(workspaceDir, "bootstrap/stage1/intern.ax")
	parserPath := filepath.Join(workspaceDir, "bootstrap/stage1/parser.ax")
	resolverPath := filepath.Join(workspaceDir, "bootstrap/stage1/resolver.ax")
	typecheckPath := filepath.Join(workspaceDir, "bootstrap/stage1/typecheck.ax")
	airPath := filepath.Join(workspaceDir, "bootstrap/stage1/air.ax")
	airBuilderPath := filepath.Join(workspaceDir, "bootstrap/stage1/air_builder.ax")
	x86RegsPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_regs.ax")
	x86SelectorPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_selector.ax")
	x86RegallocPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_regalloc.ax")
	x86AsmEmitterPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_asm_emitter.ax")
	x86ModrmPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_modrm.ax")
	x86EncodingPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_encoding.ax")
	x86EmitterPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_emitter.ax")
	x86CoffPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_coff.ax")
	x86Elf64Path := filepath.Join(workspaceDir, "bootstrap/stage1/x86_elf64.ax")
	linkerPath := filepath.Join(workspaceDir, "bootstrap/stage1/linker.ax")
	mainPath := filepath.Join(workspaceDir, "bootstrap/stage1/main_air.ax")

	paths := []string{
		tokenPath, lexerPath, astPath, internPath, parserPath, resolverPath, typecheckPath,
		airPath, airBuilderPath, x86RegsPath, x86SelectorPath, x86RegallocPath, x86AsmEmitterPath,
		x86ModrmPath, x86EncodingPath, x86EmitterPath, x86Elf64Path, x86CoffPath,
		linkerPath, mainPath,
	}
	
	var imports []string
	var body []string
	for _, p := range paths {
		content, err := os.ReadFile(p)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", p, err)
			os.Exit(1)
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
	
	outputPath := filepath.Join(workspaceDir, "bootstrap/stage1/tmp_concatenated_air.ax")
	err := os.WriteFile(outputPath, []byte(result), 0644)
	if err != nil {
		fmt.Printf("Error writing output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully generated tmp_concatenated_air.ax")
}
