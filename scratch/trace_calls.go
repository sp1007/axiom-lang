package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"syscall"
	"unsafe"
)

const (
	DEBUG_ONLY_THIS_PROCESS = 0x00000002
	CREATE_NEW_CONSOLE      = 0x00000010
	
	EXCEPTION_DEBUG_EVENT   = 1
	CREATE_PROCESS_DEBUG_EVENT = 3
	EXIT_PROCESS_DEBUG_EVENT   = 5
	
	DBG_CONTINUE            = 0x00010002
	DBG_EXCEPTION_NOT_HANDLED = 0x80010001
	
	STATUS_BREAKPOINT       = 0x80000003
	STATUS_SINGLE_STEP      = 0x80000004
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

type DEBUG_EVENT struct {
	dwDebugEventCode uint32
	dwProcessId      uint32
	dwThreadId       uint32
	padding          uint32
	u                [160]byte
}

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procWaitForDebugEv  = kernel32.NewProc("WaitForDebugEvent")
	procContinueDebugEv = kernel32.NewProc("ContinueDebugEvent")
	procReadProcessMem  = kernel32.NewProc("ReadProcessMemory")
	procWriteProcessMem = kernel32.NewProc("WriteProcessMemory")
	procGetThreadCtx    = kernel32.NewProc("GetThreadContext")
	procSetThreadCtx    = kernel32.NewProc("SetThreadContext")
	procOpenThread      = kernel32.NewProc("OpenThread")
)

func waitForDebugEvent(event *DEBUG_EVENT, timeout uint32) bool {
	r, _, _ := procWaitForDebugEv.Call(uintptr(unsafe.Pointer(event)), uintptr(timeout))
	return r != 0
}

func continueDebugEvent(pid, tid, status uint32) bool {
	r, _, _ := procContinueDebugEv.Call(uintptr(pid), uintptr(tid), uintptr(status))
	return r != 0
}

func readProcessMemory(hProcess syscall.Handle, addr uintptr, buf uintptr, size uintptr) bool {
	var read uintptr
	r, _, _ := procReadProcessMem.Call(uintptr(hProcess), addr, buf, size, uintptr(unsafe.Pointer(&read)))
	return r != 0
}

func writeProcessMemory(hProcess syscall.Handle, addr uintptr, buf uintptr, size uintptr) bool {
	var written uintptr
	r, _, _ := procWriteProcessMem.Call(uintptr(hProcess), addr, buf, size, uintptr(unsafe.Pointer(&written)))
	return r != 0
}

