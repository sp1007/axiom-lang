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
	
	DBG_CONTINUE            = 0x00010002
	DBG_EXCEPTION_NOT_HANDLED = 0x80010001
	
	STATUS_BREAKPOINT       = 0x80000003
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
	procGetThreadCtx    = kernel32.NewProc("GetThreadContext")
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
	exePath := "d:\\projects\\compiler\\Axiom\\torture_gen_ref.exe"
	
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
	
	for {
		if !waitForDebugEvent(&event, syscall.INFINITE) {
			break
		}
		
		status := uint32(DBG_CONTINUE)
		
		switch event.dwDebugEventCode {
		case CREATE_PROCESS_DEBUG_EVENT:
			lpBaseOfImage := *(*uintptr)(unsafe.Pointer(&event.u[24]))
			imageBase = lpBaseOfImage
			
		case EXCEPTION_DEBUG_EVENT:
			exceptionInfo := (*EXCEPTION_DEBUG_INFO)(unsafe.Pointer(&event.u[0]))
			rec := exceptionInfo.ExceptionRecord
			
			if rec.ExceptionCode == 0xC0000005 {
				fmt.Printf("\n!!! Access Violation at 0x%X (Offset: 0x%X) !!!\n", rec.ExceptionAddress, rec.ExceptionAddress - imageBase)
				fmt.Printf("  Access Type:      %v (0=read, 1=write)\n", rec.ExceptionInformation[0])
				fmt.Printf("  Accessed Address: 0x%X\n", rec.ExceptionInformation[1])
				
				// Open thread to read stack
				r, _, _ := procOpenThread.Call(0x0008, 0, uintptr(event.dwThreadId)) // GET_CONTEXT
				hThread := syscall.Handle(r)
				if r != 0 {
					defer syscall.CloseHandle(hThread)
					var ctx [1232]byte
					binary.LittleEndian.PutUint32(ctx[0x30:], 0x10001F) // CONTEXT_ALL
					rCtx, _, _ := procGetThreadCtx.Call(uintptr(hThread), uintptr(unsafe.Pointer(&ctx[0])))
					if rCtx != 0 {
						rsp := binary.LittleEndian.Uint64(ctx[0x98:])
						fmt.Printf("  RSP: 0x%X\n", rsp)
						
						// Read 256 bytes from stack
						var stackBytes [256]byte
						if readProcessMemory(pi.Process, uintptr(rsp), uintptr(unsafe.Pointer(&stackBytes[0])), 256) {
							fmt.Printf("\nStack values (quadwords):\n")
							for i := 0; i < 256; i += 8 {
								val := binary.LittleEndian.Uint64(stackBytes[i : i+8])
								offset := val - uint64(imageBase)
								if offset < 0x20000 {
									fmt.Printf("  RSP + %3d: 0x%X (Offset in .text: 0x%X) <--- CALLER RETURN ADDRESS!\n", i, val, offset)
								} else {
									fmt.Printf("  RSP + %3d: 0x%X\n", i, val)
								}
							}
						}
					}
				}
				
				status = uint32(DBG_EXCEPTION_NOT_HANDLED)
				continueDebugEvent(event.dwProcessId, event.dwThreadId, status)
				return
			}
		}
		
		continueDebugEvent(event.dwProcessId, event.dwThreadId, status)
	}
}
