package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

const (
	DEBUG_ONLY_THIS_PROCESS = 0x00000002
	CREATE_NEW_CONSOLE      = 0x00000010
	EXCEPTION_DEBUG_EVENT   = 1
	LOAD_DLL_DEBUG_EVENT    = 6
)

type EXCEPTION_RECORD64 struct {
	ExceptionCode    uint32
	ExceptionFlags   uint32
	ExceptionRecord  uint64
	ExceptionAddress uint64
	NumberParameters uint32
	__unusedAlignment  uint32
	ExceptionInformation [15]uint64
}

type EXCEPTION_DEBUG_INFO64 struct {
	ExceptionRecord EXCEPTION_RECORD64
	dwFirstChance   uint32
}

type DEBUG_EVENT64 struct {
	dwDebugEventCode uint32
	dwProcessId      uint32
	dwThreadId       uint32
	padding          uint32 // 4 bytes padding to align union in 64-bit
	u                [160]byte // Padding to fit the union
}

// x64 Context structure
type CONTEXT64 struct {
	P1Home uint64
	P2Home uint64
	P3Home uint64
	P4Home uint64
	P5Home uint64
	P6Home uint64

	ContextFlags uint32
	MxCsr        uint32

	SegCs  uint16
	SegDs  uint16
	SegEs  uint16
	SegFs  uint16
	SegGs  uint16
	SegSs  uint16
	EFlags uint32

	Dr0 uint64
	Dr1 uint64
	Dr2 uint64
	Dr3 uint64
	Dr6 uint64
	Dr7 uint64

	Rax uint64
	Rcx uint64
	Rdx uint64
	Rbx uint64
	Rsp uint64
	Rbp uint64
	Rsi uint64
	Rdi uint64
	R8  uint64
	R9  uint64
	R10 uint64
	R11 uint64
	R12 uint64
	R13 uint64
	R14 uint64
	R15 uint64

	Rip uint64

	// Floating point state omitted for brevity
}

func main() {
	if runtime.GOOS != "windows" {
		log.Fatalf("This script must be run on Windows")
	}

	target := `d:\projects\compiler\Axiom\test_malloc.exe`
	if len(os.Args) > 1 {
		target = os.Args[1]
	}

	argvPtr, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		log.Fatalf("Error converting string: %v", err)
	}

	var si syscall.StartupInfo
	var pi syscall.ProcessInformation

	si.Cb = uint32(unsafe.Sizeof(si))

	err = syscall.CreateProcess(
		nil,
		argvPtr,
		nil,
		nil,
		false,
		DEBUG_ONLY_THIS_PROCESS|CREATE_NEW_CONSOLE,
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

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	waitForDebugEvent := kernel32.NewProc("WaitForDebugEvent")
	continueDebugEvent := kernel32.NewProc("ContinueDebugEvent")
	getThreadContext := kernel32.NewProc("GetThreadContext")

	fmt.Printf("Started debugging process %d (thread %d)...\n", pi.ProcessId, pi.ThreadId)

	for {
		var event DEBUG_EVENT64
		res, _, _ := waitForDebugEvent.Call(uintptr(unsafe.Pointer(&event)), 0xFFFFFFFF)
		if res == 0 {
			break
		}

		continueStatus := uintptr(0x00010002) // DBG_CONTINUE

		switch event.dwDebugEventCode {
		case EXCEPTION_DEBUG_EVENT:
			info := (*EXCEPTION_DEBUG_INFO64)(unsafe.Pointer(&event.u[0]))
			rec := info.ExceptionRecord
			
			// 0x80000003 is breakpoint (normally hit at startup), ignore first chance
			if rec.ExceptionCode != 0x80000003 {
				fmt.Printf("\n--- EXCEPTION DETECTED ---\n")
				fmt.Printf("Exception Code:    0x%08X\n", rec.ExceptionCode)
				fmt.Printf("Exception Address: 0x%016X\n", rec.ExceptionAddress)

				// Get Thread Context
				var ctx CONTEXT64
				ctx.ContextFlags = 0x00100007 // CONTEXT_FULL (Control | Integer | Segments)
				const THREAD_GET_CONTEXT = 0x0008
				openThread := kernel32.NewProc("OpenThread")
				hThreadRes, _, _ := openThread.Call(uintptr(THREAD_GET_CONTEXT), 0, uintptr(event.dwThreadId))
				if hThreadRes != 0 {
					hThread := syscall.Handle(hThreadRes)
					resCtx, _, _ := getThreadContext.Call(uintptr(hThread), uintptr(unsafe.Pointer(&ctx)))
					if resCtx != 0 {
						fmt.Printf("\nRegisters:\n")
						fmt.Printf("  RIP: 0x%016X   RSP: 0x%016X\n", ctx.Rip, ctx.Rsp)
						fmt.Printf("  RAX: 0x%016X   RCX: 0x%016X\n", ctx.Rax, ctx.Rcx)
						fmt.Printf("  RDX: 0x%016X   RBX: 0x%016X\n", ctx.Rdx, ctx.Rbx)
						fmt.Printf("  RBP: 0x%016X   RSI: 0x%016X   RDI: 0x%016X\n", ctx.Rbp, ctx.Rsi, ctx.Rdi)
						fmt.Printf("  R8:  0x%016X   R9:  0x%016X   R10: 0x%016X\n", ctx.R8, ctx.R9, ctx.R10)
						fmt.Printf("  R11: 0x%016X   R12: 0x%016X   R13: 0x%016X\n", ctx.R11, ctx.R12, ctx.R13)
						fmt.Printf("  R14: 0x%016X   R15: 0x%016X\n", ctx.R14, ctx.R15)
					}
					syscall.CloseHandle(hThread)
				}
				continueStatus = 0x80010001 // DBG_EXCEPTION_NOT_HANDLED
				
				// Let's terminate the process after printing exception
				syscall.TerminateProcess(pi.Process, 1)
				return
			}
		}

		continueDebugEvent.Call(uintptr(event.dwProcessId), uintptr(event.dwThreadId), continueStatus)
	}
}
