package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\print.exe")
	if err != nil {
		log.Fatalf("Error opening PE: %v", err)
	}
	defer file.Close()

	var secText *pe.Section
	for _, sec := range file.Sections {
		if sec.Name == ".text" {
			secText = sec
			break
		}
	}

	if secText == nil {
		log.Fatalf("No .text section found")
	}

	data, err := secText.Data()
	if err != nil {
		log.Fatalf("Error reading .text section: %v", err)
	}

	fmt.Printf("Scanning .text section (size: %d bytes) for LEA or MOV referring to RVA >= 0x6000\n", len(data))
	// Look for LEA reg, [rip + offset] -> 48 8D / 4C 8D
	for i := 0; i < len(data)-7; i++ {
		b := data[i]
		// 48 8D 05 / 48 8D 15 / etc.
		if b == 0x48 && data[i+1] == 0x8D && (data[i+2]&0xC7) == 0x05 {
			offset := int32(uint32(data[i+3]) | (uint32(data[i+4]) << 8) | (uint32(data[i+5]) << 16) | (uint32(data[i+6]) << 24))
			rip := secText.VirtualAddress + uint32(i) + 7
			targetRVA := uint32(int32(rip) + offset)
			fmt.Printf("  Found LEA at RVA 0x%X (Offset: 0x%X): rip + (%d) -> Target RVA: 0x%X\n",
				secText.VirtualAddress+uint32(i), i, offset, targetRVA)
		}
	}
}
