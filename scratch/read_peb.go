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

type PROCESS_BASIC_INFORMATION struct {
	Reserved1      uintptr
	PebBaseAddress uintptr
	Reserved2      [2]uintptr
	UniqueProcessId uintptr
	Reserved3      uintptr
}

var (
	modntdll             = syscall.NewLazyDLL("ntdll.dll")
	procNtQueryInfoProc  = modntdll.NewProc("NtQueryInformationProcess")
	modkernel32          = syscall.NewLazyDLL("kernel32.dll")
	procReadProcessMem   = modkernel32.NewProc("ReadProcessMemory")
)

func ntQueryInformationProcess(hProcess syscall.Handle, processInformationClass int, processInformation uintptr, processInformationLength uint32, returnLength *uint32) uintptr {
	r, _, _ := procNtQueryInfoProc.Call(
		uintptr(hProcess),
		uintptr(processInformationClass),
		processInformation,
		uintptr(processInformationLength),
		uintptr(unsafe.Pointer(returnLength)),
	)
	return r
}

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

	const PROCESS_VM_READ = 0x0010
	const PROCESS_QUERY_INFORMATION = 0x0400
	hProcess, err := syscall.OpenProcess(PROCESS_VM_READ|PROCESS_QUERY_INFORMATION, false, uint32(cmd.Process.Pid))
	if err != nil {
		log.Fatalf("OpenProcess failed: %v", err)
	}
	defer syscall.CloseHandle(hProcess)

	var pbi PROCESS_BASIC_INFORMATION
	var retLen uint32
	status := ntQueryInformationProcess(hProcess, 0, uintptr(unsafe.Pointer(&pbi)), uint32(unsafe.Sizeof(pbi)), &retLen)
	if status != 0 {
		log.Fatalf("NtQueryInformationProcess failed with status 0x%X", status)
	}

	fmt.Printf("PEB Base Address: 0x%X\n", pbi.PebBaseAddress)

	// Read ImageBaseAddress from PEB (offset 0x10 in x64 PEB)
	var imageBase uintptr
	var read uintptr
	ok := readProcessMemory(hProcess, pbi.PebBaseAddress+0x10, uintptr(unsafe.Pointer(&imageBase)), uintptr(unsafe.Sizeof(imageBase)), &read)
	if !ok {
		log.Fatalf("Failed to read ImageBaseAddress from PEB")
	}

	fmt.Printf("Actual Image Base Address at Runtime: 0x%X\n", imageBase)

	iatRVA := uintptr(0x6040) // VirtualAlloc IAT entry RVA in test_malloc.exe
	iatAddr := imageBase + iatRVA

	var val uint64
	ok = readProcessMemory(hProcess, iatAddr, uintptr(unsafe.Pointer(&val)), 8, &read)
	if !ok {
		fmt.Printf("ReadProcessMemory failed at IAT address 0x%X\n", iatAddr)
	} else {
		fmt.Printf("Value at VirtualAlloc IAT entry (0x%X): 0x%X (Relative to base: 0x%X)\n", iatAddr, val, val - uint64(imageBase))
	}

	// Also read first few IAT entries
	for i := uintptr(0); i < 5; i++ {
		var v uint64
		if readProcessMemory(hProcess, iatAddr+i*8, uintptr(unsafe.Pointer(&v)), 8, &read) {
			fmt.Printf("  IAT[%d] (0x%X): 0x%X (Relative: 0x%X)\n", i, iatAddr+i*8, v, v - uint64(imageBase))
		}
	}
}
