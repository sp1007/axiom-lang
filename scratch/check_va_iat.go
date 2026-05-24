package main

import (
	"debug/pe"
	"encoding/binary"
	"fmt"
	"log"
)

func main() {
	file, err := pe.Open("d:\\projects\\compiler\\Axiom\\scratch\\test_va.exe")
	if err != nil {
		log.Fatalf("Error: %v", err)
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

	// idata starts at RVA 0x8000
	rvaToOffset := func(rva uint32) uint32 {
		return rva - 0x8000
	}

	fmt.Println("KERNEL32 IAT (RVA: 0x8260):")
	for i := 0; i < 5; i++ {
		off := rvaToOffset(0x8260) + uint32(i)*8
		val := binary.LittleEndian.Uint64(data[off : off+8])
		fmt.Printf("  [%d] Offset: 0x%04X, Value: 0x%X\n", i, 0x8260+i*8, val)
	}

	fmt.Println("\nKERNEL32 ILT (RVA: 0x80C8):")
	for i := 0; i < 5; i++ {
		off := rvaToOffset(0x80C8) + uint32(i)*8
		val := binary.LittleEndian.Uint64(data[off : off+8])
		fmt.Printf("  [%d] Offset: 0x%04X, Value: 0x%X\n", i, 0x80C8+i*8, val)
	}
}
