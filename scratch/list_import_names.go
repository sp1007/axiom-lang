package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	f, err := pe.Open("minimal.exe")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer f.Close()

	imports, err := f.ImportedSymbols()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Imported Symbols:")
	for i, sym := range imports {
		fmt.Printf("  [%3d] Name: %s\n", i, sym)
	}
}
