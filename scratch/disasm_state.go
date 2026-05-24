package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\test_malloc.exe")
	if err != nil {
		log.Fatalf("Error opening PE: %v", err)
	}
	defer file.Close()

	var textSec *pe.Section
	for _, sec := range file.Sections {
		if sec.Name == ".text" {
			textSec = sec
		}
	}

	if textSec == nil {
		log.Fatalf(".text section not found")
	}

	data, err := textSec.Data()
	if err != nil {
		log.Fatalf("Error reading section data: %v", err)
	}

	// Print 128 bytes from offset 0x50 in .text
	offset := uint32(0x50)
	fmt.Printf("\nBytes around get_global_state (RVA 0x%X):\n", 0x1000 + offset)

	end := offset + 128
	for i := offset; i < end; i++ {
		if (i-offset)%16 == 0 {
			fmt.Printf("\n  0x%04X: ", i+0x1000)
		}
		fmt.Printf("%02X ", data[i])
	}
	fmt.Println()
}
