package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\axiom_temp.obj")
	if err != nil {
		log.Fatalf("Error opening COFF object: %v", err)
	}
	defer file.Close()

	fmt.Printf("Sections of axiom_temp.obj:\n")
	for _, sec := range file.Sections {
		fmt.Printf("  Section Name: %s, Size: %d, Offset: 0x%X\n", sec.Name, sec.Size, sec.Offset)
	}

	fmt.Printf("\nSymbols in axiom_temp.obj:\n")
	for i, sym := range file.COFFSymbols {
		name, _ := sym.FullName(file.StringTable)
		if sym.SectionNumber > 0 {
			fmt.Printf("  Sym %d: %s, SectionNumber: %d, Value: 0x%X\n", i, name, sym.SectionNumber, sym.Value)
		}
	}
}
