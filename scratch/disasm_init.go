package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\print_stage1.exe")
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

	// __ax_runtime_init is at RVA 0x5483 (Offset: 0x4483)
	start := uint32(0x4483)
	end := uint32(0x452B)

	fmt.Printf("Bytes in __ax_runtime_init (RVA 0x5483 to 0x552B):\n")
	for i := start; i < end && i < uint32(len(data)); i++ {
		rva := textSec.VirtualAddress + i
		fmt.Printf("RVA 0x%04X (Offset: 0x%04X): %02X\n", rva, i, data[i])
	}
}
