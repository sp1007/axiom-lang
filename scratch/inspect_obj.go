package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("axiom_temp.obj")
	if err != nil {
		log.Fatalf("Error opening COFF: %v", err)
	}
	defer file.Close()

	fmt.Printf("COFF Sections:\n")
	var textSec *pe.Section
	for _, sec := range file.Sections {
		fmt.Printf("  Section Name: %s, Size: %d, Offset: 0x%X\n", sec.Name, sec.Size, sec.Offset)
		if sec.Name == ".text" {
			textSec = sec
		}
	}

	if textSec != nil {
		fmt.Printf("\n.text Relocations (count %d):\n", len(textSec.Relocs))
		for i, rel := range textSec.Relocs {
			sym := file.COFFSymbols[rel.SymbolTableIndex]
			name, err := sym.FullName(file.StringTable)
			if err != nil {
				name = string(sym.Name[:])
			}
			fmt.Printf("  [%3d] VirtualAddress: 0x%X, SymbolTableIndex: %d (%s), Type: 0x%X\n", i, rel.VirtualAddress, rel.SymbolTableIndex, name, rel.Type)
		}
	}

	fmt.Printf("\nCOFF Symbols:\n")
	for i, sym := range file.COFFSymbols {
		name, err := sym.FullName(file.StringTable)
		if err != nil {
			name = string(sym.Name[:])
		}
		if sym.SectionNumber > 0 {
			fmt.Printf("  [%3d] Name: %s, Value: %d, Section: %d, Type: 0x%X, Class: %d\n", i, name, sym.Value, sym.SectionNumber, sym.Type, sym.StorageClass)
		}
	}

	// Read .rdata section data
	var rdataSec *pe.Section
	for _, sec := range file.Sections {
		if sec.Name == ".rdata" {
			rdataSec = sec
			break
		}
	}

	if rdataSec != nil {
		data, err := rdataSec.Data()
		if err == nil {
			fmt.Printf("\n.rdata section contents (size %d):\n", len(data))
			for i := 0; i < len(data); i += 16 {
				end := i + 16
				if end > len(data) {
					end = len(data)
				}
				for j := i; j < end; j++ {
					fmt.Printf("%02X ", data[j])
				}
				// print ascii
				fmt.Printf("  | ")
				for j := i; j < end; j++ {
					ch := data[j]
					if ch >= 32 && ch < 127 {
						fmt.Printf("%c", ch)
					} else {
						fmt.Printf(".")
					}
				}
				fmt.Println()
			}
		} else {
			fmt.Printf("Error reading .rdata: %v\n", err)
		}
	} else {
		fmt.Println(".rdata section not found")
	}
}
