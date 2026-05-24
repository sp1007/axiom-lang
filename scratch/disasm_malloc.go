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

	// Read AddressOfEntryPoint from optional header
	var entryPoint uint32
	switch oh := file.OptionalHeader.(type) {
	case *pe.OptionalHeader64:
		entryPoint = oh.AddressOfEntryPoint
		fmt.Printf("OptionalHeader64: AddressOfEntryPoint=0x%X ImageBase=0x%X\n", oh.AddressOfEntryPoint, oh.ImageBase)
	case *pe.OptionalHeader32:
		entryPoint = oh.AddressOfEntryPoint
		fmt.Printf("OptionalHeader32: AddressOfEntryPoint=0x%X ImageBase=0x%X\n", oh.AddressOfEntryPoint, oh.ImageBase)
	default:
		log.Fatalf("Unknown optional header type")
	}

	var textSec *pe.Section
	for _, sec := range file.Sections {
		fmt.Printf("Section: %-8s VA=0x%08X Size=%d\n", sec.Name, sec.VirtualAddress, sec.Size)
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

	// Calculate offset of entryPoint in .text data
	offset := entryPoint - textSec.VirtualAddress
	fmt.Printf("\nBytes at Entry Point (RVA 0x%X, offset 0x%X):\n", entryPoint, offset)

	end := offset + 128
	if end > uint32(len(data)) {
		end = uint32(len(data))
	}

	for i := offset; i < end; i++ {
		if (i-offset)%16 == 0 {
			fmt.Printf("\n  0x%04X: ", i+textSec.VirtualAddress)
		}
		fmt.Printf("%02X ", data[i])
	}
	fmt.Println()
}
