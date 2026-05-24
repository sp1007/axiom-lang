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

	textData, err := secText.Data()
	if err != nil {
		log.Fatalf("Error reading .text section data: %v", err)
	}

	// compiler_intrinsic starts at RVA 0x59B5, which is offset 0x59B5 - 0x1000 = 0x49B5
	startOffset := uint32(0x59B5 - 0x1000)
	endOffset := uint32(0x5A37 - 0x1000)

	fmt.Printf("Hex bytes of compiler_intrinsic in print_stage1.exe (offset: 0x%X to 0x%X):\n", startOffset, endOffset)
	for i := startOffset; i < endOffset && i < uint32(len(textData)); i++ {
		fmt.Printf("%02X ", textData[i])
		if (i-startOffset)%16 == 15 {
			fmt.Println()
		}
	}
	fmt.Println()
}
