package main

import (
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	CREATE_SUSPENDED = 0x00000004
)

type MEMORY_BASIC_INFORMATION struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	PartitionId       uint16
	RegionSize        uintptr
	State             uint32
	Protect           uint32
	Type              uint32
}

var (
	modkernel32       = syscall.NewLazyDLL("kernel32.dll")
	procVirtualQueryEx = modkernel32.NewProc("VirtualQueryEx")
)

func virtualQueryEx(hProcess syscall.Handle, lpAddress uintptr, lpBuffer uintptr, dwLength uintptr) uintptr {
	r, _, _ := procVirtualQueryEx.Call(
		uintptr(hProcess),
		lpAddress,
		lpBuffer,
		dwLength,
	)
	return r
}

func main() {
	cmd := exec.Command("d:\\projects\\compiler\\Axiom\\test_malloc.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: CREATE_SUSPENDED,
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start process suspended: %v", err)
	}
	defer cmd.Process.Kill()

	const PROCESS_VM_READ = 0x0010
	const PROCESS_QUERY_INFORMATION = 0x0400
	hProcess, err := syscall.OpenProcess(PROCESS_VM_READ|PROCESS_QUERY_INFORMATION, false, uint32(cmd.Process.Pid))
	if err != nil {
		log.Fatalf("OpenProcess failed: %v", err)
	}
	defer syscall.CloseHandle(hProcess)

	var mbi MEMORY_BASIC_INFORMATION
	addr := uintptr(0x14000603C)
	r := virtualQueryEx(hProcess, addr, uintptr(unsafe.Pointer(&mbi)), uintptr(unsafe.Sizeof(mbi)))
	if r == 0 {
		fmt.Printf("VirtualQueryEx failed\n")
	} else {
		fmt.Printf("Memory Info for 0x%X:\n", addr)
		fmt.Printf("  BaseAddress:       0x%X\n", mbi.BaseAddress)
		fmt.Printf("  AllocationProtect: 0x%X\n", mbi.AllocationProtect)
		fmt.Printf("  RegionSize:        0x%X\n", mbi.RegionSize)
		fmt.Printf("  State:             0x%X (Commit=0x1000, Reserve=0x2000, Free=0x10000)\n", mbi.State)
		fmt.Printf("  Protect:           0x%X (ReadWrite=0x04, ReadOnly=0x02, ExecuteRead=0x20)\n", mbi.Protect)
		fmt.Printf("  Type:              0x%X (Image=0x1000000)\n", mbi.Type)
	}
}
