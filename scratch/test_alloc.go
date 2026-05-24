package main

import (
	"fmt"
	"syscall"
)

func main() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	virtualAlloc := kernel32.NewProc("VirtualAlloc")
	getLastError := kernel32.NewProc("GetLastError")

	addr := uintptr(0x30000000)
	size := uintptr(4096)

	// First call: MEM_COMMIT | MEM_RESERVE = 0x3000
	res1, _, err1 := virtualAlloc.Call(addr, size, 0x3000, 0x04)
	fmt.Printf("1st call: res=0x%X, err=%v\n", res1, err1)

	// Second call: MEM_COMMIT | MEM_RESERVE = 0x3000
	res2, _, err2 := virtualAlloc.Call(addr, size, 0x3000, 0x04)
	fmt.Printf("2nd call (0x3000): res=0x%X, err=%v\n", res2, err2)

	// Third call: MEM_COMMIT = 0x1000
	res3, _, err3 := virtualAlloc.Call(addr, size, 0x1000, 0x04)
	fmt.Printf("3rd call (0x1000): res=0x%X, err=%v\n", res3, err3)

	// Call to GetLastError
	errCode, _, _ := getLastError.Call()
	fmt.Printf("GetLastError: %d\n", errCode)
}

