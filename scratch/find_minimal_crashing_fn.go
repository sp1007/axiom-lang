package main

import (
	"debug/pe"
	"fmt"
	"log"
	"sort"
)

type FuncSym struct {
	Name   string
	Offset uint32 // section offset
	RVA    uint32 // VirtualAddress (offset + 0x1000)
}

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\minimal.exe")
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
		// In COFF files, function symbols have SectionNumber > 0 and Type 0x20
		if sym.SectionNumber > 0 && sym.Type == 0x20 {
			fns = append(fns, FuncSym{Name: name, Offset: sym.Value, RVA: sym.Value + 0x1000})
		}
	}

	// If no COFFSymbols with Type 0x20 are found, let's just dump all symbols in section 1 (.text)
	if len(fns) == 0 {
		fmt.Println("No function symbols found via Type 0x20. Dumping all symbols in Section 1:")
		for _, sym := range file.COFFSymbols {
			if sym.SectionNumber == 1 {
				name, err := sym.FullName(file.StringTable)
				if err != nil {
					name = string(sym.Name[:])
				}
				fns = append(fns, FuncSym{Name: name, Offset: sym.Value, RVA: sym.Value + 0x1000})
			}
		}
	}

	// Sort by RVA
	sort.Slice(fns, func(i, j int) bool {
		return fns[i].RVA < fns[j].RVA
	})

	crashRVA := uint32(0x22B7)
	var crashingFn FuncSym
	found := false

	fmt.Println("Functions/Symbols in minimal.exe sorted by RVA:")
	for i, fn := range fns {
		fmt.Printf("  [%3d] %-40s Offset: 0x%04X RVA: 0x%04X\n",
			i, fn.Name, fn.Offset, fn.RVA)
		if fn.RVA <= crashRVA {
			crashingFn = fn
			found = true
		}
	}

	if found {
		fmt.Printf("\n>>> CRASH occurred at RVA 0x%04X (RIP Offset: 0x%04X)\n", crashRVA, crashRVA)
		fmt.Printf(">>> Enclosing function: %s (starts at RVA 0x%04X)\n", crashingFn.Name, crashingFn.RVA)
		fmt.Printf(">>> Relative offset inside function: %d bytes (0x%X)\n", crashRVA-crashingFn.RVA, crashRVA-crashingFn.RVA)
	} else {
		fmt.Printf("\n>>> CRASH RVA 0x%04X is before any function\n", crashRVA)
	}
}
