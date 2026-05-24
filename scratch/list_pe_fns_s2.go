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
	Sec    int16
	Class  uint8
}

func main() {
	file, err := pe.Open("bin/axc_stage2.exe")
	if err != nil {
		log.Fatalf("Error opening PE: %v", err)
	}
	defer file.Close()

	var syms []Symbol
	for _, sym := range file.COFFSymbols {
		name, err := sym.FullName(file.StringTable)
		if err != nil {
			name = string(sym.Name[:])
		}
		syms = append(syms, Symbol{
			Name:   name,
			Offset: sym.Value,
			Sec:    sym.SectionNumber,
			Class:  sym.StorageClass,
		})
	}

	// Sort by offset
	sort.Slice(syms, func(i, j int) bool {
		if syms[i].Offset != syms[j].Offset {
			return syms[i].Offset < syms[j].Offset
		}
		return syms[i].Name < syms[j].Name
	})

	fmt.Println("Sorted symbols in bin/axc_stage2.exe:")
	for _, s := range syms {
		fmt.Printf("  Offset: 0x%05X  Sec: %d  Name: %s\n", s.Offset, s.Sec, s.Name)
	}
}
