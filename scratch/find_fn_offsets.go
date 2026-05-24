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

	fmt.Printf("Scanning .text section (size: 0x%X)...\n", len(data))
	
	// Scan for prologues: 55 48 89 E5
	for i := 0; i < len(data)-4; i++ {
		if data[i] == 0x55 && data[i+1] == 0x48 && data[i+2] == 0x89 && data[i+3] == 0xE5 {
			fmt.Printf("Function Prologue found at offset 0x%04X (RVA: 0x%04X)\n", i, textSec.VirtualAddress+uint32(i))
			
			// Print first 16 bytes
			fmt.Printf("  Bytes: ")
			end := i + 16
			if end > len(data) {
				end = len(data)
			}
			for j := i; j < end; j++ {
				fmt.Printf("%02X ", data[j])
			}
			fmt.Println()
		}
	}
}
