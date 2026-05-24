package main

import (
	"debug/pe"
	"fmt"
	"log"
)

func main() {
	f, err := pe.Open("hello.exe")
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

	// Scan for function prologues: 55 48 89 E5
	fmt.Printf("Scanning for functions in hello.exe (VirtualAddress: 0x%X):\n", textSec.VirtualAddress)
	for i := 0; i < len(data)-4; i++ {
		if data[i] == 0x55 && data[i+1] == 0x48 && data[i+2] == 0x89 && data[i+3] == 0xE5 {
			fmt.Printf("\n--- Function at Offset 0x%X (RVA 0x%X) ---\n", i, textSec.VirtualAddress+uint32(i))
			// Disassemble exactly 300 bytes for the function at Offset 0x52C2
			count := uint32(0)
			for j := uint32(i); j < uint32(len(data)) && count < 300; {
				fmt.Printf("  0x%04X (RVA 0x%04X): ", j, textSec.VirtualAddress+j)
				instLen := getInstLen(data[j:])
				for k := uint32(0); k < instLen; k++ {
					fmt.Printf("%02X ", data[j+k])
				}
				for k := instLen; k < 8; k++ {
					fmt.Printf("   ")
				}
				decodeInst(data[j : j+instLen])
				
				count += instLen
				j += instLen
			}
		}
	}
}

func decodeInst(bytes []byte) {
	if len(bytes) == 0 {
		return
	}
	b := bytes[0]
	if b == 0x55 {
		fmt.Println("push rbp")
		return
	}
	if b == 0x5D {
		fmt.Println("pop rbp")
		return
	}
	if b >= 0x50 && b <= 0x57 {
		regs := []string{"rax", "rcx", "rdx", "rbx", "rsp", "rbp", "rsi", "rdi"}
		fmt.Printf("push %s\n", regs[b-0x50])
		return
	}
	if b >= 0x58 && b <= 0x5F {
		regs := []string{"rax", "rcx", "rdx", "rbx", "rsp", "rbp", "rsi", "rdi"}
		fmt.Printf("pop %s\n", regs[b-0x58])
		return
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0x89 && bytes[2] == 0xE5 {
		fmt.Println("mov rbp, rsp")
		return
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0x89 {
		fmt.Printf("mov rm64, r64 (ModRM %02X)\n", bytes[2])
		return
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0x8D {
		fmt.Printf("lea r64, [rm64] (ModRM %02X)\n", bytes[2])
		return
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0x83 {
		op := "add"
		if bytes[2]&0x38 == 0x28 {
			op = "sub"
		} else if bytes[2]&0x38 == 0x38 {
			op = "cmp"
		}
		fmt.Printf("%s reg, %d (ModRM %02X)\n", op, int8(bytes[3]), bytes[2])
		return
	}
	if b == 0xE8 {
		fmt.Println("call rel32")
		return
	}
	if b == 0xE9 {
		fmt.Println("jmp rel32")
		return
	}
	if b == 0xC3 {
		fmt.Println("ret")
		return
	}
	if b == 0x7C {
		fmt.Printf("jl rel8 (%d)\n", int8(bytes[1]))
		return
	}
	if b == 0x0F && len(bytes) > 1 && bytes[1] == 0x8C {
		fmt.Println("jl rel32")
		return
	}
	fmt.Println()
}

func getInstLen(bytes []byte) uint32 {
	if len(bytes) == 0 {
		return 1
	}
	b := bytes[0]
	if b == 0x55 || b == 0x5D || (b >= 0x50 && b <= 0x5F) {
		return 1
	}
	if b == 0x41 && len(bytes) > 1 && bytes[1] >= 0x50 && bytes[1] <= 0x5F {
		return 2
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0x89 {
		return 3
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0x8D {
		if len(bytes) > 3 && bytes[2] == 0x85 {
			return 8
		}
		if len(bytes) > 3 && bytes[2] >= 0x40 && bytes[2] <= 0x7F {
			return 4
		}
		return 3
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0x83 {
		return 4
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0x81 {
		return 7
	}
	if b == 0xE8 || b == 0xE9 {
		return 5
	}
	if b == 0xFF && len(bytes) > 1 && bytes[1] == 0x25 {
		return 6
	}
	if b == 0x89 || b == 0x8B {
		if len(bytes) > 2 && bytes[1] >= 0x40 && bytes[1] <= 0x7F {
			return 3
		}
		return 2
	}
	if b == 0x48 && len(bytes) > 1 && bytes[1] >= 0xB8 && bytes[1] <= 0xBF {
		return 10
	}
	if b == 0x48 && len(bytes) > 2 && bytes[1] == 0xC7 {
		if len(bytes) > 3 && bytes[2] >= 0x40 && bytes[2] <= 0x7F {
			return 8
		}
		return 7
	}
	if b == 0xC3 || b == 0x90 || b == 0xCC {
		return 1
	}
	if b == 0x7C || b == 0x7D || b == 0x7E || b == 0x7F || b == 0x74 || b == 0x75 {
		return 2
	}
	if b == 0x0F && len(bytes) > 1 && bytes[1] >= 0x80 && bytes[1] <= 0x8F {
		return 6
	}
	return 1
}
