package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\axiom_temp.obj")
	if err != nil {
		log.Fatalf("Error opening COFF: %v", err)
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
		log.Fatalf("Error reading section data: %v", err)
	}

	fmt.Printf("Raw Bytes of axiom_temp.obj (first 160 bytes of .text):\n")
	for i := 0; i < 160; i++ {
		if i%16 == 0 {
			fmt.Printf("\n  0x%04X: ", i)
		}
		fmt.Printf("%02X ", data[i])
	}
	fmt.Println()
}
