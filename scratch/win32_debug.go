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

type CONTEXT struct {
	P1Home               uint64
	P2Home               uint64
	P3Home               uint64
	P4Home               uint64
	P5Home               uint64
	P6Home               uint64
	ContextFlags         uint32
	MxCsr                uint32
	SegCs                uint16
	SegDs                uint16
	SegEs                uint16
	SegFs                uint16
	SegGs                uint16
	SegSs                uint16
	EFlags               uint32
	Dr0                  uint64
	Dr1                  uint64
	Dr2                  uint64
	Dr3                  uint64
	Dr6                  uint64
	Dr7                  uint64
	Rax                  uint64
	Rcx                  uint64
	Rdx                  uint64
	Rbx                  uint64
	Rsp                  uint64
	Rbp                  uint64
	Rsi                  uint64
	Rdi                  uint64
	R8                   uint64
	R9                   uint64
	R10                  uint64
	R11                  uint64
	R12                  uint64
	R13                  uint64
	R14                  uint64
	R15                  uint64
	Rip                  uint64
}

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
	procOpenThread         = kernel32.NewProc("OpenThread")
	procGetThreadContext   = kernel32.NewProc("GetThreadContext")
	procReadProcessMemory  = kernel32.NewProc("ReadProcessMemory")
	procGetFinalPathNameByHandle = kernel32.NewProc("GetFinalPathNameByHandleW")
)

