package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\axiom_temp.obj")
	if err != nil {
		log.Fatalf("Error opening PE obj: %v", err)
	}
	defer file.Close()

	fmt.Printf("Total COFF Symbols: %d\n", len(file.COFFSymbols))
	for i, sym := range file.COFFSymbols {
		name, err := sym.FullName(file.StringTable)
		if err == nil {
			fmt.Printf("Sym [%d]: Name=%s, Value=%d, Section=%d, Class=%d, Type=%d\n",
				i, name, sym.Value, sym.SectionNumber, sym.StorageClass, sym.Type)
		}
	}
}
