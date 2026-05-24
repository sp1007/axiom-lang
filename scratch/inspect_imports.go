package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\valid_actor_spawn.exe")
	if err != nil {
		log.Fatalf("Error opening PE: %v", err)
	}
	defer file.Close()

	fmt.Printf("File machine: 0x%X\n", file.Machine)

	// Dump imports
	imports, err := file.ImportedSymbols()
	if err != nil {
		fmt.Printf("Error getting imported symbols: %v\n", err)
		return
	}

	fmt.Printf("\nImported Symbols (%d):\n", len(imports))
	for _, imp := range imports {
		fmt.Println("  ", imp)
	}

	// Also print sections to see what sections exist
	fmt.Println("\nSections:")
	for _, sec := range file.Sections {
		fmt.Printf("  Name: %-8s  VirtSize: 0x%X  VirtAddr: 0x%X  RawSize: %d\n",
			sec.Name, sec.VirtualSize, sec.VirtualAddress, sec.Size)
	}
}
