package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\axiom_temp.obj")
	if err != nil {
		log.Fatalf("Error opening PE: %v", err)
	}
	defer file.Close()

	for _, sec := range file.Sections {
		if sec.Name == ".text" {
			fmt.Printf("Relocations in .text section (total %d):\n", len(sec.Relocs))
			for i, r := range sec.Relocs {
				sym := file.COFFSymbols[r.SymbolTableIndex]
				name, err := sym.FullName(file.StringTable)
				if err != nil {
					name = string(sym.Name[:])
				}
				fmt.Printf("  [%3d] Offset: 0x%04X (RVA: 0x%04X) Symbol: %-30s Type: 0x%X\n",
					i, r.VirtualAddress, r.VirtualAddress+0x1000, name, r.Type)
			}
		}
	}
}