func main() {
	runtime.LockOSThread()

	targetExe := "bin/axc_stage2.exe"
	mutArgs := " build scratch/test_simple_args.ax -o scratch/test_simple_args_s2.exe"
	if len(os.Args) > 1 {
		targetExe = os.Args[1]
		if len(os.Args) > 2 {
			mutArgs = ""
			for _, arg := range os.Args[2:] {
				mutArgs += " " + arg
			}
		}
	}

	// Start target process suspended and debugged
	var si syscall.StartupInfo
	var pi syscall.ProcessInformation
	argv := syscall.StringToUTF16Ptr(targetExe + mutArgs)
	
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
			
			// Resolve DLL name from hFile
			dllName := "Unknown"
			if info.hFile != 0 && info.hFile != syscall.InvalidHandle {
				var buf [512]uint16
				r, _, _ := procGetFinalPathNameByHandle.Call(
					uintptr(info.hFile),
					uintptr(unsafe.Pointer(&buf[0])),
					512,
					0, // VOLUME_NAME_DOS
				)
				if r > 0 && r < 512 {
					dllName = syscall.UTF16ToString(buf[:r])
				}
			}
			fmt.Printf("DLL Loaded: Base Address 0x%X -> %s\n", info.lpBaseOfDll, dllName)

		case 5: // EXIT_PROCESS_DEBUG_EVENT
			fmt.Println("Process Exited.")
			return

		case EXCEPTION_DEBUG_EVENT:
			info := (*EXCEPTION_DEBUG_INFO)(unsafe.Pointer(&event.u[0]))
			rec := info.ExceptionRecord
			fmt.Printf("Exception Event: Code=0x%X Address=0x%X\n", rec.ExceptionCode, rec.ExceptionAddress)
			if rec.ExceptionCode == EXCEPTION_ACCESS_VIOLATION || rec.ExceptionCode == 0xC00000FD || rec.ExceptionCode == 0xC0000374 {
				if rec.ExceptionCode == 0xC00000FD {
					fmt.Printf("\n>>> STACK OVERFLOW (0xC00000FD) occurred at address: 0x%X\n", rec.ExceptionAddress)
				} else if rec.ExceptionCode == 0xC0000374 {
					fmt.Printf("\n>>> HEAP CORRUPTION (0xC0000374) occurred at address: 0x%X\n", rec.ExceptionAddress)
				} else {
					fmt.Printf("\n>>> ACCESS VIOLATION (0xC0000005) occurred at address: 0x%X\n", rec.ExceptionAddress)
				}
				
				// Map Exception Address to function
				rva := rec.ExceptionAddress - baseAddress
				fmt.Printf(">>> Crash RVA (Relative Virtual Address): 0x%X\n", rva)

				mapRVAToFunction(targetExe, uint32(rva))
				
				readWrite := "read"
				if rec.ExceptionInformation[0] == 1 {
					readWrite = "write"
				} else if rec.ExceptionInformation[0] == 8 {
					readWrite = "execute DEP"
				}
				fmt.Printf(">>> Access type: %s at target address: 0x%X\n", readWrite, rec.ExceptionInformation[1])
				
				// Dump instruction bytes at crash address
				var instBytes [16]byte
				var nReadInst uintptr
				procReadProcessMemory.Call(
					uintptr(pi.Process),
					uintptr(rec.ExceptionAddress),
					uintptr(unsafe.Pointer(&instBytes[0])),
					16,
					uintptr(unsafe.Pointer(&nReadInst)),
				)
				if nReadInst > 0 {
					fmt.Printf(">>> Instruction bytes at RIP: ")
					for i := uintptr(0); i < nReadInst; i++ {
						fmt.Printf("%02X ", instBytes[i])
					}
					fmt.Println()
				}

				// Dump Thread Context Registers
				var ctx CONTEXT
				ctx.ContextFlags = 0x10000B // CONTEXT_CONTROL | CONTEXT_INTEGER
				hThread, _, _ := procOpenThread.Call(0x0008, 0, uintptr(event.dwThreadId))
				if hThread != 0 {
					defer syscall.CloseHandle(syscall.Handle(hThread))
					rContext, _, _ := procGetThreadContext.Call(hThread, uintptr(unsafe.Pointer(&ctx)))
					if rContext != 0 {
						fmt.Println("\n>>> Register State at Crash:")
						fmt.Printf("  RAX: 0x%016X  RCX: 0x%016X\n", ctx.Rax, ctx.Rcx)
						fmt.Printf("  RDX: 0x%016X  RBX: 0x%016X\n", ctx.Rdx, ctx.Rbx)
						fmt.Printf("  RSP: 0x%016X  RBP: 0x%016X\n", ctx.Rsp, ctx.Rbp)
						fmt.Printf("  RSI: 0x%016X  RDI: 0x%016X\n", ctx.Rsi, ctx.Rdi)
						fmt.Printf("  R8 : 0x%016X  R9 : 0x%016X\n", ctx.R8, ctx.R9)
						fmt.Printf("  R10: 0x%016X  R11: 0x%016X\n", ctx.R10, ctx.R11)
						fmt.Printf("  R12: 0x%016X  R13: 0x%016X\n", ctx.R12, ctx.R13)
						fmt.Printf("  R14: 0x%016X  R15: 0x%016X\n", ctx.R14, ctx.R15)
						fmt.Printf("  RIP: 0x%016X  EFLAGS: 0x%08X\n", ctx.Rip, ctx.EFlags)
						
						// Read 1024 stack values (8KB of stack) starting from ctx.Rsp
						var stackVals [1024]uint64
						var nReadRet uintptr
						rRead, _, _ := procReadProcessMemory.Call(
							uintptr(pi.Process),
							uintptr(ctx.Rsp),
							uintptr(unsafe.Pointer(&stackVals[0])),
							8192,
							uintptr(unsafe.Pointer(&nReadRet)),
						)
						
						if rRead != 0 && nReadRet > 0 {
							numVals := nReadRet / 8
							fmt.Printf("\n>>> Call Stack trace (.text section matches in top %d stack values starting at RSP):\n", numVals)
							foundAny := false
							for i := uintptr(0); i < numVals; i++ {
								val := stackVals[i]
								// Check if val is inside our .text section (RVA [0x1000, 0x1000 + 0x200000])
								if val >= uint64(baseAddress+0x1000) && val < uint64(baseAddress+0x1000+0x200000) {
									rva := val - uint64(baseAddress)
									fmt.Printf("  RSP+0x%X: 0x%016X (RVA 0x%X) -> \n", i*8, val, rva)
									mapRVAToFunction(targetExe, uint32(rva))
									foundAny = true
								}
							}
							if !foundAny {
								fmt.Println("  No addresses inside .text section found on the stack. Top stack values:")
								for i := uintptr(0); i < 16 && i < numVals; i++ {
									fmt.Printf("  RSP+0x%X: 0x%016X\n", i*8, stackVals[i])
								}
							}
						} else {
							fmt.Println("\n>>> Failed to read stack memory from RSP.")
						}

						// Read IAT slot of GetCommandLineW (RVA 0x990B8)
						var iatVal uint64
						var nRead uintptr
						rReadIAT, _, _ := procReadProcessMemory.Call(
							uintptr(pi.Process),
							baseAddress + 0x990B8,
							uintptr(unsafe.Pointer(&iatVal)),
							8,
							uintptr(unsafe.Pointer(&nRead)),
						)
						if rReadIAT != 0 && nRead == 8 {
							fmt.Printf("\n>>> GetCommandLineW IAT Slot (RVA 0x990B8) Value: 0x%016X\n", iatVal)
						} else {
							fmt.Println("\n>>> Failed to read IAT slot memory.")
						}
					} else {
						fmt.Println(">>> Failed to get thread context.")
					}
				} else {
					fmt.Println(">>> Failed to OpenThread.")
				}
				
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
