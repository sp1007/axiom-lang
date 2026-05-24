package main

import (
	"debug/pe"
	"fmt"
	"os"
)

func main() {
	f, err := pe.Open("axiom_temp.obj")
	if err != nil {
		fmt.Printf("Failed to open COFF: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	fmt.Println("Symbols in axiom_temp.obj:")
	for _, sym := range f.COFFSymbols {
		name, err := sym.FullName(f.StringTable)
		if err != nil {
			name = string(sym.Name[:])
		}
		if sym.SectionNumber > 0 {
			fmt.Printf("  [DEF] Section %d: %s\n", sym.SectionNumber, name)
		} else if sym.SectionNumber == 0 {
			if sym.StorageClass == 2 { // External
				fmt.Printf("  [UNDEF]: %s\n", name)
			}
		} else {
			fmt.Printf("  [DEBUG/SPECIAL] %d: %s\n", sym.SectionNumber, name)
		}
	}
}
