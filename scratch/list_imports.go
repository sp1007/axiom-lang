package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\scratch\\test_sizeof.exe")
	if err != nil {
		log.Fatalf("Error opening PE: %v", err)
	}
	defer file.Close()

	imports, err := file.ImportedLibraries()
	if err != nil {
		log.Fatalf("Error getting imports: %v", err)
	}

	fmt.Printf("Imported Libraries in test_sizeof.exe (%d):\n", len(imports))
	for i, lib := range imports {
		fmt.Printf("  [%d] %s\n", i, lib)
	}

	symbols, err := file.ImportedSymbols()
	if err != nil {
		log.Fatalf("Error getting imported symbols: %v", err)
	}

	fmt.Printf("\nImported Symbols (%d):\n", len(symbols))
	for i, sym := range symbols {
		fmt.Printf("  [%3d] %s\n", i, sym)
	}
}
