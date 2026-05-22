//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	files := []string{
		"token.ax", "lexer.ax", "ast.ax", "intern.ax", "parser.ax", "resolver.ax",
		"typecheck.ax", "air.ax", "air_builder.ax", "main_air.ax",
	}
	cumulative := 0
	for _, f := range files {
		path := filepath.Join("bootstrap", "stage1", f)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", f, err)
			continue
		}
		size := len(data)
		cumulative += size
		fmt.Printf("%s: size=%d, cumulative_end=%d\n", f, size, cumulative)
	}
}
