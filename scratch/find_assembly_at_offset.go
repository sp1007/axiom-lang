package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\minimal.exe")
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

	fmt.Printf("Thunk bytes at RVA 0x5FA0 (Offset: 0x4FA0):\n")
	off := 0x4FA0
	for i := 0; i < 6; i++ {
		fmt.Printf("%02X ", data[off+i])
	}
	fmt.Println()
}
