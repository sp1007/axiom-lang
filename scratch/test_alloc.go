package main

import (
	"fmt"
	"syscall"
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	virtualAlloc = kernel32.NewProc("VirtualAlloc")
)

func testAlloc(addr uintptr) {
	// MEM_COMMIT = 0x1000, MEM_RESERVE = 0x2000 -> 0x3000
	// PAGE_READWRITE = 0x04
	r, _, err := virtualAlloc.Call(addr, 4096, 0x3000, 0x04)
	if r == 0 {
		fmt.Printf("Address 0x%X: FAILED (Error: %v)\n", addr, err)
	} else {
		fmt.Printf("Address 0x%X: SUCCESS (Allocated address: 0x%X)\n", addr, r)
		// Free it
		virtualFree := kernel32.NewProc("VirtualFree")
		virtualFree.Call(r, 0, 0x8000) // MEM_RELEASE = 0x8000
	}
}

func main() {
	addresses := []uintptr{
		uintptr(0x50000) << 20, // 320 GB
		uintptr(0x60000) << 20, // 384 GB
		uintptr(0x70000) << 20, // 448 GB
		uintptr(0x20000) << 20, // 128 GB
		uintptr(0x10000) << 20, // 64 GB
		uintptr(0x8000) << 20,  // 32 GB
		uintptr(0x4000) << 20,  // 16 GB
		uintptr(0x2000) << 20,  // 8 GB
		uintptr(0x1000) << 20,  // 4 GB
		uintptr(0x800) << 20,   // 2 GB
		uintptr(0x400) << 20,   // 1 GB
	}

	for _, addr := range addresses {
		testAlloc(addr)
	}
}
