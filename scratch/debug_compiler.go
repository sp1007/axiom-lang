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

func main() {
	exePath := "bin\\axc_stage1_debug.exe"
	
	var si syscall.StartupInfo
	var pi syscall.ProcessInformation
	si.Cb = uint32(unsafe.Sizeof(si))
	
	argv, _ := syscall.UTF16PtrFromString(exePath + " build scratch\\test_print.ax -o print.exe")
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
	
	for {
		if !waitForDebugEvent(&event, syscall.INFINITE) {
			break
		}
		
		status := uint32(DBG_CONTINUE)
		
		switch event.dwDebugEventCode {
		case CREATE_PROCESS_DEBUG_EVENT:
			lpBaseOfImage := *(*uintptr)(unsafe.Pointer(&event.u[24]))
			imageBase = lpBaseOfImage
			fmt.Printf("[DEBUG] Target image base: 0x%X\n", imageBase)
			
		case EXIT_PROCESS_DEBUG_EVENT:
			exitCode := *(*uint32)(unsafe.Pointer(&event.u[0]))
			fmt.Printf("[DEBUG] Process exited with code: %d\n", exitCode)
			continueDebugEvent(event.dwProcessId, event.dwThreadId, status)
			return
		
		case EXCEPTION_DEBUG_EVENT:
			exceptionInfo := (*EXCEPTION_DEBUG_INFO)(unsafe.Pointer(&event.u[0]))
			rec := exceptionInfo.ExceptionRecord
			
			if rec.ExceptionCode == STATUS_BREAKPOINT {
				// Continue past breakpoint
			} else if rec.ExceptionCode == STATUS_SINGLE_STEP {
				// Continue past single step
			} else {
				fmt.Printf("\n!!! Exception 0x%X at RIP Address 0x%X (Offset from image base: 0x%X) !!!\n", rec.ExceptionCode, rec.ExceptionAddress, rec.ExceptionAddress - imageBase)
				r, _, _ := procOpenThread.Call(0x0008, 0, uintptr(event.dwThreadId)) // GET_CONTEXT
				hThread := syscall.Handle(r)
				if r != 0 {
					var ctx [1232]byte
					binaryPutUint32(&ctx, 0x30, 0x10001F)
					rCtx, _, _ := procGetThreadCtx.Call(uintptr(hThread), uintptr(unsafe.Pointer(&ctx[0])))
					if rCtx != 0 {
						fmt.Printf("Registers:\n")
						fmt.Printf("  RAX: 0x%016X  RCX: 0x%016X  RDX: 0x%016X  RBX: 0x%016X\n",
							binaryGetUint64(&ctx, 0x78), binaryGetUint64(&ctx, 0x80), binaryGetUint64(&ctx, 0x88), binaryGetUint64(&ctx, 0x90))
						fmt.Printf("  RSP: 0x%016X  RBP: 0x%016X  RSI: 0x%016X  RDI: 0x%016X\n",
							binaryGetUint64(&ctx, 0x98), binaryGetUint64(&ctx, 0xA0), binaryGetUint64(&ctx, 0xA8), binaryGetUint64(&ctx, 0xB0))
						fmt.Printf("  R8:  0x%016X  R9:  0x%016X  R10: 0x%016X  R11: 0x%016X\n",
							binaryGetUint64(&ctx, 0xB8), binaryGetUint64(&ctx, 0xC0), binaryGetUint64(&ctx, 0xC8), binaryGetUint64(&ctx, 0xD0))
						fmt.Printf("  R12: 0x%016X  R13: 0x%016X  R14: 0x%016X  R15: 0x%016X\n",
							binaryGetUint64(&ctx, 0xD8), binaryGetUint64(&ctx, 0xE0), binaryGetUint64(&ctx, 0xE8), binaryGetUint64(&ctx, 0xF0))
						fmt.Printf("  RIP: 0x%016X (Offset: 0x%X)\n",
							binaryGetUint64(&ctx, 0xF8), binaryGetUint64(&ctx, 0xF8) - uint64(imageBase))
						
						fmt.Printf("\nStack Dump (at RSP 0x%X):\n", binaryGetUint64(&ctx, 0x98))
						var stackVal uint64
						for i := 0; i < 32; i++ {
							addr := uintptr(binaryGetUint64(&ctx, 0x98)) + uintptr(i * 8)
							if readProcessMemory(pi.Process, addr, uintptr(unsafe.Pointer(&stackVal)), 8) {
								fmt.Printf("  [RSP+%02X] 0x%016X (Offset: 0x%X)\n", i * 8, stackVal, stackVal - uint64(imageBase))
							}
						}
					}
					syscall.CloseHandle(hThread)
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

func binaryGetUint64(buf *[1232]byte, offset int) uint64 {
	var val uint64
	for i := 0; i < 8; i++ {
		val |= uint64(buf[offset+i]) << (i * 8)
	}
	return val
}
