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
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\axiom_temp.obj")
	if err != nil {
		log.Fatalf("Error opening PE: %v", err)
	}
	defer file.Close()

	var fns []FuncSym
	for _, sym := range file.COFFSymbols {
		name, err := sym.FullName(file.StringTable)
		if err != nil {
			name = string(sym.Name[:])
		}
		// In COFF object files, function symbols have SectionNumber > 0 and Type 0x20
		if sym.SectionNumber > 0 && sym.Type == 0x20 {
			fns = append(fns, FuncSym{Name: name, Offset: sym.Value})
		}
	}

	// Sort by offset
	sort.Slice(fns, func(i, j int) bool {
		return fns[i].Offset < fns[j].Offset
	})

	fmt.Println("Functions in axiom_temp.obj sorted by offset:")
	crashOffset := uint32(0x3B46)
	var crashingFn FuncSym
	found := false

	for i, fn := range fns {
		fmt.Printf("  [%3d] %-30s Offset: 0x%04X (RVA: 0x%04X)\n",
			i, fn.Name, fn.Offset, fn.Offset+0x1000)
		if fn.Offset <= crashOffset {
			crashingFn = fn
			found = true
		}
	}

	if found {
		fmt.Printf("\n>>> CRASH occurred at offset 0x%04X (RVA: 0x%04X)\n", crashOffset, crashOffset+0x1000)
		fmt.Printf(">>> This is inside function: %s (starts at offset 0x%04X)\n", crashingFn.Name, crashingFn.Offset)
		fmt.Printf(">>> Relative offset inside function: %d bytes (0x%X)\n", crashOffset-crashingFn.Offset, crashOffset-crashingFn.Offset)
	} else {
		fmt.Printf("\n>>> CRASH offset 0x%04X is before any function\n", crashOffset)
	}
}
