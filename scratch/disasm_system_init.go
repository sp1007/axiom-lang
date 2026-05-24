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

	// ax_actor_system_init starts at Value 11513 (RVA 0x3CF9, Offset: 0x2CF9)
	start := uint32(0x2CF9)
	end := uint32(0x2E69)

	fmt.Printf("Bytes in ax_actor_system_init (RVA 0x3CF9 to 0x3E69):\n")
	for i := start; i < end && i < uint32(len(data)); i++ {
		rva := textSec.VirtualAddress + i
		fmt.Printf("RVA 0x%04X (Offset: 0x%04X): %02X\n", rva, i, data[i])
	}
}
