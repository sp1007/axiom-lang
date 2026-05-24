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

var (
	modkernel32       = syscall.NewLazyDLL("kernel32.dll")
	procReadProcessMem = modkernel32.NewProc("ReadProcessMemory")
)

func readProcessMemory(hProcess syscall.Handle, lpBaseAddress uintptr, lpBuffer uintptr, nSize uintptr, lpNumberOfBytesRead *uintptr) bool {
	r, _, _ := procReadProcessMem.Call(
		uintptr(hProcess),
		lpBaseAddress,
		lpBuffer,
		nSize,
		uintptr(unsafe.Pointer(lpNumberOfBytesRead)),
	)
	return r != 0
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

	hProcess := syscall.Handle(cmd.Process.Pid) // wait, Process.Pid is PID, not process handle.
	const PROCESS_VM_READ = 0x0010
	const PROCESS_QUERY_INFORMATION = 0x0400
	hProcess, err := syscall.OpenProcess(PROCESS_VM_READ|PROCESS_QUERY_INFORMATION, false, uint32(cmd.Process.Pid))
	if err != nil {
		log.Fatalf("OpenProcess failed: %v", err)
	}
	defer syscall.CloseHandle(hProcess)

	imageBase := uintptr(0x140000000)
	iatRVA := uintptr(0x603C) // VirtualAlloc IAT entry RVA in test_malloc.exe
	iatAddr := imageBase + iatRVA

	var val uint64
	var read uintptr
	ok := readProcessMemory(hProcess, iatAddr, uintptr(unsafe.Pointer(&val)), 8, &read)
	if !ok {
		fmt.Printf("ReadProcessMemory failed at 0x%X\n", iatAddr)
	} else {
		fmt.Printf("Value at VirtualAlloc IAT entry (0x%X): 0x%X\n", iatAddr, val)
	}

	// Also read first few IAT entries
	for i := uintptr(0); i < 5; i++ {
		var v uint64
		if readProcessMemory(hProcess, iatAddr+i*8, uintptr(unsafe.Pointer(&v)), 8, &read) {
			fmt.Printf("  IAT[%d] (0x%X): 0x%X\n", i, iatAddr+i*8, v)
		}
	}

	// Let's resume process and wait
	// To resume we can do ResumeThread on the main thread, or simply let it exit/kill
}
