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
	var secRdata *pe.Section
	for _, sec := range file.Sections {
		if sec.Name == ".text" {
			secText = sec
		} else if sec.Name == ".rdata" || sec.Name == ".idata" {
			secRdata = sec
		}
	}

	if secText == nil {
		log.Fatalf("No .text section found")
	}

	textData, err := secText.Data()
	if err != nil {
		log.Fatalf("Error reading .text section data: %v", err)
	}

	fmt.Printf(".text section VirtualAddress: 0x%X, Size: %d\n", secText.VirtualAddress, len(textData))
	
	// Print bytes at RVA 0x4810 to 0x4850 (offset: RVA - VirtualAddress)
	startRVA := uint32(0x4810)
	endRVA := uint32(0x4850)
	startOff := startRVA - secText.VirtualAddress
	endOff := endRVA - secText.VirtualAddress

	fmt.Printf("Bytes from RVA 0x%X to 0x%X:\n", startRVA, endRVA)
	for i := startOff; i < endOff && i < uint32(len(textData)); i++ {
		fmt.Printf("%02X ", textData[i])
		if (i-startOff)%16 == 15 {
			fmt.Println()
		}
	}
	fmt.Println()

	// Search for "Hello" in .text section data
	for i := 0; i < len(textData)-5; i++ {
		if textData[i] == 'H' && textData[i+1] == 'e' && textData[i+2] == 'l' && textData[i+3] == 'l' && textData[i+4] == 'o' {
			fmt.Printf("Found 'Hello' in .text section at RVA 0x%X (offset %d):\n", secText.VirtualAddress+uint32(i), i)
			for j := i; j < i+20 && j < len(textData); j++ {
				c := textData[j]
				if c >= 32 && c <= 126 {
					fmt.Printf("%c", c)
				} else {
					fmt.Printf(".")
				}
			}
			fmt.Println()
		}
	}


	if secRdata != nil {
		rdataData, err := secRdata.Data()
		if err == nil {
			fmt.Printf(".rdata/.idata section VirtualAddress: 0x%X, Size: %d\n", secRdata.VirtualAddress, len(rdataData))
			fmt.Printf("Hex dump of first 128 bytes:\n")
			for i := 0; i < 128 && i < len(rdataData); i++ {
				fmt.Printf("%02X ", rdataData[i])
				if i%16 == 15 {
					fmt.Println()
				}
			}
			fmt.Println()
			fmt.Printf("ASCII dump of .rdata:\n")
			for i := 0; i < len(rdataData); i++ {
				c := rdataData[i]
				if c >= 32 && c <= 126 {
					fmt.Printf("%c", c)
				} else {
					fmt.Printf(".")
				}
			}
			fmt.Println()
		}
	}
}
