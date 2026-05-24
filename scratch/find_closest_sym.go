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

	fmt.Println("First 30 symbols in axiom_temp.obj:")
	for i := 0; i < len(syms) && i < 30; i++ {
		fmt.Printf("  Offset: 0x%05X  Sec: %d  Type: 0x%04X  Name: %s\n", syms[i].Offset, syms[i].Sec, syms[i].Type, syms[i].Name)
	}
}
