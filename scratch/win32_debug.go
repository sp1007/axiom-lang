package main

import (
	"debug/pe"
	"fmt"
	"log"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

const (
	DEBUG_ONLY_THIS_PROCESS = 0x00000002
	EXCEPTION_DEBUG_EVENT   = 1
	CREATE_PROCESS_DEBUG_EVENT = 3
	LOAD_DLL_DEBUG_EVENT    = 6
	EXCEPTION_ACCESS_VIOLATION = 0xC0000005
)

type EXCEPTION_RECORD struct {
	ExceptionCode        uint32
	ExceptionFlags       uint32
	ExceptionRecord      uintptr
	ExceptionAddress     uintptr
	NumberParameters     uint32
	ExceptionInformation [15]uintptr
}

type EXCEPTION_DEBUG_INFO struct {
	ExceptionRecord EXCEPTION_RECORD
	dwFirstChance   uint32
}

type CREATE_PROCESS_DEBUG_INFO struct {
	hFile                 syscall.Handle
	hProcess              syscall.Handle
	hThread               syscall.Handle
	lpBaseOfImage         uintptr
	dwDebugInfoFileOffset uint32
	nDebugInfoSize        uint32
	lpThreadLocalBase     uintptr
	lpStartAddress        uintptr
	lpImageName           uintptr
	fUnicode              uint16
}

type DEBUG_EVENT struct {
	dwDebugEventCode uint32
	dwProcessId      uint32
	dwThreadId       uint32
	padding          uint32
	u                [160]byte // large enough buffer for union
}

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procDebugActiveProcess = kernel32.NewProc("DebugActiveProcess")
	procWaitForDebugEvent  = kernel32.NewProc("WaitForDebugEvent")
	procContinueDebugEvent = kernel32.NewProc("ContinueDebugEvent")
)

func main() {
	runtime.LockOSThread()

	targetExe := "bin/axc_stage2.exe"
	if len(os.Args) > 1 {
		targetExe = os.Args[1]
	}

	// Start target process suspended and debugged
	var si syscall.StartupInfo
	var pi syscall.ProcessInformation
	argv := syscall.StringToUTF16Ptr(targetExe + " build scratch/test_simple_args.ax -o scratch/test_simple_args_s2.exe")
	
	err := syscall.CreateProcess(
		nil,
		argv,
		nil,
		nil,
		false,
		DEBUG_ONLY_THIS_PROCESS,
		nil,
		nil,
		&si,
		&pi,
	)
	if err != nil {
		log.Fatalf("CreateProcess failed: %v", err)
	}
	defer syscall.CloseHandle(pi.Process)
	defer syscall.CloseHandle(pi.Thread)

	var baseAddress uintptr

	// Debug loop
	var event DEBUG_EVENT
	for {
		r, _, _ := procWaitForDebugEvent.Call(uintptr(unsafe.Pointer(&event)), 0xFFFFFFFF)
		if r == 0 {
			break
		}

		continueStatus := uint32(0x00010002) // DBG_EXCEPTION_NOT_HANDLED by default

		switch event.dwDebugEventCode {
		case CREATE_PROCESS_DEBUG_EVENT:
			info := (*CREATE_PROCESS_DEBUG_INFO)(unsafe.Pointer(&event.u[0]))
			baseAddress = info.lpBaseOfImage
			fmt.Printf("Process Created. Image Base Address: 0x%X\n", baseAddress)

		case LOAD_DLL_DEBUG_EVENT:
			type LOAD_DLL_DEBUG_INFO struct {
				hFile                 syscall.Handle
				lpBaseOfDll           uintptr
				dwDebugInfoFileOffset uint32
				nDebugInfoSize        uint32
				lpImageName           uintptr
				fUnicode              uint16
			}
			info := (*LOAD_DLL_DEBUG_INFO)(unsafe.Pointer(&event.u[0]))
			fmt.Printf("DLL Loaded: Base Address 0x%X\n", info.lpBaseOfDll)

		case EXCEPTION_DEBUG_EVENT:
			info := (*EXCEPTION_DEBUG_INFO)(unsafe.Pointer(&event.u[0]))
			rec := info.ExceptionRecord
			if rec.ExceptionCode == EXCEPTION_ACCESS_VIOLATION {
				fmt.Printf("\n>>> ACCESS VIOLATION (0xC0000005) occurred at address: 0x%X\n", rec.ExceptionAddress)
				
				// Map Exception Address to function
				rva := rec.ExceptionAddress - baseAddress
				fmt.Printf(">>> Crash RVA (Relative Virtual Address): 0x%X\n", rva)

				mapRVAToFunction(targetExe, uint32(rva))
				
				// Kill the process after capturing crash info
				syscall.TerminateProcess(pi.Process, 1)
				return
			}
		}

		// Continue execution of the thread that raised the debug event
		procContinueDebugEvent.Call(
			uintptr(event.dwProcessId),
			uintptr(event.dwThreadId),
			uintptr(continueStatus),
		)
	}
}

