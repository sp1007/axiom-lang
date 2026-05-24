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

	var textSec *pe.Section
	for _, sec := range f.Sections {
		if sec.Name == ".text" {
			textSec = sec
		}
	}
	if textSec == nil {
		log.Fatalf(".text not found")
	}

	data, err := textSec.Data()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Scanning .text for jmp [rip + disp32] (FF 25)")
	for i := 0; i < len(data)-6; i++ {
		if data[i] == 0xFF && data[i+1] == 0x25 {
			disp := int32(binary.LittleEndian.Uint32(data[i+2 : i+6]))
			rva := textSec.VirtualAddress + uint32(i)
			next_rip := rva + 6
			target_rva := uint32(int32(next_rip) + disp)
			fmt.Printf("Thunk at RVA 0x%X (Offset: 0x%X) -> jmp [0x%X]\n", rva, i, target_rva)
		}
	}
}
