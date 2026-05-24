package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\torture_gen_ref.exe")
	if err != nil {
		log.Fatalf("Error opening PE: %v", err)
	}
	defer file.Close()

	fmt.Println("Sections:")
	var textSec *pe.Section
	for _, sec := range file.Sections {
		fmt.Printf("  - Name: %-8s VirtualAddress: 0x%08X Size: %d\n", sec.Name, sec.VirtualAddress, sec.Size)
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

	// Address: ImageBase + 0x1627.
	// Since .text is mapped at VirtualAddress (usually 0x1000), 0x1627 is 0x627 bytes from the start of .text!
	// Let's print bytes from 0x600 to 0x660 in .text
	offset := uint32(0x1627 - textSec.VirtualAddress)
	fmt.Printf("\nBytes around offset 0x%X in .text (Image Offset 0x1627):\n", offset)
	
	start := int(offset) - 32
	if start < 0 {
		start = 0
	}
	end := int(offset) + 32
	if end > len(data) {
		end = len(data)
	}

	for i := start; i < end; i++ {
		prefix := "  "
		if i == int(offset) {
			prefix = "=>"
		}
		fmt.Printf("%s 0x%04X: %02X\n", prefix, i+int(textSec.VirtualAddress), data[i])
	}
}
