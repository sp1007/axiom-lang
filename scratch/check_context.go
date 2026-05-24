package main

import (
	"fmt"
	"unsafe"
)

type M128A struct {
	Low  uint64
	High int64
}

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
	Header               [2 * 512]byte
	VectorRegister       [26]M128A
	VectorControl        uint64
	DebugControl         uint64
	LastBranchToRip      uint64
	LastBranchFromRip    uint64
	LastExceptionToRip   uint64
	LastExceptionFromRip uint64
}

func main() {
	var ctx CONTEXT
	fmt.Printf("Size of CONTEXT: %d\n", unsafe.Sizeof(ctx))
	fmt.Printf("Offset of ContextFlags: 0x%X\n", unsafe.Offsetof(ctx.ContextFlags))
	fmt.Printf("Offset of EFlags: 0x%X\n", unsafe.Offsetof(ctx.EFlags))
	fmt.Printf("Offset of Rip: 0x%X\n", unsafe.Offsetof(ctx.Rip))
	fmt.Printf("Offset of Rax: 0x%X\n", unsafe.Offsetof(ctx.Rax))
	fmt.Printf("Offset of Rcx: 0x%X\n", unsafe.Offsetof(ctx.Rcx))
	fmt.Printf("Offset of Rdx: 0x%X\n", unsafe.Offsetof(ctx.Rdx))
	fmt.Printf("Offset of Rsp: 0x%X\n", unsafe.Offsetof(ctx.Rsp))
	fmt.Printf("Offset of Rbp: 0x%X\n", unsafe.Offsetof(ctx.Rbp))
	fmt.Printf("Offset of Rsi: 0x%X\n", unsafe.Offsetof(ctx.Rsi))
	fmt.Printf("Offset of Rdi: 0x%X\n", unsafe.Offsetof(ctx.Rdi))
}
