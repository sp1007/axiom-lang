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
		log.Fatalf("Error reading section data: %v", err)
	}

	fmt.Printf(".text section VirtualAddress: 0x%X\n", secText.VirtualAddress)
	startRVA := uint32(0x1100)
	endRVA := uint32(0x1200)

	startOff := startRVA - secText.VirtualAddress
	endOff := endRVA - secText.VirtualAddress

	fmt.Printf("Bytes from RVA 0x%X to 0x%X:\n", startRVA, endRVA)
	for i := startOff; i < endOff && i < uint32(len(data)); i++ {
		rva := secText.VirtualAddress + i
		fmt.Printf("RVA 0x%04X (Offset: 0x%04X): %02X\n", rva, i, data[i])
	}
}
