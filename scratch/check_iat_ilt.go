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

	var idataSec *pe.Section
	for _, sec := range file.Sections {
		if sec.Name == ".idata" {
			idataSec = sec
			break
		}
	}

	if idataSec == nil {
		log.Fatalf(".idata section not found")
	}

	data, err := idataSec.Data()
	if err != nil {
		log.Fatalf("Error reading section data: %v", err)
	}

	fmt.Println("=== FULL .idata SECTION HEX & ASCII DUMP ===")
	for i := 0; i < len(data); i += 16 {
		fmt.Printf("  0x%04X: ", 0x6000+uint32(i))
		
		// Hex
		end := i + 16
		if end > len(data) {
			end = len(data)
		}
		for j := i; j < i+16; j++ {
			if j < end {
				fmt.Printf("%02X ", data[j])
			} else {
				fmt.Printf("   ")
			}
		}
		fmt.Printf(" | ")
		
		// ASCII
		for j := i; j < end; j++ {
			b := data[j]
			if b >= 32 && b <= 126 {
				fmt.Printf("%c", b)
			} else {
				fmt.Printf(".")
			}
		}
		fmt.Println()
	}
}
