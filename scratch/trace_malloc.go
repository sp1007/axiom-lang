package main

import (
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
	
	epRVA := uintptr(0x5DEE) // Entry Point RVA from optional header
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
							// Adjust RIP back by 1 (since breakpoint hit advances RIP)
							rip := binaryGetUint64(&ctx, 0xF8) - 1
							binaryPutUint64(&ctx, 0xF8, rip)
							
							// EFlags is at offset 0x44 (68 decimal)
							eflags := binaryGetUint32(&ctx, 0x44)
							eflags |= 0x100 // Set TF (Trap Flag)
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
							if offset < 0x20000 {
								fmt.Printf("  RIP: 0x%X (Offset from base: 0x%X)\n", rip, offset)
							} else {
								// DLL code
								fmt.Printf("  RIP: 0x%X (External DLL)\n", rip)
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
				
				// Read a few bytes at the instruction pointer that crashed
				var codeBytes [16]byte
				if readProcessMemory(pi.Process, rec.ExceptionAddress, uintptr(unsafe.Pointer(&codeBytes[0])), 16) {
					fmt.Printf("  Code bytes at crash RIP: %02X\n", codeBytes)
				}
				
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
