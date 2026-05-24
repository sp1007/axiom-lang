package main

import (
	"debug/pe"
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

	// We know from find_crash_fn_stage1 that ax_os_alloc starts at RVA 0x10A0 (Offset 0xA0)
	// Let's print 300 bytes of instructions starting at offset 0xA0
	start := uint32(0xA0)
	fmt.Printf("Disassembling ax_os_alloc in minimal.exe (Offset: 0x%X):\n", start)
	for i := start; i < start+300 && i < uint32(len(data)); {
		fmt.Printf("  0x%04X (RVA 0x%04X): ", i, textSec.VirtualAddress+i)
		// Print bytes
		instLen := getInstLen(data[i:])
		for j := uint32(0); j < instLen; j++ {
			fmt.Printf("%02X ", data[i+j])
		}
		for j := instLen; j < 8; j++ {
			fmt.Printf("   ")
		}
		// Print basic decoded info
		fmt.Println()
		i += instLen
	}
}

// A very basic instruction length decoder for the specific instructions we generate
func getInstLen(bytes []byte) uint32 {
	if len(bytes) == 0 {
		return 1
	}
	b := bytes[0]
	if b == 0x55 || b == 0x5D || (b >= 0x50 && b <= 0x5F) { // push/pop GPR
		return 1
	}
	if b == 0x41 && len(bytes) > 1 && bytes[1] >= 0x50 && bytes[1] <= 0x5F { // push/pop R8-R15
		return 2
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0x89 { // mov reg, reg (64-bit)
		return 3
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0x8D { // lea (64-bit)
		if len(bytes) > 3 && bytes[2] == 0x85 { // disp32
			return 8
		}
		if len(bytes) > 3 && bytes[2] >= 0x40 && bytes[2] <= 0x7F { // disp8
			return 4
		}
		return 3
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0x83 { // add/sub reg, imm8
		return 4
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0x81 { // add/sub reg, imm32
		return 7
	}
	if b == 0xE8 || b == 0xE9 { // call/jmp rel32
		return 5
	}
	if b == 0xFF && len(bytes) > 1 && bytes[1] == 0x25 { // jmp indirect (thunk)
		return 6
	}
	if b == 0x89 || b == 0x8B { // mov reg, reg or load/store (32-bit/64-bit ModRM)
		if len(bytes) > 2 && bytes[1] >= 0x40 && bytes[1] <= 0x7F { // disp8
			return 3
		}
		return 2
	}
	if b == 0x48 && len(bytes) > 1 && bytes[1] >= 0xB8 && bytes[1] <= 0xBF { // mov reg, imm64
		return 10
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0xC7 { // mov reg/mem, imm32
		if len(bytes) > 3 && bytes[2] >= 0x40 && bytes[2] <= 0x7F { // disp8
			return 8
		}
		return 7
	}
	if b == 0xC3 || b == 0x90 || b == 0xCC { // ret, nop, int3
		return 1
	}
	// Fallback
	return 1
}
