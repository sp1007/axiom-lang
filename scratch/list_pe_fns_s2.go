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
	file, err := pe.Open("bin/axc_stage2_gcc.exe")
	if err != nil {
		log.Fatalf("Error opening PE: %v", err)
	}
	defer file.Close()

	// Get section 1 VirtualAddress
	var sec1VA uint32
	for _, sec := range file.Sections {
		if sec.Name == ".text" {
			sec1VA = sec.VirtualAddress
			fmt.Printf(".text section VA is 0x%X\n", sec1VA)
			break
		}
	}

	var syms []Symbol
	for _, sym := range file.COFFSymbols {
		name, err := sym.FullName(file.StringTable)
		if err != nil {
			name = string(sym.Name[:])
		}
		// In PE executables, sym.SectionNumber == 1 is the .text section (or we match by name)
		if sym.SectionNumber == 1 {
			syms = append(syms, Symbol{
				Name:   name,
				Offset: sym.Value, // offset relative to the section start
				Sec:    sym.SectionNumber,
				Class:  sym.StorageClass,
			})
		}
	}

	// Sort by offset
	sort.Slice(syms, func(i, j int) bool {
		return syms[i].Offset < syms[j].Offset
	})

	fmt.Println("Sorted symbols in Section 1 (.text):")
	targetOffset := uint32(0x1490) - sec1VA // offset from section 1 start
	fmt.Printf("Target offset from section start: 0x%X\n", targetOffset)
	for i, s := range syms {
		if s.Offset <= targetOffset && (i == len(syms)-1 || syms[i+1].Offset > targetOffset) {
			fmt.Printf("MATCH (Starts before offset 0x%X):\n", targetOffset)
			for k := i - 3; k <= i+3; k++ {
				if k >= 0 && k < len(syms) {
					prefix := "  "
					if k == i {
						prefix = "=>"
					}
					fmt.Printf("%s Offset: 0x%05X  Name: %s\n", prefix, s.Offset, syms[k].Name)
				}
			}
		}
	}
}
