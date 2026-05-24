package main

import (
	"debug/pe"
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run disasm_stage2_offset.go <offset_or_rva>")
	}
	
	valStr := os.Args[1]
	var val uint64
	var err error
	if len(valStr) > 2 && valStr[:2] == "0x" {
		val, err = strconv.ParseUint(valStr[2:], 16, 64)
	} else {
		val, err = strconv.ParseUint(valStr, 10, 64)
	}
	if err != nil {
		log.Fatalf("Invalid number: %v", err)
	}

	exePath := "axiom_temp.obj"
	f, err := pe.Open(exePath)
	if err != nil {
		log.Fatalf("Error opening PE: %v", err)
	}
	defer f.Close()

	// Assume val is offset directly
	offset := uint32(val)
	rva := offset + 0x1000

	fmt.Printf("Searching for RVA 0x%X (Offset in .text: 0x%X):\n", rva, offset)

	type FuncSym struct {
		Name   string
		Offset uint32
	}
	var fns []FuncSym
	for _, sym := range f.COFFSymbols {
		name, err := sym.FullName(f.StringTable)
		if err != nil {
			name = string(sym.Name[:])
		}
		if sym.SectionNumber > 0 && sym.Type == 0x20 {
			fns = append(fns, FuncSym{Name: name, Offset: sym.Value})
		}
	}

	var bestFunc FuncSym
	found := false
	bestDiff := uint32(0xFFFFFFFF)

	for _, fn := range fns {
		if fn.Offset <= offset {
			diff := offset - fn.Offset
			if diff < bestDiff {
				bestDiff = diff
				bestFunc = fn
				found = true
			}
		}
	}

	if found {
		fmt.Printf(">>> Found inside function: %s (starts at Offset 0x%X, RVA 0x%X)\n", bestFunc.Name, bestFunc.Offset, bestFunc.Offset+0x1000)
		fmt.Printf(">>> Relative offset: %d bytes (0x%X)\n", bestDiff, bestDiff)
	} else {
		fmt.Println(">>> RVA is before any function symbol")
	}
}
