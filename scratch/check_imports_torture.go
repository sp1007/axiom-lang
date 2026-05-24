package main

import (
	"debug/pe"
	"fmt"
	"os"
)

func main() {
	f, err := pe.Open("torture_gen_ref.exe")
	if err != nil {
		fmt.Printf("Error opening PE: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	imports, err := f.ImportedLibraries()
	if err != nil {
		fmt.Printf("Error reading imported libraries: %v\n", err)
	}
	fmt.Println("Imported Libraries:")
	for _, lib := range imports {
		fmt.Printf("  - %s\n", lib)
	}

	symbols, err := f.ImportedSymbols()
	if err != nil {
		fmt.Printf("Error reading imported symbols: %v\n", err)
	}
	fmt.Println("\nImported Symbols:")
	for _, sym := range symbols {
		fmt.Printf("  - %s\n", sym)
	}
}
