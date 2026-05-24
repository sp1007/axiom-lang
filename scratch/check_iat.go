package main

import (
	"debug/pe"
	"encoding/binary"
	"fmt"
	"log"
)

func main() {
	f, err := pe.Open("minimal.exe")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer f.Close()

	var idataSec *pe.Section
	for _, sec := range f.Sections {
		if sec.Name == ".idata" {
			idataSec = sec
		}
	}
	if idataSec == nil {
		log.Fatalf(".idata not found")
	}

	data, err := idataSec.Data()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf(".idata RVA: 0x%X\n", idataSec.VirtualAddress)
	// Let's print the first 256 bytes of .idata
	for i := 0; i < len(data) && i < 256; i += 8 {
		val := binary.LittleEndian.Uint64(data[i : i+8])
		fmt.Printf("  Offset 0x%02X (RVA 0x%X): 0x%016X\n", i, idataSec.VirtualAddress+uint32(i), val)
	}
}