func mapRVAToFunction(exePath string, crashRVA uint32) {
	file, err := pe.Open(exePath)
	if err != nil {
		fmt.Printf("Failed to open PE %s: %v\n", exePath, err)
		return
	}
	defer file.Close()

	var textSec *pe.Section
	for _, sec := range file.Sections {
		if sec.Name == ".text" {
			textSec = sec
			break
		}
	}

	if textSec == nil {
		fmt.Println(".text section not found in PE")
		return
	}

	fmt.Printf("  .text section RVA: 0x%X (Size: 0x%X)\n", textSec.VirtualAddress, textSec.VirtualSize)
	
	if crashRVA < textSec.VirtualAddress || crashRVA >= textSec.VirtualAddress+textSec.VirtualSize {
		fmt.Println("  Crash address is OUTSIDE the .text section!")
		return
	}

	crashOffset := crashRVA - textSec.VirtualAddress
	fmt.Printf("  Crash Offset inside .text: 0x%X\n", crashOffset)

	// Dump symbols from axiom_temp.obj
	type FuncSym struct {
		Name   string
		Offset uint32
	}
	var fns []FuncSym

	objFile, err := pe.Open("axiom_temp.obj")
	if err == nil {
		defer objFile.Close()
		for _, sym := range objFile.COFFSymbols {
			name, err := sym.FullName(objFile.StringTable)
			if err != nil {
				name = string(sym.Name[:])
			}
			// In COFF object files, function symbols have SectionNumber > 0 and Type 0x20
			if sym.SectionNumber > 0 && sym.Type == 0x20 {
				fns = append(fns, FuncSym{Name: name, Offset: sym.Value})
			}
		}
	} else {
		fmt.Printf("Failed to open axiom_temp.obj: %v\n", err)
	}

	if len(fns) == 0 {
		fmt.Println("  No symbols found in PE. Scanning for prologues (55 48 89 E5 / 55 48 89 E5)...")
		data, err := textSec.Data()
		if err == nil {
			for i := 0; i < len(data)-4; i++ {
				if data[i] == 0x55 && data[i+1] == 0x48 && data[i+2] == 0x89 && data[i+3] == 0xE5 {
					fns = append(fns, FuncSym{
						Name:   fmt.Sprintf("fn_at_0x%X", i),
						Offset: uint32(i),
					})
				}
			}
		}
	}

	// Sort fns by offset
	// Let's find the function immediately preceding the crash offset
	var crashingFn FuncSym
	found := false
	bestDiff := uint32(0xFFFFFFFF)

	for _, f := range fns {
		if f.Offset <= crashOffset {
			diff := crashOffset - f.Offset
			if diff < bestDiff {
				bestDiff = diff
				crashingFn = f
				found = true
			}
		}
	}

	if found {
		fmt.Printf("\n>>> Crashing Function: %s (starts at Offset 0x%X)\n", crashingFn.Name, crashingFn.Offset)
		fmt.Printf(">>> Crash happened at Offset + 0x%X (%d) bytes inside %s\n", bestDiff, bestDiff, crashingFn.Name)
	} else {
		fmt.Println("  Could not find preceding function symbol.")
	}
}
