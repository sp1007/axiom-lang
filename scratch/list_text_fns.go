package main

import (
	"debug/pe"
	"fmt"
	"log"
	"sort"
)

type Symbol struct {
	Name   string
	Offset uint32
}

func main() {
	objFile, err := pe.Open("axiom_temp.obj")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer objFile.Close()

	var syms []Symbol
	for _, sym := range objFile.COFFSymbols {
		name, err := sym.FullName(objFile.StringTable)
		if err != nil {
			name = string(sym.Name[:])
		}
		if sym.SectionNumber == 1 && sym.Type == 0x20 {
			syms = append(syms, Symbol{
				Name:   name,
				Offset: sym.Value,
			})
		}
	}

	sort.Slice(syms, func(i, j int) bool {
		return syms[i].Offset < syms[j].Offset
	})

	fmt.Println("All function symbols in axiom_temp.obj sorted by offset:")
	for i, s := range syms {
		fmt.Printf("  [%3d] Offset: 0x%05X (RVA: 0x%05X)  Name: %s\n", i, s.Offset, s.Offset+0x1000, s.Name)
	}
}
