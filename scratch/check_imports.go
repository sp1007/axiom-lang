package main

import (
	"debug/pe"
	"fmt"
	"os"
)

func main() {
	target := "minimal.exe"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}
	f, err := pe.Open(target)
	if err != nil {
		fmt.Printf("Failed to open PE %s: %v\n", target, err)
		os.Exit(1)
	}
	defer f.Close()

	imports, err := f.ImportedLibraries()
	if err != nil {
		fmt.Printf("Failed to get imported libraries: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Imported Libraries:")
	for _, lib := range imports {
		fmt.Printf("  - %s\n", lib)
	}

	importedSymbols, err := f.ImportedSymbols()
	if err != nil {
		fmt.Printf("Failed to get imported symbols: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nImported Symbols:")
	for _, sym := range importedSymbols {
		fmt.Printf("  - %s\n", sym)
	}
}
