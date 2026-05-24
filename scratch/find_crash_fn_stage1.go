package main

import (
	"debug/pe"
	"fmt"
	"log"
)

type Func struct {
	Name   string
	Offset uint32
	RVA    uint32
}

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\minimal.exe")
	if err != nil {
		log.Fatalf("Error opening PE: %v", err)
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
		log.Fatalf(".text section not found")
	}

	data, err := textSec.Data()
	if err != nil {
		log.Fatalf("Error reading section: %v", err)
	}

	crashOffset := uint32(0x2DF2)
	fmt.Printf("Crash Offset: 0x%X (RVA: 0x%X)\n", crashOffset, textSec.VirtualAddress+crashOffset)

	// Read symbols from the COFF symbol table if available
	var functions []Func
	for _, sym := range file.COFFSymbols {
		// Section number is 1-based, text section is usually 1
		if sym.SectionNumber == 1 && sym.StorageClass == 2 { // External symbol (function)
			name, err := sym.FullName(file.StringTable)
			if err == nil {
				functions = append(functions, Func{
					Name:   name,
					Offset: sym.Value,
					RVA:    textSec.VirtualAddress + sym.Value,
				})
			}
		}
	}

	// If no COFF symbols found, scan for prologues (55 48 89 E5)
	if len(functions) == 0 {
		fmt.Println("No COFF symbols found, scanning for prologues...")
		for i := 0; i < len(data)-4; i++ {
			if data[i] == 0x55 && data[i+1] == 0x48 && data[i+2] == 0x89 && data[i+3] == 0xE5 {
				functions = append(functions, Func{
					Name:   fmt.Sprintf("fn_at_0x%X", i),
					Offset: uint32(i),
					RVA:    textSec.VirtualAddress + uint32(i),
				})
			}
		}
	}

	// Find the function immediately preceding the crash offset
	var bestFunc Func
	bestDiff := uint32(0xFFFFFFFF)

	for _, f := range functions {
		if f.Offset <= crashOffset {
			diff := crashOffset - f.Offset
			if diff < bestDiff {
				bestDiff = diff
				bestFunc = f
			}
		}
	}

	if bestDiff != 0xFFFFFFFF {
		fmt.Printf("Crashing Function: %s at Offset 0x%X (RVA: 0x%X)\n", bestFunc.Name, bestFunc.Offset, bestFunc.RVA)
		fmt.Printf("Crash happened at Offset + 0x%X bytes inside %s\n", bestDiff, bestFunc.Name)

		start := crashOffset - 32
		end := crashOffset + 32
		if end > uint32(len(data)) {
			end = uint32(len(data))
		}
		for j := start; j < end; j++ {
			prefix := "   "
			if j == crashOffset {
				prefix = "=> "
			}
			fmt.Printf("%s RVA 0x%04X (Offset: 0x%04X): %02X\n", prefix, textSec.VirtualAddress+j, j, data[j])
		}
	} else {
		fmt.Println("Could not determine crashing function")
	}
}
