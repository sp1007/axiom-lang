package main

import (
	"debug/pe"
	"fmt"
	"log"
	"sort"
)

type FuncSym struct {
	Name   string
	Offset uint32
}

func main() {
	file, err := pe.Open("axiom_temp.obj")
	if err != nil {
		log.Fatalf("Error opening axiom_temp.obj: %v", err)
	}
	defer file.Close()

	var textSec *pe.Section
	for _, sec := range file.Sections {
		if sec.Name == ".text" {
			textSec = sec
			break
		}
	}

	if textSec == nil {
		log.Fatal(".text section not found in axiom_temp.obj")
	}

	data, err := textSec.Data()
	if err != nil {
		log.Fatalf("Error reading .text: %v", err)
	}

	fmt.Printf(".text section size in object: %d bytes (0x%X)\n", len(data), len(data))

	var fns []FuncSym
	for _, sym := range file.COFFSymbols {
		name, err := sym.FullName(file.StringTable)
		if err != nil {
			name = string(sym.Name[:])
		}
		if sym.SectionNumber == 1 && sym.Type == 0x20 {
			fns = append(fns, FuncSym{Name: name, Offset: sym.Value})
		}
	}

	sort.Slice(fns, func(i, j int) bool {
		return fns[i].Offset < fns[j].Offset
	})

	fmt.Println("Functions sorted by offset:")
	for i, fn := range fns {
		fmt.Printf("  [%3d] %-40s Offset: 0x%X\n", i, fn.Name, fn.Offset)
	}
}
