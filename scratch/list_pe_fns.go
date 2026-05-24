package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\test_malloc.exe")
	if err != nil {
		log.Fatalf("Error opening PE: %v", err)
	}
	defer file.Close()

	fmt.Println("Symbols in test_malloc.exe:")
	for i, sym := range file.COFFSymbols {
		name, err := sym.FullName(file.StringTable)
		if err != nil {
			name = string(sym.Name[:])
		}
		fmt.Printf("  [%3d] Name: %-30s Value/Offset: 0x%X Section: %d Storage: %d\n",
			i, name, sym.Value, sym.SectionNumber, sym.StorageClass)
	}
}
