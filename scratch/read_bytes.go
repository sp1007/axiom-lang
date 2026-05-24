package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("bin/axc_stage2.exe")
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
		log.Fatal(".text section not found")
	}

	crashRVA := uint32(0x90B3B)
	if crashRVA < textSec.VirtualAddress || crashRVA >= textSec.VirtualAddress+textSec.VirtualSize {
		log.Fatalf("RVA 0x%X is outside .text section", crashRVA)
	}

	offset := crashRVA - textSec.VirtualAddress
	data, err := textSec.Data()
	if err != nil {
		log.Fatalf("Error reading .text data: %v", err)
	}

	fmt.Printf("Bytes at RVA 0x%X (Offset inside .text 0x%X):\n", crashRVA, offset)
	for i := -48; i < 16; i++ {
		idx := int(offset) + i
		if idx >= 0 && idx < len(data) {
			if i == 0 {
				fmt.Printf("  -> %02X (crashing instruction start)\n", data[idx])
			} else {
				fmt.Printf("     %02X\n", data[idx])
			}
		}
	}
}
