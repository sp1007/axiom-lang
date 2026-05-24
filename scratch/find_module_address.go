package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

type MODULEINFO struct {
	lpBaseOfDll uintptr
	SizeOfImage uint32
	EntryPoint  uintptr
}

var (
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	procGetModHandle  = kernel32.NewProc("GetModuleHandleW")
	procGetModInfo    = syscall.NewLazyDLL("psapi.dll").NewProc("GetModuleInformation")
	procGetCurrentProc = kernel32.NewProc("GetCurrentProcess")
)

func getModuleInfo(name string) (uintptr, uint32) {
	utf16, _ := syscall.UTF16PtrFromString(name)
	hMod, _, _ := procGetModHandle.Call(uintptr(unsafe.Pointer(utf16)))
	if hMod == 0 {
		return 0, 0
	}

	hProc, _, _ := procGetCurrentProc.Call()
	var info MODULEINFO
	r, _, _ := procGetModInfo.Call(hProc, hMod, uintptr(unsafe.Pointer(&info)), unsafe.Sizeof(info))
	if r == 0 {
		return hMod, 0
	}
	return info.lpBaseOfDll, info.SizeOfImage
}

func main() {
	// Preload ax_runtime.dll
	syscall.LoadLibrary("bin\\ax_runtime.dll")

	modules := []string{"kernel32.dll", "ntdll.dll", "ucrtbase.dll", "ax_runtime.dll"}
	fmt.Println("Loaded Modules:")
	for _, m := range modules {
		base, size := getModuleInfo(m)
		fmt.Printf("  %-16s Base: 0x%X - 0x%X (Size: 0x%X)\n", m, base, base+uintptr(size), size)
	}

	fmt.Println("\nFunction Addresses:")
	k32 := syscall.NewLazyDLL("kernel32.dll")
	fmt.Printf("  VirtualAlloc:     0x%X\n", k32.NewProc("VirtualAlloc").Addr())
	fmt.Printf("  VirtualFree:      0x%X\n", k32.NewProc("VirtualFree").Addr())
	fmt.Printf("  ExitProcess:      0x%X\n", k32.NewProc("ExitProcess").Addr())
	fmt.Printf("  TerminateProcess: 0x%X\n", k32.NewProc("TerminateProcess").Addr())

	ucrt := syscall.NewLazyDLL("ucrtbase.dll")
	fmt.Printf("  memset:           0x%X\n", ucrt.NewProc("memset").Addr())
	fmt.Printf("  memcpy:           0x%X\n", ucrt.NewProc("memcpy").Addr())

	targetAddr := uintptr(0x7FFC01D9D710)
	fmt.Printf("\nSearching module containing target address 0x%X:\n", targetAddr)
	found := false
	for _, m := range modules {
		base, size := getModuleInfo(m)
		if targetAddr >= base && targetAddr < base+uintptr(size) {
			fmt.Printf("  >>> Address 0x%X is inside %s (Offset: 0x%X)\n", targetAddr, m, targetAddr-base)
			found = true
			break
		}
	}
	if !found {
		fmt.Println("  >>> Address not found in the standard loaded modules.")
	}
}
