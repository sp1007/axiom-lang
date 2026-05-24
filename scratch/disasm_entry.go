package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	exePath := "bin/axc_stage2.exe"
	f, err := pe.Open(exePath)
	if err != nil {
		log.Fatalf("Error opening PE: %v", err)
	}
	defer f.Close()

	var entryPointRVA uint32
	switch oh := f.OptionalHeader.(type) {
	case *pe.OptionalHeader64:
		entryPointRVA = oh.AddressOfEntryPoint
	default:
		log.Fatalf("Only PE32+ supported")
	}

	fmt.Printf("Entry Point RVA: 0x%X\n", entryPointRVA)

	var textSec *pe.Section
	for _, sec := range f.Sections {
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
		log.Fatalf("Failed to read section data: %v", err)
	}

	offset := entryPointRVA - textSec.VirtualAddress
	fmt.Printf("Entry Point Offset in .text: 0x%X\n", offset)
	
	if offset+64 > uint32(len(data)) {
		log.Fatalf("Entry point out of section bounds")
	}

	fmt.Println("First 64 bytes at Entry Point:")
	for i := uint32(0); i < 64; i++ {
		fmt.Printf("%02X ", data[offset+i])
		if (i+1)%16 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()
}
