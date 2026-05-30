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
	Type   uint16
	Sec    int16
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
		if sym.SectionNumber > 0 {
			syms = append(syms, Symbol{
				Name:   name,
				Offset: sym.Value,
				Type:   sym.Type,
				Sec:    sym.SectionNumber,
			})
		}
	}

	sort.Slice(syms, func(i, j int) bool {
		return syms[i].Offset < syms[j].Offset
	})

	fmt.Println("Section 1 (.text) symbols around 0x2C08:")
	for _, s := range syms {
		if s.Sec == 1 && s.Offset >= 0x2A00 && s.Offset <= 0x2E00 {
			fmt.Printf("  Offset: 0x%05X  Type: 0x%04X  Name: %s\n", s.Offset, s.Type, s.Name)
		}
	}
}