func main() {
	exePath := "d:\\projects\\compiler\\Axiom\\test_malloc.exe"
	
	var si syscall.StartupInfo
	var pi syscall.ProcessInformation
	si.Cb = uint32(unsafe.Sizeof(si))
	
	argv, _ := syscall.UTF16PtrFromString(exePath)
	err := syscall.CreateProcess(
		nil,
		argv,
		nil,
		nil,
		false,
		DEBUG_ONLY_THIS_PROCESS | CREATE_NEW_CONSOLE,
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
	
	var event DEBUG_EVENT
	var imageBase uintptr
	
	epRVA := uintptr(0x5DEE) // Entry Point RVA
	var epAddr uintptr
	var originalByte byte
	bpSet := false
	tracing := false
	
	for {
		if !waitForDebugEvent(&event, syscall.INFINITE) {
			break
		}
		
		status := uint32(DBG_CONTINUE)
		
		switch event.dwDebugEventCode {
		case CREATE_PROCESS_DEBUG_EVENT:
			lpBaseOfImage := *(*uintptr)(unsafe.Pointer(&event.u[24]))
			imageBase = lpBaseOfImage
			epAddr = imageBase + epRVA
			
			// Set breakpoint at Entry Point
			var b byte
			if readProcessMemory(pi.Process, epAddr, uintptr(unsafe.Pointer(&b)), 1) {
				originalByte = b
				bpByte := byte(0xCC)
				writeProcessMemory(pi.Process, epAddr, uintptr(unsafe.Pointer(&bpByte)), 1)
				bpSet = true
				fmt.Printf("[DEBUG] Set Entry Point Breakpoint at 0x%X\n", epAddr)
			}
			
		case EXIT_PROCESS_DEBUG_EVENT:
			exitCode := *(*uint32)(unsafe.Pointer(&event.u[0]))
			fmt.Printf("[DEBUG] Process exited with code: %d\n", exitCode)
			continueDebugEvent(event.dwProcessId, event.dwThreadId, status)
			return

		case EXCEPTION_DEBUG_EVENT:
			exceptionInfo := (*EXCEPTION_DEBUG_INFO)(unsafe.Pointer(&event.u[0]))
			rec := exceptionInfo.ExceptionRecord
			
			if rec.ExceptionCode == STATUS_BREAKPOINT {
				if bpSet && rec.ExceptionAddress == epAddr {
					fmt.Printf("[DEBUG] Hit Entry Point Breakpoint! Starting execution trace...\n")
					
					// Restore original byte
					writeProcessMemory(pi.Process, epAddr, uintptr(unsafe.Pointer(&originalByte)), 1)
					
					// Set trap flag (single step)
					r, _, _ := procOpenThread.Call(0x0008|0x0010, 0, uintptr(event.dwThreadId)) // GET_CONTEXT | SET_CONTEXT
					hThread := syscall.Handle(r)
					if r != 0 {
						var ctx [1232]byte
						binaryPutUint32(&ctx, 0x30, 0x10001F) // CONTEXT_ALL
						rCtx, _, _ := procGetThreadCtx.Call(uintptr(hThread), uintptr(unsafe.Pointer(&ctx[0])))
						if rCtx != 0 {
							rip := binaryGetUint64(&ctx, 0xF8) - 1
							binaryPutUint64(&ctx, 0xF8, rip)
							
							eflags := binaryGetUint32(&ctx, 0x44)
							eflags |= 0x100 // Set TF
							binaryPutUint32(&ctx, 0x44, eflags)
							
							procSetThreadCtx.Call(uintptr(hThread), uintptr(unsafe.Pointer(&ctx[0])))
							tracing = true
						}
						syscall.CloseHandle(hThread)
					}
				}
			} else if rec.ExceptionCode == STATUS_SINGLE_STEP && tracing {
				// Single step hit
				r, _, _ := procOpenThread.Call(0x0008|0x0010, 0, uintptr(event.dwThreadId))
				hThread := syscall.Handle(r)
				if r != 0 {
					var ctx [1232]byte
					binaryPutUint32(&ctx, 0x30, 0x10001F)
					rCtx, _, _ := procGetThreadCtx.Call(uintptr(hThread), uintptr(unsafe.Pointer(&ctx[0])))
					if rCtx != 0 {
						rip := binaryGetUint64(&ctx, 0xF8)
						offset := rip - uint64(imageBase)
						
						// If the instruction is a call/jmp thunk to DLL, or a syscall, print it!
						// Thunks are located at the end of the code section, e.g., starting at offset 0x5E5A.
						if offset >= 0x5E5A && offset < 0x6000 {
							// Check what instruction is executed
							var instr [6]byte
							if readProcessMemory(pi.Process, uintptr(rip), uintptr(unsafe.Pointer(&instr[0])), 6) {
								if instr[0] == 0xFF && instr[1] == 0x25 { // jmp [rip + disp32]
									disp32 := binary.LittleEndian.Uint32(instr[2:6])
									iat_addr := uintptr(rip) + 6 + uintptr(disp32)
									var api_ptr uintptr
									if readProcessMemory(pi.Process, iat_addr, uintptr(unsafe.Pointer(&api_ptr)), 8) {
										fmt.Printf("--- CALLING DLL API at IAT 0x%X -> Target Address: 0x%X\n", iat_addr, api_ptr)
									}
								}
							}
						}
						
						if offset == 0x1097 {
							fmt.Printf("[TRACE] get_global_state calling VirtualAlloc\n")
							fmt.Printf("  RCX (Address): 0x%X\n", binaryGetUint64(&ctx, 0x80))
							fmt.Printf("  RDX (Size):    0x%X\n", binaryGetUint64(&ctx, 0x88))
							fmt.Printf("  R8  (Alloc):   0x%X\n", binaryGetUint64(&ctx, 0xB8))
							fmt.Printf("  R9  (Protect): 0x%X\n", binaryGetUint64(&ctx, 0xC0))
						}
						if offset == 0x109C {
							fmt.Printf("  -> VirtualAlloc returned: RAX = 0x%X\n", binaryGetUint64(&ctx, 0x78))
						}
						if offset == 0x10FA {
							fmt.Printf("[TRACE] get_global_state calling VirtualAlloc (attempt 2)\n")
						}
						if offset == 0x10FF {
							fmt.Printf("  -> VirtualAlloc returned (attempt 2): RAX = 0x%X\n", binaryGetUint64(&ctx, 0x78))
						}
						if offset == 0x113E {
							fmt.Printf("[TRACE] get_global_state calling GetLastError\n")
						}
						if offset == 0x1143 {
							fmt.Printf("  -> GetLastError returned: RAX = 0x%X\n", binaryGetUint64(&ctx, 0x78))
						}
						if offset == 0x1158 {
							fmt.Printf("[TRACE] get_global_state calling ExitProcess with error code: RAX = 0x%X\n", binaryGetUint64(&ctx, 0x78))
						}
						if offset == 0x5E1F {
							fmt.Printf("[TRACE] main calling ExitProcess with code: RCX = 0x%X\n", binaryGetUint64(&ctx, 0x80))
						}
						
						// Keep tracing
						eflags := binaryGetUint32(&ctx, 0x44)
						eflags |= 0x100
						binaryPutUint32(&ctx, 0x44, eflags)
						procSetThreadCtx.Call(uintptr(hThread), uintptr(unsafe.Pointer(&ctx[0])))
					}
					syscall.CloseHandle(hThread)
				}
			} else if rec.ExceptionCode == 0xC0000005 {
				fmt.Printf("\n!!! Access Violation at 0x%X (Offset: 0x%X) !!!\n", rec.ExceptionAddress, rec.ExceptionAddress - imageBase)
				status = uint32(DBG_EXCEPTION_NOT_HANDLED)
				continueDebugEvent(event.dwProcessId, event.dwThreadId, status)
				return
			}
		}
		
		continueDebugEvent(event.dwProcessId, event.dwThreadId, status)
	}
}

func binaryPutUint32(buf *[1232]byte, offset int, val uint32) {
	buf[offset] = byte(val)
	buf[offset+1] = byte(val >> 8)
	buf[offset+2] = byte(val >> 16)
	buf[offset+3] = byte(val >> 24)
}

func binaryPutUint64(buf *[1232]byte, offset int, val uint64) {
	for i := 0; i < 8; i++ {
		buf[offset+i] = byte(val >> (i * 8))
	}
}

func binaryGetUint32(buf *[1232]byte, offset int) uint32 {
	return uint32(buf[offset]) | (uint32(buf[offset+1]) << 8) | (uint32(buf[offset+2]) << 16) | (uint32(buf[offset+3]) << 24)
}

func binaryGetUint64(buf *[1232]byte, offset int) uint64 {
	var val uint64
	for i := 0; i < 8; i++ {
		val |= uint64(buf[offset+i]) << (i * 8)
	}
	return val
}
